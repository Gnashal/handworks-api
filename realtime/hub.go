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

type AdminHub struct {
	clients    map[*websocket.Conn]bool
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	broadcast  chan []byte
}


func NewAdminHub() *AdminHub {
	return &AdminHub{
		clients:    make(map[*websocket.Conn]bool),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		broadcast:  make(chan []byte),
	}
}
func (h *AdminHub) Run() {
	for {
		select {
		case conn := <-h.register:
			h.clients[conn] = true
		case conn := <-h.unregister:
			delete(h.clients, conn)
			conn.Close()

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
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		hub.register <- conn
	}
}

func (h *AdminHub) Broadcast(event string, payload any) {
	msg, err := json.Marshal(map[string]any{
		"event": event,
		"data":  payload,
	})
	if err != nil {
		return
	}
	h.broadcast <- msg
}




