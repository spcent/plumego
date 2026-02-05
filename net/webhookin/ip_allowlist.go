package webhookin

import (
	"fmt"
	"net"
	"strings"
)

// IPAllowlist holds a list of IPs/CIDRs that are allowed.
type IPAllowlist struct {
	ips  []net.IP
	nets []*net.IPNet
}

// NewIPAllowlist parses IP/CIDR entries into a matcher.
func NewIPAllowlist(entries []string) (*IPAllowlist, error) {
	allow := &IPAllowlist{}
	for _, raw := range entries {
		item := strings.TrimSpace(raw)
		if item == "" {
			continue
		}
		if strings.Contains(item, "/") {
			_, cidr, err := net.ParseCIDR(item)
			if err != nil {
				return nil, fmt.Errorf("invalid cidr %q: %w", item, err)
			}
			allow.nets = append(allow.nets, cidr)
			continue
		}
		ip := net.ParseIP(item)
		if ip == nil {
			return nil, fmt.Errorf("invalid ip %q", item)
		}
		allow.ips = append(allow.ips, ip)
	}
	return allow, nil
}

// Allow reports whether the provided IP string is allowed.
func (a *IPAllowlist) Allow(ip string) bool {
	if a == nil {
		return true
	}
	parsed := net.ParseIP(strings.TrimSpace(ip))
	if parsed == nil {
		return false
	}
	return a.AllowIP(parsed)
}

// AllowIP reports whether the provided IP is allowed.
func (a *IPAllowlist) AllowIP(ip net.IP) bool {
	if a == nil {
		return true
	}
	if ip == nil {
		return false
	}
	for _, allow := range a.ips {
		if allow.Equal(ip) {
			return true
		}
	}
	for _, cidr := range a.nets {
		if cidr.Contains(ip) {
			return true
		}
	}
	return false
}
