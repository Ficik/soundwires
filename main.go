package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/spf13/pflag"

	"soundwires/internal/pipewire"
	"soundwires/internal/server"
)

//go:embed web/dist
var staticFiles embed.FS

func main() {
	port := pflag.IntP("port", "p", 8080, "HTTP port to listen on")
	host := pflag.StringP("host", "H", "0.0.0.0", "bind address")
	interval := pflag.IntP("interval", "i", 500, "PipeWire poll interval in milliseconds")
	pwDump := pflag.String("pw-dump", "pw-dump", "pw-dump binary path")
	pwCat := pflag.String("pw-cat", "pw-cat", "pw-cat binary path (used for both record and playback)")
	pwLink := pflag.String("pw-link", "pw-link", "pw-link binary path (used for manual port linking in monitor)")
	pflag.Parse()

	watcher := pipewire.NewWatcher(*pwDump, time.Duration(*interval)*time.Millisecond)
	go watcher.Run()

	cfg := server.Config{
		PwCatBin:  *pwCat,
		PwLinkBin: *pwLink,
		PwDumpBin: *pwDump,
	}

	addr := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("soundwires listening on http://%s", addr)

	h := server.New(staticFiles, watcher, cfg)
	if err := http.ListenAndServe(addr, h); err != nil {
		log.Fatal(err)
	}
}
