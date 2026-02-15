// Package calendar provides a calendar/events widget with mock data.
package calendar

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

var mockEvents = []struct {
	Title    string
	Duration int // minutes
}{
	{"Daily Standup", 15},
	{"Sprint Planning", 60},
	{"1:1 with Manager", 30},
	{"Design Review", 45},
	{"Lunch Break", 60},
	{"Code Review Session", 30},
	{"Architecture Discussion", 60},
	{"Team Retro", 45},
	{"Product Demo", 30},
	{"Focus Time", 120},
	{"Interview — Senior Engineer", 60},
	{"All Hands Meeting", 45},
}

var partialTemplate = template.Must(template.New("calendar").Parse(`
{{range .Events}}
<div class="calendar-event">
  <span class="calendar-time">{{.Time}}</span>
  <div>
    <div class="calendar-event-title">{{.Title}}</div>
    <div class="calendar-countdown">{{.Countdown}}</div>
  </div>
</div>
{{end}}
`))

// Widget provides calendar event data with mock values.
type Widget struct {
	config config.WidgetConfig
}

// New creates a new calendar widget.
func New(cfg config.WidgetConfig) *Widget {
	return &Widget{config: cfg}
}

func (w *Widget) Name() string                  { return "calendar" }
func (w *Widget) RefreshInterval() time.Duration { return w.config.RefreshInterval.Duration }
func (w *Widget) Enabled() bool                  { return w.config.IsEnabled() }

func (w *Widget) Fetch(_ context.Context) (*widget.Data, error) {
	type calendarEvent struct {
		Time      string
		Title     string
		Countdown string
	}

	now := time.Now()
	count := 5 + rand.IntN(3) // 5-7 events
	perm := rand.Perm(len(mockEvents))

	events := make([]calendarEvent, 0, count)
	for index := 0; index < count && index < len(perm); index++ {
		mockEvent := mockEvents[perm[index]]

		// Spread events across the next 8 hours
		offsetMinutes := rand.IntN(480)
		eventTime := now.Add(time.Duration(offsetMinutes) * time.Minute)

		var countdown string
		if offsetMinutes < 0 {
			countdown = "Now"
		} else if offsetMinutes < 60 {
			countdown = fmt.Sprintf("in %dm", offsetMinutes)
		} else {
			countdown = fmt.Sprintf("in %dh %dm", offsetMinutes/60, offsetMinutes%60)
		}

		events = append(events, calendarEvent{
			Time:      eventTime.Format("15:04"),
			Title:     mockEvent.Title,
			Countdown: countdown,
		})
	}

	data := map[string]any{"Events": events}

	var buf strings.Builder
	if err := partialTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering calendar partial: %w", err)
	}

	return &widget.Data{
		WidgetName: "calendar",
		HTML:       template.HTML(buf.String()),
		FetchedAt:  time.Now(),
		CacheTTL:   w.config.RefreshInterval.Duration,
	}, nil
}
