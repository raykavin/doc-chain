package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	apiURL  string
)

// NewRootCmd creates the root command for the CLI
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "blockchain-cli",
		Short: "Doc Chain CLI",
		Long: `A command-line interface for interacting with the Doc Chain system.
This CLI allows you to create wallets, sign documents, mine blocks, and more.`,
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.blockchain-cli.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiURL, "api-url", "http://localhost:8080/api", "blockchain API URL")
	
	// Bind flags to viper
	viper.BindPFlag("api-url", rootCmd.PersistentFlags().Lookup("api-url"))

	// Add subcommands
	rootCmd.AddCommand(newWalletCmd())
	rootCmd.AddCommand(newDocumentCmd())
	rootCmd.AddCommand(newBlockchainCmd())

	cobra.OnInitialize(initConfig)

	return rootCmd
}

// initConfig reads in config file and ENV variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".blockchain-cli" (without extension)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".blockchain-cli")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}