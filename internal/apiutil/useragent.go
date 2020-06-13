package apiutil

import (
	"fmt"

	"github.com/mssola/user_agent"
)

// FormatUserAgent formats a user agent string into a readable os/browser pair.
func FormatUserAgent(s string) string {
	ua := user_agent.New(s)
	osInfo := ua.OSInfo()
	browserName, browserVersion := ua.Browser()
	if osInfo.Name != "" {
		return fmt.Sprintf("%s %s (%s %s)", browserName, browserVersion, osInfo.Name, osInfo.Version)
	}
	return s
}
