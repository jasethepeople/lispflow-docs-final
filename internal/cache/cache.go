package cache

import (
    "context"
    "encoding/json"
    "time"
    "github.com/redis/go-redis/v9"
    "github.com/livininthelast/lispflow/pkg/billing"
)

type Cache struct {
    rdb *redis.Client
    ttl time.Duration
}

func New(redisURL string) *Cache {
    if redisURL=="" { return &Cache{ttl: 5*time.Minute} }
    opt, err := redis.ParseURL(redisURL)
    if err != nil { return &Cache{ttl: 5*time.Minute} }
    return &Cache{rdb: redis.NewClient(opt), ttl: 5*time.Minute}
}

func (c *Cache) GetPlan(ctx context.Context, customerID string) (*billing.Plan, bool) {
    if c.rdb==nil { return nil, false }
    data, err := c.rdb.Get(ctx, "plan:"+customerID).Result()
    if err != nil { return nil, false }
    var p billing.Plan
    if err := json.Unmarshal([]byte(data), &p); err != nil { return nil, false }
    return &p, true
}

func (c *Cache) SetPlan(ctx context.Context, p billing.Plan) {
    if c.rdb==nil { return }
    b, _ := json.Marshal(p)
    c.rdb.Set(ctx, "plan:"+p.CustomerID, b, c.ttl)
}

func (c *Cache) Invalidate(ctx context.Context, customerID string){
    if c.rdb==nil { return }
    c.rdb.Del(ctx, "plan:"+customerID)
}
