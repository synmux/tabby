// Package uptime provides an uptime monitoring widget with mock data.
package uptime

import (
	"context"
	"fmt"
	"html/template"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/daveio/tabby/internal/config"
	"github.com/daveio/tabby/internal/widget"
)

var mockServices = []string{
	"api.example.com",
	"dashboard.example.com",
	"auth.example.com",
	"cdn.example.com",
	"db-primary.internal",
}

var partialTemplate = template.Must(template.New("uptime").Parse(`
{{range .Services}}
<div class="uptime-service">
  <span class="uptime-name">{{.Name}}</span>
  <span class="uptime-status">
    <span class="uptime-dot uptime-dot--{{.StatusClass}}"></span>
    {{.ResponseTime}}ms
  </span>
</div>
{{end}}
`))

// Widget provides uptime monitoring data with mock values.
type Widget struct {
	config config.WidgetConfig
}

// New creates a new uptime widget.
func New(cfg config.WidgetConfig) *Widget {
	return &Widget{config: cfg}
}

func (w *Widget) Name() string                  { return "uptime" }
func (w *Widget) RefreshInterval() time.Duration { return w.config.RefreshInterval.Duration }
func (w *Widget) Enabled() bool                  { return w.config.IsEnabled() }

func (w *Widget) Fetch(_ context.Context) (*widget.Data, error) {
	type serviceStatus struct {
		Name         string
		StatusClass  string
		ResponseTime int
	}

	services := make([]serviceStatus, len(mockServices))
	for index, name := range mockServices {
		// 80% healthy, 12% degraded, 8% down
		roll := rand.IntN(100)
		var statusClass string
		var responseTime int

		switch {
		case roll < 80:
			statusClass = "healthy"
			responseTime = 20 + rand.IntN(150)
		case roll < 92:
			statusClass = "degraded"
			responseTime = 500 + rand.IntN(2000)
		default:
			statusClass = "down"
			responseTime = 0
		}

		services[index] = serviceStatus{
			Name:         name,
			StatusClass:  statusClass,
			ResponseTime: responseTime,
		}
	}

	data := map[string]any{"Services": services}

	var buf strings.Builder
	if err := partialTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering uptime partial: %w", err)
	}

	return &widget.Data{
		WidgetName: "uptime",
		HTML:       template.HTML(buf.String()),
		FetchedAt:  time.Now(),
		CacheTTL:   w.config.RefreshInterval.Duration,
	}, nil
}
