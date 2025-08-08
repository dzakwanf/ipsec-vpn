package tunnel

import (
	"errors"
	"fmt"
	"hash/fnv"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"time"
	"github.com/dzakwan/ipsec-vpn/pkg/crypto"
	"github.com/dzakwan/ipsec-vpn/pkg/logger"
	"github.com/spf13/viper"
	"github.com/vishvananda/netlink"
)

// Status represents the current state of a tunnel
type Status string

const (
	StatusDown    Status = "DOWN"
	StatusUp      Status = "UP"
	StatusError   Status = "ERROR"
	StatusUnknown Status = "UNKNOWN"
)

// Config represents the configuration for an IPsec tunnel
type Config struct {
	Name         string
	LocalIP      string
	RemoteIP     string
	LocalSubnet  string
	RemoteSubnet string
	Encryption   string
	PostQuantum  bool
}

// Tunnel represents an IPsec tunnel
type Tunnel struct {
	Name         string    `json:"name"`
	LocalIP      string    `json:"local_ip"`
	RemoteIP     string    `json:"remote_ip"`
	LocalSubnet  string    `json:"local_subnet"`
	RemoteSubnet string    `json:"remote_subnet"`
	Encryption   string    `json:"encryption"`
	PostQuantum  bool      `json:"post_quantum"`
	Status       Status    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Create creates a new IPsec tunnel with the given configuration
func Create(config Config) (*Tunnel, error) {
	// Validate configuration
	if err := validateConfig(config); err != nil {
		logger.Error("Failed to validate tunnel configuration: %v", err)
		return nil, err
	}

	// Check if tunnel already exists
	if _, err := Get(config.Name); err == nil {
		logger.Error("Tunnel with name '%s' already exists", config.Name)
		return nil, fmt.Errorf("tunnel with name '%s' already exists", config.Name)
	}

	logger.Info("Creating new tunnel '%s' from %s to %s", config.Name, config.LocalIP, config.RemoteIP)
	logger.Debug("Tunnel details: local subnet %s, remote subnet %s, encryption %s, post-quantum %v", 
		config.LocalSubnet, config.RemoteSubnet, config.Encryption, config.PostQuantum)

	// Create tunnel object
	tunnel := &Tunnel{
		Name:         config.Name,
		LocalIP:      config.LocalIP,
		RemoteIP:     config.RemoteIP,
		LocalSubnet:  config.LocalSubnet,
		RemoteSubnet: config.RemoteSubnet,
		Encryption:   config.Encryption,
		PostQuantum:  config.PostQuantum,
		Status:       StatusDown,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save tunnel configuration
	if err := saveTunnel(tunnel); err != nil {
		logger.Error("Failed to save tunnel configuration: %v", err)
		return nil, err
	}

	// Create the actual tunnel interface
	logger.Debug("Creating tunnel interface for '%s'", config.Name)
	if err := createTunnelInterface(tunnel); err != nil {
		// Cleanup on failure
		logger.Error("Failed to create tunnel interface: %v", err)
		_ = deleteTunnelConfig(config.Name)
		return nil, err
	}

	// Start the tunnel
	logger.Debug("Starting tunnel '%s'", config.Name)
	if err := startTunnel(tunnel); err != nil {
		// Cleanup on failure
		logger.Error("Failed to start tunnel: %v", err)
		_ = deleteTunnelInterface(tunnel)
		_ = deleteTunnelConfig(config.Name)
		return nil, err
	}

	// Update status
	tunnel.Status = StatusUp
	if err := saveTunnel(tunnel); err != nil {
		return nil, err
	}

	return tunnel, nil
}

// Get retrieves a tunnel by name
func Get(name string) (*Tunnel, error) {
	// Load tunnel configuration
	tunnel, err := loadTunnel(name)
	if err != nil {
		return nil, err
	}

	// Update status
	status, err := getTunnelStatus(tunnel)
	if err != nil {
		tunnel.Status = StatusUnknown
	} else {
		tunnel.Status = status
	}

	return tunnel, nil
}

// ListAll returns all configured tunnels
func ListAll() ([]*Tunnel, error) {
	// Get config directory
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	// List all tunnel config files
	pattern := filepath.Join(configDir, "tunnels", "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	// Load each tunnel
	tunnels := make([]*Tunnel, 0, len(files))
	for _, file := range files {
		name := filepath.Base(file)
		name = name[:len(name)-5] // Remove .json extension

		tunnel, err := Get(name)
		if err != nil {
			// Skip tunnels with errors
			continue
		}

		tunnels = append(tunnels, tunnel)
	}

	return tunnels, nil
}

// Start starts an existing tunnel
func Start(name string) error {
	// Get tunnel
	tunnel, err := Get(name)
	if err != nil {
		return err
	}

	// Check if tunnel is already up
	if tunnel.Status == StatusUp {
		return nil
	}

	// Start the tunnel
	if err := startTunnel(tunnel); err != nil {
		return err
	}

	// Update status
	tunnel.Status = StatusUp
	tunnel.UpdatedAt = time.Now()
	if err := saveTunnel(tunnel); err != nil {
		return err
	}

	return nil
}

// Stop stops an active tunnel
func Stop(name string) error {
	// Get tunnel
	logger.Debug("Attempting to stop tunnel '%s'", name)
	tunnel, err := Get(name)
	if err != nil {
		logger.Error("Failed to get tunnel '%s': %v", name, err)
		return err
	}

	// Check if tunnel is already down
	if tunnel.Status == StatusDown {
		logger.Info("Tunnel '%s' is already down, no action needed", name)
		return nil
	}

	// Stop the tunnel
	logger.Info("Stopping tunnel '%s'", name)
	if err := stopTunnel(tunnel); err != nil {
		logger.Error("Failed to stop tunnel '%s': %v", name, err)
		return err
	}

	// Update status
	tunnel.Status = StatusDown
	tunnel.UpdatedAt = time.Now()
	if err := saveTunnel(tunnel); err != nil {
		logger.Error("Failed to update tunnel status: %v", err)
		return err
	}

	logger.Info("Tunnel '%s' stopped successfully", name)
	return nil
}

// Delete removes a tunnel
func Delete(name string, force bool) error {
	// Get tunnel
	tunnel, err := Get(name)
	if err != nil {
		if force {
			// If forced, try to delete config even if tunnel doesn't exist
			return deleteTunnelConfig(name)
		}
		return err
	}

	// Stop the tunnel if it's running
	if tunnel.Status == StatusUp && !force {
		return errors.New("tunnel is active, stop it first or use --force")
	} else if tunnel.Status == StatusUp {
		_ = stopTunnel(tunnel)
	}

	// Delete tunnel interface
	if err := deleteTunnelInterface(tunnel); err != nil && !force {
		return err
	}

	// Delete tunnel configuration
	return deleteTunnelConfig(name)
}

// Helper functions

// validateConfig validates the tunnel configuration
func validateConfig(config Config) error {
	if config.Name == "" {
		return errors.New("tunnel name cannot be empty")
	}

	if config.LocalIP == "" {
		return errors.New("local IP cannot be empty")
	}

	if config.RemoteIP == "" {
		return errors.New("remote IP cannot be empty")
	}

	if config.LocalSubnet == "" {
		return errors.New("local subnet cannot be empty")
	}

	if config.RemoteSubnet == "" {
		return errors.New("remote subnet cannot be empty")
	}

	// Validate encryption algorithm
	if config.Encryption == "" {
		config.Encryption = "aes256gcm" // Default
	} else {
		valid := false
		for _, algo := range crypto.ListClassicAlgorithms() {
			if algo.Name == config.Encryption {
				valid = true
				break
			}
		}

		if !valid && config.PostQuantum {
			for _, algo := range crypto.ListPostQuantumAlgorithms() {
				if algo.Name == config.Encryption {
					valid = true
					break
				}
			}
		}

		if !valid {
			return fmt.Errorf("invalid encryption algorithm: %s", config.Encryption)
		}
	}

	return nil
}

// getConfigDir returns the configuration directory
func getConfigDir() (string, error) {
	// Check if config directory is set in viper
	configDir := viper.GetString("config_dir")
	if configDir != "" {
		return configDir, nil
	}

	// Use default config directory
	home := ""
	// Check if running with sudo
	if sudoUser := os.Getenv("SUDO_USER"); sudoUser != "" {
		u, err := user.Lookup(sudoUser)
		if err != nil {
			return "", err
		}
		home = u.HomeDir
	} else {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return "", err
		}
	}

	configDir = filepath.Join(home, ".ipsec-vpn")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Join(configDir, "tunnels"), 0755); err != nil {
		return "", err
	}

	return configDir, nil
}

// saveTunnel saves the tunnel configuration to disk
func saveTunnel(tunnel *Tunnel) error {
	// Get config directory
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	// Create tunnels directory if it doesn't exist
	tunnelsDir := filepath.Join(configDir, "tunnels")
	if err := os.MkdirAll(tunnelsDir, 0755); err != nil {
		return err
	}

	// Create a new viper instance for this tunnel
	v := viper.New()
	v.SetConfigType("json")

	// Set tunnel configuration
	v.Set("name", tunnel.Name)
	v.Set("local_ip", tunnel.LocalIP)
	v.Set("remote_ip", tunnel.RemoteIP)
	v.Set("local_subnet", tunnel.LocalSubnet)
	v.Set("remote_subnet", tunnel.RemoteSubnet)
	v.Set("encryption", tunnel.Encryption)
	v.Set("post_quantum", tunnel.PostQuantum)
	v.Set("status", string(tunnel.Status))
	v.Set("created_at", tunnel.CreatedAt)
	v.Set("updated_at", tunnel.UpdatedAt)

	// Save configuration to file
	configFile := filepath.Join(tunnelsDir, tunnel.Name+".json")
	return v.WriteConfigAs(configFile)
}

// loadTunnel loads a tunnel configuration from disk
func loadTunnel(name string) (*Tunnel, error) {
	// Get config directory
	configDir, err := getConfigDir()
	if err != nil {
		return nil, err
	}

	// Check if tunnel config exists
	configFile := filepath.Join(configDir, "tunnels", name+".json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("tunnel '%s' not found", name)
	}

	// Create a new viper instance for this tunnel
	v := viper.New()
	v.SetConfigType("json")
	v.SetConfigFile(configFile)

	// Read configuration
	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	// Create tunnel object
	tunnel := &Tunnel{
		Name:         v.GetString("name"),
		LocalIP:      v.GetString("local_ip"),
		RemoteIP:     v.GetString("remote_ip"),
		LocalSubnet:  v.GetString("local_subnet"),
		RemoteSubnet: v.GetString("remote_subnet"),
		Encryption:   v.GetString("encryption"),
		PostQuantum:  v.GetBool("post_quantum"),
		Status:       Status(v.GetString("status")),
	}

	// Parse timestamps
	if v.IsSet("created_at") {
		tunnel.CreatedAt = v.GetTime("created_at")
	} else {
		tunnel.CreatedAt = time.Now()
	}

	if v.IsSet("updated_at") {
		tunnel.UpdatedAt = v.GetTime("updated_at")
	} else {
		tunnel.UpdatedAt = time.Now()
	}

	return tunnel, nil
}

// deleteTunnelConfig deletes the tunnel configuration from disk
func deleteTunnelConfig(name string) error {
	// Get config directory
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}

	// Delete tunnel config file
	configFile := filepath.Join(configDir, "tunnels", name+".json")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to delete
	}

	return os.Remove(configFile)
}

