package cache

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const driverGeoKey = "drivers:locations"

type DriverGeoCache struct {
	rdb *redis.Client
}

func NewDriverGeoCache(rdb *redis.Client) *DriverGeoCache {
	return &DriverGeoCache{rdb: rdb}
}

// UpdateLocation adds/updates a driver's position in the Redis GeoSet
func (c *DriverGeoCache) UpdateLocation(ctx context.Context, driverID uuid.UUID, lat, lng float64) error {
	return c.rdb.GeoAdd(ctx, driverGeoKey, &redis.GeoLocation{
		Name:      driverID.String(),
		Longitude: lng,
		Latitude:  lat,
	}).Err()
}

// RemoveDriver removes a driver from the GeoSet (when going offline)
func (c *DriverGeoCache) RemoveDriver(ctx context.Context, driverID uuid.UUID) error {
	return c.rdb.ZRem(ctx, driverGeoKey, driverID.String()).Err()
}

// FindNearby returns driver IDs within radius (km) of a point
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

// FindNearbyWithDist returns driver IDs with distances
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

// SetDriverStatus stores driver online status in a hash
func (c *DriverGeoCache) SetDriverStatus(ctx context.Context, driverID uuid.UUID, status string) error {
	key := fmt.Sprintf("driver:%s:status", driverID.String())
	return c.rdb.Set(ctx, key, status, 0).Err()
}

// GetDriverStatus gets driver status
func (c *DriverGeoCache) GetDriverStatus(ctx context.Context, driverID uuid.UUID) (string, error) {
	key := fmt.Sprintf("driver:%s:status", driverID.String())
	return c.rdb.Get(ctx, key).Result()
}
