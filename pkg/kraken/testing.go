package kraken

import (
	"os"
	"testing"
	"time"
)

// TestConfig holds test configuration loaded from environment
type TestConfig struct {
	DemoAPIKey    string
	DemoAPISecret string
	DemoAPIURL    string
	DemoWSURL     string
	TestPair      string
	Timeouts      struct {
		WebSocket time.Duration
		REST      time.Duration
	}
}

// LoadTestConfig loads test configuration from environment variables
func LoadTestConfig(t *testing.T) *TestConfig {
	cfg := &TestConfig{
		DemoAPIKey:    os.Getenv("KRAKEN_DEMO_KEY"),
		DemoAPISecret: os.Getenv("KRAKEN_DEMO_SECRET"),
		DemoAPIURL:    "https://demo-futures.kraken.com/derivatives",
		DemoWSURL:     "wss://demo-futures.kraken.com/ws/v1",
		TestPair:      "PI_XBTUSD",
	}

	cfg.Timeouts.WebSocket = 10 * time.Second
	cfg.Timeouts.REST = 5 * time.Second

	return cfg
}

// NewTestClient creates a client configured for testing
func NewTestClient(t *testing.T, cfg *TestConfig) *Client {
	client := NewClient(cfg.DemoAPIKey, cfg.DemoAPISecret)
	client.apiURL = cfg.DemoAPIURL
	return client
}
