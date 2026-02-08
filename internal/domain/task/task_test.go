package task

import (
	"testing"
	"time"
)

func TestTask_IsApproachingDeadline(t *testing.T) {
	tests := []struct {
		name               string
		dueDate            *time.Time
		daysBeforeDeadline int
		want               bool
	}{
		{
			name:               "no due date",
			dueDate:            nil,
			daysBeforeDeadline: 3,
			want:               false,
		},
		{
			name:               "due date is today",
			dueDate:            timePtr(time.Now()),
			daysBeforeDeadline: 3,
			want:               true,
		},
		{
			name:               "due date is tomorrow",
			dueDate:            timePtr(time.Now().Add(24 * time.Hour)),
			daysBeforeDeadline: 3,
			want:               true,
		},
		{
			name:               "due date is in 3 days",
			dueDate:            timePtr(time.Now().Add(3 * 24 * time.Hour)),
			daysBeforeDeadline: 3,
			want:               true,
		},
		{
			name:               "due date is in 4 days (outside range)",
			dueDate:            timePtr(time.Now().Add(4 * 24 * time.Hour)),
			daysBeforeDeadline: 3,
			want:               false,
		},
		{
			name:               "due date was yesterday (overdue)",
			dueDate:            timePtr(time.Now().Add(-24 * time.Hour)),
			daysBeforeDeadline: 3,
			want:               false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask("1", "Test Task", tt.dueDate, StatusNotStarted)
			got := task.IsApproachingDeadline(tt.daysBeforeDeadline)
			if got != tt.want {
				t.Errorf("IsApproachingDeadline() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_IsOverdue(t *testing.T) {
	tests := []struct {
		name    string
		dueDate *time.Time
		want    bool
	}{
		{
			name:    "no due date",
			dueDate: nil,
			want:    false,
		},
		{
			name:    "due date is today",
			dueDate: timePtr(time.Now()),
			want:    false,
		},
		{
			name:    "due date was yesterday",
			dueDate: timePtr(time.Now().Add(-24 * time.Hour)),
			want:    true,
		},
		{
			name:    "due date is tomorrow",
			dueDate: timePtr(time.Now().Add(24 * time.Hour)),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask("1", "Test Task", tt.dueDate, StatusNotStarted)
			got := task.IsOverdue()
			if got != tt.want {
				t.Errorf("IsOverdue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_IsNotificationTarget(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		want   bool
	}{
		{name: "NotStarted is target", status: StatusNotStarted, want: true},
		{name: "InProgress is target", status: StatusInProgress, want: true},
		{name: "Done is not target", status: StatusDone, want: false},
		{name: "Archived is not target", status: StatusArchived, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask("1", "Test Task", nil, tt.status)
			got := task.IsNotificationTarget()
			if got != tt.want {
				t.Errorf("IsNotificationTarget() = %v, want %v", got, tt.want)
			}
		})
	}
}

func timePtr(t time.Time) *time.Time {
	return &t
}
