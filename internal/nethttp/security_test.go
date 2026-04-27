package http

import (
	"errors"
	"net"
	"strconv"
	"testing"
)

func TestPrivateIPv4Ranges(t *testing.T) {
	want := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
		"0.0.0.0/8",
		"100.64.0.0/10",
		"192.0.0.0/24",
		"192.0.2.0/24",
		"198.18.0.0/15",
		"198.51.100.0/24",
		"203.0.113.0/24",
		"224.0.0.0/4",
		"240.0.0.0/4",
		"255.255.255.255/32",
	}
	if len(privateIPv4Ranges) != len(want) {
		t.Fatalf("privateIPv4Ranges length = %d, want %d", len(privateIPv4Ranges), len(want))
	}
	for i, ipRange := range privateIPv4Ranges {
		got := ipv4RangeCIDR(ipRange)
		if got != want[i] {
			t.Fatalf("privateIPv4Ranges[%d] = %q, want %q", i, got, want[i])
		}
	}
}

func ipv4RangeCIDR(r ipv4Range) string {
	ones, _ := net.IPMask(r.mask[:]).Size()
	return net.IPv4(r.network[0], r.network[1], r.network[2], r.network[3]).String() + "/" + strconv.Itoa(ones)
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		name      string
		ip        string
		isPrivate bool
	}{
		// Private IPv4 ranges
		{"RFC1918 10.x", "10.0.0.1", true},
		{"RFC1918 172.16-31.x", "172.16.0.1", true},
		{"RFC1918 192.168.x", "192.168.1.1", true},
		{"Loopback", "127.0.0.1", true},
		{"Link-local", "169.254.1.1", true},
		{"Current network", "0.0.0.1", true},
		{"Shared address space", "100.64.0.1", true},
		{"Broadcast", "255.255.255.255", true},
		{"Multicast", "224.0.0.1", true},

		// Public IPv4
		{"Public IP 1", "8.8.8.8", false},
		{"Public IP 2", "1.1.1.1", false},
		{"Public IP 3", "151.101.1.1", false},

		// IPv6
		{"IPv6 loopback", "::1", true},
		{"IPv6 link-local", "fe80::1", true},
		{"IPv6 unique local", "fc00::1", true},
		{"IPv6 unique local fd", "fd00::1", true},
		{"IPv6 public", "2001:4860:4860::8888", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ip := net.ParseIP(tc.ip)
			if ip == nil {
				t.Fatalf("invalid IP: %s", tc.ip)
			}

			got := isPrivateIP(ip)
			if got != tc.isPrivate {
				t.Fatalf("isPrivateIP(%s) = %v, want %v", tc.ip, got, tc.isPrivate)
			}
		})
	}
}

