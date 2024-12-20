package cmd

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	apiKey  string
	apiSec  string
	side    string
	pair    string
)

var rootCmd = &cobra.Command{
	Use:   "kraken-trader",
	Short: "A CLI tool for trading on Kraken exchange",
	Long: `kraken-trader is a CLI application that helps with trading on Kraken exchange.
It provides various trading utilities including order placement and trailing entries.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kraken-trader.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "Kraken API Key")
	rootCmd.PersistentFlags().StringVar(&apiSec, "api-secret", "", "Kraken API Secret")

	viper.BindPFlag("api.key", rootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("api.secret", rootCmd.PersistentFlags().Lookup("api-secret"))

	// Map environment variables
	viper.SetEnvPrefix("KRAKEN")
	viper.BindEnv("api.key", "KRAKEN_API_KEY")
	viper.BindEnv("api.secret", "KRAKEN_API_SECRET")

	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found")
	}

	rootCmd.AddCommand(orderCmd)
	rootCmd.AddCommand(trailingCmd)
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".kraken-trader")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		// fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