// createTunnelInterface creates the actual tunnel interface
func createTunnelInterface(tunnel *Tunnel) error {
	// This is a simplified implementation
	// In a real implementation, this would use netlink to create the tunnel interface

	// Check if running on macOS
	if runtime.GOOS == "darwin" {
		logger.Info("Tunnel interface creation is not supported on macOS")
		// For macOS, we'll simulate success for development purposes
		return nil
	}

	// Check for root privileges
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root to create tunnel interfaces")
	}

	// Check if ip_vti kernel module is loaded
	if _, err := os.Stat("/proc/net/vti"); os.IsNotExist(err) {
		return fmt.Errorf("ip_vti kernel module not loaded. Run 'sudo modprobe ip_vti'")
	}

	// Validate IP addresses
	localIP := net.ParseIP(tunnel.LocalIP)
	remoteIP := net.ParseIP(tunnel.RemoteIP)
	if localIP == nil || remoteIP == nil {
		return fmt.Errorf("invalid LocalIP or RemoteIP for tunnel: %s, %s", tunnel.LocalIP, tunnel.RemoteIP)
	}

	// Create VTI tunnel interface
	attrs := netlink.NewLinkAttrs()
	attrs.Name = fmt.Sprintf("ipsec-%s", tunnel.Name)

	// Generate a unique key for the tunnel based on its name
	h := fnv.New32a()
	h.Write([]byte(tunnel.Name))
	key := h.Sum32()

	vti := &netlink.Vti{
		LinkAttrs: attrs,
		IKey:      key,
		OKey:      key,
		Local:     localIP,
		Remote:    remoteIP,
	}

	// Add the tunnel interface
	if err := netlink.LinkAdd(vti); err != nil {
		return fmt.Errorf("failed to create tunnel interface: %v", err)
	}

	return nil
}

