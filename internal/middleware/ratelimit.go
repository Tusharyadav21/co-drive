package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"co-drive/pkg/response"
)

type RateLimiter struct {
	mu       sync.Mutex
	requests map[string][]time.Time
	rate     int
	window   time.Duration
}

func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	l := &RateLimiter{
		requests: make(map[string][]time.Time),
		rate:     rate,
		window:   window,
	}

	go func() {
		for {
			time.Sleep(window)
			l.mu.Lock()
			now := time.Now()
			for ip, times := range l.requests {
				var valid []time.Time
				for _, t := range times {
					if now.Sub(t) < window {
						valid = append(valid, t)
					}
				}
				if len(valid) == 0 {
					delete(l.requests, ip)
				} else {
					l.requests[ip] = valid
				}
			}
			l.mu.Unlock()
		}
	}()

	return l
}

func (l *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := getClientIP(r)

		l.mu.Lock()
		now := time.Now()
		times := l.requests[ip]

		var valid []time.Time
		for _, t := range times {
			if now.Sub(t) < l.window {
				valid = append(valid, t)
			}
		}

		if len(valid) >= l.rate {
			oldest := valid[0]
			retryAfter := int(l.window.Seconds() - now.Sub(oldest).Seconds())
			if retryAfter <= 0 {
				retryAfter = 1
			}
			l.mu.Unlock()

			w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
			response.Error(w, http.StatusTooManyRequests, fmt.Sprintf("Too many requests. Please retry in %d seconds.", retryAfter))
			return
		}

		l.requests[ip] = append(valid, now)
		l.mu.Unlock()

		next.ServeHTTP(w, r)
	})
}

func getClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}
	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return strings.TrimSpace(xrip)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
