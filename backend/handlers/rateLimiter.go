package handlers

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	users    map[string]*clientData
	mu       sync.Mutex
	limit    int
	duration time.Duration
}

type clientData struct {
	count     int
	resetTime time.Time
}

func NewRateLimiter(limit int, duration time.Duration) *RateLimiter {
	rl := &RateLimiter{
		users:    make(map[string]*clientData),
		limit:    limit, // todoTODO: verify session ID is valid from DB
		duration: duration,
	}
	go func() {
		for {
			time.Sleep(duration)
			rl.mu.Lock()
			now := time.Now()
			for ip, client := range rl.users {
				if now.After(client.resetTime) {
					delete(rl.users, ip)
				}
			}
			rl.mu.Unlock()
		}
	}()
	return rl
}

func (rl *RateLimiter) Allow(r *http.Request) bool {
	ip := r.RemoteAddr

	rl.mu.Lock()
	defer rl.mu.Unlock()
	now := time.Now()

	client, exists := rl.users[ip]
	if !exists || now.After(client.resetTime) {
		rl.users[ip] = &clientData{
			count:     1,
			resetTime: now.Add(rl.duration),
		}
		return true
	}
	if client.count >= rl.limit {
		return false
	}
	client.count++
	return true
}

func (rl *RateLimiter) RLMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !rl.Allow(r) {
			encodeJson(w, http.StatusTooManyRequests, nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}
