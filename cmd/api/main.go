package main

import (
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
    log.Fatal(http.ListenAndServe(":"+cfg.Port, srv))
}
