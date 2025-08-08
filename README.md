# IPsec VPN with Post-Quantum Encryption

A production-grade IPsec VPN solution with post-quantum cryptography capabilities. This project provides secure tunneling for network traffic with advanced encryption options including post-quantum algorithms.

## Features

- IPsec tunnel and transport modes
- Post-quantum encryption algorithms (Kyber)
- Network advertisement capabilities
- Cisco-like CLI configuration interface
- Comprehensive logging and monitoring
- End-to-end server configuration support

## Requirements

- Go 1.21 or higher
- Linux kernel 4.19 or higher (for IPsec and network functionality)
- Root/sudo privileges (for creating network interfaces and configuring IPsec)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/dzakwan/ipsec-vpn.git
cd ipsec-vpn

# Build the binary
go build -o ipsec-vpn

# Install (optional)
sudo mv ipsec-vpn /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/dzakwan/ipsec-vpn@latest
```

## Quick Start

### Creating a Simple Tunnel

```bash
# Create a tunnel
sudo ipsec-vpn tunnel create mytunnel \
  --local-ip 192.168.1.1 \
  --remote-ip 10.0.0.1 \
  --local-subnet 192.168.0.0/24 \
  --remote-subnet 10.0.0.0/24

# Check tunnel status
sudo ipsec-vpn tunnel show mytunnel

# Stop the tunnel
sudo ipsec-vpn tunnel stop mytunnel

# Delete the tunnel
sudo ipsec-vpn tunnel delete mytunnel
```

### Setting Up End-to-End Server Connection

For a quick end-to-end server setup, configure both endpoints:

```bash
# On Server A (203.0.113.10)
sudo ipsec-vpn tunnel create server-a \
  --local-ip 203.0.113.10 \
  --remote-ip 198.51.100.20 \
  --local-subnet 10.0.1.0/24 \
  --remote-subnet 10.0.2.0/24

# On Server B (198.51.100.20)
sudo ipsec-vpn tunnel create server-b \
  --local-ip 198.51.100.20 \
  --remote-ip 203.0.113.10 \
  --local-subnet 10.0.2.0/24 \
  --remote-subnet 10.0.1.0/24

# Start tunnels on both servers
sudo ipsec-vpn tunnel start server-a    # On Server A
sudo ipsec-vpn tunnel start server-b    # On Server B
```

### Using Post-Quantum Encryption

```bash
# Create a tunnel with post-quantum encryption
sudo ipsec-vpn tunnel create secure-tunnel \
  --local-ip 192.168.1.1 \
  --remote-ip 10.0.0.1 \
  --local-subnet 192.168.0.0/24 \
  --remote-subnet 10.0.0.0/24 \
  --encryption kyber768 \
  --post-quantum

# View available post-quantum algorithms
ipsec-vpn crypto show --post-quantum
```

### Network Advertisement

```bash
# Advertise a network through a tunnel
sudo ipsec-vpn network advertise 172.16.0.0/24 --tunnel mytunnel

# View advertised networks
sudo ipsec-vpn network show --advertised

