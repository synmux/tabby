// Package rss provides an RSS/Atom feed widget with mock data.
package rss

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

var mockHeadlines = []struct {
	Title  string
	Source string
}{
	{"Scientists Discover New Method for Carbon Capture Using Seaweed", "Nature"},
	{"Major Tech Company Announces Breakthrough in Quantum Computing", "Ars Technica"},
	{"Open Source AI Model Surpasses Commercial Alternatives in Benchmark", "HN"},
	{"Global Chip Shortage Expected to Ease by Mid-Year", "Reuters"},
	{"New Programming Language Promises 10x Developer Productivity", "HN"},
	{"Space Agency Confirms Water Ice Deposits on Mars Surface", "BBC"},
	{"Cybersecurity Researchers Find Critical Vulnerability in IoT Devices", "Wired"},
	{"Electric Vehicle Battery Costs Drop Below $100/kWh Milestone", "Bloomberg"},
	{"Remote Work Policies Shift: Major Companies Adopt Hybrid Models", "FT"},
	{"New Study Links Sleep Quality to Cognitive Performance in Adults", "Nature"},
	{"Breakthrough in Fusion Energy: Sustained Reaction Achieved", "Science"},
	{"Browser Wars Heat Up as New Competitor Enters the Market", "The Verge"},
	{"Global Internet Traffic Hits Record High During Solar Eclipse", "Cloudflare"},
	{"Machine Learning Model Predicts Earthquake Aftershocks Accurately", "MIT Review"},
	{"New CSS Features Enable Complex Layouts Without JavaScript", "Smashing Mag"},
	{"Rust Overtakes C++ in Systems Programming Popularity Survey", "Stack Overflow"},
	{"European Union Passes Comprehensive AI Regulation Framework", "Reuters"},
	{"Astronomers Detect Radio Signal from Exoplanet Atmosphere", "ESA"},
	{"Database Performance Comparison: PostgreSQL vs Modern Alternatives", "HN"},
	{"Undersea Cable Repair Completes, Restoring Connectivity to Region", "BBC"},
}

var partialTemplate = template.Must(template.New("rss").Parse(`
{{range .Items}}
<div class="rss-item">
  <div class="rss-item-title">{{.Title}}</div>
  <div class="rss-item-meta">{{.Source}} · {{.TimeAgo}}</div>
</div>
{{end}}
`))

// Widget provides RSS feed data with mock values.
type Widget struct {
	config config.WidgetConfig
}

// New creates a new RSS widget.
func New(cfg config.WidgetConfig) *Widget {
	return &Widget{config: cfg}
}

func (w *Widget) Name() string                  { return "rss" }
func (w *Widget) RefreshInterval() time.Duration { return w.config.RefreshInterval.Duration }
func (w *Widget) Enabled() bool                  { return w.config.IsEnabled() }

func (w *Widget) Fetch(_ context.Context) (*widget.Data, error) {
	// Randomly select 5-8 headlines in shuffled order
	count := 5 + rand.IntN(4)
	perm := rand.Perm(len(mockHeadlines))

	type feedItem struct {
		Title   string
		Source  string
		TimeAgo string
	}

	items := make([]feedItem, 0, count)
	for index := 0; index < count && index < len(perm); index++ {
		headline := mockHeadlines[perm[index]]
		minutesAgo := rand.IntN(180) + 1
		var timeAgo string
		if minutesAgo < 60 {
			timeAgo = fmt.Sprintf("%dm ago", minutesAgo)
		} else {
			timeAgo = fmt.Sprintf("%dh ago", minutesAgo/60)
		}
		items = append(items, feedItem{
			Title:   headline.Title,
			Source:  headline.Source,
			TimeAgo: timeAgo,
		})
	}

	data := map[string]any{"Items": items}

	var buf strings.Builder
	if err := partialTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering rss partial: %w", err)
	}

	return &widget.Data{
		WidgetName: "rss",
		HTML:       template.HTML(buf.String()),
		FetchedAt:  time.Now(),
		CacheTTL:   w.config.RefreshInterval.Duration,
	}, nil
}
