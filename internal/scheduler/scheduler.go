package scheduler

import (
    "context"
    "log"
    "time"
)

type Job func(ctx context.Context) error

type Scheduler struct {
    interval time.Duration
    jobs []Job
}

func New(interval time.Duration) *Scheduler {
    return &Scheduler{interval: interval}
}

func (s *Scheduler) Add(j Job){ s.jobs = append(s.jobs, j) }

func (s *Scheduler) Start(ctx context.Context){
    ticker := time.NewTicker(s.interval)
    go func(){
        for {
            select {
            case <-ctx.Done(): ticker.Stop(); return
            case <-ticker.C:
                for _, j := range s.jobs {
                    if err := j(ctx); err != nil { log.Printf("scheduler job failed: %v", err) }
                }
            }
        }
    }()
}

// Example jobs
func ExpireOldPlansJob() Job {
    return func(ctx context.Context) error {
        log.Println("Checking for expired plans...")
        // In real impl: UPDATE plans SET active=false WHERE expires_at < NOW()
        return nil
    }
}

func CleanupLedgerJob(retentionDays int) Job {
    return func(ctx context.Context) error {
        log.Printf("Cleaning ledger older than %d days", retentionDays)
        // DELETE FROM ledger WHERE created_at < NOW() - interval
        return nil
    }
}
