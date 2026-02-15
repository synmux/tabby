// Package sse provides a Server-Sent Events broker for broadcasting widget updates.
package sse

import (
	"fmt"
	"log/slog"
	"net/http"
	"sync"
)

// Event represents a single SSE event to send to clients.
type Event struct {
	// Name is the SSE event type (used by htmx sse-swap attribute).
	Name string
	// Data is the HTML partial content.
	Data string
}

// Client represents a connected SSE client.
type Client struct {
	events chan Event
	done   chan struct{}
}

// Broker manages SSE client connections and broadcasts events.
type Broker struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
}

// NewBroker creates a new SSE broker.
func NewBroker() *Broker {
	return &Broker{
		clients: make(map[*Client]struct{}),
	}
}

// Broadcast sends an event to all connected clients.
// Clients that are too slow (full buffer) are skipped for this event.
func (b *Broker) Broadcast(event Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	for client := range b.clients {
		select {
		case client.events <- event:
		default:
			slog.Debug("dropping SSE event for slow client", "event", event.Name)
		}
	}
}

// ClientCount returns the number of connected SSE clients.
func (b *Broker) ClientCount() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.clients)
}

func (b *Broker) addClient(client *Client) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.clients[client] = struct{}{}
	slog.Info("SSE client connected", "total", len(b.clients))
}

func (b *Broker) removeClient(client *Client) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.clients, client)
	close(client.events)
	slog.Info("SSE client disconnected", "total", len(b.clients))
}

// ServeHTTP handles an SSE connection. It registers the client, streams events,
// and cleans up when the client disconnects.
func (b *Broker) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	flusher, ok := writer.(http.Flusher)
	if !ok {
		http.Error(writer, "streaming not supported", http.StatusInternalServerError)
		return
	}

	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")
	writer.Header().Set("X-Accel-Buffering", "no")

	client := &Client{
		events: make(chan Event, 32),
		done:   make(chan struct{}),
	}

	b.addClient(client)
	defer b.removeClient(client)

	// Send initial keepalive comment so the client knows the connection is alive
	fmt.Fprintf(writer, ": connected\n\n")
	flusher.Flush()

	ctx := request.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-client.events:
			if !ok {
				return
			}
			fmt.Fprintf(writer, "event: %s\ndata: %s\n\n", event.Name, event.Data)
			flusher.Flush()
		}
	}
}
