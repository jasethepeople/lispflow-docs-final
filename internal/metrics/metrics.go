package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
    EvaluationsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "lispflow_evaluations_total", Help: "Total evaluations"}, []string{"customer_id","status"})
    EvaluationDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{Name: "lispflow_evaluation_duration_seconds", Help: "Evaluation duration", Buckets: prometheus.DefBuckets}, []string{"customer_id"})
    LedgerEntries = prometheus.NewCounterVec(prometheus.CounterOpts{Name: "lispflow_ledger_entries_total", Help: "Ledger entries"}, []string{"customer_id"})
    BatchSize = prometheus.NewHistogram(prometheus.HistogramOpts{Name: "lispflow_batch_size", Help: "Batch size", Buckets: []float64{1,10,50,100,500,1000}})
    ActivePlans = prometheus.NewGauge(prometheus.GaugeOpts{Name: "lispflow_active_plans", Help: "Active plans"})
)

func init(){
    prometheus.MustRegister(EvaluationsTotal, EvaluationDuration, LedgerEntries, BatchSize, ActivePlans)
}