// deleteTunnelInterface deletes the tunnel interface
func deleteTunnelInterface(tunnel *Tunnel) error {
	// This is a simplified implementation
	// In a real implementation, this would use netlink to delete the tunnel interface
	
	// Check if running on macOS
	if runtime.GOOS == "darwin" {
		logger.Info("Tunnel interface deletion is not supported on macOS")
		// For macOS, we'll simulate success for development purposes
		return nil
	}
	
	// Get the tunnel interface
	link, err := netlink.LinkByName(fmt.Sprintf("ipsec-%s", tunnel.Name))
	if err != nil {
		return nil // Interface doesn't exist, nothing to delete
	}

	// Delete the tunnel interface
	if err := netlink.LinkDel(link); err != nil {
		return fmt.Errorf("failed to delete tunnel interface: %v", err)
	}

	return nil
}

// startTunnel starts the tunnel
func startTunnel(tunnel *Tunnel) error {
	// This is a simplified implementation
	// In a real implementation, this would configure IPsec policies and start the tunnel
	
	// Check if running on macOS
	if runtime.GOOS == "darwin" {
		logger.Info("Tunnel interface activation is not supported on macOS")
		// For macOS, we'll simulate success for development purposes
		return nil
	}
	
	// Get the tunnel interface
	link, err := netlink.LinkByName(fmt.Sprintf("ipsec-%s", tunnel.Name))
	if err != nil {
		return fmt.Errorf("tunnel interface not found: %v", err)
	}

	// Bring the interface up
	if err := netlink.LinkSetUp(link); err != nil {
		return fmt.Errorf("failed to bring tunnel interface up: %v", err)
	}

	return nil
}

