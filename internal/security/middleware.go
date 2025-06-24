package security

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ktappdev/cicd-thing/internal/config"
)

// Middleware provides security middleware functions
type Middleware struct {
	config *config.Config
	rateLimiter *RateLimiter
}

// RateLimiter provides simple in-memory rate limiting
type RateLimiter struct {
	mu       sync.RWMutex
	clients  map[string]*ClientInfo
	cleanup  time.Duration
}

// ClientInfo tracks request information for a client
type ClientInfo struct {
	requests []time.Time
	lastSeen time.Time
}

// New creates a new security middleware instance
func New(cfg *config.Config) *Middleware {
	return &Middleware{
		config: cfg,
		rateLimiter: NewRateLimiter(),
	}
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*ClientInfo),
		cleanup: 5 * time.Minute,
	}
	
	// Start cleanup goroutine
	go rl.cleanupRoutine()
	
	return rl
}

// IPAllowlistMiddleware checks if the request IP is in the allowlist
func (m *Middleware) IPAllowlistMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// If no allowlist is configured, allow all IPs
		if len(m.config.IPAllowlist) == 0 {
			next(w, r)
			return
		}

		clientIP := getClientIP(r)
		if !m.isIPAllowed(clientIP) {
			http.Error(w, "Forbidden: IP not allowed", http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

// AuthMiddleware checks API key authentication
func (m *Middleware) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		expectedAuth := "Bearer " + m.config.APIKey

		if authHeader != expectedAuth {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		next(w, r)
	}
}

// getClientIP extracts the client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

// isIPAllowed checks if an IP is in the allowlist
func (m *Middleware) isIPAllowed(clientIP string) bool {
	for _, allowedIP := range m.config.IPAllowlist {
		if m.matchesIPOrCIDR(clientIP, allowedIP) {
			return true
		}
	}
	return false
}

// matchesIPOrCIDR checks if an IP matches an IP address or CIDR block
func (m *Middleware) matchesIPOrCIDR(clientIP, allowedIP string) bool {
	// Check if it's a CIDR block
	if strings.Contains(allowedIP, "/") {
		_, ipNet, err := net.ParseCIDR(allowedIP)
		if err != nil {
			return false
		}
		ip := net.ParseIP(clientIP)
		return ip != nil && ipNet.Contains(ip)
	}

	// Direct IP comparison
	return clientIP == allowedIP
}

// RateLimitMiddleware provides rate limiting for log viewing
// Allows 30 requests per minute per IP (reasonable for human log viewing)
func (m *Middleware) RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clientIP := getClientIP(r)
		
		if !m.rateLimiter.Allow(clientIP, 30, time.Minute) {
			http.Error(w, "Rate limit exceeded. Please wait before making more requests.", http.StatusTooManyRequests)
			return
		}
		
		next(w, r)
	}
}

// Allow checks if a client is allowed to make a request
func (rl *RateLimiter) Allow(clientIP string, maxRequests int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	
	now := time.Now()
	client, exists := rl.clients[clientIP]
	
	if !exists {
		client = &ClientInfo{
			requests: make([]time.Time, 0),
			lastSeen: now,
		}
		rl.clients[clientIP] = client
	}
	
	client.lastSeen = now
	
	// Remove old requests outside the window
	cutoff := now.Add(-window)
	validRequests := make([]time.Time, 0)
	for _, reqTime := range client.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	client.requests = validRequests
	
	// Check if under limit
	if len(client.requests) >= maxRequests {
		return false
	}
	
	// Add current request
	client.requests = append(client.requests, now)
	return true
}

// cleanupRoutine removes old client entries
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(rl.cleanup)
	defer ticker.Stop()
	
	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-2 * time.Hour) // Remove clients not seen for 2 hours
		
		for ip, client := range rl.clients {
			if client.lastSeen.Before(cutoff) {
				delete(rl.clients, ip)
			}
		}
		rl.mu.Unlock()
	}
}
