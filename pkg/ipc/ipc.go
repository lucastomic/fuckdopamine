package ipc

import (
	"encoding/json"
	"net"
	"time"

	"github.com/lucastomic/miniDNS/pkg/stats"
)

const SocketPath = "/tmp/minidns.sock"

// Request represents a client request
type Request struct {
	Type string `json:"type"` // "get_stats", "ping"
}

// Response represents a server response
type Response struct {
	Type            string             `json:"type"` // "stats", "pong", "error"
	TotalRequests   uint64             `json:"total_requests,omitempty"`
	BlockedRequests uint64             `json:"blocked_requests,omitempty"`
	AllowedRequests uint64             `json:"allowed_requests,omitempty"`
	TopDomains      []stats.DomainInfo `json:"top_domains,omitempty"`
	Uptime          string             `json:"uptime,omitempty"`
	Error           string             `json:"error,omitempty"`
}

// SendRequest sends a request to the daemon and returns the response
func SendRequest(req Request) (*Response, error) {
	conn, err := net.DialTimeout("unix", SocketPath, 2*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Set deadline for operations
	conn.SetDeadline(time.Now().Add(2 * time.Second))

	// Send request
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(req); err != nil {
		return nil, err
	}

	// Receive response
	var resp Response
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// HandleConnection handles a single IPC connection
func HandleConnection(conn net.Conn, s *stats.Stats, blockedSites map[string]bool) {
	defer conn.Close()

	// Set deadline for operations
	conn.SetDeadline(time.Now().Add(5 * time.Second))

	// Decode request
	var req Request
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&req); err != nil {
		sendError(conn, "invalid request")
		return
	}

	// Process request
	var resp Response
	switch req.Type {
	case "get_stats":
		total, blocked, allowed := s.GetCounts()
		topDomains := s.GetTopDomains(10, blockedSites)
		uptime := s.GetUptime()

		resp = Response{
			Type:            "stats",
			TotalRequests:   total,
			BlockedRequests: blocked,
			AllowedRequests: allowed,
			TopDomains:      topDomains,
			Uptime:          formatUptime(uptime),
		}
	case "ping":
		resp = Response{Type: "pong"}
	default:
		sendError(conn, "unknown request type")
		return
	}

	// Send response
	encoder := json.NewEncoder(conn)
	encoder.Encode(resp)
}

func sendError(conn net.Conn, message string) {
	resp := Response{Type: "error", Error: message}
	encoder := json.NewEncoder(conn)
	encoder.Encode(resp)
}

func formatUptime(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return string([]byte{
		byte('0' + hours/10),
		byte('0' + hours%10),
		':',
		byte('0' + minutes/10),
		byte('0' + minutes%10),
		':',
		byte('0' + seconds/10),
		byte('0' + seconds%10),
	})
}
