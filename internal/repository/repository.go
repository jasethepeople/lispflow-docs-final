package repository

import (
    "database/sql"
    "fmt"
    "time"

    "github.com/livininthelast/lispflow/pkg/billing"
    _ "github.com/lib/pq"
)

type Repo struct {
    db *sql.DB
}

func New(dsn string) (*Repo, error) {
    db, err := sql.Open("postgres", dsn)
    if err != nil { return nil, err }
    db.SetMaxOpenConns(20)
    db.SetMaxIdleConns(5)
    if err := db.Ping(); err != nil { return nil, err }
    return &Repo{db: db}, nil
}

func NewWithDB(db *sql.DB) *Repo { return &Repo{db: db} }

func (r *Repo) CreatePlan(p billing.Plan) error {
    _, err := r.db.Exec(`INSERT INTO plans(id, customer_id, version, expression, active, created_at) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT (customer_id, version) DO UPDATE SET expression=$4, active=$5`, p.ID, p.CustomerID, p.Version, p.Expression, p.Active, time.Now())
    return err
}

func (r *Repo) GetActivePlan(customerID string) (*billing.Plan, error) {
    var p billing.Plan
    err := r.db.QueryRow(`SELECT id, customer_id, version, expression, active, created_at FROM plans WHERE customer_id=$1 AND active=true ORDER BY version DESC LIMIT 1`, customerID).Scan(&p.ID, &p.CustomerID, &p.Version, &p.Expression, &p.Active, &p.CreatedAt)
    if err != nil { return nil, err }
    return &p, nil
}

func (r *Repo) DeactivateOldPlans(customerID string) error {
    _, err := r.db.Exec(`UPDATE plans SET active=false WHERE customer_id=$1 AND active=true`, customerID)
    return err
}

func (r *Repo) InsertLedger(e billing.LedgerEntry) error {
    _, err := r.db.Exec(`INSERT INTO ledger(id, customer_id, plan_id, amount, metrics, created_at) VALUES($1,$2,$3,$4,$5,$6)`, e.ID, e.CustomerID, e.PlanID, e.Amount, fmt.Sprintf("%v", e.Metrics), e.CreatedAt)
    return err
}

func (r *Repo) ListLedger(customerID string, limit, offset int) ([]billing.LedgerEntry, error) {
    rows, err := r.db.Query(`SELECT id, customer_id, plan_id, amount, created_at FROM ledger WHERE customer_id=$1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`, customerID, limit, offset)
    if err != nil { return nil, err }
    defer rows.Close()
    var out []billing.LedgerEntry
    for rows.Next() {
        var le billing.LedgerEntry
        if err := rows.Scan(&le.ID, &le.CustomerID, &le.PlanID, &le.Amount, &le.CreatedAt); err != nil { return nil, err }
        out = append(out, le)
    }
    return out, nil
}
