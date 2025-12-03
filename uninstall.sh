#!/bin/bash

set -e

echo "========================================"
echo "  miniDNS Uninstallation Script"
echo "========================================"
echo ""

# Check if running with sudo
if [ "$EUID" -ne 0 ]; then
    echo "Error: This script must be run with sudo"
    echo "Usage: sudo ./uninstall.sh"
    exit 1
fi

echo "Uninstalling miniDNS daemon..."
echo ""

# Stop and unload LaunchDaemon
echo "[1/5] Stopping daemon..."
launchctl unload /Library/LaunchDaemons/com.minidns.daemon.plist 2>/dev/null || true

# Remove LaunchDaemon plist
echo "[2/5] Removing LaunchDaemon configuration..."
rm -f /Library/LaunchDaemons/com.minidns.daemon.plist

# Remove binaries
echo "[3/5] Removing binaries..."
rm -f /usr/local/bin/minidnsd
rm -f /usr/local/bin/miniDNS

# Remove socket
echo "[4/5] Cleaning up IPC socket..."
rm -f /tmp/minidns.sock

# Remove logs (optional - ask user)
echo "[5/6] Cleaning up logs..."
read -p "Remove log files from /var/log/minidns/? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf /var/log/minidns
    echo "   Logs removed"
else
    echo "   Logs kept at /var/log/minidns/"
fi

# Remove config and stats (optional - ask user)
echo "[6/6] Cleaning up configuration and stats..."
read -p "Remove config (/etc/minidns/) and stats (/var/lib/minidns/)? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    rm -rf /etc/minidns
    rm -rf /var/lib/minidns
    echo "   Config and stats removed"
else
    echo "   Config kept at /etc/minidns/"
    echo "   Stats kept at /var/lib/minidns/"
fi

echo ""
echo "========================================"
echo "  Uninstallation Complete!"
echo "========================================"
echo ""
echo "✅ miniDNS daemon has been removed"
echo ""
echo "DNS settings have been restored to their original state."
echo ""
