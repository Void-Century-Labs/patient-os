package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/void-century-labs/patient-os/backend/internal/ws"
)

type WebSocketHandler struct {
	Hub *ws.Hub
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// SubscribeQueue upgrades the request to a WebSocket and streams queue
// snapshots (waiting entries with recalculated positions) to the client
// whenever the queue changes.
func (h *WebSocketHandler) SubscribeQueue(c *gin.Context) {
	queueID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid queue id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	client := &ws.Client{Send: make(chan []byte, 8)}
	h.Hub.Subscribe(uint(queueID), client)
	defer h.Hub.Unsubscribe(uint(queueID), client)

	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				close(client.Send)
				return
			}
		}
	}()

	for msg := range client.Send {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			return
		}
	}
}
