package jobs

import (
	"context"
	"log"
	"sync"
	"time"

	"api/internal/di"
)

// Job represents a scheduled job that runs periodically
type Job interface {
	Name() string
	Run(ctx context.Context) error
	Interval() time.Duration
}

// Scheduler manages and runs scheduled jobs
type Scheduler struct {
	jobs      []Job
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	container *di.Container
}

// NewScheduler creates a new job scheduler
func NewScheduler(container *di.Container) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		jobs:      []Job{},
		ctx:       ctx,
		cancel:    cancel,
		container: container,
	}
}

// RegisterJob adds a job to the scheduler
func (s *Scheduler) RegisterJob(job Job) {
	s.jobs = append(s.jobs, job)
	log.Printf("[SCHEDULER] Registered job: %s (interval: %v)", job.Name(), job.Interval())
}

// Start begins running all registered jobs
func (s *Scheduler) Start() {
	log.Printf("[SCHEDULER] Starting scheduler with %d jobs", len(s.jobs))

	for _, job := range s.jobs {
		s.wg.Add(1)
		go s.runJob(job)
	}
}

// Stop gracefully stops all running jobs
func (s *Scheduler) Stop() {
	log.Printf("[SCHEDULER] Stopping scheduler...")
	s.cancel()
	s.wg.Wait()
	log.Printf("[SCHEDULER] All jobs stopped")
}

// runJob runs a single job on its interval
func (s *Scheduler) runJob(job Job) {
	defer s.wg.Done()

	ticker := time.NewTicker(job.Interval())
	defer ticker.Stop()

	// Run immediately on startup
	s.executeJob(job)

	for {
		select {
		case <-s.ctx.Done():
			log.Printf("[SCHEDULER] Job %s stopped", job.Name())
			return
		case <-ticker.C:
			s.executeJob(job)
		}
	}
}

// executeJob runs a job and handles errors
func (s *Scheduler) executeJob(job Job) {
	log.Printf("[SCHEDULER] Running job: %s", job.Name())
	start := time.Now()

	if err := job.Run(s.ctx); err != nil {
		log.Printf("[SCHEDULER] Job %s failed: %v", job.Name(), err)
		// TODO: Add metrics for job failures
	} else {
		duration := time.Since(start)
		log.Printf("[SCHEDULER] Job %s completed successfully (took %v)", job.Name(), duration)
		// TODO: Add metrics for job success
	}
}
