CREATE TABLE IF NOT EXISTS plans (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    version INT NOT NULL,
    expression TEXT NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL,
    UNIQUE(customer_id, version)
);
CREATE INDEX IF NOT EXISTS idx_plans_customer_active ON plans(customer_id, active);

CREATE TABLE IF NOT EXISTS ledger (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL,
    plan_id TEXT NOT NULL,
    amount DOUBLE PRECISION NOT NULL,
    metrics TEXT,
    created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_ledger_customer_created ON ledger(customer_id, created_at DESC);
