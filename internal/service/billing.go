package service

import (
    "fmt"
    "sync"
    "time"

    "github.com/google/uuid"
    "github.com/livininthelast/lispflow/pkg/billing"
    "github.com/livininthelast/lispflow/internal/repository"
)

type Service struct {
    engine *billing.Engine
    repo   *repository.Repo
    mu     sync.Mutex
    ledger []billing.LedgerEntry
    plans  map[string][]billing.Plan // customerID -> versions
}

func New(engine *billing.Engine, repo *repository.Repo) *Service {
    return &Service{engine: engine, repo: repo, plans: make(map[string][]billing.Plan)}
}

func (s *Service) SetPlan(customerID, expr string) (billing.Plan, error) {
    if err := s.engine.Validate(expr); err != nil { return billing.Plan{}, err }
    s.mu.Lock()
    defer s.mu.Unlock()
    vers := len(s.plans[customerID]) + 1
    // deactivate old
    for i := range s.plans[customerID] { s.plans[customerID][i].Active = false }
    p := billing.Plan{ID: uuid.New().String(), CustomerID: customerID, Version: vers, Expression: expr, Active: true, CreatedAt: time.Now()}
    s.plans[customerID] = append(s.plans[customerID], p)
    if s.repo != nil {
        _ = s.repo.DeactivateOldPlans(customerID)
        _ = s.repo.CreatePlan(p)
    }
    return p, nil
}

func (s *Service) Evaluate(customerID string, metrics map[string]float64) (billing.EvaluationResult, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    plans := s.plans[customerID]
    if len(plans)==0 { return billing.EvaluationResult{}, fmt.Errorf("no plan for customer %s", customerID) }
    active := plans[len(plans)-1]
    for i := len(plans)-1; i>=0; i-- { if plans[i].Active { active = plans[i]; break } }
    amount, breakdown, err := s.engine.Evaluate(active.Expression, metrics)
    if err != nil { return billing.EvaluationResult{}, err }
    res := billing.EvaluationResult{CustomerID: customerID, Amount: amount, Breakdown: breakdown, PlanID: active.ID, Version: active.Version, Timestamp: time.Now()}
    entry := billing.LedgerEntry{ID: uuid.New().String(), CustomerID: customerID, PlanID: active.ID, Amount: amount, Metrics: metrics, CreatedAt: res.Timestamp}
    s.ledger = append(s.ledger, entry)
    if s.repo != nil { _ = s.repo.InsertLedger(entry) }
    return res, nil
}

func (s *Service) Batch(customerID string, batch []map[string]float64) ([]billing.EvaluationResult, error) {
    var out []billing.EvaluationResult
    agg := 0.0
    for _, m := range batch {
        r, err := s.Evaluate(customerID, m)
        if err != nil { return nil, err }
        out = append(out, r)
        agg += r.Amount
    }
    return out, nil
}

func (s *Service) History(customerID string, limit, offset int) ([]billing.LedgerEntry, error) {
    s.mu.Lock()
    defer s.mu.Unlock()
    // immutable ledger, paginated newest first
    var filtered []billing.LedgerEntry
    for i := len(s.ledger)-1; i>=0; i-- {
        if s.ledger[i].CustomerID == customerID { filtered = append(filtered, s.ledger[i]) }
    }
    if offset >= len(filtered) { return []billing.LedgerEntry{}, nil }
    end := offset+limit
    if end > len(filtered) { end = len(filtered) }
    return filtered[offset:end], nil
}

func (s *Service) Simulate(req billing.SimulateRequest) (billing.SimulateResponse, error) {
    if err := s.engine.Validate(req.PlanExpr); err != nil { return billing.SimulateResponse{}, fmt.Errorf("invalid plan: %w", err) }
    var results []billing.EvaluationResult
    total := 0.0
    breakdown := map[string]float64{}
    for _, scenario := range req.Scenarios {
        amt, bd, err := s.engine.Evaluate(req.PlanExpr, scenario)
        if err != nil { return billing.SimulateResponse{}, err }
        results = append(results, billing.EvaluationResult{Amount: amt, Breakdown: bd, Timestamp: time.Now()})
        total += amt
        for k,v := range bd { breakdown[k] += v }
    }
    avg := 0.0
    if len(results)>0 { avg = total/float64(len(results)) }
    return billing.SimulateResponse{Results: results, TotalRevenue: total, AvgRevenue: avg, CostBreakdown: breakdown}, nil
}
