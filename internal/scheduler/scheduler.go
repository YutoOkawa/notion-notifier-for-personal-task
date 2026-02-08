package scheduler

import (
	"context"
	"log"
	"time"

	"github.com/robfig/cron/v3"
)

type Job interface {
	Run(ctx context.Context) error
}

type Scheduler struct {
	cron     *cron.Cron
	job      Job
	schedule string
}

func New(schedule string, job Job) *Scheduler {
	return &Scheduler{
		cron:     cron.New(cron.WithLocation(time.FixedZone("JST", 9*60*60))),
		job:      job,
		schedule: schedule,
	}
}

func (s *Scheduler) Start() error {
	_, err := s.cron.AddFunc(s.schedule, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		log.Println("Running scheduled job...")
		if err := s.job.Run(ctx); err != nil {
			log.Printf("job error: %v", err)
		} else {
			log.Println("Job completed successfully")
		}
	})
	if err != nil {
		return err
	}

	s.cron.Start()
	log.Printf("Scheduler started. Schedule: %s, Next run: %v", s.schedule, s.cron.Entries()[0].Next)
	return nil
}

func (s *Scheduler) Stop() {
	s.cron.Stop()
	log.Println("Scheduler stopped")
}

func (s *Scheduler) RunNow() error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	return s.job.Run(ctx)
}
