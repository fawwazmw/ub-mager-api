package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wardayadev/ub-mager-api/internal/model"
)

func TestCalculateFare_Motorcycle(t *testing.T) {
	breakdown := CalculateFare(model.VehicleMotorcycle, 3000, 600, 1.0)

	assert.Equal(t, 8000.0, breakdown.BaseFare)
	assert.Equal(t, 2000.0, breakdown.PlatformFee)
	assert.Equal(t, 1.0, breakdown.SurgeMultiplier)
	assert.Equal(t, 0.0, breakdown.SurgeAmount)
	assert.Greater(t, breakdown.TotalEstimate, 10000.0)
}

func TestCalculateFare_Car(t *testing.T) {
	breakdown := CalculateFare(model.VehicleCar, 5000, 900, 1.0)

	assert.Equal(t, 15000.0, breakdown.BaseFare)
	assert.Equal(t, 3000.0, breakdown.PlatformFee)
	assert.Greater(t, breakdown.TotalEstimate, 25000.0)
}

func TestCalculateFare_WithSurge(t *testing.T) {
	noSurge := CalculateFare(model.VehicleMotorcycle, 5000, 600, 1.0)
	withSurge := CalculateFare(model.VehicleMotorcycle, 5000, 600, 1.5)

	assert.Greater(t, withSurge.TotalEstimate, noSurge.TotalEstimate)
	assert.Equal(t, 1.5, withSurge.SurgeMultiplier)
	assert.Greater(t, withSurge.SurgeAmount, 0.0)
}

func TestCalculateFare_SurgeCappedAt2x(t *testing.T) {
	breakdown := CalculateFare(model.VehicleMotorcycle, 5000, 600, 5.0)

	assert.Equal(t, 2.0, breakdown.SurgeMultiplier)
}

func TestCalculateFare_SurgeMinimum1x(t *testing.T) {
	breakdown := CalculateFare(model.VehicleMotorcycle, 5000, 600, 0.5)

	assert.Equal(t, 1.0, breakdown.SurgeMultiplier)
	assert.Equal(t, 0.0, breakdown.SurgeAmount)
}

func TestCalculateFare_MinimumFare(t *testing.T) {
	breakdown := CalculateFare(model.VehicleMotorcycle, 100, 30, 1.0)

	assert.GreaterOrEqual(t, breakdown.TotalEstimate, 10000.0+2000.0)
}

func TestCalculateFare_UnknownVehicleDefaultsToMotorcycle(t *testing.T) {
	breakdown := CalculateFare("UNKNOWN", 3000, 600, 1.0)
	motorcycleBreakdown := CalculateFare(model.VehicleMotorcycle, 3000, 600, 1.0)

	assert.Equal(t, motorcycleBreakdown.TotalEstimate, breakdown.TotalEstimate)
}
