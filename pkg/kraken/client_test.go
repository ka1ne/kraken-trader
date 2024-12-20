package kraken

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/websocket"
)

// Mock WebSocket server
type mockWSServer struct {
	*httptest.Server
	URL string
}

func newMockWSServer() *mockWSServer {
	var upgrader = websocket.Upgrader{}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer c.Close()

		// Echo messages back
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			c.WriteMessage(mt, message)
		}
	}))

	return &mockWSServer{
		Server: s,
		URL:    "ws" + strings.TrimPrefix(s.URL, "http"),
	}
}

func TestClient_AddOrder(t *testing.T) {
	// Mock HTTP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify API key
		if r.Header.Get("API-Key") != "test" {
			w.Write([]byte(`{"error":["EAPI:Invalid key"]}`))
			return
		}
		if r.URL.Path != "/0/private/AddOrder" {
			t.Errorf("Expected to request '/0/private/AddOrder', got: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got: %s", r.Method)
		}

		response := `{
			"error": [],
			"result": {
				"descr": {"order": "buy 1.00000000 XBTUSD @ limit 50000"},
				"txid": ["ABCD-1234"]
			}
		}`
		w.Write([]byte(response))
	}))
	defer server.Close()

	// Create client with mock server URL
	client := &Client{
		apiKey:    "test",
		apiSecret: "test",
		httpClient: &http.Client{
			Timeout: REST_TIMEOUT,
		},
		apiURL: server.URL,
	}

	req := OrderRequest{
		Pair:   "XBTUSD",
		Type:   LimitOrder,
		Side:   "buy",
		Volume: "1.0",
		Price:  "50000",
	}

	resp, err := client.AddOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("AddOrder() error = %v", err)
	}

	if len(resp.TransactionIds) != 1 || resp.TransactionIds[0] != "ABCD-1234" {
		t.Errorf("Expected txid ABCD-1234, got %v", resp.TransactionIds)
	}
}

func TestClient_WebSocketConnection(t *testing.T) {
	ws := newMockWSServer()
	defer ws.Close()

	client := NewClient("test", "test")
	var err error
	client.ws, _, err = websocket.DefaultDialer.Dial(ws.URL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to test server: %v", err)
	}
	ctx := context.Background()

	// Test subscription
	priceChan := make(chan float64)
	if err := client.SubscribeToTicker(ctx, "XBTUSD", priceChan); err != nil {
		t.Fatalf("SubscribeToTicker() error = %v", err)
	}

	// Test cleanup
	if err := client.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestClient_ParseTickerMessage(t *testing.T) {
	// Sample ticker message from Kraken WebSocket
	message := `[
		{
			"c": ["50000.0", "1.0"]
		},
		"ticker",
		"XBT/USD"
	]`

	var tickerData []interface{}
	if err := json.Unmarshal([]byte(message), &tickerData); err != nil {
		t.Fatalf("Failed to unmarshal test message: %v", err)
	}

	if len(tickerData) < 3 || tickerData[1] != "ticker" {
		t.Errorf("Invalid ticker message format")
	}

	data := tickerData[0].(map[string]interface{})
	price, ok := data["c"].([]interface{})
	if !ok || len(price) == 0 {
		t.Errorf("Failed to extract price from ticker message")
	}

	lastPrice, ok := price[0].(string)
	if !ok {
		t.Errorf("Price is not a string")
	}

	p, err := strconv.ParseFloat(lastPrice, 64)
	if err != nil {
		t.Errorf("Failed to parse price: %v", err)
	}

	if p != 50000.0 {
		t.Errorf("Expected price 50000.0, got %v", p)
	}
}

func TestCalculateOrderVolumes(t *testing.T) {
	tests := []struct {
		name   string
		config TrailingEntryConfig
		want   []float64
	}{
		{
			name: "even distribution",
			config: TrailingEntryConfig{
				TotalVolume:  1.0,
				NumOrders:    4,
				Distribution: EvenDistribution,
			},
			want: []float64{0.25, 0.25, 0.25, 0.25},
		},
		{
			name: "normal distribution",
			config: TrailingEntryConfig{
				TotalVolume:  1.0,
				NumOrders:    3,
				Distribution: NormalDistribution,
			},
			want: []float64{0.25, 0.5, 0.25}, // Approximate values
		},
		{
			name: "custom distribution",
			config: TrailingEntryConfig{
				TotalVolume:  1.0,
				NumOrders:    3,
				Distribution: CustomDistribution,
				Weights:      []float64{1, 2, 1},
			},
			want: []float64{0.25, 0.5, 0.25},
		},
		{
			name: "invalid custom weights falls back to even",
			config: TrailingEntryConfig{
				TotalVolume:  1.0,
				NumOrders:    3,
				Distribution: CustomDistribution,
				Weights:      []float64{1, 2}, // Wrong length
			},
			want: []float64{0.333, 0.333, 0.333}, // Approximate values
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateOrderVolumes(tt.config)
			if len(got) != len(tt.want) {
				t.Errorf("calculateOrderVolumes() len = %v, want %v", len(got), len(tt.want))
				return
			}

			// Check total volume is correct
			total := 0.0
			for _, v := range got {
				total += v
			}
			if math.Abs(total-tt.config.TotalVolume) > 0.001 {
				t.Errorf("total volume = %v, want %v", total, tt.config.TotalVolume)
			}

			// Check individual volumes (with tolerance for floating point)
			for i := range got {
				if math.Abs(got[i]-tt.want[i]) > 0.001 {
					t.Errorf("volume[%d] = %v, want %v", i, got[i], tt.want[i])
				}
			}
		})
	}
}
