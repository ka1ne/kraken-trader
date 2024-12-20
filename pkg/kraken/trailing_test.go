package kraken

import (
	"context"
	"testing"
	"time"
)

func TestExecuteTrailingEntry(t *testing.T) {
	client := NewClient("test", "test")
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

	err := client.ExecuteTrailingEntry(ctx, config)
	if err != nil && err != context.DeadlineExceeded {
		t.Errorf("ExecuteTrailingEntry() error = %v", err)
	}
}
