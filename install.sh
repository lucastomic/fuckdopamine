#!/bin/bash

set -e

echo "========================================"
echo "  fuckdopamine Installation Script"
echo "========================================"
echo ""

# Check if running with sudo
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run with sudo"
    echo "Usage: sudo ./install.sh"
    exit 1
fi

echo "Installing fuckdopamine daemon..."
echo ""

# Build binaries
echo "[1/8] Building daemon binary..."
go build -o /usr/local/bin/fuckdopamined ./cmd/fuckdopamined

echo "[2/8] Building CLI binary..."
go build -o /usr/local/bin/fuckdopamine ./cmd/fuckdopamine

# Set permissions
echo "[3/8] Setting permissions..."
chmod +x /usr/local/bin/fuckdopamined
chmod +x /usr/local/bin/fuckdopamine

# Create log directory
echo "[4/8] Creating log directory..."
mkdir -p /var/log/fuckdopamine
chmod 755 /var/log/fuckdopamine

# Create stats directory
echo "[5/8] Creating stats directory..."
mkdir -p /var/lib/fuckdopamine
chmod 755 /var/lib/fuckdopamine

# Create config directory
echo "[6/8] Setting up configuration..."
CONFIG_DIR="/etc/fuckdopamine"
mkdir -p "$CONFIG_DIR"
chmod 755 "$CONFIG_DIR"

# Create default config if it doesn't exist
CONFIG_FILE="$CONFIG_DIR/config.json"
if [ ! -f "$CONFIG_FILE" ]; then
    cat > "$CONFIG_FILE" <<EOF
{
  "blocked_sites": [
    "example.com"
  ],
  "log_file_path": "/var/log/fuckdopamine/dns_requests.json"
}
EOF
    chmod 644 "$CONFIG_FILE"
    echo "   Created default config at $CONFIG_FILE"
    echo "   Edit this file to configure blocked sites"
else
    echo "   Config already exists at $CONFIG_FILE"
fi

# Install LaunchDaemon
echo "[7/8] Installing LaunchDaemon..."
cp com.fuckdopamine.daemon.plist /Library/LaunchDaemons/
chmod 644 /Library/LaunchDaemons/com.fuckdopamine.daemon.plist
chown root:wheel /Library/LaunchDaemons/com.fuckdopamine.daemon.plist

# Load LaunchDaemon
echo "[8/8] Starting daemon..."
launchctl unload /Library/LaunchDaemons/com.fuckdopamine.daemon.plist 2>/dev/null || true
launchctl load /Library/LaunchDaemons/com.fuckdopamine.daemon.plist

echo ""
echo "========================================"
echo "  Installation Complete!"
echo "========================================"
echo ""
echo "✅ fuckdopamine daemon is now running"
echo ""
echo "Configuration:"
echo "  Config file: /etc/fuckdopamine/config.json"
echo "  Stats file:  /var/lib/fuckdopamine/stats.json"
echo "  Log files:   /var/log/fuckdopamine/"
echo ""
echo "Usage:"
echo "  fuckdopamine              - Show dashboard"
echo "  fuckdopamine status       - Check daemon status"
echo ""
echo "To edit blocked sites:"
echo "  sudo nano /etc/fuckdopamine/config.json"
echo "  sudo launchctl unload /Library/LaunchDaemons/com.fuckdopamine.daemon.plist"
echo "  sudo launchctl load /Library/LaunchDaemons/com.fuckdopamine.daemon.plist"
echo ""
echo "To uninstall:"
echo "  sudo ./uninstall.sh"
echo ""
