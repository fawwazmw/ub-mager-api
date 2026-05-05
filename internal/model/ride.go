package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RideStatus string

const (
	RideStatusSearching       RideStatus = "SEARCHING"
	RideStatusMatched         RideStatus = "MATCHED"
	RideStatusDriverEnRoute   RideStatus = "DRIVER_EN_ROUTE"
	RideStatusArrivedAtPickup RideStatus = "ARRIVED_AT_PICKUP"
	RideStatusInProgress      RideStatus = "IN_PROGRESS"
	RideStatusCompleted       RideStatus = "COMPLETED"
	RideStatusCancelled       RideStatus = "CANCELLED"
)

var TerminalStatuses = []RideStatus{RideStatusCompleted, RideStatusCancelled}

type PaymentMethod string

const (
	PaymentCash    PaymentMethod = "CASH"
	PaymentEwallet PaymentMethod = "EWALLET"
)

type Ride struct {
	ID               uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	PassengerID      uuid.UUID      `gorm:"type:uuid;not null;index" json:"passenger_id"`
	DriverID         *uuid.UUID     `gorm:"type:uuid;index" json:"driver_id,omitempty"`
	Status           RideStatus     `gorm:"type:varchar(30);not null;default:'SEARCHING';index" json:"status"`
	VehicleType      VehicleType    `gorm:"type:varchar(20);not null" json:"vehicle_type"`

	// Locations
	PickupLat        float64        `gorm:"not null" json:"pickup_lat"`
	PickupLng        float64        `gorm:"not null" json:"pickup_lng"`
	PickupAddress    string         `gorm:"type:text;not null" json:"pickup_address"`
	DropoffLat       float64        `gorm:"not null" json:"dropoff_lat"`
	DropoffLng       float64        `gorm:"not null" json:"dropoff_lng"`
	DropoffAddress   string         `gorm:"type:text;not null" json:"dropoff_address"`

	// Route & fare
	EstimatedDistanceM float64     `json:"estimated_distance_m"`
	EstimatedDurationS int         `json:"estimated_duration_s"`
	ActualDistanceM    float64     `json:"actual_distance_m"`
	ActualDurationS    int         `json:"actual_duration_s"`
	BaseFare           float64     `gorm:"type:numeric(12,2);not null;default:0" json:"base_fare"`
	SurgeMultiplier    float64     `gorm:"type:numeric(4,2);not null;default:1.00" json:"surge_multiplier"`
	TotalFare          float64     `gorm:"type:numeric(12,2);not null;default:0" json:"total_fare"`

	// Payment
	PaymentMethod    PaymentMethod  `gorm:"type:varchar(20);not null;default:'CASH'" json:"payment_method"`

	// Notes
	Notes            string         `gorm:"type:text" json:"notes,omitempty"`

	// Timestamps
	RequestedAt      time.Time      `gorm:"not null;default:now()" json:"requested_at"`
	MatchedAt        *time.Time     `json:"matched_at,omitempty"`
	DriverArrivedAt  *time.Time     `json:"driver_arrived_at,omitempty"`
	PickedUpAt       *time.Time     `json:"picked_up_at,omitempty"`
	CompletedAt      *time.Time     `json:"completed_at,omitempty"`
	CancelledAt      *time.Time     `json:"cancelled_at,omitempty"`
	CancellationReason string       `gorm:"type:text" json:"cancellation_reason,omitempty"`

	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Ride) TableName() string {
	return "rides"
}

type Rating struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	RideID    uuid.UUID `gorm:"type:uuid;not null;index" json:"ride_id"`
	RaterID   uuid.UUID `gorm:"type:uuid;not null" json:"rater_id"`
	RateeID   uuid.UUID `gorm:"type:uuid;not null;index" json:"ratee_id"`
	Score     int       `gorm:"type:smallint;not null" json:"score"`
	Comment   string    `gorm:"type:text" json:"comment,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

func (Rating) TableName() string {
	return "ratings"
}
