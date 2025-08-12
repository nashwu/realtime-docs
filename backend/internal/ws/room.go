package ws

import "sync"

type Room struct {
	mu ync.RWMutex
	clients map[*Conn]struct{} // active connections in this room
}

// NewRoom creates an empty room
func NewRoom() *Room { return &Room{clients: map[*Conn]struct{}{}} }

// Run is a placeholder could handle cleanup, ticks, etc
func (r *Room) Run() {}

// Join adds a connection to the room
func (r *Room) Join(c *Conn) {
	r.mu.Lock()
	r.clients[c] = struct{}{}
	r.mu.Unlock()
}

// Leave removes a connection from the room
func (r *Room) Leave(c *Conn) {
	r.mu.Lock()
	delete(r.clients, c)
	r.mu.Unlock()
}

// Broadcast sends a message to all connections without blocking
func (r *Room) Broadcast(b []byte) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for c := range r.clients {
		select {
		case c.out <- b:
		default: // skip if send buffer is full
		}
	}
}
