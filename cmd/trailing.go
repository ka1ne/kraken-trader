package cmd

import (
	"context"
	"fmt"

	"github.com/ka1ne/kraken-trader/pkg/kraken"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	upper        float64
	lower        float64
	orders       int
	distribution string
	tradeVolume  float64
)

var trailingCmd = &cobra.Command{
	Use:   "trailing",
	Short: "Execute a trailing entry strategy",
	RunE: func(cmd *cobra.Command, args []string) error {
		if !kraken.IsValidLeverage(leverage) {
			return fmt.Errorf("invalid leverage: must be none, 2, 3, 4, or 5")
		}

		config := kraken.TrailingEntryConfig{
			Pair:         pair,
			Side:         side,
			UpperBand:    upper,
			LowerBand:    lower,
			TotalVolume:  tradeVolume,
			NumOrders:    orders,
			Distribution: kraken.VolumeDistribution(distribution),
			Leverage:     leverage,
		}

		client := kraken.NewClient(
			viper.GetString("api.key"),
			viper.GetString("api.secret"),
		)

		return client.ExecuteTrailingEntry(context.Background(), config)
	},
}

func init() {
	rootCmd.AddCommand(trailingCmd)

	trailingCmd.Flags().StringVar(&pair, "pair", "", "Trading pair (e.g., ETH/USD)")
	trailingCmd.Flags().StringVar(&side, "side", "", "Order side (buy/sell)")
	trailingCmd.Flags().Float64Var(&upper, "upper", 0, "Upper price band")
	trailingCmd.Flags().Float64Var(&lower, "lower", 0, "Lower price band")
	trailingCmd.Flags().Float64Var(&tradeVolume, "volume", 0, "Total volume to trade")
	trailingCmd.Flags().IntVar(&orders, "orders", 5, "Number of orders to place")
	trailingCmd.Flags().StringVar(&distribution, "distribution", "even", "Volume distribution (even, normal)")
	trailingCmd.Flags().StringVar(&leverage, "leverage", "none", "Leverage (none, 2, 3, 4, 5)")

	trailingCmd.MarkFlagRequired("pair")
	trailingCmd.MarkFlagRequired("side")
	trailingCmd.MarkFlagRequired("upper")
	trailingCmd.MarkFlagRequired("lower")
	trailingCmd.MarkFlagRequired("volume")
}
