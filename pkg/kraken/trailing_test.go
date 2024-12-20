package kraken

import (
	"context"
	"testing"
	"time"
)

func TestExecuteTrailingEntry(t *testing.T) {
	t.Parallel()

	// Create test server and client
	server := newMockWSServer()
	defer server.Close()

	client := NewTestClient(t, &TestConfig{
		WebSocketURL: server.URL,
		TestPair:     "XBT/USD",
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Connect to websocket
	if err := client.ConnectWebSocket(ctx); err != nil {
		t.Fatalf("ConnectWebSocket() error = %v", err)
	}

	// Setup price channel and subscribe
	priceChan := make(chan float64)
	defer close(priceChan)

	if err := client.SubscribeToTicker(ctx, "XBT/USD", priceChan); err != nil {
		t.Fatalf("SubscribeToTicker() error = %v", err)
	}

	// Test parameters
	params := TrailingEntryConfig{
		Pair:         "XBT/USD",
		Side:         "buy",
		UpperBand:    1100.0,
		LowerBand:    900.0,
		TotalVolume:  1.0,
		Distribution: "even",
		NumOrders:    3,
		Interval:     time.Second,
	}

	// Create done channel to signal test completion
	done := make(chan struct{})
	defer close(done)

	// Run test with timeout
	go func() {
		err := client.ExecuteTrailingEntry(ctx, params)
		if err != nil {
			t.Errorf("ExecuteTrailingEntry() error = %v", err)
		}
		done <- struct{}{}
	}()

	// Wait for test completion or timeout
	select {
	case <-done:
		// Test completed successfully
	case <-time.After(8 * time.Second):
		t.Fatal("Test timed out")
	}
}
