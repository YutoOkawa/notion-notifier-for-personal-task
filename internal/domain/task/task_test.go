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

	tests := []struct {
		name       string
		taskType   string
		dueDate    *time.Time
		totalPages int
		readPages  int
		status     Status
		want       bool
	}{
		{
			name:       "not study type",
			taskType:   "Personal",
			dueDate:    &today,
			totalPages: 100,
			readPages:  0,
			status:     StatusInProgress,
			want:       false,
		},
		{
			name:       "due today, read 70 pages (expected 100)",
			taskType:   "Study",
			dueDate:    &today,
			totalPages: 100,
			readPages:  70,
			status:     StatusInProgress,
			want:       true,
		},
		{
			name:       "due today, read 100 pages (expected 100)",
			taskType:   "Study",
			dueDate:    &today,
			totalPages: 100,
			readPages:  100,
			status:     StatusInProgress,
			want:       false,
		},
		{
			name:       "due tomorrow, read 70 pages (expected 70)",
			taskType:   "Study",
			dueDate:    timePtr(today.Add(24 * time.Hour)),
			totalPages: 100,
			readPages:  70,
			status:     StatusInProgress,
			want:       false,
		},
		{
			name:       "due tomorrow, read 40 pages (expected 70)",
			taskType:   "Study",
			dueDate:    timePtr(today.Add(24 * time.Hour)),
			totalPages: 100,
			readPages:  40,
			status:     StatusInProgress,
			want:       true,
		},
		{
			name:       "done task is not delayed",
			taskType:   "Study",
			dueDate:    &today,
			totalPages: 100,
			readPages:  30,
			status:     StatusDone,
			want:       false,
		},
		{
			name:       "read pages meets total pages (not delayed)",
			taskType:   "Study",
			dueDate:    timePtr(today.Add(2 * 24 * time.Hour)),
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
				DueDate:    tt.dueDate,
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
	tomorrow := today.Add(24 * time.Hour)

	task := &Task{
		DueDate:    &tomorrow,
		TotalPages: 100,
	}

	expected := task.ExpectedReadPages()
	// Tomorrow is 1 day until due. 100 - (1 * 30) = 70.
	if expected != 70 {
		t.Errorf("ExpectedReadPages() = %d, want %d", expected, 70)
	}

	task.DueDate = &today
	expected = task.ExpectedReadPages()
	// Today is 0 days until due. 100 - (0 * 30) = 100.
	if expected != 100 {
		t.Errorf("ExpectedReadPages() = %d, want %d", expected, 100)
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
