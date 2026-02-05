package realtime

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func (h *ChatHub) Run() {
	h.log.Info("ChatHub started")
	for {
		select {
		case cc := <-h.register:
			if h.rooms[cc.roomID] == nil {
				h.rooms[cc.roomID] = make(map[*websocket.Conn]bool)
			}
			h.rooms[cc.roomID][cc.conn] = true

		case cc := <-h.unregister:
			if conns, ok := h.rooms[cc.roomID]; ok {
				delete(conns, cc.conn)
				cc.conn.Close()
				if len(conns) == 0 {
					delete(h.rooms, cc.roomID)
				}
				h.log.Info("Chat participant disconnected")
			}

		case msg := <-h.broadcast:
			for conn := range h.rooms[msg.roomID] {
				if err := conn.WriteMessage(websocket.TextMessage, msg.data); err != nil {
					delete(h.rooms[msg.roomID], conn)
					conn.Close()
				}
			}
		}
	}
}
func ChatWS(hub *ChatHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		roomID := c.Param("roomId")
		if roomID == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		cc := chatConn{roomID, conn}
		hub.register <- cc

		go func() {
			defer func() {
				hub.unregister <- cc
			}()
			for {
				_, msg, err := conn.ReadMessage()
				if err != nil {
					break
				}
				hub.broadcast <- chatMessage{
					roomID: roomID,
					data:   msg,
				}
			}
		}()
	}
}