func TestValidateURL_Scheme(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		protection SSRFProtection
		wantErr    bool
	}{
		{
			name: "HTTP allowed by default",
			url:  "http://example.com",
			protection: SSRFProtection{
				AllowedSchemes:    []string{"http", "https"},
				SkipDNSResolution: true,
			},
			wantErr: false,
		},
		{
			name: "HTTPS allowed by default",
			url:  "https://example.com",
			protection: SSRFProtection{
				AllowedSchemes:    []string{"http", "https"},
				SkipDNSResolution: true,
			},
			wantErr: false,
		},
		{
			name: "FTP blocked",
			url:  "ftp://example.com",
			protection: SSRFProtection{
				AllowedSchemes:    []string{"http", "https"},
				SkipDNSResolution: true,
			},
			wantErr: true,
		},
		{
			name: "File protocol blocked",
			url:  "file:///etc/passwd",
			protection: SSRFProtection{
				AllowedSchemes:    []string{"http", "https"},
				SkipDNSResolution: true,
			},
			wantErr: true,
		},
		{
			name: "Gopher blocked",
			url:  "gopher://example.com",
			protection: SSRFProtection{
				AllowedSchemes:    []string{"http", "https"},
				SkipDNSResolution: true,
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateURL(tc.url, tc.protection)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateURL() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateURL_PrivateIP(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		protection SSRFProtection
		wantErr    bool
	}{
		{
			name: "Localhost blocked",
			url:  "http://localhost",
			protection: SSRFProtection{
				BlockLoopback:  true,
				AllowedSchemes: []string{"http"},
			},
			wantErr: true,
		},
		{
			name: "127.0.0.1 blocked",
			url:  "http://127.0.0.1",
			protection: SSRFProtection{
				BlockLoopback:  true,
				AllowedSchemes: []string{"http"},
			},
			wantErr: true,
		},
		{
			name: "Private IP 10.x blocked",
			url:  "http://10.0.0.1",
			protection: SSRFProtection{
				BlockPrivateIPs: true,
				AllowedSchemes:  []string{"http"},
			},
			wantErr: true,
		},
		{
			name: "Private IP 192.168.x blocked",
			url:  "http://192.168.1.1",
			protection: SSRFProtection{
				BlockPrivateIPs: true,
				AllowedSchemes:  []string{"http"},
			},
			wantErr: true,
		},
		{
			name: "Link-local blocked",
			url:  "http://169.254.1.1",
			protection: SSRFProtection{
				BlockLinkLocal: true,
				AllowedSchemes: []string{"http"},
			},
			wantErr: true,
		},
		{
			name: "Public IP allowed",
			url:  "http://8.8.8.8",
			protection: SSRFProtection{
				BlockPrivateIPs: true,
				BlockLoopback:   true,
				AllowedSchemes:  []string{"http"},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateURL(tc.url, tc.protection)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateURL() error = %v, wantErr %v", err, tc.wantErr)
			}
			if tc.wantErr && err != nil {
				if !errors.Is(err, ErrSSRFDetected) && !errors.Is(err, ErrPrivateIP) {
					t.Fatalf("expected SSRF error, got %v", err)
				}
			}
		})
	}
}

func TestValidateURL_Allowlist(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		protection SSRFProtection
		wantErr    bool
	}{
		{
			name: "Allowed host accepted",
			url:  "https://api.example.com/endpoint",
			protection: SSRFProtection{
				AllowedHosts:      []string{"api.example.com"},
				SkipDNSResolution: true,
				AllowedSchemes:    []string{"https"},
			},
			wantErr: false,
		},
		{
			name: "Allowed host accepts trailing dot",
			url:  "https://api.example.com./endpoint",
			protection: SSRFProtection{
				AllowedHosts:      []string{"api.example.com"},
				SkipDNSResolution: true,
				AllowedSchemes:    []string{"https"},
			},
			wantErr: false,
		},
		{
			name: "Not in allowlist rejected",
			url:  "https://evil.com/endpoint",
			protection: SSRFProtection{
				AllowedHosts:      []string{"api.example.com"},
				SkipDNSResolution: true,
				AllowedSchemes:    []string{"https"},
			},
			wantErr: true,
		},
		{
			name: "Multiple allowed hosts",
			url:  "https://api2.example.com/endpoint",
			protection: SSRFProtection{
				AllowedHosts:      []string{"api.example.com", "api2.example.com"},
				SkipDNSResolution: true,
				AllowedSchemes:    []string{"https"},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateURL(tc.url, tc.protection)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateURL() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateURL_Blocklist(t *testing.T) {
	tests := []struct {
		name       string
		url        string
		protection SSRFProtection
		wantErr    bool
	}{
		{
			name: "Blocked host rejected",
			url:  "http://metadata.google.internal",
			protection: SSRFProtection{
				BlockedHosts:      []string{"metadata.google.internal"},
				SkipDNSResolution: true,
				AllowedSchemes:    []string{"http"},
			},
			wantErr: true,
		},
		{
			name: "Blocked host rejects trailing dot",
			url:  "http://metadata.google.internal.",
			protection: SSRFProtection{
				BlockedHosts:      []string{"metadata.google.internal"},
				SkipDNSResolution: true,
				AllowedSchemes:    []string{"http"},
			},
			wantErr: true,
		},
		{
			name: "AWS metadata blocked",
			url:  "http://169.254.169.254/latest/meta-data/",
			protection: SSRFProtection{
				BlockedHosts:   []string{"169.254.169.254"},
				AllowedSchemes: []string{"http"},
			},
			wantErr: true,
		},
		{
			name: "Normal host allowed",
			url:  "http://example.com",
			protection: SSRFProtection{
				BlockedHosts:      []string{"metadata.google.internal"},
				SkipDNSResolution: true,
				AllowedSchemes:    []string{"http"},
			},
			wantErr: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateURL(tc.url, tc.protection)
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateURL() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestValidateURL_UserinfoRejected(t *testing.T) {
	err := ValidateURL("https://169.254.169.254@api.example.com/endpoint", SSRFProtection{
		AllowedHosts:      []string{"api.example.com"},
		SkipDNSResolution: true,
		AllowedSchemes:    []string{"https"},
	})
	if !errors.Is(err, ErrSSRFDetected) {
		t.Fatalf("expected ErrSSRFDetected for userinfo URL, got %v", err)
	}
}

func TestDefaultSSRFProtection(t *testing.T) {
	protection := DefaultSSRFProtection()

	if !protection.BlockPrivateIPs {
		t.Fatal("default should block private IPs")
	}
	if !protection.BlockLoopback {
		t.Fatal("default should block loopback")
	}
	if !protection.BlockLinkLocal {
		t.Fatal("default should block link-local")
	}
	if len(protection.AllowedSchemes) != 2 {
		t.Fatal("default should allow http and https")
	}
}

func TestValidateURL_EmptyURL(t *testing.T) {
	err := ValidateURL("", SSRFProtection{})
	if !errors.Is(err, ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL for empty URL, got %v", err)
	}
}

func TestValidateURL_MalformedURL(t *testing.T) {
	err := ValidateURL("://invalid", SSRFProtection{})
	if !errors.Is(err, ErrInvalidURL) {
		t.Fatalf("expected ErrInvalidURL for malformed URL, got %v", err)
	}
}
