package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"soundwires/internal/pipewire"
	"soundwires/internal/server"
)

//go:embed web/dist
var staticFiles embed.FS

func main() {
	port := flag.Int("port", 8080, "HTTP port to listen on")
	host := flag.String("host", "0.0.0.0", "bind address")
	interval := flag.Int("interval", 500, "PipeWire poll interval in milliseconds")
	pwDump := flag.String("pw-dump", "pw-dump", "pw-dump binary path")
	pwCat := flag.String("pw-cat", "pw-cat", "pw-cat binary path (used for both record and playback)")
	flag.Parse()

	watcher := pipewire.NewWatcher(*pwDump, time.Duration(*interval)*time.Millisecond)
	go watcher.Run()

	cfg := server.Config{
		PwCatBin: *pwCat,
	}

	addr := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("soundwires listening on http://%s", addr)

	h := server.New(staticFiles, watcher, cfg)
	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatal(err)
	}
}
