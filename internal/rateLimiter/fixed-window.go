package rateLimiter

import (
	"sync"
	"time"
)

// FixedWindowRateLimiter implements a fixed window rate limiting algorithm.
type FixedWindowRateLimiter struct {
	sync.RWMutex
	clients map[string]int // map to track requests per IP
	limit   int            // maximum requests allowed per IP
	window  time.Duration  // time window in seconds for rate limiting
}

func NewFixedWindowRateLimiter(limit int, window time.Duration) *FixedWindowRateLimiter {
	return &FixedWindowRateLimiter{
		clients: make(map[string]int),
		limit:   limit,
		window:  window,
	}
}

// Allow checks if the request from the given IP is allowed based on the rate limit.
func (rl *FixedWindowRateLimiter) Allow(ip string) (bool, time.Duration) {
	rl.Lock()
	count, exists := rl.clients[ip] // get the current request count for the IP
	rl.Unlock()

	// If the IP does not exist in the map, it means it's the first request or if the count is less than the limit
	if !exists || count < rl.limit {
		rl.Lock()
		if !exists {
			go rl.resetCount(ip) // start a goroutine to reset the count after the time window expires
		}

		rl.clients[ip]++ // increment the request count for the IP
		rl.Unlock()
		return true, 0 // allow the request and return no retry duration
	}

	return false, rl.window // else deny the request and return the time window for retry
}

// resetCount resets the request count for the given IP after the time window expires.
func (rl *FixedWindowRateLimiter) resetCount(ip string) {
	time.Sleep(rl.window)

	rl.Lock()
	delete(rl.clients, ip)
	rl.Unlock()
}
