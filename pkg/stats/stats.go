package stats

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
	"time"
)

// Stats holds all DNS request statistics
type Stats struct {
	mu              sync.RWMutex
	TotalRequests   uint64            `json:"total_requests"`
	BlockedRequests uint64            `json:"blocked_requests"`
	AllowedRequests uint64            `json:"allowed_requests"`
	DomainCounts    map[string]uint64 `json:"domain_counts"`
	StartTime       time.Time         `json:"start_time"`
}

// DomainInfo represents domain statistics with block status
type DomainInfo struct {
	Domain  string `json:"domain"`
	Count   uint64 `json:"count"`
	Blocked bool   `json:"blocked"`
}

// New creates a new Stats instance
func New() *Stats {
	return &Stats{
		DomainCounts: make(map[string]uint64),
		StartTime:    time.Now(),
	}
}

// RecordRequest records a DNS request
func (s *Stats) RecordRequest(domain string, blocked bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalRequests++
	if blocked {
		s.BlockedRequests++
	} else {
		s.AllowedRequests++
	}
	s.DomainCounts[domain]++
}

// GetTopDomains returns the top N most requested domains
func (s *Stats) GetTopDomains(n int, blockedSites map[string]bool) []DomainInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	domains := make([]DomainInfo, 0, len(s.DomainCounts))
	for domain, count := range s.DomainCounts {
		_, isBlocked := blockedSites[domain]
		domains = append(domains, DomainInfo{
			Domain:  domain,
			Count:   count,
			Blocked: isBlocked,
		})
	}

	sort.Slice(domains, func(i, j int) bool {
		return domains[i].Count > domains[j].Count
	})

	if len(domains) > n {
		domains = domains[:n]
	}

	return domains
}

// GetCounts returns total, blocked, and allowed counts
func (s *Stats) GetCounts() (total, blocked, allowed uint64) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.TotalRequests, s.BlockedRequests, s.AllowedRequests
}

// GetUptime returns the uptime duration
func (s *Stats) GetUptime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return time.Since(s.StartTime)
}

// Save saves stats to a file
func (s *Stats) Save(path string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Load loads stats from a file
func Load(path string) (*Stats, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return New(), nil // Return new stats if file doesn't exist
	}

	var s Stats
	if err := json.Unmarshal(data, &s); err != nil {
		return New(), nil
	}

	return &s, nil
}
