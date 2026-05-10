package audio

import (
	"bytes"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

// PlayerHandler streams a continuous sine wave to a PipeWire target node.
// The stream runs until the client closes the connection (AbortController.abort()).
// If pw-play exits on its own (bad target / unsupported format), the response
// carries the stderr text as a 500 error so the browser UI can display it.
func PlayerHandler(pwPlayBin string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		targetID := q.Get("targetId")
		if targetID == "" {
			http.Error(w, "targetId required", http.StatusBadRequest)
			return
		}

		freq, _ := strconv.Atoi(q.Get("freq"))
		if freq <= 0 {
			freq = 440
		}
		channels, _ := strconv.Atoi(q.Get("channels"))
		if channels <= 0 {
			channels = 2
		}
		rate, _ := strconv.Atoi(q.Get("rate"))
		if rate <= 0 {
			rate = 48000
		}
		amp, _ := strconv.ParseFloat(q.Get("amp"), 64)
		if amp <= 0 || amp > 1 {
			amp = 0.5
		}

		var stderrBuf bytes.Buffer
		// pw-cat (and pw-play) use libsndfile, which requires a container format.
		// We wrap our raw PCM in a streaming WAV header — the WAV header carries
		// all format metadata, so no --format/--channels/--rate flags are needed.
		cmd := exec.Command(pwPlayBin,
			"--playback",
			"--target", targetID,
			"--properties", "{\"node.name\":\"Soundwires test sound\"}",
			"-",
		)
		cmd.Stderr = &stderrBuf

		stdin, err := cmd.StdinPipe()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := cmd.Start(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// WAV header must arrive before any PCM data.
		writeStreamingWAVHeader(stdin, channels, rate)

		// Kill pw-cat exactly once — either on client disconnect or after sine stops.
		var killOnce sync.Once
		kill := func() { killOnce.Do(func() { cmd.Process.Kill() }) }

		ctx := r.Context()
		go func() {
			<-ctx.Done()
			kill()
		}()

		// Generate sine until context is cancelled or stdin breaks (pw-cat died).
		generateSine(ctx, stdin, freq, rate, channels, amp)
		stdin.Close()
		kill() // ensure pw-play is stopped even if context is not yet cancelled

		waitErr := cmd.Wait()

		if ctx.Err() != nil {
			// Normal: client disconnected / AbortController.abort() — not an error.
			return
		}

		// pw-play exited before the client stopped us — surface the failure.
		if waitErr != nil {
			msg := strings.TrimSpace(stderrBuf.String())
			if msg == "" {
				msg = waitErr.Error()
			}
			http.Error(w, msg, http.StatusInternalServerError)
		}
	}
}
