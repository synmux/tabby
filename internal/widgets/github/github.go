// Package github provides a GitHub activity widget with mock data.
package github

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

var mockCommitMessages = []string{
	"feat: add user authentication flow",
	"fix: resolve race condition in worker pool",
	"refactor: extract config parser to separate package",
	"chore: update dependencies to latest versions",
	"docs: add API endpoint documentation",
	"feat: implement SSE live updates",
	"fix: correct timezone handling in scheduler",
	"test: add integration tests for cache layer",
	"feat: add dark mode toggle",
	"fix: handle nil pointer in widget registry",
	"perf: optimise database query with index",
	"feat: implement bento grid layout",
	"fix: resolve CSS grid overflow on mobile",
	"chore: configure CI pipeline",
}

var mockBranches = []string{
	"main", "feat/auth", "fix/sse-reconnect", "refactor/config",
	"feat/dark-mode", "fix/cache-ttl", "feat/widgets", "chore/deps",
}

var mockAuthors = []string{
	"alice", "bob", "charlie", "dave", "eve",
}

var mockPRTitles = []string{
	"Add user authentication with JWT",
	"Fix SSE reconnection logic",
	"Implement widget caching layer",
	"Update to Go 1.25",
	"Add responsive bento grid",
	"Refactor config parsing",
}

var partialTemplate = template.Must(template.New("github").Parse(`
{{range .Items}}
<div class="github-item">
  <div class="github-item-title">{{.Title}}</div>
  <div class="github-item-meta">
    {{.Author}} · <span class="github-branch">{{.Branch}}</span> · {{.TimeAgo}}
  </div>
</div>
{{end}}
`))

// Widget provides GitHub activity data with mock values.
type Widget struct {
	config config.WidgetConfig
}

// New creates a new GitHub widget.
func New(cfg config.WidgetConfig) *Widget {
	return &Widget{config: cfg}
}

func (w *Widget) Name() string                   { return "github" }
func (w *Widget) RefreshInterval() time.Duration { return w.config.RefreshInterval.Duration }
func (w *Widget) Enabled() bool                  { return w.config.IsEnabled() }

func (w *Widget) Fetch(_ context.Context) (*widget.Data, error) {
	type githubItem struct {
		Title   string
		Author  string
		Branch  string
		TimeAgo string
	}

	count := 4 + rand.IntN(3) // 4-6 items
	items := make([]githubItem, 0, count)

	for range count {
		// Mix commits and PRs
		isPR := rand.IntN(3) == 0
		var title string
		if isPR {
			title = "PR: " + mockPRTitles[rand.IntN(len(mockPRTitles))]
		} else {
			title = mockCommitMessages[rand.IntN(len(mockCommitMessages))]
		}

		minutesAgo := rand.IntN(300) + 1
		var timeAgo string
		if minutesAgo < 60 {
			timeAgo = fmt.Sprintf("%dm ago", minutesAgo)
		} else {
			timeAgo = fmt.Sprintf("%dh ago", minutesAgo/60)
		}

		items = append(items, githubItem{
			Title:   title,
			Author:  mockAuthors[rand.IntN(len(mockAuthors))],
			Branch:  mockBranches[rand.IntN(len(mockBranches))],
			TimeAgo: timeAgo,
		})
	}

	data := map[string]any{"Items": items}

	var buf strings.Builder
	if err := partialTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering github partial: %w", err)
	}

	return &widget.Data{
		WidgetName: "github",
		HTML:       template.HTML(buf.String()),
		FetchedAt:  time.Now(),
		CacheTTL:   w.config.RefreshInterval.Duration,
	}, nil
}
