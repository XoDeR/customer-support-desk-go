package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Limiter struct{ client *redis.Client }

func NewLimiter(client *redis.Client) *Limiter { return &Limiter{client} }

// Allow uses a Redis sorted-set sliding window.
func (l *Limiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	now := time.Now()
	member := fmt.Sprintf("%d-%d", now.UnixNano(), time.Now().UnixMicro())
	pipe := l.client.TxPipeline()
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprint(now.Add(-window).UnixMilli()))
	count := pipe.ZCard(ctx, key)
	pipe.ZAdd(ctx, key, redis.Z{Score: float64(now.UnixMilli()), Member: member})
	pipe.Expire(ctx, key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, err
	}
	return count.Val() < int64(limit), nil
}
