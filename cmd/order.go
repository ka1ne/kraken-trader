package cmd

import (
	"context"
	"fmt"

	"github.com/ka1ne/kraken-trader/pkg/kraken"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	orderType string
	side      string
	pair      string
	volume    string
	price     string
	leverage  string
)

var orderCmd = &cobra.Command{
	Use:   "order",
	Short: "Place an order on Kraken",
	Long: `Place a new order on Kraken exchange with specified parameters.
Supports various order types including market, limit, and stop orders.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Validate required flags
		if side != "buy" && side != "sell" {
			return fmt.Errorf("side must be either 'buy' or 'sell'")
		}

		apiKey := viper.GetString("api.key")
		apiSecret := viper.GetString("api.secret")

		if apiKey == "" || apiSecret == "" {
			return fmt.Errorf("API key and secret are required. Set them via flags or config file")
		}

		client := kraken.NewClient(apiKey, apiSecret)

		req := kraken.OrderRequest{
			Pair:     pair,
			Type:     orderType,
			Side:     side,
			Volume:   volume,
			Price:    price,
			Leverage: leverage,
		}

		if err := client.PlaceOrder(context.Background(), req); err != nil {
			return fmt.Errorf("failed to place order: %w", err)
		}

		fmt.Printf("Successfully placed %s %s order for %s %s at %s\n",
			side, orderType, volume, pair, price)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(orderCmd)

	orderCmd.Flags().StringVar(&orderType, "type", "limit", "Order type (market, limit, stop-loss, etc.)")
	orderCmd.Flags().StringVar(&side, "side", "", "Order side (buy/sell)")
	orderCmd.Flags().StringVar(&pair, "pair", "", "Trading pair (e.g., XBTUSD)")
	orderCmd.Flags().StringVar(&volume, "volume", "", "Order volume")
	orderCmd.Flags().StringVar(&price, "price", "", "Order price")
	orderCmd.Flags().StringVar(&leverage, "leverage", "", "Leverage (optional)")

	orderCmd.MarkFlagRequired("side")
	orderCmd.MarkFlagRequired("pair")
	orderCmd.MarkFlagRequired("volume")
}
