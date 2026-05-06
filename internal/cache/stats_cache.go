package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	dashboardStatsKey = "cache:dashboard_stats"
	statsTTL          = 10 * time.Second
)

type StatsCache struct {
	rdb *redis.Client
}

func NewStatsCache(rdb *redis.Client) *StatsCache {
	return &StatsCache{rdb: rdb}
}

func (c *StatsCache) Get(ctx context.Context) ([]byte, error) {
	return c.rdb.Get(ctx, dashboardStatsKey).Bytes()
}

func (c *StatsCache) Set(ctx context.Context, data any) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return c.rdb.Set(ctx, dashboardStatsKey, bytes, statsTTL).Err()
}

func (c *StatsCache) Invalidate(ctx context.Context) error {
	return c.rdb.Del(ctx, dashboardStatsKey).Err()
}