# Withdraw a network advertisement
sudo ipsec-vpn network withdraw 172.16.0.0/24 --tunnel mytunnel
```

## Command Reference

### Global Commands

- `ipsec-vpn version`: Display version information
- `ipsec-vpn --help`: Show help information
- `ipsec-vpn --config <file>`: Use a specific configuration file
- `ipsec-vpn --verbose`: Enable verbose output

### Tunnel Management

- `ipsec-vpn tunnel create [name]`: Create a new IPsec tunnel
  - `--local-ip`: Local IP address for the tunnel
  - `--remote-ip`: Remote IP address for the tunnel
  - `--local-subnet`: Local subnet to be tunneled (CIDR notation)
  - `--remote-subnet`: Remote subnet to be tunneled (CIDR notation)
  - `--encryption`: Encryption algorithm (default: aes256gcm)
  - `--post-quantum`: Enable post-quantum cryptography

- `ipsec-vpn tunnel show [name]`: Show tunnel details or list all tunnels
- `ipsec-vpn tunnel start [name]`: Start an IPsec tunnel
- `ipsec-vpn tunnel stop [name]`: Stop an IPsec tunnel
- `ipsec-vpn tunnel delete [name]`: Delete an IPsec tunnel
  - `--force`: Force deletion even if tunnel is active

### Cryptographic Settings

- `ipsec-vpn crypto show`: Show available cryptographic algorithms
  - `--post-quantum`: Show post-quantum algorithms only
  - `--classic`: Show classic algorithms only

- `ipsec-vpn crypto test [algorithm]`: Test a cryptographic algorithm
  - `--data`: Data to use for testing encryption

- `ipsec-vpn crypto set-default [algorithm]`: Set the default encryption algorithm
  - `--post-quantum`: Set as default post-quantum algorithm

### Network Management

- `ipsec-vpn network show`: Show network configuration
  - `--interfaces`: Show network interfaces
  - `--routes`: Show routing table
  - `--advertised`: Show advertised networks

- `ipsec-vpn network advertise [network]`: Advertise a network
  - `--tunnel`: Tunnel to advertise the network through
  - `--metric`: Metric for the advertised route (default: 100)

- `ipsec-vpn network withdraw [network]`: Withdraw a network advertisement
  - `--tunnel`: Tunnel to withdraw the network from

- `ipsec-vpn network route [add|delete] [destination] [gateway]`: Manage routes
  - `--interface`: Network interface for the route
  - `--metric`: Metric for the route (default: 100)

## Configuration File

The configuration file uses YAML format and can be placed in the following locations:
- `$HOME/.ipsec-vpn.yaml`
- `/etc/ipsec-vpn/.ipsec-vpn.yaml`
- Current directory: `.ipsec-vpn.yaml`

Example configuration for end-to-end server setup:

```yaml
# Global settings
verbose: true
config_dir: "/etc/ipsec-vpn"

# Crypto settings
crypto:
  default_classic: aes256gcm
  default_post_quantum: kyber768

# Tunnel defaults
tunnel_defaults:
  encryption: aes256gcm
  post_quantum: false
  mtu: 1400
  key_rotation_interval: 86400  # 24 hours in seconds

# Pre-configured tunnels
tunnels:
  # Standard tunnel with classic encryption
  office:
    local_ip: 192.168.1.1
    remote_ip: 10.0.0.1
    local_subnet: 192.168.0.0/24
    remote_subnet: 10.0.0.0/24
    encryption: chacha20poly1305
    post_quantum: false
    description: "Office VPN connection"
  
  # Secure tunnel with post-quantum encryption
  datacenter:
    local_ip: 192.168.1.1
    remote_ip: 172.16.0.1
    local_subnet: 192.168.0.0/24
    remote_subnet: 172.16.0.0/16
    encryption: kyber768
    post_quantum: true
    description: "Datacenter connection with post-quantum security"

# Network advertisement settings
network_advertisement:
  enabled: true
  advertised_networks:
    - cidr: 192.168.10.0/24
      tunnel: office
      metric: 100
    - cidr: 192.168.20.0/24
      tunnel: datacenter
      metric: 200

# Security settings
security:
  perfect_forward_secrecy: true
  key_rotation_enabled: true
  replay_protection: true
  authentication_method: psk  # pre-shared key
  psk_file: "/etc/ipsec-vpn/psk.key"

# Advanced settings
advanced:
  ike_version: 2  # IKEv2
  esp_proposals:
    - aes256gcm-sha256
    - chacha20poly1305-sha256
  dpd_delay: 30  # seconds
  dpd_timeout: 120  # seconds
