package stripe

import (
	"net/http"
	"strings"
)

// GetSuccessURLFromRequest determines the appropriate success URL based on the request origin
func GetSuccessURLFromRequest(r *http.Request) string {
	// Try to get origin from the Origin header first
	origin := r.Header.Get("Origin")

	// If Origin header is not present, try Referer
	if origin == "" {
		referer := r.Header.Get("Referer")
		if referer != "" {
			// Extract origin from referer URL
			if strings.HasPrefix(referer, "https://") {
				parts := strings.Split(referer, "/")
				if len(parts) >= 3 {
					origin = parts[0] + "//" + parts[2]
				}
			} else if strings.HasPrefix(referer, "http://") {
				parts := strings.Split(referer, "/")
				if len(parts) >= 3 {
					origin = parts[0] + "//" + parts[2]
				}
			}
		}
	}

	// Map known domains to their success URLs
	// Normalize origin by removing trailing slash
	origin = strings.TrimSuffix(origin, "/")

	switch {
	case strings.Contains(origin, "risesportscomplex.com"):
		return "https://www.risesportscomplex.com/success"
	case strings.Contains(origin, "rise-basketball.com"):
		return "https://www.rise-basketball.com/success"
	case strings.Contains(origin, "riseup-hoops.com"):
		return "https://www.riseup-hoops.com/success"
	default:
		// Default to rise-basketball.com if origin is unknown or localhost
		return "https://www.rise-basketball.com/success"
	}
}
