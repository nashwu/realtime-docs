package ws

import (
	"context"
	"net/http"
	"time"

	"nhooyr.io/websocket"
)

type Conn struct {
	ws    *websocket.Conn
	out   chan []byte
	saveQ chan []byte
	docID string
	rm    *Room
}

// Accept upgrades HTTP to websocket (allow all origins)
func Accept(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	return websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: []string{"*"},
		CompressionMode: websocket.CompressionDisabled,
	})
}

// NewConn wraps a WS connection for a specific doc + room
func NewConn(ws *websocket.Conn, docID string, rm *Room) *Conn {
	return &Conn{
		ws: ws, docID: docID, rm: rm,
		out:   make(chan []byte, 256),
		saveQ: make(chan []byte, 64),
	}
}

// Read blocks until it receives a text/binary message
// Returns false if connection is closed
func (c *Conn) Read(ctx context.Context) ([]byte, bool) {
	for {
		typ, data, err := c.ws.Read(ctx)
		if err != nil {
			return nil, false
		}
		if typ == websocket.MessageText || typ == websocket.MessageBinary {
			return []byte(data), true
		}
	}
}

// WriteLoop sends outbound messages + periodic pings
// Exits when ctx is cancelled
func (c *Conn) WriteLoop(ctx context.Context) {
	p := 20 * time.Second
	t := time.NewTicker(p)
	defer t.Stop()

	for {
		select {
		case b := <-c.out:
			_ = c.ws.Write(ctx, websocket.MessageBinary, b)
		case <-t.C:
			_ = c.ws.Ping(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// Saves returns a read-only channel of queued save events
func (c *Conn) Saves() <-chan []byte { return c.saveQ }

// QueueSave adds to save queue without blocking if full
func (c *Conn) QueueSave(b []byte) { select { case c.saveQ <- b: default: } }

// Close closes the WS connection normally
func (c *Conn) Close() error { return c.ws.Close(websocket.StatusNormalClosure, "bye") }