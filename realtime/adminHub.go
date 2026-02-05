package realtime

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}


func (h *AdminHub) Run() {
	h.log.Info("AdminHub started")
	for {
		select {
		case conn := <-h.register:
			h.clients[conn] = true
		case conn := <-h.unregister:
			delete(h.clients, conn)
			conn.Close()
			h.log.Info("Admin disconnected")
		case msg := <-h.broadcast:
			for conn := range h.clients {
				if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
					delete(h.clients, conn)
					conn.Close()
				}
			}
		}
	}
}

func AdminWS(hub *AdminHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminId := c.Param("adminId")
		if adminId == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		hub.register <- conn
		go func() {
			defer func() {
				hub.unregister <- conn
			}()
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					break
				}
			}
		}()
	}
}

func (h *AdminHub) SendToAdmin(event string, payload any) {
	msg, err := json.Marshal(map[string]any{
		"event": event,
		"data":  payload,
	})
	if err != nil {
		return
	}
	h.broadcast <- msg
}




