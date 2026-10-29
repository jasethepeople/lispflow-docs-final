package webhook

import (
    "bytes"
    "context"
    "encoding/json"
    "log"
    "net/http"
    "time"
    "sync"
    "github.com/livininthelast/lispflow/pkg/billing"
)

type Dispatcher struct {
    endpoints []string
    client *http.Client
    mu sync.Mutex
    retries int
}

func New(endpoints []string) *Dispatcher {
    return &Dispatcher{
        endpoints: endpoints,
        client: &http.Client{Timeout: 5*time.Second},
        retries: 3,
    }
}

type Event struct {
    Type string `json:"type"`
    CustomerID string `json:"customer_id"`
    Result billing.EvaluationResult `json:"result"`
    Timestamp time.Time `json:"timestamp"`
}

func (d *Dispatcher) Dispatch(ctx context.Context, ev Event){
    if len(d.endpoints)==0 { return }
    data, _ := json.Marshal(ev)
    var wg sync.WaitGroup
    for _, ep := range d.endpoints {
        wg.Add(1)
        go func(url string){
            defer wg.Done()
            for i:=0;i<d.retries;i++{
                req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(data))
                req.Header.Set("Content-Type","application/json")
                resp, err := d.client.Do(req)
                if err==nil && resp.StatusCode < 300 {
                    resp.Body.Close()
                    return
                }
                if resp != nil { resp.Body.Close() }
                time.Sleep(time.Duration(1<<i) * time.Second)
            }
            log.Printf("webhook failed after %d retries: %s", d.retries, url)
        }(ep)
    }
    wg.Wait()
}
