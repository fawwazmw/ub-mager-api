package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/repository"
)

type UserAdminHandler struct {
	userRepo *repository.UserRepository
}

func NewUserAdminHandler(userRepo *repository.UserRepository) *UserAdminHandler {
	return &UserAdminHandler{userRepo: userRepo}
}

func (h *UserAdminHandler) ListUsers(c *gin.Context) {
	page, perPage := ParsePagination(c)
	role := c.DefaultQuery("role", "")
	search := c.DefaultQuery("search", "")

	users, total, err := h.userRepo.List(c.Request.Context(), page, perPage, role, search)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to list users")
		return
	}

	PaginatedSuccess(c, users, page, perPage, total)
}

func (h *UserAdminHandler) SuspendUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	adminID, _ := GetUserID(c)
	if userID == adminID {
		Error(c, http.StatusBadRequest, "SELF_SUSPEND", "Cannot suspend yourself")
		return
	}

	if err := h.userRepo.SetActive(c.Request.Context(), userID, false); err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to suspend user")
		return
	}

	SuccessMessage(c, "User suspended")
}

func (h *UserAdminHandler) UnsuspendUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid user ID")
		return
	}

	if err := h.userRepo.SetActive(c.Request.Context(), userID, true); err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to unsuspend user")
		return
	}

	SuccessMessage(c, "User reactivated")
}
