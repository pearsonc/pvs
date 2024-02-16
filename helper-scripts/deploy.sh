#!/bin/bash

# Before running this script please create /config at the root of the filesystem and add openvpn-credentials.txt
# This should simply contain username on the first line and password on the second line

# Install dependencies
sudo apt-get update
sudo apt-get install -y resolvconf

# Set variables for clarity
REPO_DIR="/tmp/pearson-vpn-service"
VPN_CONFIGS_DIR="/tmp/pearson-vpn-service/vpn-configs"
BINARY_PATH="/etc/pvs/pvs.bin"
SERVICE_PATH="/etc/pvs"
SERVICE_FILE="pvs.service"
SYSTEMD_PATH="/etc/systemd/system"

# Stop the service if it is running
sudo systemctl stop "$SERVICE_FILE"

# Start with a clean install
sudo rm -rf "$SERVICE_PATH"
sudo mkdir "$SERVICE_PATH"

export PATH=$PATH:/usr/local/go/bin

# Navigate to the repository directory and update the develop branch
cd "$REPO_DIR" || { echo "Failed to change directory to $REPO_DIR. Exiting."; exit 1; }

# Checkout the develop branch and pull the latest changes
/usr/bin/git checkout develop
/usr/bin/git pull origin develop

# Build the project
/usr/bin/make build

# Move the binary to the desired location (requires sudo)
sudo mv bin/wd40vpn.bin "$BINARY_PATH"

# Copy the helper scripts to the desired location (requires sudo)
sudo cp helper-scripts/* "$SERVICE_PATH"

# Copy the VPN to the desired location (requires sudo)
sudo cp -r "$VPN_CONFIGS_DIR" "$SERVICE_PATH"

# Deploy the service file (requires sudo)
sudo cp "$REPO_DIR/$SERVICE_FILE" "$SYSTEMD_PATH"

# Remove any existing 'nameserver' entries
sudo sed -i '/nameserver/d' /etc/resolvconf/resolv.conf.d/base

# Add the new default DNS servers
echo "nameserver 1.1.1.1" | sudo tee -a /etc/resolvconf/resolv.conf.d/base
echo "nameserver 1.0.0.1" | sudo tee -a /etc/resolvconf/resolv.conf.d/base

# Update resolvconf
sudo resolvconf -u


# Reload the systemctl daemon to pick up changes and restart the service (requires sudo)
sudo systemctl daemon-reload
sudo systemctl restart "$SERVICE_FILE"

echo "Script completed successfully."
