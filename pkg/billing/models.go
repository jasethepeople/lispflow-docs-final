package billing

import (
    "time"
)

type Plan struct {
    ID         string    `json:"id"`
    CustomerID string    `json:"customer_id"`
    Version    int       `json:"version"`
    Expression string    `json:"expression"`
    Active     bool      `json:"active"`
    CreatedAt  time.Time `json:"created_at"`
}

type UsageEvent struct {
    CustomerID string            `json:"customer_id"`
    Metrics    map[string]float64 `json:"metrics"`
    Timestamp  time.Time          `json:"timestamp"`
    EventID    string             `json:"event_id"`
}

type EvaluationResult struct {
    CustomerID string             `json:"customer_id"`
    Amount     float64             `json:"amount"`
    Breakdown  map[string]float64 `json:"breakdown"`
    PlanID     string             `json:"plan_id"`
    Version    int                `json:"version"`
    Timestamp  time.Time          `json:"timestamp"`
}

type LedgerEntry struct {
    ID         string    `json:"id"`
    CustomerID string    `json:"customer_id"`
    PlanID     string    `json:"plan_id"`
    Amount     float64   `json:"amount"`
    Metrics    map[string]float64 `json:"metrics"`
    CreatedAt  time.Time `json:"created_at"`
}

type SimulateRequest struct {
    CustomerID string            `json:"customer_id"`
    PlanExpr   string            `json:"plan_expression"`
    Scenarios  []map[string]float64 `json:"scenarios"`
    StartTime  time.Time         `json:"start_time"`
    EndTime    time.Time         `json:"end_time"`
}

type SimulateResponse struct {
    Results       []EvaluationResult `json:"results"`
    TotalRevenue  float64            `json:"total_revenue"`
    AvgRevenue    float64            `json:"avg_revenue"`
    CostBreakdown map[string]float64 `json:"cost_breakdown"`
}
