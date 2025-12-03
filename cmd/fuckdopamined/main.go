package main

import (
	"encoding/json"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/lucastomic/fuckdopamine/pkg/config"
	"github.com/lucastomic/fuckdopamine/pkg/ipc"
	"github.com/lucastomic/fuckdopamine/pkg/stats"
	"github.com/miekg/dns"
)

var (
	forbidden   map[string]bool
	statsData   *stats.Stats
	logFile     *os.File
	logMutex    sync.Mutex
	logFilePath string

	// Pause state
	pauseMutex sync.RWMutex
	isPaused   bool
	pauseUntil time.Time

	// Activity tracking for sparkline (last 60 seconds)
	activityBuffer [60]uint64
	activityIndex  int
	activityMutex  sync.RWMutex
)

// DNSLogEntry represents a DNS request log entry for Grafana
type DNSLogEntry struct {
	Timestamp string `json:"timestamp"`
	Domain    string `json:"domain"`
	Blocked   bool   `json:"blocked"`
	QueryType string `json:"query_type"`
}

func logToFile(domain string, blocked bool, queryType string) {
	if logFile == nil {
		return
	}

	logMutex.Lock()
	defer logMutex.Unlock()

	entry := DNSLogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Domain:    domain,
		Blocked:   blocked,
		QueryType: queryType,
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return
	}

	logFile.Write(jsonData)
	logFile.Write([]byte("\n"))
	logFile.Sync()
}

// recordActivity increments the current second's activity counter
func recordActivity() {
	activityMutex.Lock()
	activityBuffer[activityIndex]++
	activityMutex.Unlock()
}

// getActivityData returns the last 60 seconds of activity in chronological order
func getActivityData() []float64 {
	activityMutex.RLock()
	defer activityMutex.RUnlock()

	result := make([]float64, 60)
	for i := 0; i < 60; i++ {
		// Start from oldest data point (activityIndex+1) and wrap around
		idx := (activityIndex + 1 + i) % 60
		result[i] = float64(activityBuffer[idx])
	}
	return result
}

// startActivityTicker advances the ring buffer index every second
func startActivityTicker() {
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		for range ticker.C {
			activityMutex.Lock()
			activityIndex = (activityIndex + 1) % 60
			activityBuffer[activityIndex] = 0 // Reset the new slot
			activityMutex.Unlock()
		}
	}()
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	// Record activity for sparkline
	recordActivity()

	// Check if pause expired and auto-resume
	checkAndResumePause()

	// Check if currently paused
	pauseMutex.RLock()
	paused := isPaused
	pauseMutex.RUnlock()

	for _, q := range r.Question {
		host := q.Name
		queryType := dns.TypeToString[q.Qtype]

		// Remove trailing dot for display
		cleanDomain := strings.TrimSuffix(host, ".")

		// If paused, allow all requests
		blocked := false
		if !paused {
			// Check if this domain or any parent domain is blocked
			if forbidden[host] {
				// Exact match (e.g., linkedin.com.)
				blocked = true
			} else {
				// Check if it's a subdomain of any blocked domain
				// e.g., perf.linkedin.com. should match linkedin.com.
				for blockedDomain := range forbidden {
					if host != blockedDomain && strings.HasSuffix(host, "."+blockedDomain) {
						blocked = true
						break
					}
				}
			}
		}

		if blocked {
			m.Rcode = dns.RcodeRefused
			statsData.RecordRequest(cleanDomain, true)
			logToFile(cleanDomain, true, queryType)
			w.WriteMsg(m)
			return
		} else {
			resp, err := forwardDNSQuery(r)
			if err == nil {
				m = resp
			}
			statsData.RecordRequest(cleanDomain, false)
			logToFile(cleanDomain, false, queryType)
		}
	}

	w.WriteMsg(m)
}

func forwardDNSQuery(r *dns.Msg) (*dns.Msg, error) {
	c := new(dns.Client)
	resp, _, err := c.Exchange(r, "8.8.8.8:53")
	return resp, err
}

// Pause functions
func pauseBlocking() {
	pauseMutex.Lock()
	defer pauseMutex.Unlock()

	if !isPaused {
		isPaused = true
		pauseUntil = time.Now().Add(10 * time.Minute)
		statsData.StartPause()
		log.Printf("[PAUSE] Blocking paused for 10 minutes until %s", pauseUntil.Format("15:04:05"))
	}
}

func checkAndResumePause() {
	pauseMutex.Lock()
	defer pauseMutex.Unlock()

	if isPaused && time.Now().After(pauseUntil) {
		isPaused = false
		pauseUntil = time.Time{}
		statsData.EndPause()
		log.Println("[PAUSE] Blocking resumed")
	}
}

