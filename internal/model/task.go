package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TaskCategory string

const (
	TaskCategoryJastipMakanan TaskCategory = "JASTIP_MAKANAN"
	TaskCategoryJastipBarang  TaskCategory = "JASTIP_BARANG"
	TaskCategoryTitipPrint    TaskCategory = "TITIP_PRINT"
	TaskCategoryAntarJemput   TaskCategory = "ANTAR_JEMPUT"
	TaskCategoryOther         TaskCategory = "OTHER"
)

type TaskStatus string

const (
	TaskStatusOpen       TaskStatus = "OPEN"
	TaskStatusAccepted   TaskStatus = "ACCEPTED"
	TaskStatusPickingUp  TaskStatus = "PICKING_UP"
	TaskStatusDelivering TaskStatus = "DELIVERING"
	TaskStatusCompleted  TaskStatus = "COMPLETED"
	TaskStatusCancelled  TaskStatus = "CANCELLED"
)

var TaskTerminalStatuses = []TaskStatus{TaskStatusCompleted, TaskStatusCancelled}

type Task struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	CreatorID   uuid.UUID    `gorm:"type:uuid;not null;index" json:"creator_id"`
	HelperID    *uuid.UUID   `gorm:"type:uuid;index" json:"helper_id,omitempty"`
	Category    TaskCategory `gorm:"type:varchar(30);not null;index" json:"category"`
	Status      TaskStatus   `gorm:"type:varchar(20);not null;default:'OPEN';index:idx_tasks_status_created" json:"status"`
	Title       string       `gorm:"type:varchar(200);not null" json:"title"`
	Description string       `gorm:"type:text;not null" json:"description"`
	Fee         float64      `gorm:"type:numeric(12,2);not null" json:"fee"`

	PickupLat     float64    `json:"pickup_lat"`
	PickupLng     float64    `json:"pickup_lng"`
	PickupAddress string     `gorm:"type:text" json:"pickup_address"`
	PickupZone    CampusZone `gorm:"type:varchar(20)" json:"pickup_zone,omitempty"`

	DeliveryLat     float64    `json:"delivery_lat"`
	DeliveryLng     float64    `json:"delivery_lng"`
	DeliveryAddress string     `gorm:"type:text" json:"delivery_address"`
	DeliveryZone    CampusZone `gorm:"type:varchar(20)" json:"delivery_zone,omitempty"`

	Notes string `gorm:"type:text" json:"notes,omitempty"`

	AcceptedAt  *time.Time `json:"accepted_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`

	CreatedAt time.Time      `gorm:"index:idx_tasks_status_created" json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (Task) TableName() string {
	return "tasks"
}
