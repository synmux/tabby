package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"tabby/internal/config"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	cfg := config.DefaultConfig()
	return New(&cfg)
}

func TestHealthEndpoint(t *testing.T) {
	srv := newTestServer(t)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	srv.mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", recorder.Code)
	}

	var body map[string]string
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status 'ok', got %q", body["status"])
	}
}

func TestIndexEndpoint(t *testing.T) {
	srv := newTestServer(t)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	srv.mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", recorder.Code)
	}

	contentType := recorder.Header().Get("Content-Type")
	if contentType != "text/html; charset=utf-8" {
		t.Errorf("expected text/html content type, got %q", contentType)
	}

	body := recorder.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty body")
	}
}

func TestWidgetListEndpointNoRefresher(t *testing.T) {
	srv := newTestServer(t)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/widgets", nil)
	srv.mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", recorder.Code)
	}
}

func TestSSEEndpointNoHandler(t *testing.T) {
	srv := newTestServer(t)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/events", nil)
	srv.mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 when SSE handler not set, got %d", recorder.Code)
	}
}

func TestWidgetGetNotFound(t *testing.T) {
	srv := newTestServer(t)
	// Set a dummy refresher that always returns not found
	srv.SetRefresher(&dummyRefresher{})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/widgets/nonexistent", nil)
	srv.mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", recorder.Code)
	}
}

func TestStaticFiles(t *testing.T) {
	srv := newTestServer(t)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/static/css/style.css", nil)
	srv.mux.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected 200 for static file, got %d", recorder.Code)
	}

	body := recorder.Body.String()
	if len(body) == 0 {
		t.Error("expected non-empty CSS")
	}
}

// dummyRefresher implements WidgetRefresher for testing.
type dummyRefresher struct{}

func (d *dummyRefresher) RefreshWidget(_ context.Context, _ string) error {
	return nil
}

func (d *dummyRefresher) EnabledWidgets() []string {
	return nil
}

func (d *dummyRefresher) WidgetData(_ string) (map[string]any, bool) {
	return nil, false
}
