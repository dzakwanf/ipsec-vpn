package tunnel

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
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

	// Create GRE tunnel interface
	logger.Debug("Creating GRE tunnel interface for '%s'", config.Name)
	if err := createGRETunnelInterface(tunnel); err != nil {
		logger.Error("Failed to create GRE tunnel interface: %v", err)
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

	// Delete GRE tunnel interface
	if err := deleteGRETunnelInterface(tunnel); err != nil && !force {
		return err
	}

	// Delete tunnel configuration
	return deleteTunnelConfig(name)
}
}

// createGRETunnelInterface creates a GRE tunnel interface
func createGRETunnelInterface(tunnel *Tunnel) error {
	// Requires root privileges
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root to create GRE tunnel interfaces")
	}

	localIP := net.ParseIP(tunnel.LocalIP)
	remoteIP := net.ParseIP(tunnel.RemoteIP)
	if localIP == nil || remoteIP == nil {
		return fmt.Errorf("invalid LocalIP or RemoteIP for tunnel: %s, %s", tunnel.LocalIP, tunnel.RemoteIP)
	}

	attrs := netlink.NewLinkAttrs()
	attrs.Name = fmt.Sprintf("gre-%s", tunnel.Name)

	gre := &netlink.Gretun{
		LinkAttrs: attrs,
		Local:     localIP,
		Remote:    remoteIP,
		IKey:      0,
		OKey:      0,
	}

	if err := netlink.LinkAdd(gre); err != nil {
		return fmt.Errorf("failed to create GRE tunnel interface: %v", err)
	}

	// Bring the interface up
	if err := netlink.LinkSetUp(gre); err != nil {
		return fmt.Errorf("failed to bring GRE tunnel interface up: %v", err)
	}

	return nil
}

// deleteGRETunnelInterface deletes a GRE tunnel interface
func deleteGRETunnelInterface(tunnel *Tunnel) error {
	if os.Geteuid() != 0 {
		return fmt.Errorf("must run as root to delete GRE tunnel interfaces")
	}
	link, err := netlink.LinkByName(fmt.Sprintf("gre-%s", tunnel.Name))
	if err != nil {
		return nil // Interface doesn't exist, nothing to delete
	}
	if err := netlink.LinkDel(link); err != nil {
		return fmt.Errorf("failed to delete GRE tunnel interface: %v", err)
	}
	return nil
}
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
	// No tunnel interface needed for standard IPsec/XFRM
	return nil
}

// deleteTunnelInterface deletes the tunnel interface
func deleteTunnelInterface(tunnel *Tunnel) error {
	// No tunnel interface to delete for standard IPsec/XFRM
	return nil
}

// startTunnel starts the tunnel
func startTunnel(tunnel *Tunnel) error {
	// Here you should configure XFRM policies and states for IPsec
	// Example: use netlink.XfrmPolicyAdd and netlink.XfrmStateAdd
	// For now, just simulate success
	logger.Info("Configured XFRM policies and states for tunnel '%s'", tunnel.Name)
	return nil
}

// stopTunnel stops the tunnel
func stopTunnel(tunnel *Tunnel) error {
	// Here you should remove XFRM policies and states for IPsec
	// Example: use netlink.XfrmPolicyDel and netlink.XfrmStateDel
	// For now, just simulate success
	logger.Info("Removed XFRM policies and states for tunnel '%s'", tunnel.Name)
	return nil
}

// getTunnelStatus returns the current status of the tunnel
func getTunnelStatus(tunnel *Tunnel) (Status, error) {
	// For standard IPsec/XFRM, status is based on config only
	return tunnel.Status, nil
}