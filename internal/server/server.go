package server

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/daveio/tabby/internal/config"
	"github.com/daveio/tabby/web"
)

// WidgetView holds the data needed to render a widget card in the index template.
type WidgetView struct {
	Name        string
	DisplayName string
	CardClass   string
	CachedHTML  template.HTML
}

// WidgetDataProvider supplies cached widget HTML for initial page render.
type WidgetDataProvider interface {
	CachedHTML(widgetName string) template.HTML
}

// SSEHandler manages the Server-Sent Events endpoint.
type SSEHandler interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// WidgetRefresher triggers a forced refresh of a specific widget.
type WidgetRefresher interface {
	RefreshWidget(ctx context.Context, name string) error
	EnabledWidgets() []string
	WidgetData(name string) (map[string]any, bool)
}

// Server is the main HTTP server for Tabby.
type Server struct {
	config        *config.Config
	mux           *http.ServeMux
	httpServer    *http.Server
	indexTemplate *template.Template
	dataProvider  WidgetDataProvider
	sseHandler    SSEHandler
	refresher     WidgetRefresher
}

// New creates a new Server with the given configuration.
func New(cfg *config.Config) *Server {
	srv := &Server{
		config: cfg,
		mux:    http.NewServeMux(),
	}

	srv.parseTemplates()
	srv.registerRoutes()

	srv.httpServer = &http.Server{
		Addr:              cfg.Server.Address(),
		Handler:           RecoveryMiddleware(LoggingMiddleware(srv.mux)),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	return srv
}

// SetDataProvider sets the widget data provider for initial page loads.
func (s *Server) SetDataProvider(provider WidgetDataProvider) {
	s.dataProvider = provider
}

// SetSSEHandler sets the SSE handler for live updates.
func (s *Server) SetSSEHandler(handler SSEHandler) {
	s.sseHandler = handler
}

// SetRefresher sets the widget refresher for force-refresh and listing.
func (s *Server) SetRefresher(refresher WidgetRefresher) {
	s.refresher = refresher
}

func (s *Server) parseTemplates() {
	tmpl, err := template.ParseFS(web.TemplateFiles, "templates/index.html")
	if err != nil {
		slog.Error("failed to parse index template", "error", err)
		// Create a minimal fallback template
		tmpl = template.Must(template.New("index.html").Parse("<html><body><h1>Tabby</h1><p>Template parse error</p></body></html>"))
	}
	s.indexTemplate = tmpl
}

func (s *Server) registerRoutes() {
	// Static files
	staticFS, err := fs.Sub(web.StaticFiles, "static")
	if err != nil {
		slog.Error("failed to create static sub-filesystem", "error", err)
	} else {
		s.mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	}

	// Pages
	s.mux.HandleFunc("GET /{$}", s.handleIndex)

	// API
	s.mux.HandleFunc("GET /api/events", s.handleSSE)
	s.mux.HandleFunc("GET /api/widgets", s.handleWidgetList)
	s.mux.HandleFunc("GET /api/widgets/{name}", s.handleWidgetGet)
	s.mux.HandleFunc("POST /api/widgets/{name}/refresh", s.handleWidgetRefresh)
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
}

func (s *Server) handleIndex(writer http.ResponseWriter, request *http.Request) {
	widgets := s.buildWidgetViews()

	data := map[string]any{
		"Widgets": widgets,
	}

	writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.indexTemplate.Execute(writer, data); err != nil {
		slog.Error("failed to render index", "error", err)
	}
}

// widgetDisplayNames maps widget names to their display titles.
var widgetDisplayNames = map[string]string{
	"weather":  "Weather",
	"rss":      "Headlines",
	"uptime":   "Uptime",
	"github":   "GitHub",
	"linear":   "Linear",
	"notion":   "Notion",
	"omi":      "OMI Memories",
	"calendar": "Calendar",
}

// widgetCardClasses assigns CSS grid classes for bento layout variety.
var widgetCardClasses = map[string]string{
	"weather":  "",
	"rss":      "widget-card--wide",
	"uptime":   "",
	"github":   "widget-card--wide",
	"linear":   "",
	"notion":   "",
	"omi":      "",
	"calendar": "",
}

func (s *Server) buildWidgetViews() []WidgetView {
	// Canonical widget order for the bento grid
	order := []string{"weather", "rss", "uptime", "github", "linear", "notion", "omi", "calendar"}

	var views []WidgetView
	for _, name := range order {
		widgetCfg, exists := s.config.Widgets[name]
		if !exists {
			continue
		}
		if !widgetCfg.IsEnabled() {
			continue
		}

		var cachedHTML template.HTML
		if s.dataProvider != nil {
			cachedHTML = s.dataProvider.CachedHTML(name)
		}
		if cachedHTML == "" {
			cachedHTML = `<div class="widget-error">Loading...</div>`
		}

		displayName := widgetDisplayNames[name]
		if displayName == "" {
			displayName = name
		}

		views = append(views, WidgetView{
			Name:        name,
			DisplayName: displayName,
			CardClass:   widgetCardClasses[name],
			CachedHTML:  cachedHTML,
		})
	}
	return views
}

func (s *Server) handleSSE(writer http.ResponseWriter, request *http.Request) {
	if s.sseHandler == nil {
		http.Error(writer, "SSE not available", http.StatusServiceUnavailable)
		return
	}
	s.sseHandler.ServeHTTP(writer, request)
}

func (s *Server) handleWidgetList(writer http.ResponseWriter, _ *http.Request) {
	if s.refresher == nil {
		writeJSON(writer, http.StatusOK, map[string]any{"widgets": []string{}})
		return
	}

	names := s.refresher.EnabledWidgets()
	result := make([]map[string]any, 0, len(names))
	for _, name := range names {
		entry := map[string]any{"name": name}
		if data, ok := s.refresher.WidgetData(name); ok {
			for key, value := range data {
				entry[key] = value
			}
		}
		result = append(result, entry)
	}
	writeJSON(writer, http.StatusOK, map[string]any{"widgets": result})
}

func (s *Server) handleWidgetGet(writer http.ResponseWriter, request *http.Request) {
	name := request.PathValue("name")
	if s.refresher == nil {
		http.Error(writer, "widget system not ready", http.StatusServiceUnavailable)
		return
	}

	data, ok := s.refresher.WidgetData(name)
	if !ok {
		http.Error(writer, fmt.Sprintf("widget %q not found", name), http.StatusNotFound)
		return
	}
	writeJSON(writer, http.StatusOK, data)
}

func (s *Server) handleWidgetRefresh(writer http.ResponseWriter, request *http.Request) {
	name := request.PathValue("name")
	if s.refresher == nil {
		http.Error(writer, "widget system not ready", http.StatusServiceUnavailable)
		return
	}

	if err := s.refresher.RefreshWidget(request.Context(), name); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writeJSON(writer, http.StatusOK, map[string]string{"status": "refreshed", "widget": name})
}

func (s *Server) handleHealth(writer http.ResponseWriter, _ *http.Request) {
	writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(writer http.ResponseWriter, status int, data any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	if err := json.NewEncoder(writer).Encode(data); err != nil {
		slog.Error("failed to write JSON response", "error", err)
	}
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	slog.Info("starting server", "address", s.config.Server.Address())
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}
