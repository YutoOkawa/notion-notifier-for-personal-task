package task

import "time"

type Status string

const (
	StatusNotStarted Status = "Not Started"
	StatusInProgress Status = "In Progress"
	StatusDone       Status = "Done"
	StatusArchived   Status = "Archived"
)

type Task struct {
	ID      string
	Name    string
	DueDate *time.Time
	Status  Status
}

func NewTask(id, name string, dueDate *time.Time, status Status) *Task {
	return &Task{
		ID:      id,
		Name:    name,
		DueDate: dueDate,
		Status:  status,
	}
}

func (t *Task) IsApproachingDeadline(daysBeforeDeadline int) bool {
	days := t.daysUntilDeadline()
	if days < 0 {
		return false
	}
	return days <= daysBeforeDeadline
}

func (t *Task) IsOverdue() bool {
	if t.DueDate == nil {
		return false
	}
	return t.daysUntilDeadline() < 0
}

// Returns -1 if no due date is set.
func (t *Task) DaysUntilDeadline() int {
	return t.daysUntilDeadline()
}

func (t *Task) IsNotificationTarget() bool {
	return t.Status == StatusNotStarted || t.Status == StatusInProgress
}

// Due date は日付のみ（"YYYY-MM-DD"）または時刻付き（RFC3339）で返される。
// 日単位で比較するため、年月日のみを抽出して日数差を計算する。
func (t *Task) daysUntilDeadline() int {
	if t.DueDate == nil {
		return -1
	}
	dueY, dueM, dueD := t.DueDate.Date()
	nowY, nowM, nowD := time.Now().Date()

	dueDate := time.Date(dueY, dueM, dueD, 0, 0, 0, 0, time.UTC)
	today := time.Date(nowY, nowM, nowD, 0, 0, 0, 0, time.UTC)

	return int(dueDate.Sub(today).Hours() / 24)
}
