package security

import (
	"net"
	"net/http"
	"strings"

	"github.com/ktappdev/cicd-thing/internal/config"
)

// Middleware provides security middleware functions
type Middleware struct {
	config *config.Config
}

// New creates a new security middleware instance
func New(cfg *config.Config) *Middleware {
	return &Middleware{
		config: cfg,
	}
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

// RateLimitMiddleware provides basic rate limiting (placeholder for future implementation)
func (m *Middleware) RateLimitMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement rate limiting if needed
		next(w, r)
	}
}
