package ws

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"

	"github.com/wardayadev/ub-mager-api/internal/handler"
	jwtpkg "github.com/wardayadev/ub-mager-api/internal/pkg/jwt"
)

func newUpgrader(corsOrigins string) websocket.Upgrader {
	return websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			if corsOrigins == "" || corsOrigins == "*" {
				return true
			}
			origin := r.Header.Get("Origin")
			for _, o := range strings.Split(corsOrigins, ",") {
				if strings.TrimSpace(o) == origin {
					return true
				}
			}
			return false
		},
	}
}

func HandleWebSocket(hub *Hub, jwtService *jwtpkg.JWTService, corsOrigins string) gin.HandlerFunc {
	upgrader := newUpgrader(corsOrigins)

	return func(c *gin.Context) {
		token := c.Query("token")
		if token == "" {
			handler.Error(c, http.StatusUnauthorized, "UNAUTHORIZED", "Missing token")
			return
		}

		claims, err := jwtService.ValidateToken(token)
		if err != nil {
			handler.Error(c, http.StatusUnauthorized, "TOKEN_EXPIRED", "Token is invalid or expired")
			return
		}

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

		hub.SendToUser(client.UserID, &Message{
			Type:      MsgTypeConnected,
			Timestamp: 0,
		})

		go client.WritePump()
		go client.ReadPump()
	}
}