func isPausedFn() bool {
	pauseMutex.RLock()
	defer pauseMutex.RUnlock()
	return isPaused
}

func pauseUntilFn() time.Time {
	pauseMutex.RLock()
	defer pauseMutex.RUnlock()
	return pauseUntil
}

func startDNSServer() error {
	dns.HandleFunc(".", handleDNSRequest)

	server := &dns.Server{
		Addr: ":53",
		Net:  "udp",
	}

	return server.ListenAndServe()
}

func backupAndModifyDNSSettings() (string, error) {
	out, err := exec.Command("networksetup", "-getdnsservers", "Wi-Fi").Output()
	if err != nil {
		return "", err
	}
	backup := strings.TrimSpace(string(out))

	err = exec.Command("networksetup", "-setdnsservers", "Wi-Fi", "127.0.0.1").Run()
	if err != nil {
		return "", err
	}

	return backup, nil
}

func restoreDNSSettings(backup string) {
	if backup == "" || backup == "There aren't any DNS Servers set on Wi-Fi." {
		exec.Command("networksetup", "-setdnsservers", "Wi-Fi", "Empty").Run()
	} else {
		exec.Command("networksetup", "-setdnsservers", "Wi-Fi", backup).Run()
	}
}

func startIPCServer(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go ipc.HandleConnection(conn, statsData, forbidden, isPausedFn, pauseUntilFn, pauseBlocking, getActivityData)
	}
}

func main() {
	// Setup logging
	logDir := "/var/log/fuckdopamine"
	os.MkdirAll(logDir, 0755)

	daemonLog, err := os.OpenFile(filepath.Join(logDir, "daemon.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		log.SetOutput(daemonLog)
		defer daemonLog.Close()
	}

	log.Println("[STARTUP] fuckdopamine daemon starting...")

	// Start activity tracking ticker for sparkline
	startActivityTicker()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Printf("[CONFIG] Failed to load config: %v, using defaults", err)
		cfg = config.Default()
		if err := config.Save(cfg); err != nil {
			log.Printf("[CONFIG] Failed to save default config: %v", err)
		}
	}

	// Initialize forbidden sites map
	forbidden = make(map[string]bool)
	for _, site := range cfg.BlockedSites {
		forbidden[site+"."] = true
	}
	log.Printf("[CONFIG] Loaded %d blocked sites", len(forbidden))

	// Load or create stats
	statsPath := config.GetStatsPath()
	// Create stats directory if it doesn't exist
	os.MkdirAll(filepath.Dir(statsPath), 0755)

	statsData, err = stats.Load(statsPath)
	if err != nil {
		log.Printf("[STATS] Failed to load stats: %v, starting fresh", err)
		statsData = stats.New()
	}

	// Setup periodic stats saving
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			if err := statsData.Save(statsPath); err != nil {
				log.Printf("[STATS] Failed to save stats: %v", err)
			}
		}
	}()

	// Open log file for DNS requests
	logFilePath = cfg.LogFilePath
	os.MkdirAll(filepath.Dir(logFilePath), 0755)
	logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("[LOG] Failed to open DNS log file: %v", err)
	}
	if logFile != nil {
		defer logFile.Close()
	}

	// Backup and modify DNS settings
	originalDNS, err := backupAndModifyDNSSettings()
	if err != nil {
		log.Fatalf("[DNS] Failed to modify DNS settings: %v", err)
	}
	log.Printf("[DNS] Backed up DNS settings: %s", originalDNS)
	defer func() {
		log.Println("[SHUTDOWN] Restoring DNS settings...")
		restoreDNSSettings(originalDNS)
	}()

	// Remove old socket if it exists
	os.Remove(ipc.SocketPath)

	// Start IPC server
	listener, err := net.Listen("unix", ipc.SocketPath)
	if err != nil {
		log.Fatalf("[IPC] Failed to create Unix socket: %v", err)
	}
	defer listener.Close()
	defer os.Remove(ipc.SocketPath)

	// Make socket accessible
	os.Chmod(ipc.SocketPath, 0666)

	log.Println("[IPC] Starting IPC server...")
	go startIPCServer(listener)

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start DNS server in goroutine
	go func() {
		log.Println("[DNS] Starting DNS server on port 53...")
		if err := startDNSServer(); err != nil {
			log.Fatalf("[DNS] Failed to start DNS server: %v", err)
		}
	}()

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("[SHUTDOWN] Received signal: %v", sig)

	// Save stats before exit
	if err := statsData.Save(statsPath); err != nil {
		log.Printf("[STATS] Failed to save stats on shutdown: %v", err)
	}

	log.Println("[SHUTDOWN] fuckdopamine daemon stopped")
}
