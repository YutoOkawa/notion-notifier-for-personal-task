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

func (s *NotificationService) NotifyDelayedReadingTasks(ctx context.Context) error {
	tasks, err := s.taskRepo.FetchIncompleteStudyTasks(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch study tasks: %w", err)
	}

	var delayedTasks []*task.Task
	for _, t := range tasks {
		if t.IsReadingPaceDelayed() {
			delayedTasks = append(delayedTasks, t)
		}
	}

	if len(delayedTasks) == 0 {
		return nil
	}

	message := s.buildReadingNotificationMessage(delayedTasks)
	if err := s.notifier.Notify(ctx, message); err != nil {
		return fmt.Errorf("failed to send reading notification: %w", err)
	}

	return nil
}

func (s *NotificationService) Run(ctx context.Context) error {
	if err := s.NotifyUpcomingDeadlines(ctx); err != nil {
		return err
	}
	return s.NotifyDelayedReadingTasks(ctx)
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

func (s *NotificationService) buildReadingNotificationMessage(tasks []*task.Task) string {
	var sb strings.Builder
	sb.WriteString("📚 **読書ペース遅延アラート**\n\n")

	for _, t := range tasks {
		expected := t.ExpectedReadPages()
		diff := expected - t.ReadPages
		projectNameStr := ""
		if t.ProjectName != "" {
			projectNameStr = fmt.Sprintf("[%s] ", t.ProjectName)
		}
		sb.WriteString(fmt.Sprintf("- %s%s: 現在 %dページ / 目標 %dページ (残り: %dp)\n",
			projectNameStr, t.Name, t.ReadPages, expected, diff))
	}

	return sb.String()
}
