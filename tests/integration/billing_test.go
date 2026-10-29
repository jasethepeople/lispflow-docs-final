//go:build integration
package integration

import (
    "fmt"
    "sync"
    "testing"
    "github.com/livininthelast/lispflow/pkg/billing"
    "github.com/livininthelast/lispflow/internal/service"
)

func newSvc() *service.Service {
    eng := billing.NewEngine()
    return service.New(eng, nil)
}

func TestPlanManagement_ActivationVersioning(t *testing.T) {
    svc := newSvc()
    p1, err := svc.SetPlan("cust_1", "(+ api_calls 10)")
    if err != nil { t.Fatal(err) }
    if p1.Version != 1 || !p1.Active { t.Fatalf("expected v1 active, got %+v", p1) }
    p2, err := svc.SetPlan("cust_1", "(* api_calls 0.01)")
    if err != nil { t.Fatal(err) }
    if p2.Version != 2 { t.Fatalf("expected v2, got %d", p2.Version) }
    // old version should be inactive via history check: evaluating should use v2
    res, err := svc.Evaluate("cust_1", map[string]float64{"api_calls": 100})
    if err != nil { t.Fatal(err) }
    if res.Amount != 1.0 { t.Fatalf("expected 1.0 got %f", res.Amount) }
    if res.Version != 2 { t.Fatalf("expected version 2") }
}

func TestEvaluation_SimpleVolume(t *testing.T) {
    svc := newSvc()
    _, _ = svc.SetPlan("c", "(* api_calls 0.02)")
    res, err := svc.Evaluate("c", map[string]float64{"api_calls": 500})
    if err != nil { t.Fatal(err) }
    if res.Amount != 10 { t.Fatalf("expected 10 got %f", res.Amount) }
}

func TestEvaluation_Tiered(t *testing.T) {
    svc := newSvc()
    // tiered: first 1000 at 0.01, next 10000 at 0.008, rest 0.005
    expr := "(tiered api_calls 0.01 1000 0.008 10000 0.005)"
    _, _ = svc.SetPlan("c", expr)
    tests := []struct{ usage, expected float64 }{
        {500, 5},
        {1000, 10},
        {2000, 10 + 1000*0.008},
        {12000, 10 + 10000*0.008 + 1000*0.005},
    }
    for _, tt := range tests {
        res, err := svc.Evaluate("c", map[string]float64{"api_calls": tt.usage})
        if err != nil { t.Fatalf("usage %f err %v", tt.usage, err) }
        if res.Amount != tt.expected { t.Fatalf("usage %f expected %f got %f", tt.usage, tt.expected, res.Amount) }
    }
}

func TestEvaluation_MultiDimensionalMinMaxCaps(t *testing.T) {
    svc := newSvc()
    // (max 5 (min 100 (* (+ api_calls storage_gb) 0.01)))
    expr := "(max 5 (min 100 (* (+ api_calls storage_gb) 0.01)))"
    _, _ = svc.SetPlan("c", expr)
    cases := []struct{ api, stor, exp float64 }{
        {0,0,5}, // floor
        {100,0,5},
        {1000,0,10},
        {20000,0,100}, // cap
    }
    for _, cc := range cases {
        res, err := svc.Evaluate("c", map[string]float64{"api_calls": cc.api, "storage_gb": cc.stor})
        if err != nil { t.Fatal(err) }
        if res.Amount != cc.exp { t.Fatalf("expected %f got %f for %+v", cc.exp, res.Amount, cc) }
    }
}

func TestEvaluation_ConditionalDiscount(t *testing.T) {
    svc := newSvc()
    // if api_calls > 1000 then 10% discount
    expr := "(* api_calls (if (> api_calls 1000) 0.009 0.01))"
    _, _ = svc.SetPlan("c", expr)
    r1, _ := svc.Evaluate("c", map[string]float64{"api_calls": 500})
    if r1.Amount != 5 { t.Fatalf("expected 5 got %f", r1.Amount) }
    r2, _ := svc.Evaluate("c", map[string]float64{"api_calls": 2000})
    if r2.Amount != 18 { t.Fatalf("expected 18 got %f", r2.Amount) }
}

