package nettest

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

const (
	defaultWaitTimeout  = 2 * time.Second
	defaultWaitInterval = 20 * time.Millisecond
)

// WriteSelfSignedTLSCertFiles writes a short-lived localhost certificate/key pair
// into temp files and returns their paths.
func WriteSelfSignedTLSCertFiles(t testing.TB, certPattern, keyPattern string) (string, string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate rsa key: %v", err)
	}

	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName: "127.0.0.1",
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}

	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}

	certFile, err := os.CreateTemp("", certPattern)
	if err != nil {
		t.Fatalf("create cert temp file: %v", err)
	}
	keyFile, err := os.CreateTemp("", keyPattern)
	if err != nil {
		t.Fatalf("create key temp file: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Remove(certFile.Name())
		_ = os.Remove(keyFile.Name())
	})

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: der}); err != nil {
		t.Fatalf("write cert pem: %v", err)
	}
	if err := pem.Encode(keyFile, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)}); err != nil {
		t.Fatalf("write key pem: %v", err)
	}
	if err := certFile.Close(); err != nil {
		t.Fatalf("close cert file: %v", err)
	}
	if err := keyFile.Close(); err != nil {
		t.Fatalf("close key file: %v", err)
	}

	return certFile.Name(), keyFile.Name()
}

// WaitForHTTPStatus polls the URL until it returns the expected HTTP status.
func WaitForHTTPStatus(t testing.TB, client *http.Client, url string, status int) {
	t.Helper()

	if client == nil {
		client = &http.Client{Timeout: 100 * time.Millisecond}
	}

	deadline := time.Now().Add(defaultWaitTimeout)
	var lastErr error
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
			if resp.StatusCode == status {
				return
			}
			lastErr = fmt.Errorf("status %d", resp.StatusCode)
		} else {
			lastErr = err
		}
		time.Sleep(defaultWaitInterval)
	}

	t.Fatalf("timed out waiting for %s from %s: %v", http.StatusText(status), url, lastErr)
}
