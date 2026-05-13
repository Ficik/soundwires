package pipewire

import (
	"encoding/json"
	"log"
	"time"
)

type EventKind string

const (
	EventInit    EventKind = "init"
	EventAdded   EventKind = "added"
	EventChanged EventKind = "changed"
	EventRemoved EventKind = "removed"
)

type Event struct {
	Type   EventKind `json:"type"`
	Object *PWObject `json:"object,omitempty"`
	ID     *int      `json:"id,omitempty"`

	// Only for EventInit
	Objects []PWObject `json:"objects,omitempty"`
}

type Watcher struct {
	pwDumpBin string
	interval  time.Duration
	Events    chan Event
	reload    chan struct{}

	current map[int]string // id → raw JSON fingerprint
}

func NewWatcher(pwDumpBin string, interval time.Duration) *Watcher {
	return &Watcher{
		pwDumpBin: pwDumpBin,
		interval:  interval,
		Events:    make(chan Event, 64),
		reload:    make(chan struct{}, 1),
		current:   make(map[int]string),
	}
}

// Reload drops the watcher's tracked state and causes the next poll to emit
// a fresh EventInit, so clients can rebuild the graph from scratch. Safe to
// call from any goroutine; coalesces if called repeatedly while pending.
func (w *Watcher) Reload() {
	select {
	case w.reload <- struct{}{}:
	default:
	}
}

func (w *Watcher) Run() {
	first := true
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	poll := func() {
		objects, err := Dump(w.pwDumpBin)
		if err != nil {
			log.Printf("pw-dump error: %v", err)
			return
		}

		if first {
			first = false
			w.current = make(map[int]string, len(objects))
			for _, obj := range objects {
				w.current[obj.ID] = fingerprint(obj)
			}
			w.Events <- Event{Type: EventInit, Objects: objects}
			return
		}

		seen := make(map[int]bool, len(objects))
		for _, obj := range objects {
			seen[obj.ID] = true
			fp := fingerprint(obj)
			prev, exists := w.current[obj.ID]
			if !exists {
				w.current[obj.ID] = fp
				o := obj
				w.Events <- Event{Type: EventAdded, Object: &o}
			} else if prev != fp {
				w.current[obj.ID] = fp
				o := obj
				w.Events <- Event{Type: EventChanged, Object: &o}
			}
		}

		for id := range w.current {
			if !seen[id] {
				delete(w.current, id)
				id := id
				w.Events <- Event{Type: EventRemoved, ID: &id}
			}
		}
	}

	for {
		select {
		case <-w.reload:
			first = true
			poll()
		case <-ticker.C:
			poll()
		}
	}
}

func fingerprint(obj PWObject) string {
	b, _ := json.Marshal(obj)
	return string(b)
}
