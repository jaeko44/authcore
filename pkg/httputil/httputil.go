package httputil

import (
	"net/url"
	"strings"
)

// GetIPAddrFromXFF gets the IP address from X-Forwarded-For header.
// It splits the X-Forwarded-For header and takes the last part and trims it.
func GetIPAddrFromXFF(xff string) string {
	// Suppose that IP_X_Y is the ip address of X identified by Y.
	// If Cloudflare is used, the X-Forwarded-For header should be in the format
	// "user-input,IP_user_CF" (normally it would be "user-input,IP_user_CF,IP_CF_nginx",
	// however IP_CF_nginx is being stripped).
	// Otherwise, the X-Forwarded-For header is in the format "user-input,IP_user_nginx".
	// If we are trusting the IP address given from nginx (in which we do), then the correct
	// address should be on the last segment.
	ips := strings.Split(xff, ",")
	return strings.Trim(ips[len(ips)-1], " ")
}

// NormalizeURI returns the normalized URI.
func NormalizeURI(uri string) (string, error) {
	uriObj, err := url.Parse(uri)
	if err != nil {
		return "", err
	}
	normalizedURI := uriObj.ResolveReference(&url.URL{}).String()
	return normalizedURI, nil
}
