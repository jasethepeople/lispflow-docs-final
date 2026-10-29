package main

import (
    "context"
    "time"
    "github.com/livininthelast/lispflow/internal/scheduler"

    "log"
    "net/http"
    "github.com/livininthelast/lispflow/internal/config"
    "github.com/livininthelast/lispflow/internal/repository"
    "github.com/livininthelast/lispflow/internal/server"
    "github.com/livininthelast/lispflow/internal/service"
    "github.com/livininthelast/lispflow/pkg/billing"
)

func main() {
    cfg := config.Load()
    engine := billing.NewEngine()
    var repo *repository.Repo
    if cfg.DatabaseURL != "" {
        r, err := repository.New(cfg.DatabaseURL)
        if err != nil { log.Printf("repo init failed, running in-memory: %v", err) } else { repo = r }
    }
    svc := service.New(engine, repo)
    srv := server.New(svc, engine)
    log.Printf("LispFlow API listening on :%s", cfg.Port)
    
    // middleware
    // rate limit: 100 rps, burst 200
    // Note: import middleware
    ctx := context.Background()
    sched := scheduler.New(1*time.Hour)
    sched.Add(scheduler.ExpireOldPlansJob())
    sched.Add(scheduler.CleanupLedgerJob(90))
    sched.Start(ctx)

    log.Fatal(http.ListenAndServe(":"+cfg.Port, srv))

}
