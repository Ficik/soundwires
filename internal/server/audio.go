package server

import (
	"net/http"

	"soundwires/internal/audio"
)

func recordHandler(pwCatBin string) http.HandlerFunc {
	return audio.RecordHandler(pwCatBin)
}

func playHandler(pwPlayBin string) http.HandlerFunc {
	return audio.PlayerHandler(pwPlayBin)
}