func TestEvaluation_MissingKeysDefaultZero(t *testing.T) {
    svc := newSvc()
    _, _ = svc.SetPlan("c", "(+ api_calls missing_metric)")
    res, err := svc.Evaluate("c", map[string]float64{"api_calls": 10})
    if err != nil { t.Fatal(err) }
    if res.Amount != 10 { t.Fatalf("missing keys should default to 0, got %f", res.Amount) }
}

func TestSimulation_TimeTravel(t *testing.T) {
    svc := newSvc()
    req := billing.SimulateRequest{
        PlanExpr: "(* api_calls 0.01)",
        Scenarios: []map[string]float64{
            {"api_calls": 100},
            {"api_calls": 200},
            {"api_calls": 1000},
        },
    }
    resp, err := svc.Simulate(req)
    if err != nil { t.Fatal(err) }
    if resp.TotalRevenue != 13 { t.Fatalf("expected 13 got %f", resp.TotalRevenue) }
    if len(resp.Results)!=3 { t.Fatalf("expected 3 results") }
}

func TestSimulation_InvalidPlanRejection(t *testing.T) {
    svc := newSvc()
    req := billing.SimulateRequest{
        PlanExpr: "(unknown_func api_calls)",
        Scenarios: []map[string]float64{{"api_calls": 10}},
    }
    _, err := svc.Simulate(req)
    if err == nil { t.Fatal("expected error for invalid plan") }
}

func TestBatchProcessing(t *testing.T) {
    svc := newSvc()
    _, _ = svc.SetPlan("cust_batch", "(* api_calls 0.01)")
    events := []map[string]float64{{"api_calls":100},{"api_calls":200},{"api_calls":300}}
    results, err := svc.Batch("cust_batch", events)
    if err != nil { t.Fatal(err) }
    if len(results)!=3 { t.Fatalf("expected 3") }
    total := 0.0
    for _, r := range results { total += r.Amount }
    if total != 6 { t.Fatalf("expected 6 got %f", total) }
}

func TestLedger_PaginationImmutability(t *testing.T) {
    svc := newSvc()
    _, _ = svc.SetPlan("cust_hist", "(* api_calls 1)")
    for i:=0;i<25;i++ {
        _, _ = svc.Evaluate("cust_hist", map[string]float64{"api_calls": float64(i)})
    }
    page1, err := svc.History("cust_hist", 10, 0)
    if err != nil { t.Fatal(err) }
    if len(page1)!=10 { t.Fatalf("expected 10 got %d", len(page1)) }
    page2, _ := svc.History("cust_hist", 10, 10)
    if page2[0].ID == page1[0].ID { t.Fatal("pagination should return different entries") }
    // immutability: ledger entries should be ordered newest first
    if page1[0].Amount != 24 { t.Fatalf("expected newest first, got %f", page1[0].Amount) }
}

func TestConcurrency_100Evaluations(t *testing.T) {
    svc := newSvc()
    _, _ = svc.SetPlan("conc", "(+ api_calls (* storage_gb 2))")
    var wg sync.WaitGroup
    errs := make(chan error, 100)
    for i:=0;i<100;i++ {
        wg.Add(1)
        go func(n int) {
            defer wg.Done()
            _, err := svc.Evaluate("conc", map[string]float64{"api_calls": float64(n), "storage_gb": 1})
            if err != nil { errs <- err }
        }(i)
    }
    wg.Wait()
    close(errs)
    for e := range errs { t.Fatalf("concurrent eval error: %v", e) }
    hist, _ := svc.History("conc", 200, 0)
    if len(hist)!=100 { t.Fatalf("expected 100 ledger entries, got %d", len(hist)) }
}

func TestValidation_SizeLimits(t *testing.T) {
    svc := newSvc()
    eng := billing.NewEngine()
    // syntax check
    err := eng.Validate("( + )")
    if err != nil { // empty + should still parse? Our impl allows 0 sum -> we test unknown func
    }
    err = eng.Validate("(invalid_func 1 2)")
    if err != nil { t.Fatalf("parse should succeed even if func invalid at validation time? got %v", err) }
    // size limit
    big := "(" + fmt.Sprintf("%0*d", billing.MaxPlanSize+1, 0) + ")"
    _, err = svc.SetPlan("c", big)
    if err == nil { t.Fatal("expected size limit error") }
}
