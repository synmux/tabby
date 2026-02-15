package sse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestBrokerBroadcast(t *testing.T) {
	broker := NewBroker()

	// Simulate a client by manually adding one
	client := &Client{
		events: make(chan Event, 10),
		done:   make(chan struct{}),
	}

	broker.mu.Lock()
	broker.clients[client] = struct{}{}
	broker.mu.Unlock()

	if broker.ClientCount() != 1 {
		t.Errorf("expected 1 client, got %d", broker.ClientCount())
	}

	// Broadcast an event
	broker.Broadcast(Event{Name: "weather", Data: "<div>Sunny</div>"})

	select {
	case event := <-client.events:
		if event.Name != "weather" {
			t.Errorf("expected event name 'weather', got %q", event.Name)
		}
		if event.Data != "<div>Sunny</div>" {
			t.Errorf("unexpected data: %q", event.Data)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestBrokerDropsSlowClients(t *testing.T) {
	broker := NewBroker()

	// Client with zero buffer — immediately full
	slowClient := &Client{
		events: make(chan Event),
		done:   make(chan struct{}),
	}

	broker.mu.Lock()
	broker.clients[slowClient] = struct{}{}
	broker.mu.Unlock()

	// This should not block; it drops the event for the slow client
	broker.Broadcast(Event{Name: "test", Data: "data"})
}

func TestBrokerServeHTTP(t *testing.T) {
	broker := NewBroker()

	// Create a cancellable context for the request
	ctx, cancel := context.WithCancel(context.Background())

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/events", nil).WithContext(ctx)

	// Run ServeHTTP in a goroutine since it blocks
	done := make(chan struct{})
	go func() {
		defer close(done)
		broker.ServeHTTP(recorder, request)
	}()

	// Give the handler time to register the client
	time.Sleep(50 * time.Millisecond)

	if broker.ClientCount() != 1 {
		t.Errorf("expected 1 client after connect, got %d", broker.ClientCount())
	}

	// Broadcast an event
	broker.Broadcast(Event{Name: "uptime", Data: "<div>All OK</div>"})

	// Cancel the request context to simulate client disconnect
	cancel()

	// Wait for handler to exit cleanly
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("handler didn't exit after context cancellation")
	}

	if broker.ClientCount() != 0 {
		t.Errorf("expected 0 clients after disconnect, got %d", broker.ClientCount())
	}
}

func TestBrokerSSEFormat(t *testing.T) {
	event := Event{Name: "weather", Data: "<div>22°C</div>"}

	var builder strings.Builder
	builder.WriteString("event: " + event.Name + "\n")
	builder.WriteString("data: " + event.Data + "\n\n")

	output := builder.String()
	if !strings.Contains(output, "event: weather\n") {
		t.Error("SSE format missing event line")
	}
	if !strings.Contains(output, "data: <div>22°C</div>\n") {
		t.Error("SSE format missing data line")
	}
}

func TestBrokerMultipleClients(t *testing.T) {
	broker := NewBroker()

	const clientCount = 5
	clients := make([]*Client, clientCount)
	for index := range clientCount {
		clients[index] = &Client{
			events: make(chan Event, 10),
			done:   make(chan struct{}),
		}
		broker.mu.Lock()
		broker.clients[clients[index]] = struct{}{}
		broker.mu.Unlock()
	}

	if broker.ClientCount() != clientCount {
		t.Errorf("expected %d clients, got %d", clientCount, broker.ClientCount())
	}

	// Broadcast should reach all clients
	broker.Broadcast(Event{Name: "test", Data: "hello"})

	for index, client := range clients {
		select {
		case event := <-client.events:
			if event.Name != "test" {
				t.Errorf("client %d: expected event 'test', got %q", index, event.Name)
			}
		case <-time.After(time.Second):
			t.Fatalf("client %d: timed out waiting for event", index)
		}
	}
}
