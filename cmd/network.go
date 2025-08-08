package cmd

import (
	"fmt"

	"github.com/dzakwan/ipsec-vpn/pkg/network"
	"github.com/spf13/cobra"
)

// networkCmd represents the network command
var networkCmd = &cobra.Command{
	Use:   "network",
	Short: "Manage network settings",
	Long:  `Configure and manage network settings, including route advertisement and interface configuration.`,
}

var networkShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show network configuration",
	Run: func(cmd *cobra.Command, args []string) {
		interfaces, _ := cmd.Flags().GetBool("interfaces")
		routes, _ := cmd.Flags().GetBool("routes")
		advertised, _ := cmd.Flags().GetBool("advertised")

		// If no flags specified, show everything
		if !interfaces && !routes && !advertised {
			interfaces = true
			routes = true
			advertised = true
		}

		if interfaces {
			fmt.Println("Network Interfaces:")
			netIfaces, err := network.ListInterfaces()
			if err != nil {
				fmt.Printf("Error listing interfaces: %v\n", err)
			} else {
				for _, iface := range netIfaces {
					fmt.Printf("- %s: %s\n", iface.Name, iface.Status)
					fmt.Printf("  MAC: %s\n", iface.MAC)
					fmt.Printf("  IP Addresses: %v\n", iface.IPAddresses)
					fmt.Printf("  MTU: %d\n", iface.MTU)
					fmt.Println()
				}
			}
		}

		if routes {
			fmt.Println("Routing Table:")
			routes, err := network.ListRoutes()
			if err != nil {
				fmt.Printf("Error listing routes: %v\n", err)
			} else {
				for _, route := range routes {
					fmt.Printf("- Destination: %s\n", route.Destination)
					fmt.Printf("  Gateway: %s\n", route.Gateway)
					fmt.Printf("  Interface: %s\n", route.Interface)
					fmt.Printf("  Metric: %d\n", route.Metric)
					fmt.Println()
				}
			}
		}

		if advertised {
			fmt.Println("Advertised Networks:")
			advNetworks, err := network.ListAdvertisedNetworks()
			if err != nil {
				fmt.Printf("Error listing advertised networks: %v\n", err)
			} else {
				for _, net := range advNetworks {
					fmt.Printf("- Network: %s\n", net.CIDR)
					fmt.Printf("  Advertised via: %s\n", net.AdvertisedVia)
					fmt.Printf("  Status: %s\n", net.Status)
					fmt.Println()
				}
			}
		}
	},
}

var networkAdvertiseCmd = &cobra.Command{
	Use:   "advertise [network]",
	Short: "Advertise a network",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		networkCIDR := args[0]
		tunnelName, _ := cmd.Flags().GetString("tunnel")
		metric, _ := cmd.Flags().GetInt("metric")

		err := network.AdvertiseNetwork(networkCIDR, tunnelName, metric)
		if err != nil {
			fmt.Printf("Error advertising network: %v\n", err)
			return
		}

		fmt.Printf("Network %s is now being advertised via tunnel %s\n", 
			networkCIDR, tunnelName)
	},
}

var networkWithdrawCmd = &cobra.Command{
	Use:   "withdraw [network]",
	Short: "Withdraw a network advertisement",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		networkCIDR := args[0]
		tunnelName, _ := cmd.Flags().GetString("tunnel")

		err := network.WithdrawNetwork(networkCIDR, tunnelName)
		if err != nil {
			fmt.Printf("Error withdrawing network: %v\n", err)
			return
		}

		fmt.Printf("Network %s advertisement withdrawn from tunnel %s\n", 
			networkCIDR, tunnelName)
	},
}

var networkRouteCmd = &cobra.Command{
	Use:   "route [add|delete] [destination] [gateway]",
	Short: "Manage routes",
	Args:  cobra.ExactArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		action := args[0]
		destination := args[1]
		gateway := args[2]
		iface, _ := cmd.Flags().GetString("interface")
		metric, _ := cmd.Flags().GetInt("metric")

		switch action {
		case "add":
			err := network.AddRoute(destination, gateway, iface, metric)
			if err != nil {
				fmt.Printf("Error adding route: %v\n", err)
				return
			}
			fmt.Printf("Route to %s via %s added successfully\n", destination, gateway)

		case "delete":
			err := network.DeleteRoute(destination, gateway, iface)
			if err != nil {
				fmt.Printf("Error deleting route: %v\n", err)
				return
			}
			fmt.Printf("Route to %s via %s deleted successfully\n", destination, gateway)

		default:
			fmt.Printf("Unknown action: %s. Use 'add' or 'delete'\n", action)
		}
	},
}

func init() {
	// Add subcommands to network command
	networkCmd.AddCommand(networkShowCmd)
	networkCmd.AddCommand(networkAdvertiseCmd)
	networkCmd.AddCommand(networkWithdrawCmd)
	networkCmd.AddCommand(networkRouteCmd)

	// Flags for show command
	networkShowCmd.Flags().Bool("interfaces", false, "Show network interfaces")
	networkShowCmd.Flags().Bool("routes", false, "Show routing table")
	networkShowCmd.Flags().Bool("advertised", false, "Show advertised networks")

	// Flags for advertise command
	networkAdvertiseCmd.Flags().String("tunnel", "", "Tunnel to advertise the network through")
	networkAdvertiseCmd.Flags().Int("metric", 100, "Metric for the advertised route")
	networkAdvertiseCmd.MarkFlagRequired("tunnel")

	// Flags for withdraw command
	networkWithdrawCmd.Flags().String("tunnel", "", "Tunnel to withdraw the network from")
	networkWithdrawCmd.MarkFlagRequired("tunnel")

	// Flags for route command
	networkRouteCmd.Flags().String("interface", "", "Network interface for the route")
	networkRouteCmd.Flags().Int("metric", 100, "Metric for the route")
}