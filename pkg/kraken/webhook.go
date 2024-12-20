package kraken

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type TradingViewAlert struct {
	Strategy  string  `json:"strategy"`
	Action    string  `json:"action"` // "buy" or "sell"
	Pair      string  `json:"pair"`
	Price     float64 `json:"price"`
	Volume    float64 `json:"volume"`
	OrderType string  `json:"orderType"` // "limit" or "market"
	StopPrice float64 `json:"stopPrice,omitempty"`
}

func WebhookHandler(client *Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var alert TradingViewAlert
		if err := json.NewDecoder(r.Body).Decode(&alert); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Create order request
		order := OrderRequest{
			Pair:   alert.Pair,
			Type:   OrderType(alert.OrderType),
			Side:   alert.Action,
			Volume: fmt.Sprintf("%.8f", alert.Volume),
		}

		if alert.OrderType == "limit" {
			order.Price = fmt.Sprintf("%.2f", alert.Price)
		}

		// Place the order
		_, err := client.AddOrder(r.Context(), order)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
