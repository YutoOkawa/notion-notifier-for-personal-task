package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type WebhookClient struct {
	httpClient *http.Client
	webhookURL string
}

func NewWebhookClient(webhookURL string) *WebhookClient {
	return &WebhookClient{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		webhookURL: webhookURL,
	}
}

func (c *WebhookClient) Notify(ctx context.Context, message string) error {
	payload := map[string]string{
		"content": message,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook responded with status: %d", resp.StatusCode)
	}

	return nil
}
