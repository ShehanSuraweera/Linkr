package service

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ValidateURL checks scheme, structure, and guards against SSRF targets.
func ValidateURL(raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fmt.Errorf("url is required")
	}

	u, err := url.ParseRequestURI(raw)
	if err != nil {
		return fmt.Errorf("url is malformed: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("url scheme must be http or https, got %q", u.Scheme)
	}

	if u.Host == "" {
		return fmt.Errorf("url must have a host")
	}

	host := u.Hostname()
	if err := blockPrivateHosts(host); err != nil {
		return err
	}

	return nil
}

// blockPrivateHosts rejects localhost and RFC-1918/link-local/loopback ranges
// to prevent the shortener from being used as an SSRF relay.
func blockPrivateHosts(host string) error {
	// Reject bare hostnames that are obviously local.
	lower := strings.ToLower(host)
	if lower == "localhost" || strings.HasSuffix(lower, ".local") ||
		strings.HasSuffix(lower, ".internal") {
		return fmt.Errorf("url target is not allowed")
	}

	ips, err := net.LookupHost(host)
	if err != nil {
		// If DNS fails we can't verify; allow it (fail open to avoid false positives).
		// A production system might fail closed or use a safe DNS resolver.
		return nil
	}

	for _, ipStr := range ips {
		ip := net.ParseIP(ipStr)
		if ip == nil {
			continue
		}
		if isPrivateIP(ip) {
			return fmt.Errorf("url target resolves to a private address")
		}
	}
	return nil
}

var privateRanges = []net.IPNet{
	parseCIDR("10.0.0.0/8"),
	parseCIDR("172.16.0.0/12"),
	parseCIDR("192.168.0.0/16"),
	parseCIDR("127.0.0.0/8"),
	parseCIDR("::1/128"),
	parseCIDR("169.254.0.0/16"),   // link-local (AWS metadata, etc.)
	parseCIDR("fe80::/10"),        // IPv6 link-local
	parseCIDR("fc00::/7"),         // IPv6 ULA
	parseCIDR("100.64.0.0/10"),    // CGNAT
}

func isPrivateIP(ip net.IP) bool {
	for _, block := range privateRanges {
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func parseCIDR(s string) net.IPNet {
	_, network, _ := net.ParseCIDR(s)
	return *network
}
