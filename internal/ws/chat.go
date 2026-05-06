package ws

import (
	"encoding/json"
	"time"
)

func (h *Hub) BroadcastChat(rideID, senderID, messageID, content string, createdAt time.Time) {
	payload, _ := json.Marshal(map[string]string{
		"message_id": messageID,
		"ride_id":    rideID,
		"sender_id":  senderID,
		"content":    content,
		"created_at": createdAt.Format(time.RFC3339),
	})

	msg := &Message{
		Type:      MsgTypeChatMessage,
		Payload:   payload,
		Timestamp: time.Now().UnixMilli(),
	}

	h.BroadcastToRide(rideID, msg, senderID)
}
