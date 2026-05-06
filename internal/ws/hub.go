package ws

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type MessageHandler interface {
	OnDriverLocation(client *Client, payload LocationPayload)
}

type Hub struct {
	clients        map[string]*Client
	register       chan *Client
	unregister     chan *Client
	broadcast      chan *Message
	waiters        map[string]chan *Message
	messageHandler MessageHandler
	mu             sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *Message, 256),
		waiters:    make(map[string]chan *Message),
	}
}

func (h *Hub) SetMessageHandler(handler MessageHandler) {
	h.messageHandler = handler
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if old, exists := h.clients[client.UserID]; exists {
				close(old.Send)
				old.Conn.Close()
				log.Info().Str("user_id", client.UserID).Msg("evicted stale ws connection")
			}
			h.clients[client.UserID] = client
			h.mu.Unlock()

			log.Info().
				Str("user_id", client.UserID).
				Str("role", client.Role).
				Int("total", len(h.clients)).
				Msg("ws client registered")

		case client := <-h.unregister:
			h.mu.Lock()
			if current, exists := h.clients[client.UserID]; exists {
				if current == client {
					delete(h.clients, client.UserID)
					close(client.Send)
				}
			}
			h.mu.Unlock()

			log.Info().
				Str("user_id", client.UserID).
				Str("role", client.Role).
				Int("total", len(h.clients)).
				Msg("ws client unregistered")

		case message := <-h.broadcast:
			h.mu.RLock()
			data, _ := json.Marshal(message)
			for userID, client := range h.clients {
				select {
				case client.Send <- data:
				default:
					log.Warn().Str("user_id", userID).Msg("ws buffer full, dropping broadcast")
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) SendToUser(userID string, msg *Message) bool {
	h.mu.RLock()
	client, exists := h.clients[userID]
	h.mu.RUnlock()

	if !exists {
		return false
	}

	data, err := json.Marshal(msg)
	if err != nil {
		log.Error().Err(err).Msg("ws marshal error")
		return false
	}

	select {
	case client.Send <- data:
		return true
	default:
		log.Warn().Str("user_id", userID).Msg("ws send buffer full")
		return false
	}
}

func (h *Hub) WaitForResponse(userID string, msg *Message, timeout time.Duration) (*Message, error) {
	correlationID := msg.CorrelationID
	ch := make(chan *Message, 1)

	h.mu.Lock()
	h.waiters[correlationID] = ch
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.waiters, correlationID)
		h.mu.Unlock()
	}()

	if !h.SendToUser(userID, msg) {
		return nil, ErrUserNotConnected
	}

	select {
	case resp := <-ch:
		return resp, nil
	case <-time.After(timeout):
		return nil, ErrTimeout
	}
}

func (h *Hub) ResolveWaiter(correlationID string, msg *Message) bool {
	h.mu.RLock()
	ch, exists := h.waiters[correlationID]
	h.mu.RUnlock()

	if !exists {
		return false
	}

	select {
	case ch <- msg:
		return true
	default:
		return false
	}
}

func (h *Hub) IsConnected(userID string) bool {
	h.mu.RLock()
	_, exists := h.clients[userID]
	h.mu.RUnlock()
	return exists
}

func (h *Hub) GetOnlineCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

func (h *Hub) GetClient(userID string) *Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.clients[userID]
}

func (h *Hub) BroadcastToRide(rideID string, msg *Message, excludeUserID string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return
	}

	for userID, client := range h.clients {
		if client.ActiveRideID == rideID && userID != excludeUserID {
			select {
			case client.Send <- data:
			default:
			}
		}
	}
}

var (
	ErrUserNotConnected = &WsError{Code: "USER_NOT_CONNECTED", Message: "User is not connected"}
	ErrTimeout          = &WsError{Code: "TIMEOUT", Message: "Request timed out"}
)

type WsError struct {
	Code    string
	Message string
}

func (e *WsError) Error() string {
	return e.Message
}
