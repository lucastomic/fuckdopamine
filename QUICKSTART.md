# miniDNS Quick Start Guide

## Installation (One Command)

```bash
sudo ./install.sh
```

That's it! The daemon is now running 24/7.

---

## Daily Usage

### View Dashboard
```bash
miniDNS
```
Press `q` or `Ctrl+C` to exit.

### Check Status
```bash
miniDNS status
```

---

## Configuration

### Edit Blocked Sites
```bash
nano /etc/minidns/config.json
```

Example config:
```json
{
  "blocked_sites": [
    "facebook.com",
    "twitter.com",
    "instagram.com"
  ],
  "log_file_path": "/var/log/minidns/dns_requests.json"
}
```

### Apply Changes
```bash
sudo launchctl unload /Library/LaunchDaemons/com.minidns.daemon.plist
sudo launchctl load /Library/LaunchDaemons/com.minidns.daemon.plist
```

---

## Daemon Management

### Start
```bash
sudo launchctl load /Library/LaunchDaemons/com.minidns.daemon.plist
```

### Stop
```bash
sudo launchctl unload /Library/LaunchDaemons/com.minidns.daemon.plist
```

### View Logs
```bash
tail -f /var/log/minidns/daemon.log
```

---

## Troubleshooting

### Daemon Not Running?
```bash
miniDNS status
sudo launchctl list | grep minidns
cat /var/log/minidns/daemon.log
```

### Dashboard Can't Connect?
```bash
ls -l /tmp/minidns.sock
sudo launchctl load /Library/LaunchDaemons/com.minidns.daemon.plist
```

### DNS Not Working?
```bash
# Reset DNS manually
sudo networksetup -setdnsservers Wi-Fi Empty
# Then reload daemon
sudo launchctl load /Library/LaunchDaemons/com.minidns.daemon.plist
```

---

## Uninstall

```bash
sudo ./uninstall.sh
```

---

## File Locations

| What | Where |
|------|-------|
| Config | `/etc/minidns/config.json` |
| Stats | `/var/lib/minidns/stats.json` |
| Logs | `/var/log/minidns/` |
| Binaries | `/usr/local/bin/miniDNS`, `/usr/local/bin/minidnsd` |
| Daemon plist | `/Library/LaunchDaemons/com.minidns.daemon.plist` |

---

## Common Tasks

### Add a Site to Block List
1. `nano /etc/minidns/config.json`
2. Add site to `blocked_sites` array
3. Save and exit
4. `sudo launchctl unload /Library/LaunchDaemons/com.minidns.daemon.plist`
5. `sudo launchctl load /Library/LaunchDaemons/com.minidns.daemon.plist`

### View Real-time Stats
```bash
miniDNS
```

### Check Uptime
```bash
miniDNS status
```

### See All DNS Requests
```bash
tail -f /var/log/minidns/dns_requests.json
```

---

**For full documentation, see README.md**
