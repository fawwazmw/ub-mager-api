package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidCoordinate(t *testing.T) {
	tests := []struct {
		name     string
		lat, lng float64
		expected bool
	}{
		{"valid UB campus", -7.9526, 112.6146, true},
		{"valid equator", 0, 0, true},
		{"valid north pole", 90, 180, true},
		{"valid south pole", -90, -180, true},
		{"invalid lat too high", 91, 112, false},
		{"invalid lat too low", -91, 112, false},
		{"invalid lng too high", -7, 181, false},
		{"invalid lng too low", -7, -181, false},
		{"both invalid", 100, 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ValidCoordinate(tt.lat, tt.lng))
		})
	}
}
