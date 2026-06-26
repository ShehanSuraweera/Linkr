package ua

import "strings"

// ParseDevice classifies a User-Agent string into one of three device categories.
// Tablet detection precedes mobile so iPad/Android-tablet UA strings are not
// misclassified as mobile.
func ParseDevice(ua string) string {
	lower := strings.ToLower(ua)
	if strings.Contains(lower, "ipad") ||
		(strings.Contains(lower, "android") && !strings.Contains(lower, "mobile")) {
		return "tablet"
	}
	if strings.Contains(lower, "mobi") ||
		strings.Contains(lower, "iphone") ||
		strings.Contains(lower, "ipod") {
		return "mobile"
	}
	return "desktop"
}

// ParseBrowser identifies the browser family from a User-Agent string.
// Order matters: Edge and OPR tokens must be tested before Chrome/Safari
// because those browsers include "Chrome" and "Safari" in their UA strings.
func ParseBrowser(ua string) string {
	switch {
	case strings.Contains(ua, "Edg/") || strings.Contains(ua, "Edge/"):
		return "edge"
	case strings.Contains(ua, "OPR/") || strings.Contains(ua, "Opera"):
		return "opera"
	case strings.Contains(ua, "Firefox/"):
		return "firefox"
	case strings.Contains(ua, "Chrome/"):
		return "chrome"
	case strings.Contains(ua, "Safari/"):
		return "safari"
	default:
		return "other"
	}
}

// ParseRefererDomain extracts the registrable hostname from a Referer URL.
// Returns "direct" for empty or unparseable values so the category is always
// a non-empty string suitable for storage.
func ParseRefererDomain(referer string) string {
	if referer == "" {
		return "direct"
	}
	s := referer
	// Strip scheme (http:// or https://)
	if i := strings.Index(s, "://"); i >= 0 {
		s = s[i+3:]
	}
	// Strip path, query, fragment
	if i := strings.IndexAny(s, "/?#"); i >= 0 {
		s = s[:i]
	}
	// Strip port
	if i := strings.LastIndex(s, ":"); i >= 0 {
		s = s[:i]
	}
	s = strings.TrimPrefix(strings.ToLower(s), "www.")
	if s == "" {
		return "direct"
	}
	return s
}
