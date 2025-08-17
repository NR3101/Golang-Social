package rateLimiter

import "time"

// Limiter interface defines the methods for a rate limiter
type Limiter interface {
	Allow(ip string) (bool, time.Duration)
}

type Config struct {
	RequestPerTimeFrame int           // number of requests allowed per time frame
	TimeFrame           time.Duration // duration of the time frame for rate limiting
	Enabled             bool          // flag to enable or disable rate limiting
}
