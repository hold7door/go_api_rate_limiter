package redis_rate_limiter

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

var (
	_ Strategy = &counterStrategy{}
)

const (
	keyThatDoesNotExist = -2
	keyWithoutExpire = -1
)

func NewCounterStrategy(client *redis.Client, now func() time.Time) *counterStrategy {
	return &counterStrategy{
		client: client,
		now:    now,
	}
}

type counterStrategy struct {
	client *redis.Client
	now func() time.Time

}

func (c *counterStrategy) Run(ctx context.Context, r*Request) (*Result, error){
	getPipeline := c.client.Pipeline()
	getResult := getPipeline.Get(ctx, r.Key)
	ttlResult := getPipeline.TTL(ctx, r.Key)

	if _, err := getPipeline.Exec(ctx); err != nil && !errors.Is(err, redis.Nil){
		return nil, errors.Wrapf(err, "failed to execute pipeline with get and ttl to key %v", r.Key)
	}

	var ttlDuration time.Duration

	if d, err := ttlResult.Result(); err != nil || d == keyWithoutExpire || d == keyThatDoesNotExist {
		ttlDuration = r.Duration
		if err := c.client.Expire(ctx, r.Key, r.Duration).Err(); err != nil {
			return nil, errors.Wrapf(err, "failed to set an expiration to key %v", r.Key)
		}
	} else {
		ttlDuration = d
	}
	
	expiresAt := c.now().Add(ttlDuration)

	if total, err := getResult.Uint64(); err != nil && errors.Is(err, redis.Nil){

	} else if total >= r.Limit {
		return &Result{
			State: Deny,
			TotalRequests: total,
			ExpiresAt: expiresAt,
		}, nil
	}

	// Increments the number stored at key by one. If the key does not exist, it is set to 0 before performing the operation
	incrResult := c.client.Incr(ctx, r.Key)

	totalRequests, err := incrResult.Uint64()

	if err != nil {
		return nil, errors.Wrapf(err, "failed to increment key %v", r.Key)
	}

	if totalRequests > r.Limit {
		return &Result{
			State: Deny,
			TotalRequests: totalRequests,
			ExpiresAt: expiresAt,
		}, nil
	}

	return &Result{
		State: Allow,
		TotalRequests: totalRequests,
		ExpiresAt: expiresAt,
	}, nil
}