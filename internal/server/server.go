package server

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net/http"

	"soundwires/internal/pipewire"
)

type Config struct {
	PwCatBin string
}

// New builds the HTTP mux, wires up WebSocket hub and event broadcast loop,
// and returns a ready-to-listen http.Handler.
func New(staticFiles embed.FS, watcher *pipewire.Watcher, cfg Config) http.Handler {
	h := newHub()
	mux := http.NewServeMux()

	// Serve embedded frontend
	distFS, err := fs.Sub(staticFiles, "web/dist")
	if err != nil {
		log.Fatalf("embed sub: %v", err)
	}
	mux.Handle("/", http.FileServer(http.FS(distFS)))

	// WebSocket — new connections get the current init snapshot on connect
	var lastInit []pipewire.PWObject
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("ws upgrade: %v", err)
			return
		}
		// Send current state immediately on connect
		if lastInit != nil {
			if err := h.sendInit(conn, lastInit); err != nil {
				conn.Close()
				return
			}
		}
		h.add(conn)
		defer h.remove(conn)
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	})

	// Audio endpoints
	mux.HandleFunc("/api/record", recordHandler(cfg.PwCatBin))
	mux.HandleFunc("/api/play", playHandler(cfg.PwCatBin))
	mux.HandleFunc("/ws/monitor", monitorWsHandler(cfg.PwCatBin))

	// Broadcast watcher events to all WebSocket clients
	go func() {
		for event := range watcher.Events {
			if event.Type == pipewire.EventInit {
				lastInit = event.Objects
			}
			msg, err := json.Marshal(event)
			if err != nil {
				continue
			}
			h.broadcast(msg)
		}
	}()

	return mux
}
