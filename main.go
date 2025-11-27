package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/miekg/dns"
)

var forbidden = map[string]any{
	"example.com.": 1,
}

var startTime time.Time

// Stats holds all DNS request statistics
type Stats struct {
	mu              sync.RWMutex
	totalRequests   uint64
	blockedRequests uint64
	allowedRequests uint64
	domainCounts    map[string]uint64
}

func newStats() *Stats {
	return &Stats{
		domainCounts: make(map[string]uint64),
	}
}

func (s *Stats) recordRequest(domain string, blocked bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalRequests++
	if blocked {
		s.blockedRequests++
	} else {
		s.allowedRequests++
	}
	s.domainCounts[domain]++
}

func (s *Stats) getTopDomains(n int) []struct {
	domain  string
	count   uint64
	blocked bool
} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	type domainCount struct {
		domain  string
		count   uint64
		blocked bool
	}

	domains := make([]domainCount, 0, len(s.domainCounts))
	for domain, count := range s.domainCounts {
		_, isBlocked := forbidden[domain]
		domains = append(domains, domainCount{domain, count, isBlocked})
	}

	sort.Slice(domains, func(i, j int) bool {
		return domains[i].count > domains[j].count
	})

	if len(domains) > n {
		domains = domains[:n]
	}

	result := make([]struct {
		domain  string
		count   uint64
		blocked bool
	}, len(domains))

	for i, d := range domains {
		result[i].domain = d.domain
		result[i].count = d.count
		result[i].blocked = d.blocked
	}

	return result
}

func (s *Stats) getCounts() (total, blocked, allowed uint64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.totalRequests, s.blockedRequests, s.allowedRequests
}

var stats *Stats

func getUptime() string {
	uptime := time.Since(startTime)
	hours := int(uptime.Hours())
	minutes := int(uptime.Minutes()) % 60
	seconds := int(uptime.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func handleDNSRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)

	for _, q := range r.Question {
		host := q.Name
		domainAndTLD := strings.Join(strings.Split(host, ".")[1:], ".")

		// Remove trailing dot for display purposes
		cleanDomain := strings.TrimSuffix(domainAndTLD, ".")
		if cleanDomain == "" {
			cleanDomain = strings.TrimSuffix(host, ".")
		}

		if _, found := forbidden[domainAndTLD]; found {
			m.Rcode = dns.RcodeRefused
			stats.recordRequest(cleanDomain, true)
			w.WriteMsg(m)
			return
		} else {
			resp, err := forwardDNSQuery(r)
			if err == nil {
				m = resp
			}
			stats.recordRequest(cleanDomain, false)
		}
	}

	w.WriteMsg(m)
}

func forwardDNSQuery(r *dns.Msg) (*dns.Msg, error) {
	c := new(dns.Client)
	resp, _, err := c.Exchange(r, "8.8.8.8:53")
	return resp, err
}

func startDNSServer(done chan struct{}) {
	dns.HandleFunc(".", handleDNSRequest)

	server := &dns.Server{
		Addr: ":53",
		Net:  "udp",
	}

	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start: %v", err)
	}
	close(done)
}

func backupAndModifyDNSSettings() (string, error) {
	out, err := exec.Command("networksetup", "-getdnsservers", "Wi-Fi").Output()
	if err != nil {
		return "", err
	}
	backup := string(out)

	err = exec.Command("networksetup", "-setdnsservers", "Wi-Fi", "127.0.0.1").Run()
	if err != nil {
		return "", err
	}

	return backup, nil
}

func restoreDNSSettings(backup string) {
	err := exec.Command("networksetup", "-setdnsservers", "Wi-Fi", backup).Run()
	if err != nil {
		// Silent error handling
	}
}

func renderDashboard() {
	total, blocked, allowed := stats.getCounts()
	uptime := getUptime()

	// Calculate percentages
	var blockedPct, allowedPct float64
	if total > 0 {
		blockedPct = float64(blocked) / float64(total) * 100
		allowedPct = float64(allowed) / float64(total) * 100
	}

	// Header
	header := widgets.NewParagraph()
	header.Text = fmt.Sprintf("[miniDNS Server - Active](fg:green,mod:bold)\nUptime: %s", uptime)
	header.Border = true
	header.SetRect(0, 0, 60, 4)

	// Stats box
	statsBox := widgets.NewParagraph()
	statsBox.Title = "Statistics"
	statsBox.Text = fmt.Sprintf(
		"[Total Requests:](fg:cyan,mod:bold)      %d\n"+
			"[Blocked:](fg:red,mod:bold)             %d (%.1f%%)\n"+
			"[Allowed:](fg:green,mod:bold)             %d (%.1f%%)",
		total, blocked, blockedPct, allowed, allowedPct,
	)
	statsBox.Border = true
	statsBox.SetRect(0, 4, 60, 9)

	// Top domains list
	topDomains := stats.getTopDomains(10)
	domainList := widgets.NewList()
	domainList.Title = "Top Requested Domains"
	domainList.Border = true

	rows := make([]string, 0, len(topDomains))
	for i, d := range topDomains {
		status := ""
		if d.blocked {
			status = " [BLOCKED](fg:red,mod:bold)"
		}
		rows = append(rows, fmt.Sprintf("%2d. %-30s %6d%s", i+1, d.domain, d.count, status))
	}
	if len(rows) == 0 {
		rows = append(rows, "No requests yet...")
	}
	domainList.Rows = rows
	domainList.TextStyle = ui.NewStyle(ui.ColorWhite)
	domainList.SetRect(0, 9, 60, 21)

	// Footer
	footer := widgets.NewParagraph()
	footer.Text = "[Press Ctrl+C to exit](fg:yellow)"
	footer.Border = false
	footer.SetRect(0, 21, 60, 23)

	ui.Render(header, statsBox, domainList, footer)
}

func startDashboard() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	// Initial render
	renderDashboard()

	// Create ticker for periodic updates
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// Handle events
	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "<Resize>":
				renderDashboard()
			}
		case <-ticker.C:
			renderDashboard()
		}
	}
}

func main() {
	startTime = time.Now()
	stats = newStats()

	if len(os.Args) < 2 {
		log.Fatalf("[CONFIG] No sites provided. Usage: %s <site1> <site2> ...", "miniDNS")
	}

	sites := os.Args[1:]
	for _, site := range sites {
		forbidden[site+"."] = 1
	}

	originalResolConf, err := backupAndModifyDNSSettings()
	if err != nil {
		log.Fatalf("[CONFIG] Failed to modify DNS settings: %v", err)
	}
	defer restoreDNSSettings(originalResolConf)

	dnsDone := make(chan struct{})

	go startDNSServer(dnsDone)
	time.Sleep(100 * time.Millisecond)

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Run dashboard in goroutine so we can catch signals
	dashboardDone := make(chan struct{})
	go func() {
		startDashboard()
		close(dashboardDone)
	}()

	// Wait for either signal or dashboard exit
	select {
	case <-sigChan:
		ui.Close()
	case <-dashboardDone:
	}

	time.Sleep(200 * time.Millisecond)
}
