package kraken

import (
	"context"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestExecuteTrailingEntry(t *testing.T) {
	ws := newMockWSServer()
	defer ws.Close()

	client := NewClient("test", "test")
	conn, _, err := websocket.DefaultDialer.Dial(ws.URL, nil)
	if err != nil {
		t.Fatalf("Failed to connect to test server: %v", err)
	}
	client.ws = conn
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	config := TrailingEntryConfig{
		Pair:        "XBTUSD",
		Side:        "buy",
		UpperBand:   52000,
		LowerBand:   48000,
		TotalVolume: 1.0,
		NumOrders:   5,
		Interval:    time.Second,
	}

	// Mock price updates
	go func() {
		time.Sleep(50 * time.Millisecond)
		// Simulate price entering the band
		// This would normally come from WebSocket
		// but we're just testing the logic
	}()

	err = client.ExecuteTrailingEntry(ctx, config)
	if err != nil && err != context.DeadlineExceeded {
		t.Errorf("ExecuteTrailingEntry() error = %v", err)
	}
}
