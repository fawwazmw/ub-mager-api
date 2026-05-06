package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/wardayadev/ub-mager-api/internal/model"
	"github.com/wardayadev/ub-mager-api/internal/repository"
)

var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrTaskAlreadyTaken   = errors.New("task already accepted by another helper")
	ErrTaskNotCancellable = errors.New("task cannot be cancelled in current state")
	ErrTaskSelfAccept     = errors.New("cannot accept your own task")
	ErrInvalidTaskStatus  = errors.New("invalid task status transition")
)

var validTaskTransitions = map[model.TaskStatus][]model.TaskStatus{
	model.TaskStatusAccepted:   {model.TaskStatusPickingUp},
	model.TaskStatusPickingUp:  {model.TaskStatusDelivering},
	model.TaskStatusDelivering: {model.TaskStatusCompleted},
}

type TaskService struct {
	taskRepo *repository.TaskRepository
}

func NewTaskService(taskRepo *repository.TaskRepository) *TaskService {
	return &TaskService{taskRepo: taskRepo}
}

type CreateTaskInput struct {
	Category        string  `json:"category" binding:"required,oneof=JASTIP_MAKANAN JASTIP_BARANG TITIP_PRINT ANTAR_JEMPUT OTHER"`
	Title           string  `json:"title" binding:"required,min=5,max=200"`
	Description     string  `json:"description" binding:"required,min=10,max=1000"`
	Fee             float64 `json:"fee" binding:"required,min=1000"`
	PickupLat       float64 `json:"pickup_lat" binding:"required"`
	PickupLng       float64 `json:"pickup_lng" binding:"required"`
	PickupAddress   string  `json:"pickup_address" binding:"required"`
	PickupZone      string  `json:"pickup_zone"`
	DeliveryLat     float64 `json:"delivery_lat" binding:"required"`
	DeliveryLng     float64 `json:"delivery_lng" binding:"required"`
	DeliveryAddress string  `json:"delivery_address" binding:"required"`
	DeliveryZone    string  `json:"delivery_zone"`
	Notes           string  `json:"notes" binding:"max=500"`
}

func (s *TaskService) Create(ctx context.Context, creatorID uuid.UUID, input CreateTaskInput) (*model.Task, error) {
	task := &model.Task{
		ID:              uuid.New(),
		CreatorID:       creatorID,
		Category:        model.TaskCategory(input.Category),
		Status:          model.TaskStatusOpen,
		Title:           input.Title,
		Description:     input.Description,
		Fee:             input.Fee,
		PickupLat:       input.PickupLat,
		PickupLng:       input.PickupLng,
		PickupAddress:   input.PickupAddress,
		PickupZone:      model.CampusZone(input.PickupZone),
		DeliveryLat:     input.DeliveryLat,
		DeliveryLng:     input.DeliveryLng,
		DeliveryAddress: input.DeliveryAddress,
		DeliveryZone:    model.CampusZone(input.DeliveryZone),
		Notes:           input.Notes,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskService) Accept(ctx context.Context, taskID, helperID uuid.UUID) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTaskNotFound
		}
		return err
	}

	if task.Status != model.TaskStatusOpen {
		return ErrTaskAlreadyTaken
	}

	if task.CreatorID == helperID {
		return ErrTaskSelfAccept
	}

	now := time.Now()
	return s.taskRepo.UpdateStatus(ctx, taskID, model.TaskStatusAccepted, map[string]any{
		"helper_id":   helperID,
		"accepted_at": now,
	})
}

func (s *TaskService) UpdateStatus(ctx context.Context, taskID, helperID uuid.UUID, newStatus model.TaskStatus) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTaskNotFound
		}
		return err
	}

	if task.HelperID == nil || *task.HelperID != helperID {
		return ErrUnauthorized
	}

	allowed, ok := validTaskTransitions[task.Status]
	if !ok {
		return ErrInvalidTaskStatus
	}

	valid := false
	for _, st := range allowed {
		if st == newStatus {
			valid = true
			break
		}
	}
	if !valid {
		return ErrInvalidTaskStatus
	}

	updates := map[string]any{}
	if newStatus == model.TaskStatusCompleted {
		now := time.Now()
		updates["completed_at"] = now
	}

	return s.taskRepo.UpdateStatus(ctx, taskID, newStatus, updates)
}

func (s *TaskService) Cancel(ctx context.Context, taskID, userID uuid.UUID) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrTaskNotFound
		}
		return err
	}

	if task.CreatorID != userID && (task.HelperID == nil || *task.HelperID != userID) {
		return ErrUnauthorized
	}

	switch task.Status {
	case model.TaskStatusOpen, model.TaskStatusAccepted, model.TaskStatusPickingUp:
	default:
		return ErrTaskNotCancellable
	}

	now := time.Now()
	return s.taskRepo.UpdateStatus(ctx, taskID, model.TaskStatusCancelled, map[string]any{
		"cancelled_at": now,
	})
}

func (s *TaskService) GetByID(ctx context.Context, taskID uuid.UUID) (*model.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTaskNotFound
		}
		return nil, err
	}
	return task, nil
}

func (s *TaskService) Feed(ctx context.Context, page, perPage int, category, zone string) ([]repository.TaskFeedItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}
	return s.taskRepo.Feed(ctx, page, perPage, category, zone)
}

func (s *TaskService) MyTasks(ctx context.Context, userID uuid.UUID, page, perPage int) ([]model.Task, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}
	return s.taskRepo.FindByCreator(ctx, userID, page, perPage)
}

func (s *TaskService) MyHelperTasks(ctx context.Context, userID uuid.UUID, page, perPage int) ([]model.Task, int64, error) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 50 {
		perPage = 20
	}
	return s.taskRepo.FindByHelper(ctx, userID, page, perPage)
}
