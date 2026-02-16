package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLoggingMiddleware(t *testing.T) {
	handler := LoggingMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("ok"))
	}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/test", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", recorder.Code)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	handler := RecoveryMiddleware(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		panic("test panic")
	}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)

	// This should not panic thanks to recovery
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("expected 500 after panic, got %d", recorder.Code)
	}
}

func TestResponseCapture(t *testing.T) {
	handler := LoggingMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusNotFound)
	}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/missing", nil)
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", recorder.Code)
	}
}

func TestResponseCaptureImplementsFlusher(t *testing.T) {
	recorder := httptest.NewRecorder()
	capture := &responseCapture{ResponseWriter: recorder, statusCode: http.StatusOK}

	flusher, ok := any(capture).(http.Flusher)
	if !ok {
		t.Fatal("responseCapture does not implement http.Flusher")
	}

	// Should delegate to the recorder's Flush without panicking.
	flusher.Flush()

	if !recorder.Flushed {
		t.Error("expected underlying recorder to be flushed")
	}
}

func TestFlusherPreservedThroughMiddleware(t *testing.T) {
	flusherAvailable := false

	handler := LoggingMiddleware(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		_, flusherAvailable = writer.(http.Flusher)
		writer.WriteHeader(http.StatusOK)
	}))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/sse", nil)
	handler.ServeHTTP(recorder, request)

	if !flusherAvailable {
		t.Error("http.Flusher not available inside handler wrapped by LoggingMiddleware")
	}
}

// nonFlushableWriter is a minimal http.ResponseWriter that does not implement http.Flusher.
type nonFlushableWriter struct {
	header     http.Header
	statusCode int
}

func (writer *nonFlushableWriter) Header() http.Header             { return writer.header }
func (writer *nonFlushableWriter) Write(bytes []byte) (int, error) { return len(bytes), nil }
func (writer *nonFlushableWriter) WriteHeader(code int)            { writer.statusCode = code }

func TestFlushWithNonFlushableWriter(t *testing.T) {
	underlying := &nonFlushableWriter{header: make(http.Header)}
	capture := &responseCapture{ResponseWriter: underlying, statusCode: http.StatusOK}

	// Flush must not panic when the underlying writer lacks http.Flusher.
	capture.Flush()
}
