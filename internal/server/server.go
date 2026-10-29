package server

import (
    "encoding/json"
    "net/http"
    "strconv"

    "github.com/gorilla/mux"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    "github.com/livininthelast/lispflow/internal/service"
    "github.com/livininthelast/lispflow/pkg/billing"
)

type Server struct {
    svc *service.Service
    engine *billing.Engine
    r *mux.Router
}

func New(svc *service.Service, engine *billing.Engine) *Server {
    s := &Server{svc: svc, engine: engine, r: mux.NewRouter()}
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
    json.NewEncoder(w).Encode(p)
}

func (s *Server) evaluate(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    var body struct{ Metrics map[string]float64 `json:"metrics"` }
    json.NewDecoder(r.Body).Decode(&body)
    res, err := s.svc.Evaluate(vars["id"], body.Metrics)
    if err != nil { http.Error(w, err.Error(), 400); return }
    json.NewEncoder(w).Encode(res)
}

func (s *Server) batch(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    var body struct{ Events []map[string]float64 `json:"events"` }
    json.NewDecoder(r.Body).Decode(&body)
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
