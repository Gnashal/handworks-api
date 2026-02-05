package realtime

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func (h *EmployeeHub) Run() {
	h.log.Info("EmployeeHub started")
	for {
		select {
		case ec := <-h.register:
			if h.clients[ec.employeeID] == nil {
				h.clients[ec.employeeID] = make(map[*websocket.Conn]bool)
			}
			h.clients[ec.employeeID][ec.conn] = true

		case ec := <-h.unregister:
			if conns, ok := h.clients[ec.employeeID]; ok {
				delete(conns, ec.conn)
				ec.conn.Close()
				if len(conns) == 0 {
					delete(h.clients, ec.employeeID)
				}
				h.log.Info("Employee disconnected")
			}
		}
	}
}

func (h *EmployeeHub) SendToEmployee(employeeID string, event string, payload any) {
	msg, _ := json.Marshal(map[string]any{
		"event": event,
		"data":  payload,
	})

	for conn := range h.clients[employeeID] {
		if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			delete(h.clients[employeeID], conn)
			conn.Close()
		}
	}
}
func EmployeeWS(hub *EmployeeHub) gin.HandlerFunc {
	return func(c *gin.Context) {
		employeeID := c.Param("employeeID")
		if employeeID == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			return
		}

		ec := employeeConn{employeeID, conn}
		hub.register <- ec

		go func() {
			defer func() {
				hub.unregister <- ec
			}()
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					break
				}
			}
		}()
	}
}

