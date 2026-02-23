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
			task := NewTask("1", "Test Task", "Test Project", tt.dueDate, StatusNotStarted)
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
			task := NewTask("1", "Test Task", "Test Project", tt.dueDate, StatusNotStarted)
			got := task.IsOverdue()
			if got != tt.want {
				t.Errorf("IsOverdue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_IsReadingPaceDelayed(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)
	yesterday := today.Add(-24 * time.Hour)
	twoDaysAgo := today.Add(-48 * time.Hour)

	tests := []struct {
		name       string
		taskType   string
		startDate  *time.Time
		totalPages int
		readPages  int
		status     Status
		want       bool
	}{
		{
			name:       "not study type",
			taskType:   "Personal",
			startDate:  &twoDaysAgo,
			totalPages: 100,
			readPages:  0,
			status:     StatusInProgress,
			want:       false,
		},
		{
			name:       "started today, read 0 pages (expected 30)",
			taskType:   "Study",
			startDate:  &today,
			totalPages: 100,
			readPages:  0,
			status:     StatusInProgress,
			want:       true,
		},
		{
			name:       "started today, read 30 pages (expected 30)",
			taskType:   "Study",
			startDate:  &today,
			totalPages: 100,
			readPages:  30,
			status:     StatusInProgress,
			want:       false,
		},
		{
			name:       "started yesterday, read 30 pages (expected 60)",
			taskType:   "Study",
			startDate:  &yesterday,
			totalPages: 100,
			readPages:  30,
			status:     StatusInProgress,
			want:       true,
		},
		{
			name:       "started yesterday, read 60 pages (expected 60)",
			taskType:   "Study",
			startDate:  &yesterday,
			totalPages: 100,
			readPages:  60,
			status:     StatusInProgress,
			want:       false,
		},
		{
			name:       "done task is not delayed",
			taskType:   "Study",
			startDate:  &twoDaysAgo,
			totalPages: 100,
			readPages:  30,
			status:     StatusDone,
			want:       false,
		},
		{
			name:       "read pages meets total pages (not delayed)",
			taskType:   "Study",
			startDate:  &twoDaysAgo,
			totalPages: 50,
			readPages:  50,
			status:     StatusInProgress,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &Task{
				TaskType:   tt.taskType,
				StartDate:  tt.startDate,
				TotalPages: tt.totalPages,
				ReadPages:  tt.readPages,
				Status:     tt.status,
			}
			got := task.IsReadingPaceDelayed()
			if got != tt.want {
				t.Errorf("IsReadingPaceDelayed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTask_ExpectedReadPages(t *testing.T) {
	today := time.Now().Truncate(24 * time.Hour)
	twoDaysAgo := today.Add(-48 * time.Hour)

	task := &Task{
		StartDate:  &twoDaysAgo,
		TotalPages: 100,
	}

	expected := task.ExpectedReadPages()
	// 2 days ago to today is 3 days (day 0, day 1, today). 3 * 30 = 90.
	if expected != 90 {
		t.Errorf("ExpectedReadPages() = %d, want %d", expected, 90)
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
			task := NewTask("1", "Test Task", "Test Project", nil, tt.status)
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
