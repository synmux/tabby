// Package notion provides a Notion recent pages widget with mock data.
package notion

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

var mockPages = []struct {
	Title string
	Icon  string
}{
	{"Project Roadmap Q1 2026", "🗺️"},
	{"Meeting Notes — Weekly Standup", "📝"},
	{"Architecture Decision Records", "🏗️"},
	{"Sprint Retrospective", "🔄"},
	{"API Design Guidelines", "📐"},
	{"Onboarding Checklist", "✅"},
	{"Bug Triage Dashboard", "🐛"},
	{"Product Requirements Document", "📋"},
	{"Team OKRs", "🎯"},
	{"Research: SSE vs WebSockets", "🔬"},
	{"Code Review Standards", "👀"},
	{"Infrastructure Runbook", "🛠️"},
}

var partialTemplate = template.Must(template.New("notion").Parse(`
{{range .Pages}}
<div class="notion-page">
  <div class="notion-page-title">{{.Icon}} {{.Title}}</div>
  <div class="notion-page-meta">Edited {{.EditedAgo}}</div>
</div>
{{end}}
`))

// Widget provides Notion page data with mock values.
type Widget struct {
	config config.WidgetConfig
}

// New creates a new Notion widget.
func New(cfg config.WidgetConfig) *Widget {
	return &Widget{config: cfg}
}

func (w *Widget) Name() string                  { return "notion" }
func (w *Widget) RefreshInterval() time.Duration { return w.config.RefreshInterval.Duration }
func (w *Widget) Enabled() bool                  { return w.config.IsEnabled() }

func (w *Widget) Fetch(_ context.Context) (*widget.Data, error) {
	type notionPage struct {
		Title     string
		Icon      string
		EditedAgo string
	}

	count := 4 + rand.IntN(3) // 4-6 pages
	perm := rand.Perm(len(mockPages))

	pages := make([]notionPage, 0, count)
	for index := 0; index < count && index < len(perm); index++ {
		page := mockPages[perm[index]]
		minutesAgo := rand.IntN(480) + 1

		var editedAgo string
		switch {
		case minutesAgo < 60:
			editedAgo = fmt.Sprintf("%dm ago", minutesAgo)
		case minutesAgo < 1440:
			editedAgo = fmt.Sprintf("%dh ago", minutesAgo/60)
		default:
			editedAgo = fmt.Sprintf("%dd ago", minutesAgo/1440)
		}

		pages = append(pages, notionPage{
			Title:     page.Title,
			Icon:      page.Icon,
			EditedAgo: editedAgo,
		})
	}

	data := map[string]any{"Pages": pages}

	var buf strings.Builder
	if err := partialTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering notion partial: %w", err)
	}

	return &widget.Data{
		WidgetName: "notion",
		HTML:       template.HTML(buf.String()),
		FetchedAt:  time.Now(),
		CacheTTL:   w.config.RefreshInterval.Duration,
	}, nil
}
