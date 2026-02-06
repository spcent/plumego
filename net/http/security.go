package http

import (
	"errors"
	"fmt"
	"net"
	"net/url"
	"strings"
)

var (
	// ErrSSRFDetected is returned when a request is blocked due to SSRF protection.
	ErrSSRFDetected = errors.New("ssrf: request blocked")

	// ErrPrivateIP is returned when a request targets a private IP address.
	ErrPrivateIP = errors.New("ssrf: private ip address blocked")

	// ErrInvalidURL is returned when a URL is malformed or empty.
	ErrInvalidURL = errors.New("invalid url")
)

// SSRFProtection defines SSRF protection configuration.
type SSRFProtection struct {
	// BlockPrivateIPs blocks requests to private IP addresses
	// (127.0.0.0/8, 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, ::1, fe80::/10, etc.)
	BlockPrivateIPs bool

	// BlockLoopback blocks requests to localhost/127.0.0.1/::1
	BlockLoopback bool

	// BlockLinkLocal blocks link-local addresses (169.254.0.0/16, fe80::/10)
	BlockLinkLocal bool

	// AllowedHosts is a whitelist of allowed hostnames (e.g., ["api.example.com"])
	// If set, only these hosts are allowed
	AllowedHosts []string

	// BlockedHosts is a blacklist of blocked hostnames (e.g., ["metadata.google.internal"])
	// Takes precedence over AllowedHosts
	BlockedHosts []string

	// AllowedSchemes restricts URL schemes (e.g., ["https"])
	// If empty, defaults to ["http", "https"]
	AllowedSchemes []string

	// SkipDNSResolution skips DNS resolution and IP checking.
	// This is useful for testing or when DNS resolution is not desired.
	// WARNING: When enabled, private IP protection is disabled.
	SkipDNSResolution bool
}

// DefaultSSRFProtection returns a secure default SSRF protection config.
func DefaultSSRFProtection() SSRFProtection {
	return SSRFProtection{
		BlockPrivateIPs: true,
		BlockLoopback:   true,
		BlockLinkLocal:  true,
		AllowedSchemes:  []string{"http", "https"},
	}
}

// isPrivateIP checks if an IP address is private.
// This prevents SSRF attacks targeting internal services.
func isPrivateIP(ip net.IP) bool {
	if ip == nil {
		return false
	}

	// Check for private IPv4 ranges
	privateIPv4Ranges := []string{
		"10.0.0.0/8",         // RFC1918
		"172.16.0.0/12",      // RFC1918
		"192.168.0.0/16",     // RFC1918
		"127.0.0.0/8",        // Loopback
		"169.254.0.0/16",     // Link-local
		"0.0.0.0/8",          // Current network
		"100.64.0.0/10",      // Shared address space (RFC6598)
		"192.0.0.0/24",       // IETF Protocol Assignments
		"192.0.2.0/24",       // TEST-NET-1
		"198.18.0.0/15",      // Benchmarking
		"198.51.100.0/24",    // TEST-NET-2
		"203.0.113.0/24",     // TEST-NET-3
		"224.0.0.0/4",        // Multicast
		"240.0.0.0/4",        // Reserved
		"255.255.255.255/32", // Broadcast
	}

	for _, cidr := range privateIPv4Ranges {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if ipNet.Contains(ip) {
			return true
		}
	}

	// Check for private IPv6 ranges
	if ip.To4() == nil {
		if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return true
		}
		// Check for unique local addresses (fc00::/7)
		if len(ip) == net.IPv6len && ip[0] == 0xfc || ip[0] == 0xfd {
			return true
		}
	}

	return false
}

// ValidateURL validates a URL against SSRF protection rules.
// This is the core SSRF防护function that should be called before making any HTTP request.
func ValidateURL(urlStr string, protection SSRFProtection) error {
	if urlStr == "" {
		return ErrInvalidURL
	}

	// Parse URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidURL, err)
	}

	// Validate scheme
	allowedSchemes := protection.AllowedSchemes
	if len(allowedSchemes) == 0 {
		allowedSchemes = []string{"http", "https"}
	}
	schemeAllowed := false
	for _, scheme := range allowedSchemes {
		if strings.EqualFold(u.Scheme, scheme) {
			schemeAllowed = true
			break
		}
	}
	if !schemeAllowed {
		return fmt.Errorf("%w: scheme %q not allowed", ErrSSRFDetected, u.Scheme)
	}

	// Extract hostname
	hostname := u.Hostname()
	if hostname == "" {
		return fmt.Errorf("%w: empty hostname", ErrInvalidURL)
	}

	// Check blocked hosts first (takes precedence)
	for _, blocked := range protection.BlockedHosts {
		if strings.EqualFold(hostname, blocked) {
			return fmt.Errorf("%w: host %q is blocked", ErrSSRFDetected, hostname)
		}
	}

	// Check allowed hosts whitelist
	if len(protection.AllowedHosts) > 0 {
		allowed := false
		for _, allowedHost := range protection.AllowedHosts {
			if strings.EqualFold(hostname, allowedHost) {
				allowed = true
				break
			}
		}
		if !allowed {
			return fmt.Errorf("%w: host %q not in allowlist", ErrSSRFDetected, hostname)
		}
	}

	// Skip DNS resolution if configured (e.g., for testing)
	if protection.SkipDNSResolution {
		return nil
	}

	// Resolve hostname to IP
	ips, err := net.LookupIP(hostname)
	if err != nil {
		return fmt.Errorf("ssrf: failed to resolve hostname: %w", err)
	}

	// Check each resolved IP
	for _, ip := range ips {
		// Check loopback
		if protection.BlockLoopback && ip.IsLoopback() {
			return fmt.Errorf("%w: loopback address %s", ErrSSRFDetected, ip.String())
		}

		// Check link-local
		if protection.BlockLinkLocal && (ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()) {
			return fmt.Errorf("%w: link-local address %s", ErrSSRFDetected, ip.String())
		}

		// Check private IP
		if protection.BlockPrivateIPs && isPrivateIP(ip) {
			return fmt.Errorf("%w: private ip address %s", ErrPrivateIP, ip.String())
		}
	}

	return nil
}
