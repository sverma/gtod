package weatherclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Forecast mirrors the weather service JSON response.
type Forecast struct {
	Timezone      string      `json:"timezone"`
	LocalDatetime string      `json:"local_datetime"`
	Location      string      `json:"location"`
	Conditions    Conditions  `json:"conditions"`
	Temperature   Temperature `json:"temperature"`
	HumidityPct   int         `json:"humidity_percent"`
	Wind          Wind        `json:"wind"`
	Provider      string      `json:"provider"`
	ObservedAt    string      `json:"observed_at"`
}

type Conditions struct {
	Summary     string `json:"summary"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
}

type Temperature struct {
	Celsius    float64 `json:"celsius"`
	Fahrenheit float64 `json:"fahrenheit"`
}

type Wind struct {
	SpeedKPH  float64 `json:"speed_kph"`
	Direction string  `json:"direction"`
}

// Client fetches weather from the weather microservice.
type Client interface {
	Enabled() bool
	Lookup(ctx context.Context, tz, at string) (Forecast, error)
}

type noopClient struct{}

func (noopClient) Enabled() bool { return false }

func (noopClient) Lookup(context.Context, string, string) (Forecast, error) {
	return Forecast{}, fmt.Errorf("weather client not configured")
}

// HTTPClient calls GET {baseURL}/weather?tz=&at= on the weather service.
type HTTPClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewFromEnv returns an HTTP client when WEATHER_SERVICE_URL is set, otherwise a noop client.
func NewFromEnv() Client {
	baseURL := strings.TrimSpace(os.Getenv("WEATHER_SERVICE_URL"))
	if baseURL == "" {
		return noopClient{}
	}
	return &HTTPClient{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *HTTPClient) Enabled() bool { return true }

func (c *HTTPClient) Lookup(ctx context.Context, tz, at string) (Forecast, error) {
	u, err := url.Parse(c.BaseURL + "/weather")
	if err != nil {
		return Forecast{}, fmt.Errorf("parse weather url: %w", err)
	}

	q := u.Query()
	q.Set("tz", tz)
	if at != "" {
		q.Set("at", at)
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Forecast{}, fmt.Errorf("create weather request: %w", err)
	}

	client := c.HTTPClient
	if client == nil {
		client = http.DefaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return Forecast{}, fmt.Errorf("call weather service: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Forecast{}, fmt.Errorf("read weather response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return Forecast{}, fmt.Errorf("weather service status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var forecast Forecast
	if err := json.Unmarshal(body, &forecast); err != nil {
		return Forecast{}, fmt.Errorf("decode weather response: %w", err)
	}
	return forecast, nil
}
