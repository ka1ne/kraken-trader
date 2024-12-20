package kraken

import (
	"context"
	"testing"
	"time"
)

var testConfig *TestConfig

func TestMain(m *testing.M) {
	// Setup test environment
	if testing.Short() {
		return
	}
	testConfig = LoadTestConfig(nil)
	m.Run()
}

func skipIfNoAPIKeys(t *testing.T) {
	if testConfig.DemoAPIKey == "" || testConfig.DemoAPISecret == "" {
		t.Skip("Skipping integration test: API credentials required")
	}
}

func TestIntegration_AddOrder(t *testing.T) {
	skipIfNoAPIKeys(t)

	tests := []struct {
		name    string
		req     OrderRequest
		wantErr bool
	}{
		{
			name: "limit buy order",
			req: OrderRequest{
				Pair:   testConfig.TestPair,
				Type:   LimitOrder,
				Side:   "buy",
				Volume: "1.0",
				Price:  "20000.0",
			},
			wantErr: false,
		},
		{
			name: "market sell order",
			req: OrderRequest{
				Pair:   testConfig.TestPair,
				Type:   MarketOrder,
				Side:   "sell",
				Volume: "0.1",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), testConfig.Timeouts.REST)
			defer cancel()

			client := NewTestClient(t, testConfig)
			resp, err := client.AddOrder(ctx, tt.req)

			if (err != nil) != tt.wantErr {
				t.Errorf("AddOrder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && len(resp.TransactionIds) == 0 {
				t.Error("Expected transaction ID")
			}
			t.Logf("Order placed: %v", resp.Description.Order)
		})
	}
}

func TestIntegration_WebSocketReconnect(t *testing.T) {
	skipIfNoAPIKeys(t)

	client := NewTestClient(t, testConfig)
	ctx, cancel := context.WithTimeout(context.Background(), testConfig.Timeouts.WebSocket)
	defer cancel()

	// Test connection and reconnection
	if err := client.ConnectWebSocket(ctx); err != nil {
		t.Fatalf("Initial connection failed: %v", err)
	}

	// Force disconnect
	if err := client.ws.Close(); err != nil {
		t.Logf("Close error (expected): %v", err)
	}

	// Wait for reconnect
	time.Sleep(2 * time.Second)

	// Verify reconnected state
	if client.getState() != Connected {
		t.Error("Expected reconnected state")
	}
}

func TestIntegration_TickerSubscription(t *testing.T) {
	skipIfNoAPIKeys(t)

	client := NewTestClient(t, testConfig)
	ctx, cancel := context.WithTimeout(context.Background(), testConfig.Timeouts.WebSocket)
	defer cancel()

	if err := client.ConnectWebSocket(ctx); err != nil {
		t.Fatalf("ConnectWebSocket() error = %v", err)
	}
	defer client.Close()

	priceChan := make(chan float64)
	if err := client.SubscribeToTicker(ctx, testConfig.TestPair, priceChan); err != nil {
		t.Fatalf("SubscribeToTicker() error = %v", err)
	}

	// Collect some price updates
	prices := make([]float64, 0)
	timeout := time.After(5 * time.Second)

	for i := 0; i < 3; {
		select {
		case price := <-priceChan:
			prices = append(prices, price)
			i++
			t.Logf("Price update %d: %v", i, price)
		case <-timeout:
			t.Fatal("Timeout waiting for price updates")
		}
	}

	if len(prices) < 3 {
		t.Errorf("Expected at least 3 price updates, got %d", len(prices))
	}
}
