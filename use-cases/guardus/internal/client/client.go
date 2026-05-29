package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"
)

// injectedHTTPClient is used for testing.
var injectedHTTPClient *http.Client

// InjectHTTPClient overrides the HTTP client used by GetHTTPClient. Test-only.
func InjectHTTPClient(c *http.Client) { injectedHTTPClient = c }

// GetHTTPClient returns the HTTP client matching the supplied config.
func GetHTTPClient(config *Config) *http.Client {
	if injectedHTTPClient != nil {
		return injectedHTTPClient
	}
	if config == nil {
		return defaultConfig.getHTTPClient()
	}
	return config.getHTTPClient()
}

// CanCreateNetworkConnection probes a TCP/UDP target. Returns whether the
// connection succeeded and the response bytes if a request body was supplied.
func CanCreateNetworkConnection(netType, address, body string, config *Config) (bool, []byte) {
	const maximumMessageSize = 1024
	conn, err := net.DialTimeout(netType, address, config.Timeout)
	if err != nil {
		return false, nil
	}
	defer conn.Close()
	if body != "" {
		body = parseLocalAddressPlaceholder(body, conn.LocalAddr())
		_ = conn.SetDeadline(time.Now().Add(config.Timeout))
		if _, err := conn.Write([]byte(body)); err != nil {
			return false, nil
		}
		buf := make([]byte, maximumMessageSize)
		n, err := conn.Read(buf)
		if err != nil {
			return false, nil
		}
		return true, buf[:n]
	}
	return true, nil
}

// CanPerformStartTLS opens a STARTTLS handshake and returns the server cert.
func CanPerformStartTLS(address string, config *Config) (bool, *x509.Certificate, error) {
	hostAndPort := strings.Split(address, ":")
	if len(hostAndPort) != 2 {
		return false, nil, errors.New("invalid address for starttls, format must be host:port")
	}
	conn, err := net.DialTimeout("tcp", address, config.Timeout)
	if err != nil {
		return false, nil, err
	}
	smtpClient, err := smtp.NewClient(conn, hostAndPort[0])
	if err != nil {
		return false, nil, err
	}
	if err := smtpClient.StartTLS(&tls.Config{
		InsecureSkipVerify: config.Insecure,
		ServerName:         hostAndPort[0],
	}); err != nil {
		return false, nil, err
	}
	state, ok := smtpClient.TLSConnectionState()
	if !ok {
		return false, nil, errors.New("could not get TLS connection state")
	}
	return true, state.PeerCertificates[0], nil
}

// CanPerformTLS opens a TLS handshake to address and returns the server cert.
func CanPerformTLS(address, body string, config *Config) (bool, []byte, *x509.Certificate, error) {
	const maximumMessageSize = 1024
	conn, err := tls.DialWithDialer(&net.Dialer{Timeout: config.Timeout}, "tcp", address, &tls.Config{
		InsecureSkipVerify: config.Insecure,
	})
	if err != nil {
		return false, nil, nil, err
	}
	defer conn.Close()
	verifiedChains := conn.ConnectionState().VerifiedChains
	var cert *x509.Certificate
	if len(verifiedChains) == 0 || len(verifiedChains[0]) == 0 {
		peerCertificates := conn.ConnectionState().PeerCertificates
		cert = peerCertificates[0]
	} else {
		cert = verifiedChains[0][0]
	}
	var response []byte
	if body != "" {
		body = parseLocalAddressPlaceholder(body, conn.LocalAddr())
		_ = conn.SetDeadline(time.Now().Add(config.Timeout))
		if _, err := conn.Write([]byte(body)); err != nil {
			return false, nil, cert, err
		}
		buf := make([]byte, maximumMessageSize)
		n, err := conn.Read(buf)
		if err != nil {
			return false, nil, cert, err
		}
		response = buf[:n]
	}
	return true, response, cert, nil
}

// dialContext returns a contextual dialer that respects the config's custom
// DNS resolver if one is set.
func (c *Config) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if c != nil && c.HasCustomDNSResolver() {
		if resolverCfg, err := c.parseDNSResolver(); err == nil {
			d := &net.Dialer{
				Timeout: c.Timeout,
				Resolver: &net.Resolver{
					PreferGo: true,
					Dial: func(ctx context.Context, _, _ string) (net.Conn, error) {
						d := net.Dialer{}
						return d.DialContext(ctx, resolverCfg.Protocol, resolverCfg.Host+":"+resolverCfg.Port)
					},
				},
			}
			return d.DialContext(ctx, network, address)
		}
	}
	d := &net.Dialer{Timeout: c.Timeout}
	return d.DialContext(ctx, network, address)
}

func parseLocalAddressPlaceholder(s string, localAddr net.Addr) string {
	return strings.ReplaceAll(s, "[LOCAL_ADDRESS]", localAddr.String())
}
