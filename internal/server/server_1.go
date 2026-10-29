package server

import (
    "context"
    "encoding/json"
    "net/http"
    "strconv"
    "time"

    "github.com/gorilla/mux"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/livininthelast/lispflow/internal/cache"
    "github.com/livininthelast/lispflow/internal/service"
    "github.com/livininthelast/lispflow/internal/webhook"
    "github.com/livininthelast/lispflow/pkg/billing"
    "github.com/livininthelast/lispflow/internal/metrics"
)

type Server struct {
    svc *service.Service
    engine *billing.Engine
    r *mux.Router
    cache *cache.Cache
    wh *webhook.Dispatcher
}

func New(svc *service.Service, engine *billing.Engine, c *cache.Cache, wh *webhook.Dispatcher) *Server {
    s := &Server{svc: svc, engine: engine, r: mux.NewRouter(), cache: c, wh: wh}
    s.routes()
    return s
}

func (s *Server) routes() {
    s.r.HandleFunc("/health", s.health).Methods("GET")
    s.r.HandleFunc("/ready", s.ready).Methods("GET")
    s.r.Handle("/metrics", promhttp.Handler()).Methods("GET")
    s.r.HandleFunc("/customers/{id}/plans", s.setPlan).Methods("POST")
    s.r.HandleFunc("/customers/{id}/evaluate", s.evaluate).Methods("POST")
    s.r.HandleFunc("/customers/{id}/batch", s.batch).Methods("POST")
    s.r.HandleFunc("/customers/{id}/history", s.history).Methods("GET")
    s.r.HandleFunc("/simulate", s.simulate).Methods("POST")
    s.r.HandleFunc("/validate", s.validate).Methods("POST")
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) { s.r.ServeHTTP(w,r) }

func (s *Server) health(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"status":"ok"}`)) }
func (s *Server) ready(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"status":"ready"}`)) }

func (s *Server) setPlan(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    var body struct{ Expression string `json:"expression"` }
    json.NewDecoder(r.Body).Decode(&body)
    p, err := s.svc.SetPlan(vars["id"], body.Expression)
    if err != nil { http.Error(w, err.Error(), 400); return }
    if s.cache != nil { s.cache.SetPlan(context.Background(), p); metrics.ActivePlans.Inc() }
    json.NewEncoder(w).Encode(p)
}

func (s *Server) evaluate(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    vars := mux.Vars(r)
    var body struct{ Metrics map[string]float64 `json:"metrics"` }
    json.NewDecoder(r.Body).Decode(&body)
    if s.cache != nil {
        if _, ok := s.cache.GetPlan(context.Background(), vars["id"]); ok {
            // cache hit
        }
    }
    res, err := s.svc.Evaluate(vars["id"], body.Metrics)
    if err != nil { 
        metrics.EvaluationsTotal.WithLabelValues(vars["id"], "error").Inc()
        http.Error(w, err.Error(), 400); return 
    }
    metrics.EvaluationsTotal.WithLabelValues(vars["id"], "success").Inc()
    metrics.EvaluationDuration.WithLabelValues(vars["id"]).Observe(time.Since(start).Seconds())
    metrics.LedgerEntries.WithLabelValues(vars["id"]).Inc()

    if s.wh != nil {
        go s.wh.Dispatch(context.Background(), webhook.Event{Type: "evaluation.completed", CustomerID: vars["id"], Result: res, Timestamp: time.Now()})
    }
    json.NewEncoder(w).Encode(res)
}

func (s *Server) batch(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    var body struct{ Events []map[string]float64 `json:"events"` }
    json.NewDecoder(r.Body).Decode(&body)
    metrics.BatchSize.Observe(float64(len(body.Events)))
    res, err := s.svc.Batch(vars["id"], body.Events)
    if err != nil { http.Error(w, err.Error(), 400); return }
    json.NewEncoder(w).Encode(res)
}

func (s *Server) history(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
    if limit==0 { limit=20 }
    offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
    h, err := s.svc.History(vars["id"], limit, offset)
    if err != nil { http.Error(w, err.Error(), 500); return }
    json.NewEncoder(w).Encode(h)
}

func (s *Server) simulate(w http.ResponseWriter, r *http.Request) {
    var req billing.SimulateRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil { http.Error(w, err.Error(), 400); return }
    res, err := s.svc.Simulate(req)
    if err != nil { http.Error(w, err.Error(), 400); return }
    json.NewEncoder(w).Encode(res)
}

func (s *Server) validate(w http.ResponseWriter, r *http.Request) {
    var body struct{ Expression string `json:"expression"` }
    json.NewDecoder(r.Body).Decode(&body)
    if err := s.engine.Validate(body.Expression); err != nil {
        json.NewEncoder(w).Encode(map[string]interface{}{"valid": false, "error": err.Error()})
        return
    }
    json.NewEncoder(w).Encode(map[string]interface{}{"valid": true})
}
