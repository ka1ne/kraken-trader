package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/ka1ne/kraken-trader/pkg/kraken"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	upperBand    string
	lowerBand    string
	numOrders    int
	totalVolume  string
	timeInterval string
	distribution string
)

var trailingCmd = &cobra.Command{
	Use:   "trailing",
	Short: "Place orders across a price band",
	Long: `Place a series of orders spread across a specified price band.
This helps in averaging out entry prices for a position.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		interval, err := time.ParseDuration(timeInterval)
		if err != nil {
			return fmt.Errorf("invalid time interval: %w", err)
		}

		// Parse bands
		upper, err := strconv.ParseFloat(upperBand, 64)
		if err != nil {
			return fmt.Errorf("invalid upper band: %w", err)
		}

		lower, err := strconv.ParseFloat(lowerBand, 64)
		if err != nil {
			return fmt.Errorf("invalid lower band: %w", err)
		}

		if upper <= lower {
			return fmt.Errorf("upper band must be greater than lower band")
		}

		volume, err := strconv.ParseFloat(totalVolume, 64)
		if err != nil {
			return fmt.Errorf("invalid volume: %w", err)
		}

		apiKey := viper.GetString("api.key")
		apiSecret := viper.GetString("api.secret")

		if apiKey == "" || apiSecret == "" {
			return fmt.Errorf("API key and secret are required. Set them via flags or config file")
		}

		client := kraken.NewClient(apiKey, apiSecret)

		config := kraken.TrailingEntryConfig{
			Pair:        pair,
			Side:        side,
			UpperBand:   upper,
			LowerBand:   lower,
			TotalVolume: volume,
			NumOrders:   numOrders,
			Interval:    interval,
		}

		fmt.Printf("Starting band entry with %d orders between %s and %s, total volume %s\n",
			numOrders, lowerBand, upperBand, totalVolume)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		go func() {
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, os.Interrupt)
			<-sigChan
			fmt.Println("\nReceived interrupt signal. Cleaning up...")
			cancel()
		}()

		if err := client.ExecuteTrailingEntry(ctx, config); err != nil {
			return fmt.Errorf("band entry failed: %w", err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(trailingCmd)

	trailingCmd.Flags().StringVar(&upperBand, "upper", "", "Upper price band")
	trailingCmd.Flags().StringVar(&lowerBand, "lower", "", "Lower price band")
	trailingCmd.Flags().IntVar(&numOrders, "orders", 5, "Number of orders to place")
	trailingCmd.Flags().StringVar(&totalVolume, "volume", "", "Total volume to be split across orders")
	trailingCmd.Flags().StringVar(&timeInterval, "interval", "1m", "Check interval")
	trailingCmd.Flags().StringVar(&distribution, "distribution", "even",
		"Volume distribution (even, normal, custom)")

	trailingCmd.MarkFlagRequired("upper")
	trailingCmd.MarkFlagRequired("lower")
	trailingCmd.MarkFlagRequired("volume")
	trailingCmd.MarkFlagRequired("side")
	trailingCmd.MarkFlagRequired("pair")
}
