package application

import (
	"context"
	"testing"
	"time"

	"github.com/YutoOkawa/notion-notifier-for-personal-task/internal/domain/task"
)

type mockTaskRepo struct {
	tasks []*task.Task
	err   error
}

func (m *mockTaskRepo) FetchTasksWithUpcomingDeadlines(ctx context.Context, days int) ([]*task.Task, error) {
	return m.tasks, m.err
}

type mockNotifier struct {
	lastMessage string
	err         error
}

func (m *mockNotifier) Notify(ctx context.Context, message string) error {
	m.lastMessage = message
	return m.err
}

func TestNotificationService_NotifyUpcomingDeadlines(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)

	tasks := []*task.Task{
		task.NewTask("1", "Task Due Today", "Personal", &today, task.StatusNotStarted),
		task.NewTask("2", "Task Due Tomorrow", "Work", &tomorrow, task.StatusInProgress),
	}

	repo := &mockTaskRepo{tasks: tasks}
	notifier := &mockNotifier{}
	service := NewNotificationService(repo, notifier, 3)

	err := service.NotifyUpcomingDeadlines(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if notifier.lastMessage == "" {
		t.Fatal("expected notification to be sent")
	}

	if !contains(notifier.lastMessage, "Task Due Today") {
		t.Errorf("expected message to contain 'Task Due Today', got: %s", notifier.lastMessage)
	}
	if !contains(notifier.lastMessage, "本日締切") {
		t.Errorf("expected message to contain '本日締切'")
	}
}

func TestNotificationService_NoTasks(t *testing.T) {
	repo := &mockTaskRepo{tasks: []*task.Task{}}
	notifier := &mockNotifier{}
	service := NewNotificationService(repo, notifier, 3)

	err := service.NotifyUpcomingDeadlines(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if notifier.lastMessage != "" {
		t.Errorf("expected no notification, got: %s", notifier.lastMessage)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
