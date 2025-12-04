package stripe

import (
	"net/http"
	"strings"
)

// GetSuccessURLFromRequest determines the appropriate success URL based on the request origin
func GetSuccessURLFromRequest(r *http.Request) string {
	origin := getOriginFromRequest(r)
	return getSuccessURLForOrigin(origin)
}

// GetCancelURLFromRequest determines the appropriate cancel URL based on the request origin
// This URL is where users are redirected when they abort the checkout process
func GetCancelURLFromRequest(r *http.Request) string {
	origin := getOriginFromRequest(r)
	return getCancelURLForOrigin(origin)
}

// GetCheckoutURLs returns both success and cancel URLs for a checkout session
func GetCheckoutURLs(r *http.Request) (successURL, cancelURL string) {
	origin := getOriginFromRequest(r)
	return getSuccessURLForOrigin(origin), getCancelURLForOrigin(origin)
}

// getOriginFromRequest extracts the origin from request headers
func getOriginFromRequest(r *http.Request) string {
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

	// Normalize origin by removing trailing slash
	return strings.TrimSuffix(origin, "/")
}

// getSuccessURLForOrigin maps origin to success URL
func getSuccessURLForOrigin(origin string) string {
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

// getCancelURLForOrigin maps origin to cancel URL
// Users are redirected here when they click "back" or close the Stripe checkout
func getCancelURLForOrigin(origin string) string {
	switch {
	case strings.Contains(origin, "risesportscomplex.com"):
		return "https://www.risesportscomplex.com/checkout/canceled"
	case strings.Contains(origin, "rise-basketball.com"):
		return "https://www.rise-basketball.com/checkout/canceled"
	case strings.Contains(origin, "riseup-hoops.com"):
		return "https://www.riseup-hoops.com/checkout/canceled"
	default:
		// Default to rise-basketball.com if origin is unknown or localhost
		return "https://www.rise-basketball.com/checkout/canceled"
	}
}
