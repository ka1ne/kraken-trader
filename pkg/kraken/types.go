package kraken

// REST API types
type OrderRequest struct {
	Pair     string
	Type     string // market, limit, etc.
	Side     string // buy or sell
	Volume   string
	Price    string
	Leverage string
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
