# miniDNS

**A lightweight DNS-based website blocker that runs 24/7 on macOS**

miniDNS is a background service that blocks access to distracting websites by intercepting DNS queries at the system level. Perfect for productivity, focus time, parental controls, or breaking social media habits.

Unlike browser extensions that can be easily disabled, miniDNS runs as a system daemon with root privileges, making it more resistant to circumvention. All DNS queries are handled locally, with blocked sites returning a DNS refusal and all other queries forwarded to Google DNS (8.8.8.8).

> **Warning:**
> This tool modifies your system DNS settings. While it safely restores them on uninstall, improper use could temporarily disrupt your internet connection. Requires administrator privileges for installation.

---

## What Can You Do With miniDNS?

- **🎯 Stay Focused:** Block social media during work hours
- **⏰ Build Better Habits:** Limit access to time-wasting websites
- **👨‍👩‍👧 Parental Controls:** Block inappropriate content on family devices
- **📊 Track Usage:** Monitor which sites are being accessed
- **⏸️ Temporary Access:** Pause blocking for 10 minutes when needed
- **📈 Analyze Patterns:** View statistics on blocked vs. allowed requests

---

## Features

- ✅ **Always-On Protection:** Runs continuously as a macOS LaunchDaemon, starting automatically on boot
- ✅ **System-Level Blocking:** Blocks all DNS queries, works across all browsers and applications
- ✅ **Subdomain Blocking:** Automatically blocks all subdomains (e.g., blocking `reddit.com` also blocks `old.reddit.com`)
- ✅ **Pause Feature:** Temporarily disable blocking for 10 minutes when you need access
- ✅ **Real-Time Dashboard:** Beautiful terminal UI showing live statistics
- ✅ **Persistent Statistics:** Track total pauses, blocking time, and request patterns
- ✅ **Easy Configuration:** Simple JSON file for managing blocked sites
- ✅ **Grafana Integration:** Export logs for advanced visualization
- ✅ **No Sudo Required:** Dashboard and pause command work without admin privileges

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
sudo mkdir -p /etc/minidns
sudo mkdir -p /var/lib/minidns
sudo mkdir -p /var/log/minidns

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
- Current pause status (active or paused with countdown)
- Current uptime
- Total, blocked, and allowed request counts
- Pause statistics (total pauses, time blocking, time paused)
- Top 10 most requested domains
- Real-time updates every second

**Press `q` or `Ctrl+C` to exit the dashboard**

### Check Status

Check if the daemon is running:

```bash
miniDNS status
```

### Pause Blocking

Temporarily pause blocking for 10 minutes:

```bash
miniDNS pause
```

This allows access to all blocked sites for 10 minutes, then automatically resumes blocking. The pause statistics are tracked persistently and displayed in the dashboard.

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
   - If paused, all queries are allowed through
   - Blocked sites receive `REFUSED` response
   - Other queries forwarded to Google DNS (8.8.8.8)
   - All requests logged and counted

3. **Pause Mechanism:**
   - Pause command sets 10-minute timer via IPC
   - Daemon allows all DNS queries during pause
   - Auto-resumes blocking when timer expires
   - Pause statistics tracked and persisted

4. **Dashboard:**
   - CLI connects to Unix socket
   - Fetches stats from daemon
   - Renders interactive UI
   - Updates every second

5. **On Shutdown:**
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

**Note:** Configuration and statistics are kept. Remove manually if desired:
```bash
sudo rm -rf /etc/minidns
sudo rm -rf /var/lib/minidns
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
