package repository

import (
	"context"
	"time"

	"github.com/wardayadev/ub-mager-api/internal/model"
)

type RevenueStats struct {
	Period       string  `json:"period"`
	TotalRevenue float64 `json:"total_revenue"`
	TotalRides   int64   `json:"total_rides"`
	AvgFare      float64 `json:"avg_fare"`
	Currency     string  `json:"currency"`
}

func (r *AnalyticsRepository) GetRevenueStats(ctx context.Context, period string) (*RevenueStats, error) {
	since, period := parsePeriod(period)

	stats := &RevenueStats{Period: period, Currency: "IDR"}

	r.db.WithContext(ctx).Model(&model.Ride{}).
		Where("status = ? AND completed_at >= ?", model.RideStatusCompleted, since).
		Count(&stats.TotalRides)

	var agg struct {
		Total *float64
		Avg   *float64
	}
	r.db.WithContext(ctx).Model(&model.Ride{}).
		Select("COALESCE(SUM(total_fare), 0) as total, COALESCE(AVG(total_fare), 0) as avg").
		Where("status = ? AND completed_at >= ?", model.RideStatusCompleted, since).
		Scan(&agg)

	if agg.Total != nil {
		stats.TotalRevenue = *agg.Total
	}
	if agg.Avg != nil {
		stats.AvgFare = *agg.Avg
	}

	return stats, nil
}

type DailyRevenue struct {
	Date    string  `json:"date"`
	Revenue float64 `json:"revenue"`
	Rides   int64   `json:"rides"`
}

func (r *AnalyticsRepository) GetDailyRevenue(ctx context.Context, days int) ([]DailyRevenue, error) {
	if days < 1 || days > 90 {
		days = 7
	}

	since := startOfDay().AddDate(0, 0, -(days - 1))

	var results []DailyRevenue
	err := r.db.WithContext(ctx).
		Table("rides").
		Select("TO_CHAR(completed_at, 'YYYY-MM-DD') as date, COALESCE(SUM(total_fare), 0) as revenue, COUNT(*) as rides").
		Where("status = ? AND completed_at >= ?", model.RideStatusCompleted, since).
		Group("TO_CHAR(completed_at, 'YYYY-MM-DD')").
		Order("date ASC").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	filled := fillMissingDays(results, since, days)
	return filled, nil
}

func fillMissingDays(data []DailyRevenue, since time.Time, days int) []DailyRevenue {
	dateMap := make(map[string]DailyRevenue)
	for _, d := range data {
		dateMap[d.Date] = d
	}

	var result []DailyRevenue
	for i := 0; i < days; i++ {
		date := since.AddDate(0, 0, i).Format("2006-01-02")
		if d, ok := dateMap[date]; ok {
			result = append(result, d)
		} else {
			result = append(result, DailyRevenue{Date: date, Revenue: 0, Rides: 0})
		}
	}
	return result
}
