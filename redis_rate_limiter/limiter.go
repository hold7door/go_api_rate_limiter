package redis_rate_limiter

import (
	"context"
	"time"
)

type Request struct {
	Key string
	Limit uint64
	Duration time.Duration
}

type State int64

const (
	Deny State = 0
	Allow State = 1
)


type Result struct {
	State State
	TotalRequests uint64
	ExpiresAt time.Time
}

type Strategy interface {
	Run(ctx context.Context, r *Request) (*Result, error)
}