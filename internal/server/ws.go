package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"soundwires/internal/pipewire"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type hub struct {
	mu      sync.Mutex
	clients map[*websocket.Conn]bool
}

func newHub() *hub {
	return &hub{clients: make(map[*websocket.Conn]bool)}
}

func (h *hub) add(c *websocket.Conn) {
	h.mu.Lock()
	h.clients[c] = true
	h.mu.Unlock()
}

func (h *hub) remove(c *websocket.Conn) {
	h.mu.Lock()
	delete(h.clients, c)
	h.mu.Unlock()
	c.Close()
}

func (h *hub) broadcast(msg []byte) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		if err := c.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Printf("ws write error: %v", err)
			delete(h.clients, c)
			c.Close()
		}
	}
}

func (h *hub) sendInit(c *websocket.Conn, objects []pipewire.PWObject) error {
	msg, err := json.Marshal(pipewire.Event{
		Type:    pipewire.EventInit,
		Objects: objects,
	})
	if err != nil {
		return err
	}
	return c.WriteMessage(websocket.TextMessage, msg)
}

// wsHandler upgrades the connection, registers with the hub, and keeps it alive.
// The hub's broadcast loop does the actual writing; this goroutine just reads
// (to handle pings and detect disconnects).
func wsHandler(h *hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("ws upgrade: %v", err)
		return
	}
	h.add(conn)
	defer h.remove(conn)

	// drain incoming messages (browser may send pings or nothing)
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
