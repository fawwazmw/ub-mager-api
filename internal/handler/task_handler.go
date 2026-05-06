package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/model"
	"github.com/wardayadev/ub-mager-api/internal/repository"
	"github.com/wardayadev/ub-mager-api/internal/service"
)

type TaskHandler struct {
	taskService *service.TaskService
	taskRepo    *repository.TaskRepository
}

func NewTaskHandler(taskService *service.TaskService, taskRepo *repository.TaskRepository) *TaskHandler {
	return &TaskHandler{taskService: taskService, taskRepo: taskRepo}
}

func (h *TaskHandler) Create(c *gin.Context) {
	userID, _ := GetUserID(c)

	var input service.CreateTaskInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	if !ValidCoordinate(input.PickupLat, input.PickupLng) || !ValidCoordinate(input.DeliveryLat, input.DeliveryLng) {
		Error(c, http.StatusBadRequest, "INVALID_COORDINATES", "Coordinates must be valid lat/lng values")
		return
	}

	task, err := h.taskService.Create(c.Request.Context(), userID, input)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to create task")
		return
	}

	Success(c, http.StatusCreated, task)
}

func (h *TaskHandler) Feed(c *gin.Context) {
	page, perPage := ParsePagination(c)
	category := c.DefaultQuery("category", "")
	zone := c.DefaultQuery("zone", "")

	tasks, total, err := h.taskService.Feed(c.Request.Context(), page, perPage, category, zone)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get task feed")
		return
	}

	PaginatedSuccess(c, tasks, page, perPage, total)
}

func (h *TaskHandler) GetByID(c *gin.Context) {
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid task ID")
		return
	}

	task, err := h.taskService.GetByID(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			Error(c, http.StatusNotFound, "TASK_NOT_FOUND", "Task not found")
		} else {
			Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get task")
		}
		return
	}

	Success(c, http.StatusOK, task)
}

func (h *TaskHandler) Accept(c *gin.Context) {
	userID, _ := GetUserID(c)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid task ID")
		return
	}

	err = h.taskService.Accept(c.Request.Context(), taskID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFound):
			Error(c, http.StatusNotFound, "TASK_NOT_FOUND", "Task not found")
		case errors.Is(err, service.ErrTaskAlreadyTaken):
			Error(c, http.StatusConflict, "TASK_TAKEN", "Task already accepted by another helper")
		case errors.Is(err, service.ErrTaskSelfAccept):
			Error(c, http.StatusBadRequest, "SELF_ACCEPT", "Cannot accept your own task")
		default:
			Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to accept task")
		}
		return
	}

	SuccessMessage(c, "Task accepted")
}

func (h *TaskHandler) UpdateStatus(c *gin.Context) {
	userID, _ := GetUserID(c)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid task ID")
		return
	}

	var input struct {
		Status string `json:"status" binding:"required,oneof=PICKING_UP DELIVERING COMPLETED"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	err = h.taskService.UpdateStatus(c.Request.Context(), taskID, userID, model.TaskStatus(input.Status))
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFound):
			Error(c, http.StatusNotFound, "TASK_NOT_FOUND", "Task not found")
		case errors.Is(err, service.ErrUnauthorized):
			Error(c, http.StatusForbidden, "FORBIDDEN", "Not authorized")
		case errors.Is(err, service.ErrInvalidTaskStatus):
			Error(c, http.StatusBadRequest, "INVALID_TRANSITION", "Invalid status transition")
		default:
			Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to update task")
		}
		return
	}

	SuccessMessage(c, "Task status updated")
}

func (h *TaskHandler) Cancel(c *gin.Context) {
	userID, _ := GetUserID(c)
	taskID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid task ID")
		return
	}

	err = h.taskService.Cancel(c.Request.Context(), taskID, userID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTaskNotFound):
			Error(c, http.StatusNotFound, "TASK_NOT_FOUND", "Task not found")
		case errors.Is(err, service.ErrUnauthorized):
			Error(c, http.StatusForbidden, "FORBIDDEN", "Not authorized to cancel this task")
		case errors.Is(err, service.ErrTaskNotCancellable):
			Error(c, http.StatusBadRequest, "NOT_CANCELLABLE", "Task cannot be cancelled in current state")
		default:
			Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to cancel task")
		}
		return
	}

	SuccessMessage(c, "Task cancelled")
}

func (h *TaskHandler) MyTasks(c *gin.Context) {
	userID, _ := GetUserID(c)
	page, perPage := ParsePagination(c)

	tasks, total, err := h.taskService.MyTasks(c.Request.Context(), userID, page, perPage)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get tasks")
		return
	}

	PaginatedSuccess(c, tasks, page, perPage, total)
}

func (h *TaskHandler) MyHelperTasks(c *gin.Context) {
	userID, _ := GetUserID(c)
	page, perPage := ParsePagination(c)

	tasks, total, err := h.taskService.MyHelperTasks(c.Request.Context(), userID, page, perPage)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get tasks")
		return
	}

	PaginatedSuccess(c, tasks, page, perPage, total)
}

func (h *TaskHandler) AdminList(c *gin.Context) {
	page, perPage := ParsePagination(c)
	category := c.DefaultQuery("category", "")
	status := c.DefaultQuery("status", "")
	search := c.DefaultQuery("search", "")

	tasks, total, err := h.taskRepo.AdminList(c.Request.Context(), page, perPage, category, status, search)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list tasks")
		return
	}

	PaginatedSuccess(c, tasks, page, perPage, total)
}
