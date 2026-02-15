package widget

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestFetchAllSuccess(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&mockWidget{name: "w1", enabled: true, interval: time.Minute, html: "<div>1</div>"})
	registry.Register(&mockWidget{name: "w2", enabled: true, interval: time.Minute, html: "<div>2</div>"})
	registry.Register(&mockWidget{name: "w3", enabled: false, interval: time.Minute, html: "<div>3</div>"})

	results := FetchAll(context.Background(), registry)

	// Only 2 enabled widgets should be fetched
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	for _, result := range results {
		if result.Error != nil {
			t.Errorf("unexpected error: %v", result.Error)
		}
		if result.Data == nil {
			t.Error("expected data, got nil")
		}
	}
}

func TestFetchAllWithError(t *testing.T) {
	registry := NewRegistry()
	registry.Register(&mockWidget{name: "good", enabled: true, interval: time.Minute, html: "<div>ok</div>"})
	registry.Register(&mockWidget{name: "bad", enabled: true, interval: time.Minute, err: fmt.Errorf("api down")})

	results := FetchAll(context.Background(), registry)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}

	// Should still return results for both, with error recorded for the bad one
	var hasError bool
	for _, result := range results {
		if result.Error != nil {
			hasError = true
			if result.Data.Error != "api down" {
				t.Errorf("expected error message 'api down', got %q", result.Data.Error)
			}
		}
	}
	if !hasError {
		t.Error("expected at least one error result")
	}
}

func TestFetchAllEmpty(t *testing.T) {
	registry := NewRegistry()
	results := FetchAll(context.Background(), registry)
	if len(results) != 0 {
		t.Errorf("expected 0 results for empty registry, got %d", len(results))
	}
}
