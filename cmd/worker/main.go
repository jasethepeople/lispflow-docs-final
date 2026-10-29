package main

import (
    "log"
    "time"
    "github.com/livininthelast/lispflow/internal/config"
    "github.com/livininthelast/lispflow/internal/ingestion"
    "github.com/livininthelast/lispflow/internal/repository"
    "github.com/livininthelast/lispflow/internal/service"
    "github.com/livininthelast/lispflow/pkg/billing"
)

func main(){
    cfg := config.Load()
    eng := billing.NewEngine()
    var repo *repository.Repo
    if cfg.DatabaseURL != "" {
        r, err := repository.New(cfg.DatabaseURL)
        if err != nil { log.Printf("repo failed: %v", err) } else { repo = r }
    }
    svc := service.New(eng, repo)
    batcher := ingestion.New(100, 5*time.Second)
    log.Println("Worker started, batch size 100, flush 5s")
    for batch := range batcher.Batches() {
        log.Printf("Flushing batch of %d events", len(batch))
        for _, ev := range batch {
            _, err := svc.Evaluate(ev.CustomerID, ev.Metrics)
            if err != nil { log.Printf("eval failed for %s: %v", ev.CustomerID, err) }
        }
    }
}
