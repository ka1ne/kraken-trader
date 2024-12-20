package kraken

import (
	"testing"
)

func TestOrderRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     OrderRequest
		wantErr bool
	}{
		{
			name: "valid limit order",
			req: OrderRequest{
				Pair:   "XBTUSD",
				Type:   LimitOrder,
				Side:   "buy",
				Volume: "1.0",
				Price:  "50000",
			},
			wantErr: false,
		},
		{
			name: "limit order without price",
			req: OrderRequest{
				Pair:   "XBTUSD",
				Type:   LimitOrder,
				Side:   "buy",
				Volume: "1.0",
			},
			wantErr: true,
		},
		{
			name: "invalid side",
			req: OrderRequest{
				Pair:   "XBTUSD",
				Type:   LimitOrder,
				Side:   "invalid",
				Volume: "1.0",
				Price:  "50000",
			},
			wantErr: true,
		},
		{
			name: "missing volume",
			req: OrderRequest{
				Pair:  "XBTUSD",
				Type:  LimitOrder,
				Side:  "buy",
				Price: "50000",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
