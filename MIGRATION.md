# Migration Guide: miniDNS v1 to v2

This guide helps you migrate from the standalone version of miniDNS to the new daemon-based architecture.

## What's Changed?

### Old Architecture (v1)
- Single binary that runs in the foreground
- Required `sudo ./miniDNS example.com test.com` to start
- Blocked sites specified as command-line arguments
- Stopped when you close the terminal

### New Architecture (v2)
- **Daemon** (`minidnsd`) runs in the background 24/7
- **CLI client** (`miniDNS`) shows dashboard without sudo
- Blocked sites configured in `/etc/minidns/config.json`
- Starts automatically on boot
- Survives terminal closure and reboots

## Migration Steps

### 1. Stop Old Version

If you have the old version running:

```bash
# Press Ctrl+C to stop it
# Or kill it if running in background
pkill miniDNS
```

### 2. Save Your Blocked Sites

If you were running the old version with specific sites, note them down:

```bash
# Example: sudo ./miniDNS facebook.com instagram.com twitter.com
```

### 3. Install New Version

```bash
sudo ./install.sh
```

### 4. Configure Blocked Sites

Edit the config file:

```bash
nano /etc/minidns/config.json
```

Add your blocked sites:

```json
{
  "blocked_sites": [
    "facebook.com",
    "instagram.com",
    "twitter.com"
  ],
  "log_file_path": "/var/log/minidns/dns_requests.json"
}
```

### 5. Restart Daemon

```bash
sudo launchctl unload /Library/LaunchDaemons/com.minidns.daemon.plist
sudo launchctl load /Library/LaunchDaemons/com.minidns.daemon.plist
```

### 6. View Dashboard

```bash
miniDNS
```

## Key Differences

### Command Changes

| Old Command | New Command | Notes |
|------------|-------------|-------|
| `sudo ./miniDNS site1 site2` | `miniDNS` | Now shows dashboard only |
| N/A | `miniDNS status` | Check if daemon is running |
| Edit command line args | `nano /etc/minidns/config.json` | Config file based |
| Ctrl+C to stop | `sudo launchctl unload ...` | Daemon management |

### File Locations

| Type | Old Location | New Location |
|------|-------------|--------------|
| Binary | Wherever you built it | `/usr/local/bin/miniDNS` (CLI)<br>`/usr/local/bin/minidnsd` (daemon) |
| Config | Command line args | `/etc/minidns/config.json` |
| Logs | `logs/dns_requests.json` | `/var/log/minidns/dns_requests.json` |
| Stats | Not persisted | `/var/lib/minidns/stats.json` |

### Behavioral Changes

1. **Auto-start**: Daemon now starts automatically on boot
2. **Persistence**: Stats and uptime persist across restarts
3. **No sudo for dashboard**: CLI client doesn't need root
4. **Background operation**: No terminal needed
5. **Config-based**: Edit config file instead of command args

## Troubleshooting Migration

### Old Binary Still Running

Check for old processes:

```bash
ps aux | grep miniDNS
sudo pkill miniDNS
```

### DNS Settings Not Working

Reset DNS manually:

```bash
sudo networksetup -setdnsservers Wi-Fi Empty
```

Then reload daemon:

```bash
sudo launchctl load /Library/LaunchDaemons/com.minidns.daemon.plist
```

### Grafana Configuration

If you were using Grafana with the old version, update the log file path in your Grafana config:

**Old:** `./logs/dns_requests.json`
**New:** `/var/log/minidns/dns_requests.json`

## Benefits of New Architecture

1. ✅ **Always Running**: Works 24/7, survives reboots
2. ✅ **Easier Management**: Config file instead of command args
3. ✅ **Better Monitoring**: Real-time dashboard without interrupting service
4. ✅ **Persistent Stats**: Data survives daemon restarts
5. ✅ **Auto-recovery**: Daemon restarts automatically if it crashes
6. ✅ **Proper Logging**: System logs in `/var/log/minidns/`
7. ✅ **Clean Uninstall**: Proper uninstallation script

## Legacy Version

The old standalone version is still available in `main.go` for reference, but it's deprecated. Use the new daemon-based architecture for production use.

## Getting Help

- Check daemon status: `miniDNS status`
- View logs: `tail -f /var/log/minidns/daemon.log`
- Check if loaded: `sudo launchctl list | grep minidns`
- Full docs: See `README.md`

---

**Ready to migrate? Run `sudo ./install.sh` to get started!**
