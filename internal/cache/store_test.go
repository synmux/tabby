package cache

import (
	"testing"
	"time"
)

func newTestStore(t *testing.T) *Store {
	t.Helper()
	store, err := NewStore(":memory:")
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestStoreSetAndGet(t *testing.T) {
	store := newTestStore(t)

	entry := &Entry{
		WidgetName: "weather",
		HTML:       "<div>Sunny 22°C</div>",
		FetchedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(10 * time.Minute),
		Error:      "",
	}

	if err := store.Set(entry); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := store.Get("weather")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got == nil {
		t.Fatal("expected entry, got nil")
	}
	if got.WidgetName != "weather" {
		t.Errorf("expected widget name 'weather', got %q", got.WidgetName)
	}
	if got.HTML != "<div>Sunny 22°C</div>" {
		t.Errorf("unexpected HTML: %q", got.HTML)
	}
}

func TestStoreGetMissing(t *testing.T) {
	store := newTestStore(t)

	got, err := store.Get("nonexistent")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for missing entry, got %+v", got)
	}
}

func TestStoreGetExpired(t *testing.T) {
	store := newTestStore(t)

	entry := &Entry{
		WidgetName: "weather",
		HTML:       "<div>old data</div>",
		FetchedAt:  time.Now().Add(-20 * time.Minute),
		ExpiresAt:  time.Now().Add(-10 * time.Minute), // Already expired
		Error:      "",
	}

	if err := store.Set(entry); err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	got, err := store.Get("weather")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if got != nil {
		t.Error("expected nil for expired entry")
	}
}

func TestStoreUpsert(t *testing.T) {
	store := newTestStore(t)

	entry1 := &Entry{
		WidgetName: "weather",
		HTML:       "<div>first</div>",
		FetchedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(10 * time.Minute),
	}
	entry2 := &Entry{
		WidgetName: "weather",
		HTML:       "<div>second</div>",
		FetchedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(10 * time.Minute),
	}

	if err := store.Set(entry1); err != nil {
		t.Fatal(err)
	}
	if err := store.Set(entry2); err != nil {
		t.Fatal(err)
	}

	got, err := store.Get("weather")
	if err != nil {
		t.Fatal(err)
	}
	if got.HTML != "<div>second</div>" {
		t.Errorf("expected upserted value, got %q", got.HTML)
	}
}

func TestStoreDelete(t *testing.T) {
	store := newTestStore(t)

	entry := &Entry{
		WidgetName: "weather",
		HTML:       "<div>data</div>",
		FetchedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(10 * time.Minute),
	}

	if err := store.Set(entry); err != nil {
		t.Fatal(err)
	}
	if err := store.Delete("weather"); err != nil {
		t.Fatal(err)
	}

	got, err := store.Get("weather")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Error("expected nil after delete")
	}
}

func TestStorePurge(t *testing.T) {
	store := newTestStore(t)

	for _, name := range []string{"weather", "rss", "uptime"} {
		entry := &Entry{
			WidgetName: name,
			HTML:       "<div>" + name + "</div>",
			FetchedAt:  time.Now(),
			ExpiresAt:  time.Now().Add(10 * time.Minute),
		}
		if err := store.Set(entry); err != nil {
			t.Fatal(err)
		}
	}

	if err := store.Purge(); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"weather", "rss", "uptime"} {
		got, err := store.Get(name)
		if err != nil {
			t.Fatal(err)
		}
		if got != nil {
			t.Errorf("expected nil for %q after purge", name)
		}
	}
}
