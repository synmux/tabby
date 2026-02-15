// Package omi provides an OMI memory/insight widget with mock data.
package omi

import (
	"context"
	"fmt"
	"html/template"
	"math/rand/v2"
	"strings"
	"time"

	"tabby/internal/config"
	"tabby/internal/widget"
)

var mockMemories = []string{
	"Discussed migrating the authentication service to JWT tokens with rotating refresh keys.",
	"Noted that SQLite WAL mode significantly improves concurrent read performance.",
	"Realised the bento grid works best with 4 columns on desktop and CSS subgrid for alignment.",
	"Reminded to review the pull request for the SSE broker implementation before EOD.",
	"Had insight about using html/template for server-side rendering of widget partials.",
	"Conversation about error handling patterns — prefer explicit errors over panics in Go.",
	"Brainstormed ideas for the uptime monitoring dashboard colour scheme.",
	"Decision: use TOML over YAML for configuration due to better type support.",
	"Explored using embed.FS for zero-dependency binary distribution.",
	"Thought about caching strategies — decided on SQLite with TTL-based expiry.",
	"Meeting note: team prefers htmx over full SPA framework for dashboard.",
	"Reflection on the importance of graceful shutdown in long-running services.",
}

var partialTemplate = template.Must(template.New("omi").Parse(`
{{range .Memories}}
<div class="omi-memory">
  <div class="omi-memory-text">"{{.Text}}"</div>
  <div class="omi-memory-time">{{.TimeAgo}}</div>
</div>
{{end}}
`))

// Widget provides OMI memory data with mock values.
type Widget struct {
	config config.WidgetConfig
}

// New creates a new OMI widget.
func New(cfg config.WidgetConfig) *Widget {
	return &Widget{config: cfg}
}

func (w *Widget) Name() string                  { return "omi" }
func (w *Widget) RefreshInterval() time.Duration { return w.config.RefreshInterval.Duration }
func (w *Widget) Enabled() bool                  { return w.config.IsEnabled() }

func (w *Widget) Fetch(_ context.Context) (*widget.Data, error) {
	type memory struct {
		Text    string
		TimeAgo string
	}

	count := 3 + rand.IntN(3) // 3-5 memories
	perm := rand.Perm(len(mockMemories))

	memories := make([]memory, 0, count)
	for index := 0; index < count && index < len(perm); index++ {
		minutesAgo := rand.IntN(720) + 10

		var timeAgo string
		switch {
		case minutesAgo < 60:
			timeAgo = fmt.Sprintf("%d minutes ago", minutesAgo)
		case minutesAgo < 1440:
			timeAgo = fmt.Sprintf("%d hours ago", minutesAgo/60)
		default:
			timeAgo = fmt.Sprintf("%d days ago", minutesAgo/1440)
		}

		memories = append(memories, memory{
			Text:    mockMemories[perm[index]],
			TimeAgo: timeAgo,
		})
	}

	data := map[string]any{"Memories": memories}

	var buf strings.Builder
	if err := partialTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering omi partial: %w", err)
	}

	return &widget.Data{
		WidgetName: "omi",
		HTML:       template.HTML(buf.String()),
		FetchedAt:  time.Now(),
		CacheTTL:   w.config.RefreshInterval.Duration,
	}, nil
}
