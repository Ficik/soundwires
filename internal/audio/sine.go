package audio

import (
	"context"
	"encoding/binary"
	"io"
	"math"
)

// generateSine writes continuous s16le PCM sine wave samples to w until ctx is done or w errors.
func generateSine(ctx context.Context, w io.Writer, freq, rate, channels int, amplitude float64) error {
	phase := 0.0
	step := 2 * math.Pi * float64(freq) / float64(rate)
	peak := amplitude * 32767

	// 1024 frames per write
	buf := make([]byte, 1024*channels*2)
	frames := len(buf) / (channels * 2)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		for i := 0; i < frames; i++ {
			s := int16(math.Sin(phase) * peak)
			phase += step
			if phase >= 2*math.Pi {
				phase -= 2 * math.Pi
			}
			for c := 0; c < channels; c++ {
				off := (i*channels + c) * 2
				binary.LittleEndian.PutUint16(buf[off:], uint16(s))
			}
		}

		if _, err := w.Write(buf); err != nil {
			return err
		}
	}
}
