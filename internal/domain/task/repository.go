package task

import "context"

type Repository interface {
	FetchTasksWithUpcomingDeadlines(ctx context.Context, daysBeforeDeadline int) ([]*Task, error)
}
