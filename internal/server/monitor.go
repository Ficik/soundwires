package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"time"

	"soundwires/internal/pipewire"

	"github.com/gorilla/websocket"
)

// monitorWsHandler taps any PipeWire node's output ports by:
//  1. Starting pw-cat with node.autoconnect=false and a unique node.name
//  2. Polling pw-dump until both the target and our node appear
//  3. Using pw-link to manually connect target's output ports → our input ports
//  4. Streaming s16le PCM as binary WebSocket frames
//
// First frame is always a JSON text message: {"channels":N,"rate":N} or {"error":"..."}.
func monitorWsHandler(pwCatBin, pwLinkBin, pwDumpBin string) http.HandlerFunc {
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

		go func() {
			for {
				if _, _, err := conn.ReadMessage(); err != nil {
					cancel()
					return
				}
			}
		}()

		meta, _ := json.Marshal(map[string]any{"channels": channels, "rate": rate})
		if err := conn.WriteMessage(websocket.TextMessage, meta); err != nil {
			return
		}

		// Unique node name so we can find our pw-cat in pw-dump.
		uid := fmt.Sprintf("sw-mon-%d", time.Now().UnixNano())
		uidJSON, _ := json.Marshal(uid)

		var stderrBuf bytes.Buffer
		cmd := exec.Command(pwCatBin,
			"--record",
			"--target", target,
			"--properties", fmt.Sprintf(`{"node.name":%s, "node.description": "Soundwires monitor", "node.autoconnect":false}`, uidJSON),
			"--format", "s16",
			"--channels", strconv.Itoa(channels),
			"--rate", strconv.Itoa(rate),
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

		// Wait for both nodes to appear in the graph, then link their ports.
		outIDs, inIDs, linkErr := waitAndMatchPorts(ctx, pwDumpBin, target, uid)
		if linkErr != nil {
			sendError(conn, "port linking failed: "+linkErr.Error())
			cancel()
			cmd.Wait()
			return
		}

		for i := range outIDs {
			lc := exec.Command(pwLinkBin, strconv.Itoa(outIDs[i]), strconv.Itoa(inIDs[i]))
			if out, lerr := lc.CombinedOutput(); lerr != nil {
				log.Printf("pw-link %d→%d: %v: %s", outIDs[i], inIDs[i], lerr, out)
			}
		}

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

type portEntry struct {
	id      int
	channel string
}

// waitAndMatchPorts polls pw-dump until the target node's output ports and our
// pw-cat node's input ports are both visible, then pairs them by audio.channel.
func waitAndMatchPorts(ctx context.Context, pwDumpBin, targetName, ourName string) (outIDs, inIDs []int, err error) {
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return nil, nil, ctx.Err()
		default:
		}

		objects, dumpErr := pipewire.Dump(pwDumpBin)
		if dumpErr != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		nodeIDs := make(map[string]int)
		for _, obj := range objects {
			if obj.Type != pipewire.TypeNode {
				continue
			}
			var info pipewire.InfoNode
			if json.Unmarshal(obj.Info, &info) != nil {
				continue
			}
			var props map[string]any
			if json.Unmarshal(info.Props, &props) == nil {
				if name, ok := props["node.name"].(string); ok {
					nodeIDs[name] = obj.ID
				}
			}
		}

		targetID, targetOK := nodeIDs[targetName]
		ourID, ourOK := nodeIDs[ourName]
		if !targetOK || !ourOK {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		var targetOuts, ourIns []portEntry
		for _, obj := range objects {
			if obj.Type != pipewire.TypePort {
				continue
			}
			var info pipewire.InfoPort
			if json.Unmarshal(obj.Info, &info) != nil {
				continue
			}
			var props map[string]any
			if json.Unmarshal(info.Props, &props) != nil {
				continue
			}
			nodeIDf, ok := props["node.id"].(float64)
			if !ok {
				continue
			}
			nodeID := int(nodeIDf)
			ch, _ := props["audio.channel"].(string)

			switch {
			case nodeID == targetID && info.Direction == "output":
				targetOuts = append(targetOuts, portEntry{obj.ID, ch})
			case nodeID == ourID && info.Direction == "input":
				ourIns = append(ourIns, portEntry{obj.ID, ch})
			}
		}

		if len(targetOuts) == 0 || len(ourIns) == 0 {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		pairs := matchPorts(targetOuts, ourIns)
		if len(pairs) == 0 {
			return nil, nil, fmt.Errorf("no matching ports between %q and %q", targetName, ourName)
		}
		for _, p := range pairs {
			outIDs = append(outIDs, p[0])
			inIDs = append(inIDs, p[1])
		}
		return outIDs, inIDs, nil
	}
	return nil, nil, fmt.Errorf("timeout: target %q not found or has no output ports", targetName)
}

// matchPorts pairs output ports to input ports by audio.channel name,
// falling back to positional order when channel names are missing or unmatched.
func matchPorts(outs, ins []portEntry) [][2]int {
	inByChannel := make(map[string]int, len(ins))
	for _, p := range ins {
		if p.channel != "" {
			inByChannel[p.channel] = p.id
		}
	}

	used := make(map[int]bool)
	var result [][2]int
	for _, o := range outs {
		if inID, ok := inByChannel[o.channel]; ok && !used[inID] {
			result = append(result, [2]int{o.id, inID})
			used[inID] = true
		}
	}
	if len(result) > 0 {
		return result
	}

	n := len(outs)
	if len(ins) < n {
		n = len(ins)
	}
	for i := range n {
		result = append(result, [2]int{outs[i].id, ins[i].id})
	}
	return result
}

func sendError(conn *websocket.Conn, msg string) {
	b, _ := json.Marshal(map[string]string{"error": msg})
	conn.WriteMessage(websocket.TextMessage, b) //nolint:errcheck
}
