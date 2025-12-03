#!/bin/bash

set -e

echo "========================================"
echo "  fuckdopamine Uninstallation Script"
echo "========================================"
echo ""

# Check if running with sudo
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run with sudo"
    echo "Usage: sudo ./uninstall.sh"
    exit 1
fi

echo "Uninstalling fuckdopamine daemon..."
echo ""

# Stop and unload LaunchDaemon
echo "[1/5] Stopping daemon..."
launchctl unload /Library/LaunchDaemons/com.fuckdopamine.daemon.plist 2>/dev/null || true

# Remove LaunchDaemon plist
echo "[2/5] Removing LaunchDaemon configuration..."
rm -f /Library/LaunchDaemons/com.fuckdopamine.daemon.plist

# Remove binaries
echo "[3/5] Removing binaries..."
rm -f /usr/local/bin/fuckdopamined
rm -f /usr/local/bin/fuckdopamine

# Remove socket
echo "[4/5] Cleaning up IPC socket..."
rm -f /tmp/fuckdopamine.sock

# Remove logs (optional - ask user)
echo "[5/6] Cleaning up logs..."
read -p "Remove log files from /var/log/fuckdopamine/? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf /var/log/fuckdopamine
    echo "   Logs removed"
else
    echo "   Logs kept at /var/log/fuckdopamine/"
fi

# Remove config and stats (optional - ask user)
echo "[6/6] Cleaning up configuration and stats..."
read -p "Remove config (/etc/fuckdopamine/) and stats (/var/lib/fuckdopamine/)? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf /etc/fuckdopamine
    rm -rf /var/lib/fuckdopamine
    echo "   Config and stats removed"
else
    echo "   Config kept at /etc/fuckdopamine/"
    echo "   Stats kept at /var/lib/fuckdopamine/"
fi

echo ""
echo "========================================"
echo "  Uninstallation Complete!"
echo "========================================"
echo ""
echo "✅ fuckdopamine daemon has been removed"
echo ""
echo "DNS settings have been restored to their original state."
echo ""
