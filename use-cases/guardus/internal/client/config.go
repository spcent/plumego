// Package client provides probe helpers for endpoint monitoring.
//
// Compared to upstream gatus, OAuth2/IAP/SSH-tunnel client features and
// SSH/SCTP/UDP probes are out of scope for guardus v1.
package client

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"time"
)

const defaultTimeout = 10 * time.Second

var (
	ErrInvalidDNSResolver     = errors.New("invalid DNS resolver specified. Required format is {proto}://{ip}:{port}")
	ErrInvalidDNSResolverPort = errors.New("invalid DNS resolver port")
	ErrInvalidClientTLSConfig = errors.New("invalid TLS configuration: certificate-file and private-key-file must be specified")

	defaultConfig = Config{
		Insecure:       false,
		IgnoreRedirect: false,
		Timeout:        defaultTimeout,
		Network:        "ip",
	}
)

// GetDefaultConfig returns a copy of the default configuration.
func GetDefaultConfig() *Config {
	cfg := defaultConfig
	return &cfg
}

// Config is the configuration used by the probe clients.
type Config struct {
	ProxyURL       string        `json:"proxy-url,omitempty"`
	Insecure       bool          `json:"insecure,omitempty"`
	IgnoreRedirect bool          `json:"ignore-redirect,omitempty"`
	Timeout        time.Duration `json:"timeout"`
	DNSResolver    string        `json:"dns-resolver,omitempty"`
	// Network is "ip", "ip4" or "ip6"; only the ICMP probe consults it.
	Network string     `json:"network,omitempty"`
	TLS     *TLSConfig `json:"tls,omitempty"`

	httpClient *http.Client
}

// UnmarshalJSON accepts Timeout as either a duration string ("10s") or a
// numeric nanosecond value, mirroring the YAML ergonomics this struct used
// to provide.
func (c *Config) UnmarshalJSON(data []byte) error {
	type configAlias Config
	aux := struct {
		Timeout any `json:"timeout,omitempty"`
		*configAlias
	}{configAlias: (*configAlias)(c)}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	switch v := aux.Timeout.(type) {
	case nil:
		c.Timeout = 0
	case string:
		if v == "" {
			c.Timeout = 0
			return nil
		}
		d, err := time.ParseDuration(v)
		if err != nil {
			return fmt.Errorf("client timeout: %w", err)
		}
		c.Timeout = d
	case float64:
		c.Timeout = time.Duration(v)
	default:
		return fmt.Errorf("client timeout: unsupported type %T", v)
	}
	return nil
}

// DNSResolverConfig is the parsed configuration from the DNSResolver string.
type DNSResolverConfig struct {
	Protocol string
	Host     string
	Port     string
}

// TLSConfig is the configuration for mTLS.
type TLSConfig struct {
	CertificateFile      string `json:"certificate-file,omitempty"`
	PrivateKeyFile       string `json:"private-key-file,omitempty"`
	RenegotiationSupport string `json:"renegotiation,omitempty"`
}

// ValidateAndSetDefaults validates the client configuration.
func (c *Config) ValidateAndSetDefaults() error {
	if c.Timeout < time.Millisecond {
		c.Timeout = defaultTimeout
	}
	if c.HasCustomDNSResolver() {
		if _, err := c.parseDNSResolver(); err != nil {
			return err
		}
	}
	if c.HasTLSConfig() {
		if err := c.TLS.isValid(); err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) HasCustomDNSResolver() bool { return len(c.DNSResolver) > 0 }
func (c *Config) HasTLSConfig() bool {
	return c.TLS != nil && len(c.TLS.CertificateFile) > 0 && len(c.TLS.PrivateKeyFile) > 0
}

func (c *Config) parseDNSResolver() (*DNSResolverConfig, error) {
	re := regexp.MustCompile(`^(?P<proto>(.*))://(?P<host>[A-Za-z0-9\-\.]+):(?P<port>[0-9]+)?(.*)$`)
	matches := re.FindStringSubmatch(c.DNSResolver)
	if len(matches) == 0 {
		return nil, ErrInvalidDNSResolver
	}
	r := make(map[string]string)
	for i, k := range re.SubexpNames() {
		if i != 0 && k != "" {
			r[k] = matches[i]
		}
	}
	port, err := strconv.Atoi(r["port"])
	if err != nil {
		return nil, err
	}
	if port < 1 || port > 65535 {
		return nil, ErrInvalidDNSResolverPort
	}
	return &DNSResolverConfig{Protocol: r["proto"], Host: r["host"], Port: r["port"]}, nil
}

func (t *TLSConfig) isValid() error {
	if len(t.CertificateFile) > 0 && len(t.PrivateKeyFile) > 0 {
		if _, err := tls.LoadX509KeyPair(t.CertificateFile, t.PrivateKeyFile); err != nil {
			return err
		}
		return nil
	}
	return ErrInvalidClientTLSConfig
}

func configureTLS(tlsConfig *tls.Config, c TLSConfig) *tls.Config {
	clientTLSCert, err := tls.LoadX509KeyPair(c.CertificateFile, c.PrivateKeyFile)
	if err != nil {
		return tlsConfig
	}
	tlsConfig.Certificates = []tls.Certificate{clientTLSCert}
	tlsConfig.Renegotiation = tls.RenegotiateNever
	support := map[string]tls.RenegotiationSupport{
		"once":   tls.RenegotiateOnceAsClient,
		"freely": tls.RenegotiateFreelyAsClient,
		"never":  tls.RenegotiateNever,
	}
	if v, ok := support[c.RenegotiationSupport]; ok {
		tlsConfig.Renegotiation = v
	}
	return tlsConfig
}

func (c *Config) getHTTPClient() *http.Client {
	if c.httpClient != nil {
		return c.httpClient
	}
	tlsConfig := &tls.Config{InsecureSkipVerify: c.Insecure}
	if c.HasTLSConfig() && c.TLS.isValid() == nil {
		tlsConfig = configureTLS(tlsConfig, *c.TLS)
	}
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		Proxy:               http.ProxyFromEnvironment,
		TLSClientConfig:     tlsConfig,
	}
	if c.ProxyURL != "" {
		if proxyURL, err := url.Parse(c.ProxyURL); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}
	if c.HasCustomDNSResolver() {
		if resolverCfg, err := c.parseDNSResolver(); err == nil {
			dialer := &net.Dialer{
				Resolver: &net.Resolver{
					PreferGo: true,
					Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
						d := net.Dialer{}
						return d.DialContext(ctx, resolverCfg.Protocol, resolverCfg.Host+":"+resolverCfg.Port)
					},
				},
			}
			transport.DialContext = dialer.DialContext
		}
	}
	c.httpClient = &http.Client{
		Timeout:   c.Timeout,
		Transport: transport,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			if c.IgnoreRedirect {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}
	return c.httpClient
}
