// Package cache provides a SQLite-backed cache for widget data.
package cache

import (
	"database/sql"
	"fmt"
	"log/slog"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Entry represents a cached widget data item.
type Entry struct {
	WidgetName string
	HTML       string
	FetchedAt  time.Time
	ExpiresAt  time.Time
	Error      string
}

// Store is a SQLite-backed cache for widget data.
type Store struct {
	db *sql.DB
	mu sync.Mutex
}

// NewStore creates a new cache store with the given SQLite database path.
// Use ":memory:" for an in-memory database (useful for testing).
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening cache database: %w", err)
	}

	// SQLite pragmas for performance
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA busy_timeout=5000",
	}
	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			db.Close()
			return nil, fmt.Errorf("setting pragma: %w", err)
		}
	}

	// Create the cache table
	createSQL := `
		CREATE TABLE IF NOT EXISTS widget_cache (
			widget_name TEXT PRIMARY KEY,
			html        TEXT NOT NULL,
			fetched_at  TEXT NOT NULL,
			expires_at  TEXT NOT NULL,
			error       TEXT NOT NULL DEFAULT ''
		)
	`
	if _, err := db.Exec(createSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("creating cache table: %w", err)
	}

	slog.Info("cache store initialised", "path", dbPath)
	return &Store{db: db}, nil
}

// Get retrieves a cached entry by widget name. Returns nil if not found or expired.
func (s *Store) Get(widgetName string) (*Entry, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	row := s.db.QueryRow(
		`SELECT widget_name, html, fetched_at, expires_at, error
		 FROM widget_cache WHERE widget_name = ?`,
		widgetName,
	)

	var entry Entry
	var fetchedAt, expiresAt string
	err := row.Scan(&entry.WidgetName, &entry.HTML, &fetchedAt, &expiresAt, &entry.Error)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning cache row: %w", err)
	}

	entry.FetchedAt, _ = time.Parse(time.RFC3339, fetchedAt)
	entry.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAt)

	// Return nil for expired entries
	if time.Now().After(entry.ExpiresAt) {
		return nil, nil
	}

	return &entry, nil
}

// Set upserts a cache entry.
func (s *Store) Set(entry *Entry) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec(
		`INSERT INTO widget_cache (widget_name, html, fetched_at, expires_at, error)
		 VALUES (?, ?, ?, ?, ?)
		 ON CONFLICT(widget_name) DO UPDATE SET
		   html = excluded.html,
		   fetched_at = excluded.fetched_at,
		   expires_at = excluded.expires_at,
		   error = excluded.error`,
		entry.WidgetName,
		entry.HTML,
		entry.FetchedAt.Format(time.RFC3339),
		entry.ExpiresAt.Format(time.RFC3339),
		entry.Error,
	)
	if err != nil {
		return fmt.Errorf("upserting cache entry: %w", err)
	}
	return nil
}

// Delete removes a cache entry by widget name.
func (s *Store) Delete(widgetName string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM widget_cache WHERE widget_name = ?", widgetName)
	if err != nil {
		return fmt.Errorf("deleting cache entry: %w", err)
	}
	return nil
}

// Purge removes all cached entries.
func (s *Store) Purge() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, err := s.db.Exec("DELETE FROM widget_cache")
	if err != nil {
		return fmt.Errorf("purging cache: %w", err)
	}
	return nil
}

// Close closes the database connection.
func (s *Store) Close() error {
	return s.db.Close()
}
