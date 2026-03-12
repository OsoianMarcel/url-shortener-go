package httputil

import (
	"net"
	"net/http"
	"strings"
)

// GetRealIP extracts the client's real IP from the request.
func GetRealIP(r *http.Request, trustProxy bool) string {
	if trustProxy {
		// Cloudflare header
		if cfIP := r.Header.Get("CF-Connecting-IP"); cfIP != "" {
			return strings.TrimSpace(cfIP)
		}

		// Standard X-Forwarded-For header
		if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
			ips := strings.Split(xff, ",")
			if len(ips) > 0 {
				return strings.TrimSpace(ips[0])
			}
		}

		// Standard X-Real-IP header
		if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
			return strings.TrimSpace(xrip)
		}

	}

	// Fallback to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
