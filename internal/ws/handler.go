package ws

import (
	"encoding/json"
	"time"

	"github.com/rs/zerolog/log"
)

// HandleMessage routes incoming messages to the appropriate handler
func (h *Hub) HandleMessage(client *Client, msg *Message) {
	switch msg.Type {

	case MsgTypePing:
		pong := &Message{
			Type:      MsgTypePong,
			Timestamp: time.Now().UnixMilli(),
		}
		data, _ := json.Marshal(pong)
		select {
		case client.Send <- data:
		default:
		}

	case MsgTypeDriverLocation:
		if client.Role != "DRIVER" {
			h.sendError(client, "Only drivers can send location updates")
			return
		}

		var payload LocationPayload
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			h.sendError(client, "Invalid location payload")
			return
		}

		if h.messageHandler != nil {
			h.messageHandler.OnDriverLocation(client, payload)
		}

	case MsgTypeRideAccept, MsgTypeRideReject:
		if msg.CorrelationID != "" {
			h.ResolveWaiter(msg.CorrelationID, msg)
		}

	case MsgTypeSubscribeRide:
		var payload struct {
			RideID string `json:"ride_id"`
		}
		if err := json.Unmarshal(msg.Payload, &payload); err == nil {
			client.ActiveRideID = payload.RideID
			log.Info().Str("user_id", client.UserID).Str("ride_id", payload.RideID).Msg("subscribed to ride")
		}

	default:
		log.Warn().Str("user_id", client.UserID).Str("type", msg.Type).Msg("unknown ws message type")
	}
}

func (h *Hub) sendError(client *Client, message string) {
	errMsg := &Message{
		Type:      MsgTypeError,
		Timestamp: time.Now().UnixMilli(),
	}
	payload, _ := json.Marshal(map[string]string{"message": message})
	errMsg.Payload = payload

	data, _ := json.Marshal(errMsg)
	select {
	case client.Send <- data:
	default:
	}
}
