package network

import (
	"errors"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/vishvananda/netlink"
)

// Interface represents a network interface
type Interface struct {
	Name        string
	MAC         string
	IPAddresses []string
	MTU         int
	Status      string
}

// Route represents a routing table entry
type Route struct {
	Destination string
	Gateway     string
	Interface   string
	Metric      int
}

// AdvertisedNetwork represents a network that is being advertised
type AdvertisedNetwork struct {
	CIDR          string
	AdvertisedVia string
	Status        string
}

// ListInterfaces returns a list of network interfaces
func ListInterfaces() ([]Interface, error) {
	// Get all network interfaces
	links, err := netlink.LinkList()
	if err != nil {
		return nil, fmt.Errorf("failed to list interfaces: %v", err)
	}

	// Convert to Interface objects
	interfaces := make([]Interface, 0, len(links))
	for _, link := range links {
		attrs := link.Attrs()

		// Get IP addresses
		addrs, err := netlink.AddrList(link, 0) // 0 means all families (AF_UNSPEC)
		if err != nil {
			continue
		}

		ipAddresses := make([]string, 0, len(addrs))
		for _, addr := range addrs {
			ipAddresses = append(ipAddresses, addr.IPNet.String())
		}

		// Determine status
		status := "DOWN"
		if attrs.Flags&net.FlagUp != 0 {
			status = "UP"
		}

		// Create Interface object
		iface := Interface{
			Name:        attrs.Name,
			MAC:         attrs.HardwareAddr.String(),
			IPAddresses: ipAddresses,
			MTU:         attrs.MTU,
			Status:      status,
		}

		interfaces = append(interfaces, iface)
	}

	return interfaces, nil
}

// ListRoutes returns the routing table
func ListRoutes() ([]Route, error) {
	// Get all routes
	netlinkRoutes, err := netlink.RouteList(nil, 0) // 0 means all families (AF_UNSPEC)
	if err != nil {
		return nil, fmt.Errorf("failed to list routes: %v", err)
	}

	// Convert to Route objects
	routes := make([]Route, 0, len(netlinkRoutes))
	for _, nlRoute := range netlinkRoutes {
		// Skip routes without a destination
		if nlRoute.Dst == nil {
			continue
		}

		// Get interface name
		link, err := netlink.LinkByIndex(nlRoute.LinkIndex)
		if err != nil {
			continue
		}

		// Create Route object
		route := Route{
			Destination: nlRoute.Dst.String(),
			Gateway:     nlRoute.Gw.String(),
			Interface:   link.Attrs().Name,
			Metric:      nlRoute.Priority,
		}

		routes = append(routes, route)
	}

	return routes, nil
}

// ListAdvertisedNetworks returns a list of networks being advertised
func ListAdvertisedNetworks() ([]AdvertisedNetwork, error) {
	// In a real implementation, this would query the IPsec daemon or routing protocol
	// For this example, we'll return a placeholder implementation
	
	// Get tunnels that might be advertising networks
	cmd := exec.Command("ip", "tunnel", "show")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to list tunnels: %v", err)
	}

	// Parse output to find tunnels
	lines := strings.Split(string(output), "\n")
	advertisedNetworks := make([]AdvertisedNetwork, 0)

	for _, line := range lines {
		if strings.Contains(line, "ipsec") {
			// This is a simplified example - in a real implementation,
			// we would query the actual advertised networks for each tunnel
			parts := strings.Fields(line)
			if len(parts) > 0 {
				tunnelName := parts[0]
				
				// For demonstration, we'll assume each tunnel advertises a network
				advertisedNetworks = append(advertisedNetworks, AdvertisedNetwork{
					CIDR:          "192.168.0.0/24", // Example network
					AdvertisedVia: tunnelName,
					Status:        "ACTIVE",
				})
			}
		}
	}

	return advertisedNetworks, nil
}

