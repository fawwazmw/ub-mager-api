package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wardayadev/ub-mager-api/internal/model"
	"github.com/wardayadev/ub-mager-api/internal/repository"
)

type ChatBroadcaster interface {
	BroadcastChat(rideID, senderID, messageID, content string, createdAt time.Time)
}

type ChatHandler struct {
	chatRepo    *repository.ChatRepository
	broadcaster ChatBroadcaster
}

func NewChatHandler(chatRepo *repository.ChatRepository, broadcaster ChatBroadcaster) *ChatHandler {
	return &ChatHandler{chatRepo: chatRepo, broadcaster: broadcaster}
}

type SendMessageInput struct {
	Content string `json:"content" binding:"required,min=1,max=500"`
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	userID, _ := GetUserID(c)
	rideID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid ride ID")
		return
	}

	var input SendMessageInput
	if err := c.ShouldBindJSON(&input); err != nil {
		Error(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	msg := &model.ChatMessage{
		ID:        uuid.New(),
		RideID:    rideID,
		SenderID:  userID,
		Content:   input.Content,
		CreatedAt: time.Now(),
	}

	if err := h.chatRepo.Create(c.Request.Context(), msg); err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to send message")
		return
	}

	h.broadcastChatMessage(userID.String(), rideID.String(), msg)

	Success(c, http.StatusCreated, msg)
}

func (h *ChatHandler) GetMessages(c *gin.Context) {
	rideID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid ride ID")
		return
	}

	limit := ParseIntQuery(c, "limit", 50, 1, 100)
	messages, err := h.chatRepo.FindByRideID(c.Request.Context(), rideID, limit)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get messages")
		return
	}

	Success(c, http.StatusOK, messages)
}

func (h *ChatHandler) AdminGetMessages(c *gin.Context) {
	rideID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		Error(c, http.StatusBadRequest, "INVALID_ID", "Invalid ride ID")
		return
	}

	messages, err := h.chatRepo.FindByRideID(c.Request.Context(), rideID, 100)
	if err != nil {
		Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get messages")
		return
	}

	Success(c, http.StatusOK, messages)
}

func (h *ChatHandler) broadcastChatMessage(senderID, rideID string, msg *model.ChatMessage) {
	if h.broadcaster != nil {
		h.broadcaster.BroadcastChat(rideID, senderID, msg.ID.String(), msg.Content, msg.CreatedAt)
	}
}
