package server

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os/exec"
	"strconv"

	"github.com/gorilla/websocket"
)

// monitorWsHandler streams raw s16le PCM from pw-cat --record as binary WebSocket frames.
// The first frame is always a JSON text message describing the format or carrying an error.
// All subsequent frames are raw s16le binary PCM (interleaved channels).
// The stream stops when the client closes the WebSocket or pw-cat exits.
func monitorWsHandler(pwCatBin string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		target := q.Get("target")
		if target == "" {
			http.Error(w, "target required", http.StatusBadRequest)
			return
		}

		channels, _ := strconv.Atoi(q.Get("channels"))
		if channels <= 0 {
			channels = 2
		}
		rate, _ := strconv.Atoi(q.Get("rate"))
		if rate <= 0 {
			rate = 48000
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("monitor ws upgrade: %v", err)
			return
		}
		defer conn.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Detect WebSocket close in background; cancel context on any read error.
		go func() {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					cancel()
					return
				}
			}
		}()

		// Announce format to the client so it can set up AudioContext correctly.
		meta, _ := json.Marshal(map[string]any{"channels": channels, "rate": rate})
		if err := conn.WriteMessage(websocket.TextMessage, meta); err != nil {
			return
		}

		var stderrBuf bytes.Buffer
		cmd := exec.Command(pwCatBin,
			"--record",
			"--target", target,
			"--format", "s16",
			"--channels", strconv.Itoa(channels),
			"--rate", strconv.Itoa(rate),
			"--properties", "{\"node.name\":\"Soundwires Monitor\"}",
			"-",
		)
		cmd.Stderr = &stderrBuf

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			sendError(conn, err.Error())
			return
		}
		if err := cmd.Start(); err != nil {
			sendError(conn, err.Error())
			return
		}

		go func() {
			<-ctx.Done()
			cmd.Process.Kill()
		}()

		buf := make([]byte, 4096)
		for {
			select {
			case <-ctx.Done():
				cmd.Wait()
				return
			default:
			}

			n, err := stdout.Read(buf)
			if n > 0 {
				if werr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); werr != nil {
					cancel()
					cmd.Wait()
					return
				}
			}
			if err != nil {
				cmd.Wait()
				if ctx.Err() == nil {
					msg := stderrBuf.String()
					if msg == "" {
						msg = "pw-cat exited unexpectedly"
					}
					sendError(conn, msg)
				}
				return
			}
		}
	}
}

func sendError(conn *websocket.Conn, msg string) {
	b, _ := json.Marshal(map[string]string{"error": msg})
	conn.WriteMessage(websocket.TextMessage, b) //nolint:errcheck
}
