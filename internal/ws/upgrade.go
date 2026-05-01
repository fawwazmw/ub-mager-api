package ws

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	jwtpkg "github.com/wardayadev/ub-mager-api/internal/pkg/jwt"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// HandleWebSocket upgrades HTTP to WebSocket and registers the client
func HandleWebSocket(hub *Hub, jwtService *jwtpkg.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authenticate via query parameter token
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		// Upgrade to WebSocket
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			log.Error().Err(err).Msg("ws upgrade failed")
			return
		}

		client := &Client{
			UserID: claims.UserID.String(),
			Role:   claims.Role,
			Conn:   conn,
			Send:   make(chan []byte, sendBufferSize),
			Hub:    hub,
		}

		hub.register <- client

		// Send CONNECTED message
		hub.SendToUser(client.UserID, &Message{
			Type:      MsgTypeConnected,
			Timestamp: 0,
		})

		// Start read/write pumps
		go client.WritePump()
		go client.ReadPump()
	}
}
