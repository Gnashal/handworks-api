package realtime

import (
	"handworks-api/utils"

	"github.com/gorilla/websocket"
)

type RealtimeHubs struct {
	EmployeeHub *EmployeeHub
	AdminHub    *AdminHub
	ChatHub     *ChatHub
}

func NewRealtimeHubs(log *utils.Logger) *RealtimeHubs {
	return &RealtimeHubs{
		EmployeeHub: NewEmployeeHub(log),
		AdminHub:    NewAdminHub(log),
		ChatHub:     NewChatHub(log),
	}
}
type AdminHub struct {
	log        *utils.Logger
	clients    map[*websocket.Conn]bool
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	broadcast  chan []byte
}

func NewAdminHub(log *utils.Logger) *AdminHub {
	return &AdminHub{
		log:        log,
		clients:    make(map[*websocket.Conn]bool),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		broadcast:  make(chan []byte),
	}
}
type EmployeeHub struct {
	log        *utils.Logger
	clients    map[string]map[*websocket.Conn]bool
	register   chan employeeConn
	unregister chan employeeConn
}

type employeeConn struct {
	employeeID string
	conn       *websocket.Conn
}

func NewEmployeeHub(log *utils.Logger) *EmployeeHub {
	return &EmployeeHub{
		log:        log,
		clients:    make(map[string]map[*websocket.Conn]bool),
		register:   make(chan employeeConn),
		unregister: make(chan employeeConn),
	}
}

type ChatHub struct {
	log        *utils.Logger
	rooms      map[string]map[*websocket.Conn]bool
	register   chan chatConn
	unregister chan chatConn
	broadcast  chan chatMessage
}

type chatConn struct {
	roomID string
	conn   *websocket.Conn
}

type chatMessage struct {
	roomID string
	data   []byte
}

func NewChatHub(log *utils.Logger) *ChatHub {
	return &ChatHub{
		log:        log,
		rooms:      make(map[string]map[*websocket.Conn]bool),
		register:   make(chan chatConn),
		unregister: make(chan chatConn),
		broadcast:  make(chan chatMessage),
	}
}

