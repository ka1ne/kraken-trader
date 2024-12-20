package kraken

import (
	"fmt"
)

// REST API types
type OrderRequest struct {
	Pair       string
	Type       OrderType
	Side       string
	Volume     string
	Price      string
	Leverage   string `json:"leverage,omitempty"`
	OrderFlags string `json:"oflags,omitempty"`
}

type OrderResponse struct {
	Description struct {
		Order string `json:"order"`
		Close string `json:"close,omitempty"`
	} `json:"descr"`
	TransactionIds []string `json:"txid"`
}

// WebSocket API types
type WSOrderRequest struct {
	OrderType  string  `json:"order_type"`
	Side       string  `json:"side"`
	OrderQty   float64 `json:"order_qty"`
	Symbol     string  `json:"symbol"`
	LimitPrice float64 `json:"limit_price,omitempty"`
	Token      string  `json:"token"`
}

type OrderType string
type Leverage string

const (
	LimitOrder      OrderType = "limit"
	MarketOrder     OrderType = "market"
	StopLossOrder   OrderType = "stop-loss"
	TakeProfitOrder OrderType = "take-profit"

	// Leverage options
	NoLeverage Leverage = "none"
	Leverage2x Leverage = "2"
	Leverage3x Leverage = "3"
	Leverage4x Leverage = "4"
	Leverage5x Leverage = "5"
)

func IsValidLeverage(l string) bool {
	switch Leverage(l) {
	case NoLeverage, Leverage2x, Leverage3x, Leverage4x, Leverage5x:
		return true
	default:
		return false
	}
}

// Add validation
func (r *OrderRequest) Validate() error {
	switch r.Type {
	case LimitOrder:
		if r.Price == "" {
			return fmt.Errorf("price is required for limit orders")
		}
	case MarketOrder:
		// Market orders don't need price
	default:
		return fmt.Errorf("invalid order type: %s", r.Type)
	}

	if r.Side != "buy" && r.Side != "sell" {
		return fmt.Errorf("invalid side: must be buy or sell")
	}

	if r.Volume == "" {
		return fmt.Errorf("volume is required")
	}

	if r.Leverage != "" && !IsValidLeverage(r.Leverage) {
		return fmt.Errorf("invalid leverage: must be none, 2, 3, 4, or 5")
	}

	return nil
}

type TrailingEntryParams struct {
	Pair           string
	Side           string
	EntryPrice     float64
	TrailingAmount float64
	Volume         float64
	Distribution   string
	OrderCount     int
}
