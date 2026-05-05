package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type VehicleType string

const (
	VehicleMotorcycle VehicleType = "MOTORCYCLE"
	VehicleCar        VehicleType = "CAR"
	VehicleCarXL      VehicleType = "CAR_XL"
)

type DriverStatus string

const (
	DriverStatusOnline  DriverStatus = "ONLINE_AVAILABLE"
	DriverStatusOffline DriverStatus = "OFFLINE"
)

type DriverProfile struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	UserID         uuid.UUID      `gorm:"type:uuid;uniqueIndex;not null" json:"user_id"`
	User           User           `gorm:"foreignKey:UserID" json:"-"`

	// Vehicle info
	VehicleType    VehicleType    `gorm:"type:varchar(20);not null" json:"vehicle_type"`
	VehicleBrand   string         `gorm:"type:varchar(100)" json:"vehicle_brand"`
	VehicleModel   string         `gorm:"type:varchar(100)" json:"vehicle_model"`
	VehicleYear    int            `gorm:"type:smallint" json:"vehicle_year"`
	VehicleColor   string         `gorm:"type:varchar(50)" json:"vehicle_color"`
	LicensePlate   string         `gorm:"type:varchar(20);uniqueIndex;not null" json:"license_plate"`
	LicenseNumber  string         `gorm:"type:varchar(50);not null" json:"license_number"`
	LicenseExpiry  time.Time      `gorm:"type:date;not null" json:"license_expiry"`

	// Real-time state
	Latitude       *float64       `gorm:"type:float" json:"latitude,omitempty"`
	Longitude      *float64       `gorm:"type:float" json:"longitude,omitempty"`
	Heading        *float64       `gorm:"type:float" json:"heading,omitempty"`
	Speed          *float64       `gorm:"type:float" json:"speed,omitempty"`
	IsOnline       bool           `gorm:"not null;default:false" json:"is_online"`
	IsAvailable    bool           `gorm:"not null;default:true" json:"is_available"`
	LastLocationAt *time.Time     `json:"last_location_at,omitempty"`

	// Stats
	RatingAvg      float64        `gorm:"type:numeric(3,2);not null;default:5.00" json:"rating_avg"`
	TotalTrips     int            `gorm:"not null;default:0" json:"total_trips"`
	AcceptanceRate float64        `gorm:"type:numeric(5,2);not null;default:100.00" json:"acceptance_rate"`

	// Verification
	IsVerified     bool           `gorm:"not null;default:false" json:"is_verified"`
	VerifiedAt     *time.Time     `json:"verified_at,omitempty"`

	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (DriverProfile) TableName() string {
	return "driver_profiles"
}