// stopTunnel stops the tunnel
func stopTunnel(tunnel *Tunnel) error {
	// This is a simplified implementation
	// In a real implementation, this would remove IPsec policies and stop the tunnel
	
	// Check if running on macOS
	if runtime.GOOS == "darwin" {
		logger.Info("Tunnel interface deactivation is not supported on macOS")
		// For macOS, we'll simulate success for development purposes
		return nil
	}
	
	// Get the tunnel interface
	link, err := netlink.LinkByName(fmt.Sprintf("ipsec-%s", tunnel.Name))
	if err != nil {
		return nil // Interface doesn't exist, nothing to stop
	}

	// Bring the interface down
	if err := netlink.LinkSetDown(link); err != nil {
		return fmt.Errorf("failed to bring tunnel interface down: %v", err)
	}

	return nil
}

// getTunnelStatus returns the current status of the tunnel
func getTunnelStatus(tunnel *Tunnel) (Status, error) {
	// This is a simplified implementation
	// In a real implementation, this would check the actual tunnel status
	
	// Check if running on macOS
	if runtime.GOOS == "darwin" {
		logger.Debug("Tunnel status check is not supported on macOS, returning stored status")
		// For macOS, we'll return the stored status
		return tunnel.Status, nil
	}
	
	// Get the tunnel interface
	link, err := netlink.LinkByName(fmt.Sprintf("ipsec-%s", tunnel.Name))
	if err != nil {
		return StatusDown, nil // Interface doesn't exist, tunnel is down
	}

	// Check if the interface is up
	if link.Attrs().Flags&net.FlagUp != 0 {
		return StatusUp, nil
	}

	return StatusDown, nil
}