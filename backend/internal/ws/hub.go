package ws

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

// Client one ws connection
type Client struct {
	Conn     *websocket.Conn
	UserID   uint
	Username string
	RoomID   string
	Send     chan []byte
	LastSeen int64 // last received message unix nano; atomic r/w
}

// Hub room manager; roomID = doc-{id}
type Hub struct {
	mu         sync.Mutex
	rooms      map[string]map[*Client]bool
	Register   chan *Client
	Unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case cli := <-h.Register:
			h.mu.Lock()
			if h.rooms[cli.RoomID] == nil {
				h.rooms[cli.RoomID] = make(map[*Client]bool)
			}
			h.rooms[cli.RoomID][cli] = true
			users := h.userList(cli.RoomID)
			h.mu.Unlock()
			// on join, broadcast join + online list
			msg, _ := json.Marshal(map[string]interface{}{
				"type":  "join",
				"user":  cli.Username,
				"users": users,
			})
			h.Broadcast(cli.RoomID, msg, nil)

		case cli := <-h.Unregister:
			h.mu.Lock()
			if set, ok := h.rooms[cli.RoomID]; ok {
				if _, ok := set[cli]; ok {
					delete(set, cli)
					close(cli.Send)
				}
				if len(set) == 0 {
					delete(h.rooms, cli.RoomID)
				}
			}
			users := h.userList(cli.RoomID)
			h.mu.Unlock()
			msg, _ := json.Marshal(map[string]interface{}{
				"type":  "leave",
				"user":  cli.Username,
				"users": users,
			})
			h.Broadcast(cli.RoomID, msg, nil)
		}
	}
}

// userList deduplicated usernames in room
func (h *Hub) userList(roomID string) []string {
	seen := map[uint]string{}
	for cli := range h.rooms[roomID] {
		seen[cli.UserID] = cli.Username
	}
	list := make([]string, 0, len(seen))
	for _, name := range seen {
		list = append(list, name)
	}
	return list
}

// Broadcast to room; except=nil means send to all
func (h *Hub) Broadcast(roomID string, msg []byte, except *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for cli := range h.rooms[roomID] {
		if cli == except {
			continue
		}
		select {
		case cli.Send <- msg:
		default:
		}
	}
}
