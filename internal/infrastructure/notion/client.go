package notion

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/YutoOkawa/notion-notifier-for-personal-task/internal/domain/task"
)

const (
	notionAPIVersion = "2022-06-28"
	notionBaseURL    = "https://api.notion.com/v1"
)

type Client struct {
	httpClient *http.Client
	apiToken   string
	databaseID string
}

func NewClient(apiToken, databaseID string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		apiToken:   apiToken,
		databaseID: databaseID,
	}
}

// Notion API でフィルタ条件を使って締切が近いタスクを取得する。
// Status が Not Started または In Progress、かつ Due が指定日数以内のタスクを返す。
func (c *Client) FetchTasksWithUpcomingDeadlines(ctx context.Context, daysBeforeDeadline int) ([]*task.Task, error) {
	now := time.Now()
	endDate := now.AddDate(0, 0, daysBeforeDeadline)

	filter := map[string]interface{}{
		"and": []map[string]interface{}{
			{
				"property": "Due",
				"date": map[string]interface{}{
					"on_or_before": endDate.Format("2006-01-02"),
				},
			},
			{
				"property": "Due",
				"date": map[string]interface{}{
					"on_or_after": now.Format("2006-01-02"),
				},
			},
			{
				"or": []map[string]interface{}{
					{
						"property": "Status",
						"status": map[string]string{
							"equals": "Not Started",
						},
					},
					{
						"property": "Status",
						"status": map[string]string{
							"equals": "In Progress",
						},
					},
				},
			},
		},
	}

	reqBody := map[string]interface{}{
		"filter": filter,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	url := fmt.Sprintf("%s/databases/%s/query", notionBaseURL, c.databaseID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Notion-Version", notionAPIVersion)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("notion API error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result queryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	projectIDs := make(map[string]bool)
	for _, p := range result.Results {
		if len(p.Properties.Project.Relation) > 0 {
			projectIDs[p.Properties.Project.Relation[0].ID] = true
		}
	}

	projectNames := make(map[string]string)
	for id := range projectIDs {
		name, err := c.fetchPageTitle(ctx, id)
		if err != nil {
			projectNames[id] = "Personal"
		} else {
			projectNames[id] = name
		}
	}

	return c.convertToTasks(result.Results, projectNames), nil
}

func (c *Client) convertToTasks(pages []page, projectNames map[string]string) []*task.Task {
	tasks := make([]*task.Task, 0, len(pages))
	for _, p := range pages {
		t := c.pageToTask(p, projectNames)
		if t != nil {
			tasks = append(tasks, t)
		}
	}
	return tasks
}

func (c *Client) pageToTask(p page, projectNames map[string]string) *task.Task {
	name := ""
	if title := p.Properties.TaskName; title.Title != nil && len(title.Title) > 0 {
		name = title.Title[0].PlainText
	}

	var dueDate *time.Time
	if p.Properties.Due.Date != nil && p.Properties.Due.Date.Start != "" {
		dueDate = parseDueDate(p.Properties.Due.Date.Start)
	}

	status := task.StatusNotStarted
	if p.Properties.Status.Status != nil {
		switch p.Properties.Status.Status.Name {
		case "In Progress":
			status = task.StatusInProgress
		case "Done":
			status = task.StatusDone
		case "Archived":
			status = task.StatusArchived
		}
	}

	projectName := "Personal"
	if len(p.Properties.Project.Relation) > 0 {
		projectName = projectNames[p.Properties.Project.Relation[0].ID]
	}

	return task.NewTask(p.ID, name, projectName, dueDate, status)
}

type queryResponse struct {
	Results []page `json:"results"`
}

type page struct {
	ID         string     `json:"id"`
	Properties properties `json:"properties"`
}

type properties struct {
	TaskName titleProperty    `json:"Task name"`
	Due      dateProperty     `json:"Due"`
	Status   statusProperty   `json:"Status"`
	Project  relationProperty `json:"Project"`
}

type relationProperty struct {
	Relation []relationValue `json:"relation"`
}

type relationValue struct {
	ID string `json:"id"`
}

type titleProperty struct {
	Title []richText `json:"title"`
}

type richText struct {
	PlainText string `json:"plain_text"`
}

type dateProperty struct {
	Date *dateValue `json:"date"`
}

type dateValue struct {
	Start string `json:"start"`
}

type statusProperty struct {
	Status *statusValue `json:"status"`
}

type statusValue struct {
	Name string `json:"name"`
}

// Notion の日付形式をパースする。
// - 日付のみ: "2026-02-10" → UTC 00:00:00 として解釈
// - 時刻付き: "2026-02-10T15:00:00.000+09:00" → 元のタイムゾーンを保持
func parseDueDate(s string) *time.Time {
	// まず時刻付き形式（RFC3339）を試行
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return &t
	}
	// 次に日付のみ形式を試行
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return &t
	}
	return nil
}

func (c *Client) fetchPageTitle(ctx context.Context, pageID string) (string, error) {
	url := fmt.Sprintf("%s/pages/%s", notionBaseURL, pageID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Notion-Version", notionAPIVersion)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to fetch page: %d", resp.StatusCode)
	}

	var pageResp struct {
		Properties map[string]struct {
			Type  string     `json:"type"`
			Title []richText `json:"title"`
		} `json:"properties"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&pageResp); err != nil {
		return "", err
	}

	for _, prop := range pageResp.Properties {
		if prop.Type == "title" && len(prop.Title) > 0 {
			return prop.Title[0].PlainText, nil
		}
	}
	return "Personal", nil
}
