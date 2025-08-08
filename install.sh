#!/bin/bash

# Installation script for IPsec VPN with Post-Quantum Encryption

set -e

echo "IPsec VPN with Post-Quantum Encryption - Installation Script"
echo "--------------------------------------------------------"

# Check if running as root
if [ "$(id -u)" -ne 0 ]; then
    echo "Error: This script must be run as root" >&2
    echo "Please run with sudo: sudo ./install.sh" >&2
    exit 1
fi

# Check for required dependencies
echo "Checking dependencies..."

DEPENDENCIES=("go" "make" "ip")
MISSING_DEPS=0

for dep in "${DEPENDENCIES[@]}"; do
    if ! command -v "$dep" &> /dev/null; then
        echo "Error: Required dependency '$dep' is not installed" >&2
        MISSING_DEPS=1
    fi
done

if [ "$MISSING_DEPS" -eq 1 ]; then
    echo "Please install the missing dependencies and try again" >&2
    echo "For Ubuntu/Debian: sudo apt-get install golang make iproute2" >&2
    echo "For CentOS/RHEL: sudo yum install golang make iproute" >&2
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
GO_MAJOR=$(echo "$GO_VERSION" | cut -d. -f1)
GO_MINOR=$(echo "$GO_VERSION" | cut -d. -f2)

if [ "$GO_MAJOR" -lt 1 ] || ([ "$GO_MAJOR" -eq 1 ] && [ "$GO_MINOR" -lt 21 ]); then
    echo "Error: Go version 1.21 or higher is required" >&2
    echo "Current version: $GO_VERSION" >&2
    exit 1
fi

# Check kernel version for IPsec support
KERNEL_VERSION=$(uname -r | cut -d- -f1)
KERNEL_MAJOR=$(echo "$KERNEL_VERSION" | cut -d. -f1)
KERNEL_MINOR=$(echo "$KERNEL_VERSION" | cut -d. -f2)

if [ "$KERNEL_MAJOR" -lt 4 ] || ([ "$KERNEL_MAJOR" -eq 4 ] && [ "$KERNEL_MINOR" -lt 19 ]); then
    echo "Warning: Kernel version 4.19 or higher is recommended for full IPsec support" >&2
    echo "Current kernel version: $(uname -r)" >&2
    echo "Continuing anyway, but some features may not work correctly" >&2
    sleep 2
fi

# Build and install
echo "Building IPsec VPN..."
make build

echo "Installing IPsec VPN..."
make install

# Create configuration directory
CONFIG_DIR="/etc/ipsec-vpn"
echo "Creating configuration directory: $CONFIG_DIR"
mkdir -p "$CONFIG_DIR"

# Install example configuration
if [ ! -f "$CONFIG_DIR/ipsec-vpn.yaml" ]; then
    echo "Installing example configuration to $CONFIG_DIR/ipsec-vpn.yaml"
    cp .ipsec-vpn.yaml "$CONFIG_DIR/ipsec-vpn.yaml"
else
    echo "Configuration file already exists, not overwriting"
    echo "Example configuration saved to $CONFIG_DIR/ipsec-vpn.yaml.example"
    cp .ipsec-vpn.yaml "$CONFIG_DIR/ipsec-vpn.yaml.example"
fi

# Set permissions
chmod 600 "$CONFIG_DIR/ipsec-vpn.yaml"*

# Create log directory
LOG_DIR="/var/log/ipsec-vpn"
echo "Creating log directory: $LOG_DIR"
mkdir -p "$LOG_DIR"
chmod 755 "$LOG_DIR"

echo ""
echo "Installation completed successfully!"
echo ""
echo "IPsec VPN is now installed in /usr/local/bin/ipsec-vpn"
echo "Configuration file: $CONFIG_DIR/ipsec-vpn.yaml"
echo "Log directory: $LOG_DIR"
echo ""
echo "To get started, run: ipsec-vpn --help"
echo ""