package ingestion

import (
    "sync"
    "time"
    "github.com/livininthelast/lispflow/pkg/billing"
)

type Batcher struct {
    mu      sync.Mutex
    buffer  []billing.UsageEvent
    size    int
    flushCh chan []billing.UsageEvent
    ticker  *time.Ticker
}

func New(size int, flushInterval time.Duration) *Batcher {
    b := &Batcher{size: size, flushCh: make(chan []billing.UsageEvent, 10), ticker: time.NewTicker(flushInterval)}
    go b.loop()
    return b
}

func (b *Batcher) loop(){
    for range b.ticker.C { b.Flush() }
}

func (b *Batcher) Add(e billing.UsageEvent){
    b.mu.Lock()
    b.buffer = append(b.buffer, e)
    should := len(b.buffer) >= b.size
    b.mu.Unlock()
    if should { b.Flush() }
}

func (b *Batcher) Flush(){
    b.mu.Lock()
    if len(b.buffer)==0 { b.mu.Unlock(); return }
    batch := make([]billing.UsageEvent, len(b.buffer))
    copy(batch, b.buffer)
    b.buffer = nil
    b.mu.Unlock()
    select { case b.flushCh <- batch: default: }
}

func (b *Batcher) Batches() <-chan []billing.UsageEvent { return b.flushCh }
