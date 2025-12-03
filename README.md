# miniDNS

miniDNS is a lightweight DNS server daemon for macOS that runs 24/7 in the background, blocking specified websites by refusing DNS queries. All other queries are forwarded to Google DNS (8.8.8.8). The daemon starts automatically on boot and includes a CLI dashboard for real-time monitoring.

> **Warning:**
> Modifying your DNS settings can disrupt your internet connection if not restored properly. Use at your own risk. The current implementation is targeted for macOS (using `networksetup` commands) and requires administrator privileges for installation.

---

## Features

- **Background Daemon:** Runs continuously as a macOS LaunchDaemon, starting automatically on boot
- **DNS Blocking:** Block specific websites by refusing DNS queries for them
- **Real-time Dashboard:** Interactive terminal UI showing statistics and top requested domains
- **Persistent Stats:** Statistics survive daemon restarts
- **Configuration File:** Easy-to-edit JSON configuration for blocked sites
- **Grafana Integration:** JSON logs for visualization with Grafana
- **IPC Communication:** Unix socket-based communication between daemon and CLI

---

## Architecture

miniDNS consists of two components:

1. **minidnsd** - Background daemon that:
   - Runs as root to bind to port 53
   - Modifies DNS settings to 127.0.0.1
   - Handles all DNS queries
   - Exposes stats via Unix socket
   - Logs requests for Grafana

2. **miniDNS** - CLI client that:
   - Connects to the daemon
   - Displays interactive dashboard
   - Shows real-time statistics
   - Requires no sudo privileges

---

## Prerequisites

- **macOS:** Uses `networksetup` command and LaunchDaemon system
- **Go:** Required to build from source (download from [golang.org](https://golang.org/dl/))
- **Administrator Privileges:** Required for installation only

---

## Installation

### Quick Install

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/miniDNS.git
   cd miniDNS
   ```

2. Run the installation script:
   ```bash
   sudo ./install.sh
   ```

The installation script will:
- Build both binaries (`minidnsd` and `miniDNS`)
- Install them to `/usr/local/bin/`
- Create configuration directory at `/etc/minidns/`
- Create a default config file
- Install and start the LaunchDaemon

### Manual Installation

If you prefer to install manually:

```bash
# Build binaries
go build -o /usr/local/bin/minidnsd ./cmd/minidnsd
go build -o /usr/local/bin/miniDNS ./cmd/miniDNS

# Create directories
mkdir -p ~/.minidns
mkdir -p /var/log/minidns

# Create config file (see Configuration section)
# Install LaunchDaemon
sudo cp com.minidns.daemon.plist /Library/LaunchDaemons/
sudo launchctl load /Library/LaunchDaemons/com.minidns.daemon.plist
```

---

## Configuration

Edit the configuration file at `/etc/minidns/config.json`:

```json
{
  "blocked_sites": [
    "example.com",
    "test.com",
    "facebook.com"
  ],
  "log_file_path": "/var/log/minidns/dns_requests.json"
}
```

**After editing the config, restart the daemon:**

```bash
sudo launchctl unload /Library/LaunchDaemons/com.minidns.daemon.plist
sudo launchctl load /Library/LaunchDaemons/com.minidns.daemon.plist
```

---

## Usage

### View Dashboard

Simply run the CLI client:

```bash
miniDNS
```

Or explicitly:

```bash
miniDNS dashboard
```

The dashboard shows:
- Current uptime
- Total, blocked, and allowed request counts
- Top 10 most requested domains
- Real-time updates every second

**Press `q` or `Ctrl+C` to exit the dashboard**

### Check Status

Check if the daemon is running:

```bash
miniDNS status
```

### Daemon Management

Start the daemon:
```bash
sudo launchctl load /Library/LaunchDaemons/com.minidns.daemon.plist
```

Stop the daemon:
```bash
sudo launchctl unload /Library/LaunchDaemons/com.minidns.daemon.plist
```

View daemon logs:
```bash
tail -f /var/log/minidns/daemon.log
```

---

## How It Works

1. **At Boot:**
   - macOS launches `minidnsd` via LaunchDaemon
   - Daemon backs up current DNS settings
   - Sets DNS to `127.0.0.1` (localhost)
   - Starts DNS server on port 53
   - Opens Unix socket at `/tmp/minidns.sock` for IPC

2. **DNS Query Handling:**
   - All DNS queries go to `minidnsd`
   - Blocked sites receive `REFUSED` response
   - Other queries forwarded to Google DNS (8.8.8.8)
   - All requests logged and counted

3. **Dashboard:**
   - CLI connects to Unix socket
   - Fetches stats from daemon
   - Renders interactive UI
   - Updates every second

4. **On Shutdown:**
   - Daemon restores original DNS settings
   - Saves statistics to disk
   - Cleans up socket

---

## File Locations

- **Binaries:** `/usr/local/bin/minidnsd`, `/usr/local/bin/miniDNS`
- **Configuration:** `/etc/minidns/config.json`
- **Statistics:** `/var/lib/minidns/stats.json`
- **LaunchDaemon:** `/Library/LaunchDaemons/com.minidns.daemon.plist`
- **Logs:** `/var/log/minidns/`
  - `daemon.log` - Daemon activity log
  - `dns_requests.json` - Grafana-compatible request logs
  - `stdout.log` / `stderr.log` - Standard streams

---

## Uninstallation

Run the uninstallation script:

```bash
sudo ./uninstall.sh
```

This will:
- Stop and remove the daemon
- Remove binaries
- Clean up socket
- Optionally remove logs
- Restore DNS settings

**Note:** User configuration (`/etc/minidns/`) is kept. Remove manually if desired:
```bash
rm -rf ~/.minidns
```

---

## Troubleshooting

### Daemon Won't Start

Check daemon logs:
```bash
cat /var/log/minidns/daemon.log
cat /var/log/minidns/stderr.log
```

Verify LaunchDaemon is loaded:
```bash
sudo launchctl list | grep minidns
```

### Dashboard Can't Connect

Ensure daemon is running:
```bash
miniDNS status
```

Check socket exists:
```bash
ls -l /tmp/minidns.sock
```

### Port 53 Already in Use

Another service may be using port 53. Check and stop conflicting services.

### DNS Not Working After Uninstall

Manually reset DNS settings:
```bash
sudo networksetup -setdnsservers Wi-Fi Empty
```

---

## Development

### Project Structure

```
miniDNS/
├── cmd/
│   ├── minidnsd/          # Daemon binary
│   │   └── main.go
│   └── miniDNS/           # CLI client binary
│       └── main.go
├── pkg/
│   ├── config/            # Configuration management
│   │   └── config.go
│   ├── stats/             # Statistics tracking
│   │   └── stats.go
│   └── ipc/               # Inter-process communication
│       └── ipc.go
├── com.minidns.daemon.plist  # LaunchDaemon configuration
├── install.sh             # Installation script
├── uninstall.sh           # Uninstallation script
└── main.go               # Legacy standalone version
```

### Building

Build both binaries:
```bash
go build -o minidnsd ./cmd/minidnsd
go build -o miniDNS ./cmd/miniDNS
```

---

## Contributing

Contributions, issues, and feature requests are welcome!
Feel free to check [Issues](https://github.com/yourusername/miniDNS/issues) and [Pull Requests](https://github.com/yourusername/miniDNS/pulls).

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

Enjoy using miniDNS to manage your browsing experience 24/7!
