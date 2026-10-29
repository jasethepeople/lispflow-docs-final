package middleware

import (
    "net/http"
    "sync"
    "time"
    "golang.org/x/time/rate"
)

func Logging(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
        start := time.Now()
        next.ServeHTTP(w, r)
        _ = start
    })
}

func Recovery(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
        defer func(){
            if rec := recover(); rec != nil {
                http.Error(w, "internal error", 500)
            }
        }()
        next.ServeHTTP(w, r)
    })
}

type IPRateLimiter struct {
    mu sync.Mutex
    limiters map[string]*rate.Limiter
    r rate.Limit
    b int
}

func NewIPRateLimiter(rps int, burst int) *IPRateLimiter {
    return &IPRateLimiter{limiters: make(map[string]*rate.Limiter), r: rate.Limit(rps), b: burst}
}

func (i *IPRateLimiter) get(ip string) *rate.Limiter {
    i.mu.Lock()
    defer i.mu.Unlock()
    if l, ok := i.limiters[ip]; ok { return l }
    l := rate.NewLimiter(i.r, i.b)
    i.limiters[ip]=l
    return l
}

func (i *IPRateLimiter) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
        ip := r.RemoteAddr
        if !i.get(ip).Allow(){
            http.Error(w, "rate limited", 429)
            return
        }
        next.ServeHTTP(w, r)
    })
}

func CORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
        if r.Method=="OPTIONS" { w.WriteHeader(200); return }
        next.ServeHTTP(w, r)
    })
}
