package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/YutoOkawa/notion-notifier-for-personal-task/internal/domain/notification"
	"github.com/YutoOkawa/notion-notifier-for-personal-task/internal/domain/task"
)

type NotificationService struct {
	taskRepo           task.Repository
	notifier           notification.Notifier
	daysBeforeDeadline int
}

func NewNotificationService(taskRepo task.Repository, notifier notification.Notifier, daysBeforeDeadline int) *NotificationService {
	return &NotificationService{
		taskRepo:           taskRepo,
		notifier:           notifier,
		daysBeforeDeadline: daysBeforeDeadline,
	}
}

func (s *NotificationService) NotifyUpcomingDeadlines(ctx context.Context) error {
	tasks, err := s.taskRepo.FetchTasksWithUpcomingDeadlines(ctx, s.daysBeforeDeadline)
	if err != nil {
		return fmt.Errorf("failed to fetch tasks: %w", err)
	}

	if len(tasks) == 0 {
		return nil
	}

	message := s.buildNotificationMessage(tasks)
	if err := s.notifier.Notify(ctx, message); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

func (s *NotificationService) Run(ctx context.Context) error {
	return s.NotifyUpcomingDeadlines(ctx)
}

func (s *NotificationService) buildNotificationMessage(tasks []*task.Task) string {
	var sb strings.Builder
	sb.WriteString("📋 **締切が近いタスク一覧**\n\n")

	for _, t := range tasks {
		days := t.DaysUntilDeadline()
		var dueText string
		switch {
		case days == 0:
			dueText = "🔴 **本日締切**"
		case days == 1:
			dueText = "🟠 明日締切"
		default:
			dueText = fmt.Sprintf("🟡 あと%d日", days)
		}

		projectNameStr := ""
		if t.ProjectName != "" {
			projectNameStr = fmt.Sprintf("[%s] ", t.ProjectName)
		}

		sb.WriteString(fmt.Sprintf("- %s%s: %s\n", projectNameStr, t.Name, dueText))
	}

	return sb.String()
}
