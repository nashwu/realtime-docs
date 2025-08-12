package ws

import (
	"context"
	"net/http"
	"sync"
	"time"

	"log/slog"
	"realtime-docs/internal/store"
)

type Hub struct {
	log *slog.Logger
	bus *RedisBus
	db  *store.Postgres

	mu    sync.RWMutex
	rooms map[string]*Room // active doc rooms by docID
}

// NewHub sets up the hub with redis bus + DB + logger
func NewHub(logger *slog.Logger, bus *RedisBus, db *store.Postgres) *Hub {
	return &Hub{log: logger, bus: bus, db: db, rooms: map[string]*Room{}}
}

// Run listens to redis bus and forwards updates to local rooms
func (h *Hub) Run(ctx context.Context) {
	go h.bus.Subscribe(ctx, func(msg BusMessage) {
		h.mu.RLock()
		rm := h.rooms[msg.DocID]
		h.mu.RUnlock()
		if rm != nil {
			rm.Broadcast(msg.Payload)
		}
	})
	<-ctx.Done()
}

// room returns the Room for a doc, creating it if needed
func (h *Hub) room(docID string) *Room {
	h.mu.Lock()
	defer h.mu.Unlock()
	rm := h.rooms[docID]
	if rm == nil {
		rm = NewRoom()
		h.rooms[docID] = rm
		go rm.Run()
	}
	return rm
}

// ServeWS handles a new /ws connection for a docId
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	docID := r.URL.Query().Get("docId")
	if docID == "" {
		http.Error(w, "docId required", http.StatusBadRequest)
		return
	}

	conn, err := Accept(w, r)
	if err != nil {
		h.log.Error("ws.accept", "err", err)
		return
	}

	rm := h.room(docID)
	c := NewConn(conn, docID, rm)
	rm.Join(c)

	// Outbound writer
	go c.WriteLoop(ctx)

	// Debounced save loop (batch full snapshots every 250ms)
	go func() {
		const debounceDur = 250 * time.Millisecond
		timer := time.NewTimer(debounceDur)
		if !timer.Stop() { <-timer.C }
		var latest []byte

		for {
			select {
			case b, ok := <-c.Saves():
				if !ok {
					return
				}
				latest = b
				if !timer.Stop() { select { case <-timer.C: default: } }
				timer.Reset(debounceDur)

			case <-timer.C:
				if latest != nil {
					_ = h.db.SaveDoc(ctx, docID, latest)
					latest = nil
				}
				timer.Reset(debounceDur)

			case <-ctx.Done():
				return
			}
		}
	}()

	// Inbound reader broadcast every frame, queue-save only snapshots
	for {
		payload, ok := c.Read(ctx)
		if !ok {
			break
		}

		// Cross-instance + local broadcast
		_ = h.bus.Publish(ctx, BusMessage{DocID: docID, Payload: payload})
		rm.Broadcast(payload)

		// Frame type 3 = snapshot; strip type byte before saving
		if len(payload) > 1 && payload[0] == 3 {
			c.QueueSave(payload[1:])
		}
	}

	rm.Leave(c)
	_ = c.Close()
}
