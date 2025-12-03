# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

miniDNS is a lightweight DNS server daemon for macOS that runs 24/7 as a background service, blocking specified websites by refusing their DNS queries. All other queries are forwarded to Google DNS (8.8.8.8). The system consists of a daemon that starts automatically on boot and a CLI client for monitoring.

## Architecture

miniDNS has been refactored into a daemon-based architecture with two main components:

### 1. Daemon (minidnsd)
**Location:** `cmd/minidnsd/main.go`

The daemon runs as a macOS LaunchDaemon with root privileges:
- Loads configuration from `/etc/minidns/config.json`
- Backs up and modifies DNS settings to `127.0.0.1`
- Listens on UDP port 53 for DNS queries
- Handles IPC via Unix socket at `/tmp/minidns.sock`
- Logs to `/var/log/minidns/daemon.log` and `/var/log/minidns/dns_requests.json`
- Saves statistics to `/var/lib/minidns/stats.json` every 5 minutes
- Restores DNS settings on shutdown

**Key Functions:**
- `handleDNSRequest()`: Checks queries against `forbidden` map, blocks or forwards
- `forwardDNSQuery()`: Forwards allowed queries to 8.8.8.8
- `backupAndModifyDNSSettings()`: Modifies Wi-Fi DNS to 127.0.0.1
- `restoreDNSSettings()`: Restores original DNS on exit
- `startIPCServer()`: Handles Unix socket connections from CLI
- `logToFile()`: Writes JSON logs for Grafana integration

### 2. CLI Client (miniDNS)
**Location:** `cmd/miniDNS/main.go`

The CLI client provides an interactive dashboard without requiring sudo:
- Connects to daemon via Unix socket
- Displays real-time statistics using termui
- Shows uptime, request counts, and top domains
- Supports `dashboard`, `status`, and `help` commands

**Key Functions:**
- `fetchStats()`: Requests stats from daemon via IPC
- `renderDashboard()`: Renders termui interface
- `checkStatus()`: Pings daemon to verify it's running

### 3. Shared Packages

**pkg/config/config.go:**
- `Config` struct with `BlockedSites` and `LogFilePath`
- `Load()`: Reads config from `/etc/minidns/config.json`
- `Save()`: Writes config to disk
- `Default()`: Returns default configuration

**pkg/stats/stats.go:**
- `Stats` struct tracking total/blocked/allowed requests and domain counts
- `RecordRequest()`: Thread-safe request recording
- `GetTopDomains()`: Returns top N domains with block status
- `Save()`/`Load()`: Persist stats across daemon restarts

**pkg/ipc/ipc.go:**
- Unix socket communication protocol
- `Request` and `Response` structs for IPC messages
- `SendRequest()`: Client-side request sender
- `HandleConnection()`: Server-side connection handler
- Supports "get_stats" and "ping" request types

## Build and Run Commands

**Build both binaries:**
```bash
go build -o minidnsd ./cmd/minidnsd
go build -o miniDNS ./cmd/miniDNS
```

**Install as daemon:**
```bash
sudo ./install.sh
```

**View dashboard:**
```bash
miniDNS
# or
miniDNS dashboard
```

**Check daemon status:**
```bash
miniDNS status
```

**Manage daemon:**
```bash
# Start
sudo launchctl load /Library/LaunchDaemons/com.minidns.daemon.plist

# Stop
sudo launchctl unload /Library/LaunchDaemons/com.minidns.daemon.plist

# View logs
tail -f /var/log/minidns/daemon.log
```

**Uninstall:**
```bash
sudo ./uninstall.sh
```

## Configuration

**Config file location:** `/etc/minidns/config.json`

```json
{
  "blocked_sites": ["example.com", "facebook.com"],
  "log_file_path": "/var/log/minidns/dns_requests.json"
}
```

After editing config, restart daemon:
```bash
sudo launchctl unload /Library/LaunchDaemons/com.minidns.daemon.plist
sudo launchctl load /Library/LaunchDaemons/com.minidns.daemon.plist
```

## File Locations

- **Binaries:** `/usr/local/bin/minidnsd`, `/usr/local/bin/miniDNS`
- **Config:** `/etc/minidns/config.json`
- **Stats:** `/var/lib/minidns/stats.json`
- **LaunchDaemon plist:** `/Library/LaunchDaemons/com.minidns.daemon.plist`
- **Logs:** `/var/log/minidns/` (daemon.log, dns_requests.json, stdout.log, stderr.log)
- **Unix socket:** `/tmp/minidns.sock`
- **Legacy standalone:** `main.go` (deprecated, kept for reference)

## Dependencies

- `github.com/miekg/dns`: DNS protocol implementation for server and client operations
- `github.com/gizak/termui/v3`: Terminal UI for dashboard
- Standard library for Unix sockets, JSON, signals, etc.

## Platform Constraints

**macOS-specific**:
- Uses `networksetup` command with hardcoded "Wi-Fi" interface
- Requires macOS LaunchDaemon system
- To support other platforms, modify:
  - DNS configuration logic in daemon's `backupAndModifyDNSSettings()` and `restoreDNSSettings()`
  - LaunchDaemon system (use systemd for Linux, etc.)

## IPC Protocol

Communication between CLI and daemon uses Unix socket at `/tmp/minidns.sock` with JSON messages:

**Request types:**
- `"ping"`: Health check, returns `"pong"`
- `"get_stats"`: Returns full statistics

**Response fields:**
- `type`: "stats", "pong", or "error"
- `total_requests`, `blocked_requests`, `allowed_requests`: Counters
- `top_domains`: Array of {domain, count, blocked}
- `uptime`: Formatted uptime string (HH:MM:SS)
- `error`: Error message if type is "error"

## Known Behaviors

- Daemon automatically restarts on crash (LaunchDaemon KeepAlive)
- Stats persist across daemon restarts via `/var/lib/minidns/stats.json`
- DNS settings restored on daemon shutdown (graceful or signal-based)
- Logs rotate automatically via Grafana file watching
- Hardcoded to "Wi-Fi" interface only
- Domain format in `forbidden` map includes trailing dot (e.g., `"example.com."`)