// AdvertiseNetwork advertises a network through a tunnel
func AdvertiseNetwork(networkCIDR, tunnelName string, metric int) error {
	// Validate network CIDR
	_, _, err := net.ParseCIDR(networkCIDR)
	if err != nil {
		return fmt.Errorf("invalid network CIDR: %v", err)
	}

	// Check if tunnel exists
	link, err := netlink.LinkByName(tunnelName)
	if err != nil {
		return fmt.Errorf("tunnel not found: %v", err)
	}

	// In a real implementation, this would configure the IPsec daemon or routing protocol
	// to advertise the network through the tunnel
	
	// For this example, we'll add a route for the network through the tunnel
	_, dst, err := net.ParseCIDR(networkCIDR)
	if err != nil {
		return fmt.Errorf("failed to parse network CIDR: %v", err)
	}

	route := netlink.Route{
		Dst:       dst,
		LinkIndex: link.Attrs().Index,
		Protocol:  30, // Protocol BIRD (for example)
		Priority:  metric,
	}

	if err := netlink.RouteAdd(&route); err != nil {
		return fmt.Errorf("failed to add route: %v", err)
	}

	return nil
}

// WithdrawNetwork withdraws a network advertisement
func WithdrawNetwork(networkCIDR, tunnelName string) error {
	// Validate network CIDR
	_, _, err := net.ParseCIDR(networkCIDR)
	if err != nil {
		return fmt.Errorf("invalid network CIDR: %v", err)
	}

	// Check if tunnel exists
	link, err := netlink.LinkByName(tunnelName)
	if err != nil {
		return fmt.Errorf("tunnel not found: %v", err)
	}

	// In a real implementation, this would configure the IPsec daemon or routing protocol
	// to withdraw the network advertisement
	
	// For this example, we'll remove the route for the network
	_, dst, err := net.ParseCIDR(networkCIDR)
	if err != nil {
		return fmt.Errorf("failed to parse network CIDR: %v", err)
	}

	route := netlink.Route{
		Dst:       dst,
		LinkIndex: link.Attrs().Index,
	}

	if err := netlink.RouteDel(&route); err != nil {
		return fmt.Errorf("failed to delete route: %v", err)
	}

	return nil
}

// AddRoute adds a route to the routing table
func AddRoute(destination, gateway, iface string, metric int) error {
	// Validate destination
	_, dst, err := net.ParseCIDR(destination)
	if err != nil {
		return fmt.Errorf("invalid destination: %v", err)
	}

	// Validate gateway
	gw := net.ParseIP(gateway)
	if gw == nil {
		return errors.New("invalid gateway IP address")
	}

	// Get interface
	var link netlink.Link
	if iface != "" {
		link, err = netlink.LinkByName(iface)
		if err != nil {
			return fmt.Errorf("interface not found: %v", err)
		}
	} else {
		// If no interface is specified, find the interface with a route to the gateway
		routes, err := netlink.RouteGet(gw)
		if err != nil || len(routes) == 0 {
			return fmt.Errorf("failed to find route to gateway: %v", err)
		}

		link, err = netlink.LinkByIndex(routes[0].LinkIndex)
		if err != nil {
			return fmt.Errorf("failed to find interface for gateway: %v", err)
		}
	}

	// Create route
	route := netlink.Route{
		Dst:       dst,
		Gw:        gw,
		LinkIndex: link.Attrs().Index,
		Priority:  metric,
	}

	// Add route
	if err := netlink.RouteAdd(&route); err != nil {
		return fmt.Errorf("failed to add route: %v", err)
	}

	return nil
}

// DeleteRoute deletes a route from the routing table
func DeleteRoute(destination, gateway, iface string) error {
	// Validate destination
	_, dst, err := net.ParseCIDR(destination)
	if err != nil {
		return fmt.Errorf("invalid destination: %v", err)
	}

	// Validate gateway
	gw := net.ParseIP(gateway)
	if gw == nil {
		return errors.New("invalid gateway IP address")
	}

	// Get interface
	var link netlink.Link
	if iface != "" {
		link, err = netlink.LinkByName(iface)
		if err != nil {
			return fmt.Errorf("interface not found: %v", err)
		}
	} else {
		// If no interface is specified, find the interface with a route to the gateway
		routes, err := netlink.RouteGet(gw)
		if err != nil || len(routes) == 0 {
			return fmt.Errorf("failed to find route to gateway: %v", err)
		}

		link, err = netlink.LinkByIndex(routes[0].LinkIndex)
		if err != nil {
			return fmt.Errorf("failed to find interface for gateway: %v", err)
		}
	}

	// Create route for deletion
	route := netlink.Route{
		Dst:       dst,
		Gw:        gw,
		LinkIndex: link.Attrs().Index,
	}

	// Delete route
	if err := netlink.RouteDel(&route); err != nil {
		return fmt.Errorf("failed to delete route: %v", err)
	}

	return nil
}