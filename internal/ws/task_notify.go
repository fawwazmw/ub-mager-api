package ws

import (
	"encoding/json"
	"time"
)

type TaskNotification struct {
	TaskID   string  `json:"task_id"`
	Title    string  `json:"title"`
	Category string  `json:"category"`
	Fee      float64 `json:"fee"`
	Zone     string  `json:"zone,omitempty"`
}

func (h *Hub) BroadcastTaskNew(notification TaskNotification) {
	payload, _ := json.Marshal(notification)
	msg := &Message{
		Type:      MsgTypeTaskNew,
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	data, _ := json.Marshal(msg)
	for _, client := range h.clients {
		select {
		case client.Send <- data:
		default:
		}
	}
}

func (h *Hub) NotifyTaskAccepted(creatorUserID string, taskID, helperName string) {
	payload, _ := json.Marshal(map[string]string{
		"task_id":     taskID,
		"helper_name": helperName,
	})
	msg := &Message{
		Type:      MsgTypeTaskAccepted,
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
	}
	h.SendToUser(creatorUserID, msg)
}

func (h *Hub) NotifyTaskCompleted(creatorUserID string, taskID string) {
	payload, _ := json.Marshal(map[string]string{
		"task_id": taskID,
	})
	msg := &Message{
		Type:      MsgTypeTaskCompleted,
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
	}
	h.SendToUser(creatorUserID, msg)
}
