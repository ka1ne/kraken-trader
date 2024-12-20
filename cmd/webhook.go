package cmd

import (
	"fmt"
	"log"
	"net/http"

	"github.com/ka1ne/kraken-trader/pkg/kraken"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	port int
)

var webhookCmd = &cobra.Command{
	Use:   "webhook",
	Short: "Start webhook server for TradingView alerts",
	Run: func(cmd *cobra.Command, args []string) {
		client := kraken.NewClient(
			viper.GetString("api.key"),
			viper.GetString("api.secret"),
		)

		http.HandleFunc("/webhook", kraken.WebhookHandler(client))

		addr := fmt.Sprintf(":%d", port)
		log.Printf("Starting webhook server on %s", addr)
		if err := http.ListenAndServe(addr, nil); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	},
}

func init() {
	webhookCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run webhook server on")
	rootCmd.AddCommand(webhookCmd)
}
