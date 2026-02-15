// Package linear provides a Linear issue tracking widget with mock data.
package linear

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

var mockTickets = []struct {
	ID    string
	Title string
}{
	{"TAB-101", "Implement SSE live updates"},
	{"TAB-102", "Add weather widget"},
	{"TAB-103", "Fix cache invalidation bug"},
	{"TAB-104", "Design bento grid layout"},
	{"TAB-105", "Add configuration validation"},
	{"TAB-106", "Implement graceful shutdown"},
	{"TAB-107", "Write integration tests"},
	{"TAB-108", "Add TLS support"},
	{"TAB-109", "Optimise SQLite queries"},
	{"TAB-110", "Create widget error states"},
	{"TAB-111", "Add mobile responsive layout"},
	{"TAB-112", "Implement RSS feed parser"},
}

type ticketStatus struct {
	Label string
	Class string
}

var statuses = []ticketStatus{
	{"Todo", "todo"},
	{"In Progress", "progress"},
	{"In Progress", "progress"},
	{"Done", "done"},
}

var partialTemplate = template.Must(template.New("linear").Parse(`
{{range .Tickets}}
<div class="linear-ticket">
  <span class="linear-status-dot linear-status-dot--{{.StatusClass}}"></span>
  <span class="linear-ticket-id">{{.ID}}</span>
  <span class="linear-ticket-title">{{.Title}}</span>
</div>
{{end}}
`))

// Widget provides Linear issue data with mock values.
type Widget struct {
	config config.WidgetConfig
}

// New creates a new Linear widget.
func New(cfg config.WidgetConfig) *Widget {
	return &Widget{config: cfg}
}

func (w *Widget) Name() string                  { return "linear" }
func (w *Widget) RefreshInterval() time.Duration { return w.config.RefreshInterval.Duration }
func (w *Widget) Enabled() bool                  { return w.config.IsEnabled() }

func (w *Widget) Fetch(_ context.Context) (*widget.Data, error) {
	type linearTicket struct {
		ID          string
		Title       string
		StatusClass string
	}

	// Select 5-7 tickets with cycling statuses
	count := 5 + rand.IntN(3)
	perm := rand.Perm(len(mockTickets))

	tickets := make([]linearTicket, 0, count)
	for index := 0; index < count && index < len(perm); index++ {
		ticket := mockTickets[perm[index]]
		status := statuses[rand.IntN(len(statuses))]

		tickets = append(tickets, linearTicket{
			ID:          ticket.ID,
			Title:       ticket.Title,
			StatusClass: status.Class,
		})
	}

	data := map[string]any{"Tickets": tickets}

	var buf strings.Builder
	if err := partialTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering linear partial: %w", err)
	}

	return &widget.Data{
		WidgetName: "linear",
		HTML:       template.HTML(buf.String()),
		FetchedAt:  time.Now(),
		CacheTTL:   w.config.RefreshInterval.Duration,
	}, nil
}
