package widget

import (
	"context"
	"log/slog"
	"sync"
)

// FetchResult holds the outcome of a single widget fetch.
type FetchResult struct {
	Data  *Data
	Error error
}

// FetchAll concurrently fetches data from all enabled widgets.
// It returns results for every widget, including those that errored.
func FetchAll(ctx context.Context, registry *Registry) []FetchResult {
	enabledWidgets := make([]Widget, 0)
	for _, widget := range registry.All() {
		if widget.Enabled() {
			enabledWidgets = append(enabledWidgets, widget)
		}
	}

	results := make([]FetchResult, len(enabledWidgets))
	var waitGroup sync.WaitGroup

	for index, widget := range enabledWidgets {
		waitGroup.Add(1)
		go func(idx int, w Widget) {
			defer waitGroup.Done()

			data, err := w.Fetch(ctx)
			if err != nil {
				slog.Warn("widget fetch failed",
					"widget", w.Name(),
					"error", err,
				)
				results[idx] = FetchResult{
					Data: &Data{
						WidgetName: w.Name(),
						Error:      err.Error(),
					},
					Error: err,
				}
				return
			}

			results[idx] = FetchResult{Data: data}
		}(index, widget)
	}

	waitGroup.Wait()
	return results
}
