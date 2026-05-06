package middleware

import (
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

type Metrics struct {
	TotalRequests   atomic.Int64
	TotalErrors     atomic.Int64
	ActiveRequests  atomic.Int64
	RequestDuration sync.Map
	StatusCounts    sync.Map
}

var GlobalMetrics = &Metrics{}

func MetricsCollector() gin.HandlerFunc {
	return func(c *gin.Context) {
		GlobalMetrics.TotalRequests.Add(1)
		GlobalMetrics.ActiveRequests.Add(1)
		start := time.Now()

		c.Next()

		GlobalMetrics.ActiveRequests.Add(-1)
		duration := time.Since(start)

		status := c.Writer.Status()
		if status >= 500 {
			GlobalMetrics.TotalErrors.Add(1)
		}

		statusKey := strconv.Itoa(status)
		if val, ok := GlobalMetrics.StatusCounts.Load(statusKey); ok {
			val.(*atomic.Int64).Add(1)
		} else {
			counter := &atomic.Int64{}
			counter.Add(1)
			GlobalMetrics.StatusCounts.Store(statusKey, counter)
		}

		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		if val, ok := GlobalMetrics.RequestDuration.Load(path); ok {
			entry := val.(*durationEntry)
			entry.add(duration)
		} else {
			entry := &durationEntry{}
			entry.add(duration)
			GlobalMetrics.RequestDuration.Store(path, entry)
		}
	}
}

type durationEntry struct {
	count atomic.Int64
	total atomic.Int64
}

func (d *durationEntry) add(dur time.Duration) {
	d.count.Add(1)
	d.total.Add(int64(dur))
}

func (d *durationEntry) avg() time.Duration {
	c := d.count.Load()
	if c == 0 {
		return 0
	}
	return time.Duration(d.total.Load() / c)
}

type MetricsSnapshot struct {
	TotalRequests  int64                      `json:"total_requests"`
	TotalErrors    int64                      `json:"total_errors"`
	ActiveRequests int64                      `json:"active_requests"`
	Uptime         string                     `json:"uptime"`
	StatusCounts   map[string]int64           `json:"status_counts"`
	Endpoints      map[string]EndpointMetrics `json:"endpoints"`
}

type EndpointMetrics struct {
	Count int64 `json:"count"`
	AvgMs int64 `json:"avg_ms"`
}

var startTime = time.Now()

func GetMetricsSnapshot() MetricsSnapshot {
	snap := MetricsSnapshot{
		TotalRequests:  GlobalMetrics.TotalRequests.Load(),
		TotalErrors:    GlobalMetrics.TotalErrors.Load(),
		ActiveRequests: GlobalMetrics.ActiveRequests.Load(),
		Uptime:         time.Since(startTime).Round(time.Second).String(),
		StatusCounts:   make(map[string]int64),
		Endpoints:      make(map[string]EndpointMetrics),
	}

	GlobalMetrics.StatusCounts.Range(func(key, value any) bool {
		snap.StatusCounts[key.(string)] = value.(*atomic.Int64).Load()
		return true
	})

	GlobalMetrics.RequestDuration.Range(func(key, value any) bool {
		entry := value.(*durationEntry)
		snap.Endpoints[key.(string)] = EndpointMetrics{
			Count: entry.count.Load(),
			AvgMs: entry.avg().Milliseconds(),
		}
		return true
	})

	return snap
}
