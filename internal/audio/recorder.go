package audio

import (
	"context"
	"encoding/binary"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strconv"
)

// RecordHandler streams audio from a PipeWire node as a WAV response.
func RecordHandler(pwCatBin string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		targetID := q.Get("targetId")
		if targetID == "" {
			http.Error(w, "targetId required", http.StatusBadRequest)
			return
		}

		rate, _ := strconv.Atoi(q.Get("rate"))
		if rate == 0 {
			rate = 48000
		}
		channels, _ := strconv.Atoi(q.Get("channels"))
		if channels == 0 {
			channels = 2
		}

		ctx, cancel := context.WithCancel(r.Context())
		defer cancel()

		cmd := exec.CommandContext(ctx, pwCatBin,
			"--record",
			"--target", targetID,
			"--format", "s16",
			"--channels", strconv.Itoa(channels),
			"--rate", strconv.Itoa(rate),
			"--properties", "{\"node.name\":\"Soundwires Monitor\"}",
			"-",
		)

		stdout, err := cmd.StdoutPipe()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := cmd.Start(); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() {
			cancel()
			cmd.Wait()
		}()

		w.Header().Set("Content-Type", "audio/wav")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Transfer-Encoding", "chunked")

		writeStreamingWAVHeader(w, channels, rate)
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				if _, werr := w.Write(buf[:n]); werr != nil {
					log.Printf("record write: %v", werr)
					return
				}
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
			if err != nil {
				if err != io.EOF {
					log.Printf("pw-cat read: %v", err)
				}
				return
			}
		}
	}
}

// writeStreamingWAVHeader writes a WAV header with an indefinite data length (0xFFFFFFFF).
// Modern browsers handle this for streaming audio.
func writeStreamingWAVHeader(w io.Writer, channels, rate int) {
	bitsPerSample := 16
	byteRate := rate * channels * bitsPerSample / 8
	blockAlign := channels * bitsPerSample / 8

	buf := make([]byte, 44)

	// RIFF chunk
	copy(buf[0:], "RIFF")
	binary.LittleEndian.PutUint32(buf[4:], 0xFFFFFFFF) // indefinite size
	copy(buf[8:], "WAVE")

	// fmt sub-chunk
	copy(buf[12:], "fmt ")
	binary.LittleEndian.PutUint32(buf[16:], 16)                    // subchunk size
	binary.LittleEndian.PutUint16(buf[20:], 1)                     // PCM
	binary.LittleEndian.PutUint16(buf[22:], uint16(channels))      //nolint:gosec
	binary.LittleEndian.PutUint32(buf[24:], uint32(rate))          //nolint:gosec
	binary.LittleEndian.PutUint32(buf[28:], uint32(byteRate))      //nolint:gosec
	binary.LittleEndian.PutUint16(buf[32:], uint16(blockAlign))    //nolint:gosec
	binary.LittleEndian.PutUint16(buf[34:], uint16(bitsPerSample)) //nolint:gosec

	// data sub-chunk
	copy(buf[36:], "data")
	binary.LittleEndian.PutUint32(buf[40:], 0xFFFFFFFF) // indefinite data size

	w.Write(buf) //nolint:errcheck
}
