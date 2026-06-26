// Package ws implements a minimal pub/sub hub for broadcasting queue
// updates to connected patient and operator clients over WebSockets.
package ws

import "sync"

type Client struct {
	Send chan []byte
}

type Hub struct {
	mu    sync.RWMutex
	rooms map[uint]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[uint]map[*Client]bool)}
}

func (h *Hub) Subscribe(queueID uint, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[queueID] == nil {
		h.rooms[queueID] = make(map[*Client]bool)
	}
	h.rooms[queueID][client] = true
}

func (h *Hub) Unsubscribe(queueID uint, client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients, ok := h.rooms[queueID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.rooms, queueID)
		}
	}
}

// Broadcast sends message to every client subscribed to queueID. Clients
// with a full send buffer are skipped rather than blocking the broadcaster.
func (h *Hub) Broadcast(queueID uint, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.rooms[queueID] {
		select {
		case client.Send <- message:
		default:
		}
	}
}
