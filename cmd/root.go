package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ipsec-vpn",
	Short: "A production-grade IPsec VPN with post-quantum encryption",
	Long: `IPsec VPN is a secure, production-grade VPN solution with post-quantum
cryptography capabilities. It provides secure tunneling for network traffic
with advanced encryption options including post-quantum algorithms.

Features:
- IPsec tunnel and transport modes
- Post-quantum encryption algorithms
- Network advertisement capabilities
- Cisco-like CLI configuration interface
- Comprehensive logging and monitoring`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.ipsec-vpn.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")

	// Add commands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(tunnelCmd)
	rootCmd.AddCommand(cryptoCmd)
	rootCmd.AddCommand(networkCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".ipsec-vpn" (without extension).
		viper.AddConfigPath(home)
		viper.AddConfigPath(filepath.Join("/etc", "ipsec-vpn"))
		viper.AddConfigPath(".") // also look in the working directory
		viper.SetConfigType("yaml")
		viper.SetConfigName(".ipsec-vpn")
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}

	// Environment variables can override config file settings
	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvPrefix("IPSEC") // will be uppercased automatically
}