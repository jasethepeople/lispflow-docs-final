package main

import (
    "log"
    "net/http"
    "os"
    "strings"
    "github.com/livininthelast/lispflow/internal/cache"
    "github.com/livininthelast/lispflow/internal/config"
    "github.com/livininthelast/lispflow/internal/repository"
    "github.com/livininthelast/lispflow/internal/server"
    "github.com/livininthelast/lispflow/internal/service"
    "github.com/livininthelast/lispflow/internal/webhook"
    "github.com/livininthelast/lispflow/pkg/billing"
    _ "github.com/livininthelast/lispflow/internal/metrics"
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
    cache := cache.New(os.Getenv("REDIS_URL"))
    whEndpoints := []string{}
    if v := os.Getenv("WEBHOOK_URLS"); v != "" {
        whEndpoints = strings.Split(v, ",")
    }
    wh := webhook.New(whEndpoints)
    srv := server.New(svc, engine, cache, wh)
    log.Printf("LispFlow API listening on :%s (cache=%v webhooks=%d)", cfg.Port, cache!=nil, len(whEndpoints))
    log.Fatal(http.ListenAndServe(":"+cfg.Port, srv))
}
