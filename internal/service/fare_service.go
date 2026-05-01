package service

import "github.com/wardayadev/ub-mager-api/internal/model"

// FareConfig holds pricing parameters per vehicle type
type FareConfig struct {
	BaseFare      float64 // IDR
	PerKmRate     float64 // IDR per km
	PerMinuteRate float64 // IDR per minute
	MinFare       float64 // Minimum fare
	PlatformFee   float64 // Flat platform fee
}

var fareConfigs = map[model.VehicleType]FareConfig{
	model.VehicleMotorcycle: {
		BaseFare:      8000,
		PerKmRate:     3000,
		PerMinuteRate: 200,
		MinFare:       10000,
		PlatformFee:   2000,
	},
	model.VehicleCar: {
		BaseFare:      15000,
		PerKmRate:     5000,
		PerMinuteRate: 500,
		MinFare:       25000,
		PlatformFee:   3000,
	},
	model.VehicleCarXL: {
		BaseFare:      20000,
		PerKmRate:     7000,
		PerMinuteRate: 700,
		MinFare:       35000,
		PlatformFee:   4000,
	},
}

type FareBreakdown struct {
	BaseFare        float64 `json:"base_fare"`
	DistanceFare    float64 `json:"distance_fare"`
	TimeFare        float64 `json:"time_fare"`
	SurgeMultiplier float64 `json:"surge_multiplier"`
	SurgeAmount     float64 `json:"surge_amount"`
	PlatformFee     float64 `json:"platform_fee"`
	TotalEstimate   float64 `json:"total_estimate"`
}

// CalculateFare computes the fare for a ride
func CalculateFare(vehicleType model.VehicleType, distanceM float64, durationS int, surgeMultiplier float64) FareBreakdown {
	cfg, ok := fareConfigs[vehicleType]
	if !ok {
		cfg = fareConfigs[model.VehicleMotorcycle]
	}

	if surgeMultiplier < 1.0 {
		surgeMultiplier = 1.0
	}
	if surgeMultiplier > 2.0 {
		surgeMultiplier = 2.0 // Cap at 2x
	}

	distanceKm := distanceM / 1000.0
	durationMin := float64(durationS) / 60.0

	distanceFare := distanceKm * cfg.PerKmRate
	timeFare := durationMin * cfg.PerMinuteRate
	subtotal := cfg.BaseFare + distanceFare + timeFare

	// Apply surge
	surgeAmount := subtotal * (surgeMultiplier - 1.0)
	surgedTotal := subtotal + surgeAmount

	// Apply minimum fare
	if surgedTotal < cfg.MinFare {
		surgedTotal = cfg.MinFare
	}

	total := surgedTotal + cfg.PlatformFee

	return FareBreakdown{
		BaseFare:        cfg.BaseFare,
		DistanceFare:    distanceFare,
		TimeFare:        timeFare,
		SurgeMultiplier: surgeMultiplier,
		SurgeAmount:     surgeAmount,
		PlatformFee:     cfg.PlatformFee,
		TotalEstimate:   total,
	}
}
