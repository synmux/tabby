// Package weather provides a weather widget with mock data.
package weather

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

var weatherConditions = []struct {
	Label string
	Icon  string
}{
	{"Sunny", "☀️"},
	{"Partly Cloudy", "⛅"},
	{"Cloudy", "☁️"},
	{"Light Rain", "🌦️"},
	{"Rain", "🌧️"},
	{"Thunderstorm", "⛈️"},
	{"Snow", "🌨️"},
	{"Foggy", "🌫️"},
	{"Windy", "💨"},
	{"Clear Night", "🌙"},
}

var partialTemplate = template.Must(template.New("weather").Parse(`
<div class="weather-current">
  <span class="weather-temp">{{.Temperature}}°C</span>
  <span class="weather-condition">{{.Icon}} {{.Condition}}</span>
</div>
<div class="weather-detail">Feels like {{.FeelsLike}}°C · Wind {{.WindSpeed}} km/h {{.WindDir}}</div>
<div class="weather-detail">Humidity {{.Humidity}}% · UV Index {{.UVIndex}}</div>
`))

// Widget provides weather data with mock values.
type Widget struct {
	config   config.WidgetConfig
	baseTemp float64
}

// New creates a new weather widget.
func New(cfg config.WidgetConfig) *Widget {
	return &Widget{
		config:   cfg,
		baseTemp: 14.0 + rand.Float64()*10, // Base temp 14-24°C
	}
}

func (w *Widget) Name() string                  { return "weather" }
func (w *Widget) RefreshInterval() time.Duration { return w.config.RefreshInterval.Duration }
func (w *Widget) Enabled() bool                  { return w.config.IsEnabled() }

func (w *Widget) Fetch(_ context.Context) (*widget.Data, error) {
	condition := weatherConditions[rand.IntN(len(weatherConditions))]
	temp := w.baseTemp + (rand.Float64()*6 - 3) // ±3°C drift
	feelsLike := temp + (rand.Float64()*4 - 2)
	windSpeed := 5 + rand.IntN(35)
	humidity := 30 + rand.IntN(60)
	uvIndex := rand.IntN(11)

	windDirs := []string{"N", "NE", "E", "SE", "S", "SW", "W", "NW"}
	windDir := windDirs[rand.IntN(len(windDirs))]

	data := map[string]any{
		"Temperature": fmt.Sprintf("%.1f", temp),
		"FeelsLike":   fmt.Sprintf("%.1f", feelsLike),
		"Condition":   condition.Label,
		"Icon":        condition.Icon,
		"WindSpeed":   windSpeed,
		"WindDir":     windDir,
		"Humidity":    humidity,
		"UVIndex":     uvIndex,
	}

	var buf strings.Builder
	if err := partialTemplate.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("rendering weather partial: %w", err)
	}

	return &widget.Data{
		WidgetName: "weather",
		HTML:       template.HTML(buf.String()),
		FetchedAt:  time.Now(),
		CacheTTL:   w.config.RefreshInterval.Duration,
	}, nil
}