```

## Security Considerations

### Post-Quantum Cryptography

This project implements post-quantum cryptographic algorithms to protect against future quantum computing threats. The implemented algorithms (Kyber) are NIST-selected candidates for standardization.

### Key Management

IPsec VPN uses secure key management practices:
- Keys are generated using cryptographically secure random number generators
- For post-quantum algorithms, hybrid modes are available that combine classical and post-quantum security
- Keys are never stored in plaintext on disk

### Network Security

- All network traffic is encrypted using strong algorithms
- Perfect Forward Secrecy is implemented to protect past communications
- Regular key rotation is enforced

## Architecture

The IPsec VPN solution follows a modular architecture:

1. **CLI Interface**: Provides a Cisco-like command-line interface for configuration
2. **Tunnel Management**: Handles creation, configuration, and lifecycle of IPsec tunnels
3. **Cryptographic Layer**: Implements encryption algorithms, including post-quantum options
4. **Network Layer**: Manages network interfaces, routing, and advertisement
5. **Configuration Management**: Handles persistent configuration storage and retrieval

## Development

### Project Structure

```
.
├── cmd/                # Command-line interface
│   ├── root.go        # Root command
│   ├── tunnel.go      # Tunnel management commands
│   ├── crypto.go      # Cryptographic settings commands
│   ├── network.go     # Network management commands
│   └── version.go     # Version information
├── pkg/               # Core packages
│   ├── tunnel/        # Tunnel implementation
│   ├── crypto/        # Cryptographic algorithms
│   └── network/       # Network management
├── go.mod             # Go module definition
├── go.sum             # Go module checksums
├── main.go            # Application entry point
└── README.md          # Documentation
```

### Building from Source

```bash
# Get dependencies
go mod download

# Build
go build -o ipsec-vpn

# Run tests
go test ./...
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## End-to-End Server Configuration

To set up an end-to-end VPN connection between two servers, follow these steps:

### 1. Server Side Configuration

```bash
# On the server (e.g., 203.0.113.10)
sudo ipsec-vpn tunnel create server-tunnel \
  --local-ip 203.0.113.10 \
  --remote-ip 198.51.100.20 \
  --local-subnet 10.0.1.0/24 \
  --remote-subnet 10.0.2.0/24 \
  --encryption aes256gcm

# Start the tunnel
sudo ipsec-vpn tunnel start server-tunnel
```

### 2. Client Side Configuration

```bash
# On the client (e.g., 198.51.100.20)
sudo ipsec-vpn tunnel create client-tunnel \
  --local-ip 198.51.100.20 \
  --remote-ip 203.0.113.10 \
  --local-subnet 10.0.2.0/24 \
  --remote-subnet 10.0.1.0/24 \
  --encryption aes256gcm

# Start the tunnel
sudo ipsec-vpn tunnel start client-tunnel
```

### 3. Verification

```bash
# Check tunnel status on both sides
sudo ipsec-vpn tunnel show server-tunnel  # On server
sudo ipsec-vpn tunnel show client-tunnel   # On client

# Test connectivity by pinging hosts on the remote subnet
```

### 4. Advanced Security Configuration

For production environments, consider these additional security measures:

```bash
# Create a strong pre-shared key
openssl rand -base64 48 > /etc/ipsec-vpn/psk.key
chmod 600 /etc/ipsec-vpn/psk.key

# Enable post-quantum cryptography
sudo ipsec-vpn crypto set-default kyber768 --post-quantum
```

### 5. Firewall Configuration

For end-to-end server connections to work properly, you need to configure your firewall to allow IPsec traffic. Here are the required firewall rules:

```bash
# For iptables-based firewalls (on both servers)

# Allow IKE (Internet Key Exchange)
sudo iptables -A INPUT -p udp --dport 500 -j ACCEPT
sudo iptables -A OUTPUT -p udp --sport 500 -j ACCEPT

# Allow NAT-Traversal (if NAT is involved)
sudo iptables -A INPUT -p udp --dport 4500 -j ACCEPT
sudo iptables -A OUTPUT -p udp --sport 4500 -j ACCEPT

# Allow ESP (Encapsulating Security Payload)
sudo iptables -A INPUT -p esp -j ACCEPT
sudo iptables -A OUTPUT -p esp -j ACCEPT

# Save the rules (depends on your distribution)
sudo service iptables save  # For some distributions
# or
sudo iptables-save > /etc/iptables/rules.v4  # For others
```

For UFW (Uncomplicated Firewall):

