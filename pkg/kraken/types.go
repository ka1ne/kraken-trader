package kraken

import (
	"fmt"
)

// REST API types
type OrderRequest struct {
	Pair      string
	Type      OrderType // Using our new OrderType
	Side      string
	Volume    string
	Price     string
	Leverage  string
	StopPrice string `json:"stop_price,omitempty"` // For stop orders
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

const (
	LimitOrder      OrderType = "limit"
	MarketOrder     OrderType = "market"
	StopLossOrder   OrderType = "stop-loss"
	TakeProfitOrder OrderType = "take-profit"
)

// Add validation
func (r *OrderRequest) Validate() error {
	switch r.Type {
	case LimitOrder:
		if r.Price == "" {
			return fmt.Errorf("price is required for limit orders")
		}
	case StopLossOrder, TakeProfitOrder:
		if r.StopPrice == "" {
			return fmt.Errorf("stop price is required for stop orders")
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

	return nil
}
