package cmd

import (
	"fmt"

	"github.com/dzakwan/ipsec-vpn/pkg/logger"
	"github.com/spf13/cobra"
)

// Version information
var (
	Version   = "0.1.0"
	Commit    = "none"
	BuildDate = "unknown"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Long:  `Display the version, commit, and build date information for the IPsec VPN application.`,
	Run: func(cmd *cobra.Command, args []string) {
		logger.Info("Displaying version information: v%s (commit: %s, built: %s)", Version, Commit, BuildDate)
		fmt.Printf("IPsec VPN v%s (commit: %s, built: %s)\n", Version, Commit, BuildDate)
	},
}