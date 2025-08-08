package cmd

import (
	"fmt"

	"github.com/dzakwan/ipsec-vpn/pkg/logger"
	"github.com/dzakwan/ipsec-vpn/pkg/tunnel"
	"github.com/spf13/cobra"
)

// tunnelCmd represents the tunnel command
var tunnelCmd = &cobra.Command{
	Use:   "tunnel",
	Short: "Manage IPsec tunnels",
	Long:  `Create, configure, and manage IPsec tunnels with various encryption options.`,
}

var tunnelCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new IPsec tunnel",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		localIP, _ := cmd.Flags().GetString("local-ip")
		remoteIP, _ := cmd.Flags().GetString("remote-ip")
		localSubnet, _ := cmd.Flags().GetString("local-subnet")
		remoteSubnet, _ := cmd.Flags().GetString("remote-subnet")
		encryption, _ := cmd.Flags().GetString("encryption")
		pqEnabled, _ := cmd.Flags().GetBool("post-quantum")

		// Create tunnel configuration
		config := tunnel.Config{
			Name:          name,
			LocalIP:       localIP,
			RemoteIP:      remoteIP,
			LocalSubnet:   localSubnet,
			RemoteSubnet:  remoteSubnet,
			Encryption:    encryption,
			PostQuantum:   pqEnabled,
		}

		// Create and start the tunnel
		logger.Info("Creating tunnel '%s' with local IP %s and remote IP %s", name, localIP, remoteIP)
		tun, err := tunnel.Create(config)
		if err != nil {
			logger.Error("Error creating tunnel: %v", err)
			fmt.Printf("Error creating tunnel: %v\n", err)
			return
		}

		logger.Info("Tunnel '%s' created successfully", tun.Name)
		fmt.Printf("Tunnel '%s' created successfully\n", tun.Name)
		fmt.Printf("Local IP: %s, Remote IP: %s\n", tun.LocalIP, tun.RemoteIP)
		fmt.Printf("Local Subnet: %s, Remote Subnet: %s\n", tun.LocalSubnet, tun.RemoteSubnet)
		fmt.Printf("Encryption: %s, Post-Quantum: %v\n", tun.Encryption, tun.PostQuantum)
	},
}

var tunnelShowCmd = &cobra.Command{
	Use:   "show [name]",
	Short: "Show tunnel details",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			// List all tunnels
			logger.Debug("Listing all configured tunnels")
			tunnels, err := tunnel.ListAll()
			if err != nil {
				logger.Error("Error listing tunnels: %v", err)
				fmt.Printf("Error listing tunnels: %v\n", err)
				return
			}

			if len(tunnels) == 0 {
				logger.Info("No tunnels configured")
				fmt.Println("No tunnels configured")
				return
			}

			logger.Info("Found %d configured tunnels", len(tunnels))
			fmt.Println("Configured tunnels:")
			for _, t := range tunnels {
				fmt.Printf("- %s: %s <-> %s (%s)\n", t.Name, t.LocalIP, t.RemoteIP, t.Status)
			}
		} else {
			// Show specific tunnel
			name := args[0]
			logger.Debug("Retrieving details for tunnel '%s'", name)
			tun, err := tunnel.Get(name)
			if err != nil {
				logger.Error("Error getting tunnel '%s': %v", name, err)
				fmt.Printf("Error getting tunnel '%s': %v\n", name, err)
				return
			}

			logger.Info("Displaying details for tunnel '%s'", tun.Name)
			fmt.Printf("Tunnel: %s\n", tun.Name)
			fmt.Printf("Status: %s\n", tun.Status)
			fmt.Printf("Local IP: %s\n", tun.LocalIP)
			fmt.Printf("Remote IP: %s\n", tun.RemoteIP)
			fmt.Printf("Local Subnet: %s\n", tun.LocalSubnet)
			fmt.Printf("Remote Subnet: %s\n", tun.RemoteSubnet)
			fmt.Printf("Encryption: %s\n", tun.Encryption)
			fmt.Printf("Post-Quantum: %v\n", tun.PostQuantum)
			fmt.Printf("Created: %s\n", tun.CreatedAt)
			fmt.Printf("Last Modified: %s\n", tun.UpdatedAt)
		}
	},
}

var tunnelDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete an IPsec tunnel",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		force, _ := cmd.Flags().GetBool("force")

		logger.Info("Deleting tunnel '%s' (force: %t)", name, force)
		err := tunnel.Delete(name, force)
		if err != nil {
			logger.Error("Error deleting tunnel '%s': %v", name, err)
			fmt.Printf("Error deleting tunnel '%s': %v\n", name, err)
			return
		}

		logger.Info("Tunnel '%s' deleted successfully", name)
		fmt.Printf("Tunnel '%s' deleted successfully\n", name)
	},
}

var tunnelStartCmd = &cobra.Command{
	Use:   "start [name]",
	Short: "Start an IPsec tunnel",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		logger.Info("Starting tunnel '%s'", name)
		err := tunnel.Start(name)
		if err != nil {
			logger.Error("Error starting tunnel '%s': %v", name, err)
			fmt.Printf("Error starting tunnel '%s': %v\n", name, err)
			return
		}

		logger.Info("Tunnel '%s' started successfully", name)
		fmt.Printf("Tunnel '%s' started successfully\n", name)
	},
}

var tunnelStopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop an IPsec tunnel",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]
		logger.Info("Stopping tunnel '%s'", name)
		err := tunnel.Stop(name)
		if err != nil {
			logger.Error("Error stopping tunnel '%s': %v", name, err)
			fmt.Printf("Error stopping tunnel '%s': %v\n", name, err)
			return
		}

		logger.Info("Tunnel '%s' stopped successfully", name)
		fmt.Printf("Tunnel '%s' stopped successfully\n", name)
	},
}

func init() {
	// Add subcommands to tunnel command
	tunnelCmd.AddCommand(tunnelCreateCmd)
	tunnelCmd.AddCommand(tunnelShowCmd)
	tunnelCmd.AddCommand(tunnelDeleteCmd)
	tunnelCmd.AddCommand(tunnelStartCmd)
	tunnelCmd.AddCommand(tunnelStopCmd)

	// Flags for create command
	tunnelCreateCmd.Flags().String("local-ip", "", "Local IP address for the tunnel")
	tunnelCreateCmd.Flags().String("remote-ip", "", "Remote IP address for the tunnel")
	tunnelCreateCmd.Flags().String("local-subnet", "", "Local subnet to be tunneled (CIDR notation)")
	tunnelCreateCmd.Flags().String("remote-subnet", "", "Remote subnet to be tunneled (CIDR notation)")
	tunnelCreateCmd.Flags().String("encryption", "aes256gcm", "Encryption algorithm (aes256gcm, chacha20poly1305)")
	tunnelCreateCmd.Flags().Bool("post-quantum", false, "Enable post-quantum cryptography")

	// Mark required flags
	tunnelCreateCmd.MarkFlagRequired("local-ip")
	tunnelCreateCmd.MarkFlagRequired("remote-ip")
	tunnelCreateCmd.MarkFlagRequired("local-subnet")
	tunnelCreateCmd.MarkFlagRequired("remote-subnet")

	// Flags for delete command
	tunnelDeleteCmd.Flags().Bool("force", false, "Force deletion even if tunnel is active")
}