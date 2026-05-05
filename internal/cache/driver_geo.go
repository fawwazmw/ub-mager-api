package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/wardayadev/ub-mager-api/internal/model"
)

const driverGeoKey = "drivers:locations"

type DriverGeoCache struct {
	rdb *redis.Client
}

func NewDriverGeoCache(rdb *redis.Client) *DriverGeoCache {
	return &DriverGeoCache{rdb: rdb}
}

func (c *DriverGeoCache) UpdateLocation(ctx context.Context, driverID uuid.UUID, lat, lng float64) error {
	return c.rdb.GeoAdd(ctx, driverGeoKey, &redis.GeoLocation{
		Name:      driverID.String(),
		Longitude: lng,
		Latitude:  lat,
	}).Err()
}

func (c *DriverGeoCache) RemoveDriver(ctx context.Context, driverID uuid.UUID) error {
	return c.rdb.ZRem(ctx, driverGeoKey, driverID.String()).Err()
}

func (c *DriverGeoCache) FindNearby(ctx context.Context, lat, lng, radiusKm float64, count int) ([]string, error) {
	return c.rdb.GeoSearch(ctx, driverGeoKey, &redis.GeoSearchQuery{
		Longitude:  lng,
		Latitude:   lat,
		Radius:     radiusKm,
		RadiusUnit: "km",
		Count:      count,
		Sort:       "ASC",
	}).Result()
}

func (c *DriverGeoCache) FindNearbyWithDist(ctx context.Context, lat, lng, radiusKm float64, count int) ([]redis.GeoLocation, error) {
	results, err := c.rdb.GeoSearchLocation(ctx, driverGeoKey, &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Longitude:  lng,
			Latitude:   lat,
			Radius:     radiusKm,
			RadiusUnit: "km",
			Count:      count,
			Sort:       "ASC",
		},
		WithCoord: true,
		WithDist:  true,
	}).Result()
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (c *DriverGeoCache) SetDriverStatus(ctx context.Context, driverID uuid.UUID, status model.DriverStatus) error {
	key := fmt.Sprintf("driver:%s:status", driverID.String())
	ttl := 24 * time.Hour
	if status == model.DriverStatusOffline {
		ttl = 1 * time.Hour
	}
	return c.rdb.Set(ctx, key, string(status), ttl).Err()
}

func (c *DriverGeoCache) GetDriverStatus(ctx context.Context, driverID uuid.UUID) (model.DriverStatus, error) {
	key := fmt.Sprintf("driver:%s:status", driverID.String())
	val, err := c.rdb.Get(ctx, key).Result()
	return model.DriverStatus(val), err
}
