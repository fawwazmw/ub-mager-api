package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SearchHandler struct {
	db *gorm.DB
}

func NewSearchHandler(db *gorm.DB) *SearchHandler {
	return &SearchHandler{db: db}
}

type SearchResult struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Subtitle  string `json:"subtitle"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

func (h *SearchHandler) GlobalSearch(c *gin.Context) {
	query := c.Query("q")
	if len(query) < 2 {
		Error(c, http.StatusBadRequest, "QUERY_TOO_SHORT", "Search query must be at least 2 characters")
		return
	}

	limit := ParseIntQuery(c, "limit", 10, 1, 50)
	pattern := "%" + query + "%"

	var results []SearchResult

	var rides []struct {
		ID             string
		PickupAddress  string
		DropoffAddress string
		Status         string
		CreatedAt      time.Time
	}
	h.db.WithContext(c.Request.Context()).
		Table("rides").
		Select("id, pickup_address, dropoff_address, status, created_at").
		Where("pickup_address ILIKE ? OR dropoff_address ILIKE ?", pattern, pattern).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Scan(&rides)

	for _, r := range rides {
		results = append(results, SearchResult{
			ID:        r.ID,
			Type:      "ride",
			Title:     r.PickupAddress + " → " + r.DropoffAddress,
			Subtitle:  r.Status,
			Status:    r.Status,
			CreatedAt: r.CreatedAt.Format(time.RFC3339),
		})
	}

	var tasks []struct {
		ID        string
		Title     string
		Category  string
		Status    string
		CreatedAt time.Time
	}
	h.db.WithContext(c.Request.Context()).
		Table("tasks").
		Select("id, title, category, status, created_at").
		Where("title ILIKE ? OR description ILIKE ?", pattern, pattern).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Limit(limit).
		Scan(&tasks)

	for _, t := range tasks {
		results = append(results, SearchResult{
			ID:        t.ID,
			Type:      "task",
			Title:     t.Title,
			Subtitle:  t.Category,
			Status:    t.Status,
			CreatedAt: t.CreatedAt.Format(time.RFC3339),
		})
	}

	var users []struct {
		ID       string
		FullName string
		Phone    string
		Role     string
	}
	h.db.WithContext(c.Request.Context()).
		Table("users").
		Select("id, full_name, phone, role").
		Where("full_name ILIKE ? OR phone ILIKE ? OR email ILIKE ?", pattern, pattern, pattern).
		Where("deleted_at IS NULL").
		Limit(limit).
		Scan(&users)

	for _, u := range users {
		results = append(results, SearchResult{
			ID:       u.ID,
			Type:     "user",
			Title:    u.FullName,
			Subtitle: u.Phone,
			Status:   u.Role,
		})
	}

	Success(c, http.StatusOK, results)
}
