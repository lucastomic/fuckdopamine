# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

miniDNS is a lightweight local DNS server for macOS that blocks specified websites by refusing their DNS queries. All other queries are forwarded to Google DNS (8.8.8.8). The server modifies system DNS settings to `127.0.0.1`, listens on UDP port 53, and automatically restores original DNS configuration on exit.

## Build and Run Commands

**Build the binary:**
```bash
go build -o miniDNS
```

**Run from source (requires sudo for port 53 and DNS modification):**
```bash
sudo go run main.go example.com test.com
```

**Run the binary:**
```bash
sudo ./miniDNS example.com test.com
```

**Build release binary:**
```bash
go build -o releases/miniDNS
```

## Architecture

### Core Components

**main.go** - Single-file application containing all logic:

- **DNS Query Handler** (`handleDNSRequest`, line 30): Checks incoming DNS queries against the `forbidden` map. Blocked domains return `dns.RcodeRefused`, others are forwarded to Google DNS.

- **Forbidden Sites Map** (line 16): Global `map[string]any{}` storing blocked domains. Domain format must include trailing dot (e.g., `"example.com."`). Populated from command-line args in `main()` (line 102-105).

- **DNS Forwarder** (`forwardDNSQuery`, line 52): Uses `github.com/miekg/dns` client to forward non-blocked queries to `8.8.8.8:53`.

- **macOS DNS Management** (`backupAndModifyDNSSettings`, `restoreDNSSettings`, lines 73-93): Uses `networksetup` command to modify Wi-Fi DNS settings. Backup is stored and restored via `defer` in `main()`.

- **Uptime Tracking** (lines 20-28, 118-126): Background goroutine prints uptime every 10 seconds using `time.Since(startTime)`.

### Signal Handling

The program uses `signal.Notify` to catch `SIGINT`/`SIGTERM` (line 128-131), ensuring DNS settings are restored via the deferred `restoreDNSSettings()` call before exit.

### Dependencies

- `github.com/miekg/dns`: DNS protocol implementation for server and client operations
- Standard library only (no additional external dependencies)

## Platform Constraints

**macOS-specific**: Uses `networksetup` command with hardcoded "Wi-Fi" interface. To support other platforms or network interfaces, modify DNS configuration logic in `backupAndModifyDNSSettings()` and `restoreDNSSettings()`.

## Known Issues

- DNS restoration uses silent error handling (line 90-92), which may leave DNS settings in incorrect state if restoration fails
- Hardcoded to "Wi-Fi" interface only
- No logging for forwarded/blocked queries (only uptime is displayed)