```bash
sudo ufw allow 500/udp
sudo ufw allow 4500/udp
sudo ufw allow esp
```

### 6. Troubleshooting End-to-End Connections

If you encounter issues with your end-to-end server connection:

```bash
# Enable verbose logging
sudo ipsec-vpn --verbose tunnel show server-tunnel

# Check system logs
cat /var/log/ipsec-vpn.log

# Verify network connectivity between servers
ping <remote_server_ip>

# Check firewall rules (IPsec requires UDP 500, UDP 4500, and ESP protocol)
sudo iptables -L -n

# Restart a problematic tunnel
sudo ipsec-vpn tunnel stop server-tunnel
sudo ipsec-vpn tunnel start server-tunnel
```

Common issues and solutions:

1. **Tunnel creation fails**: Verify IP addresses and subnet configurations
2. **Tunnel starts but no connectivity**: Check firewall rules on both servers
3. **Intermittent connection issues**: Consider enabling DPD (Dead Peer Detection) in configuration
4. **Performance issues**: Adjust MTU settings in the configuration file

### 7. Performance Tuning

To optimize your end-to-end server connection performance:

1. **MTU Optimization**:
   ```yaml
   # In your configuration file
   tunnel_defaults:
     mtu: 1400  # Adjust based on your network requirements
   ```

2. **Encryption Algorithm Selection**:
   - For maximum performance: `aes256gcm`
   - For balanced security/performance: `chacha20poly1305`
   - For maximum security: `kyber768` (post-quantum)

3. **Kernel Parameters** (add to `/etc/sysctl.conf` and apply with `sudo sysctl -p`):
   ```
   # Increase UDP buffer sizes
   net.core.rmem_max = 16777216
   net.core.wmem_max = 16777216
   
   # Enable TCP BBR congestion control
   net.core.default_qdisc = fq
   net.ipv4.tcp_congestion_control = bbr
   ```

4. **Hardware Acceleration**:
   - If available, enable AES-NI in your BIOS/UEFI
   - For cloud instances, choose instances with cryptographic acceleration

### 8. High Availability Configuration

For mission-critical end-to-end server connections, implement high availability:

1. **Multiple Tunnel Setup**:
   ```bash
   # Create primary tunnel
   sudo ipsec-vpn tunnel create primary-tunnel \
     --local-ip 203.0.113.10 \
     --remote-ip 198.51.100.20 \
     --local-subnet 10.0.1.0/24 \
     --remote-subnet 10.0.2.0/24
   
   # Create backup tunnel through alternate path
   sudo ipsec-vpn tunnel create backup-tunnel \
     --local-ip 203.0.113.11 \
     --remote-ip 198.51.100.21 \
     --local-subnet 10.0.1.0/24 \
     --remote-subnet 10.0.2.0/24
   ```

2. **Monitoring and Failover**:
   - Use a monitoring script to check tunnel status
   - Implement automatic failover when primary tunnel fails

   Example monitoring script:
   ```bash
   #!/bin/bash
   
   PRIMARY="primary-tunnel"
   BACKUP="backup-tunnel"
   
   # Check if primary tunnel is up
   if ! sudo ipsec-vpn tunnel show $PRIMARY | grep -q "Status: UP"; then
     echo "Primary tunnel down, activating backup"
     sudo ipsec-vpn tunnel start $BACKUP
   else
     # Ensure backup is ready but not active unless needed
     sudo ipsec-vpn tunnel stop $BACKUP
   fi
   ```

3. **Load Balancing**:
   For high-throughput scenarios, distribute traffic across multiple tunnels:
   ```yaml
   # In your configuration file
   load_balancing:
     enabled: true
     tunnels:
       - tunnel1
       - tunnel2
     method: round-robin  # or weighted, source-ip
   ```

## Acknowledgments

- [Cloudflare CIRCL](https://github.com/cloudflare/circl) for post-quantum cryptography implementations
- [Cobra](https://github.com/spf13/cobra) for CLI framework
- [Viper](https://github.com/spf13/viper) for configuration management
- [Netlink](https://github.com/vishvananda/netlink) for network interface management