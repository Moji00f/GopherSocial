package ratelimiter

import "time"

type Limiter interface {
	Allow(ip string) (bool, time.Duration)
}

type Config struct {
	Enable              bool
	TimeFrame           time.Duration
	RequestPerTimeFrame int
}
