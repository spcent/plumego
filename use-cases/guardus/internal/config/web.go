package config

import (
	"crypto/tls"
	"errors"
	"fmt"
	"math"
)

const (
	DefaultAddress        = "0.0.0.0"
	DefaultPort           = 8080
	DefaultReadBufferSize = 8192
	MinimumReadBufferSize = 4096
)

// WebConfig describes how the HTTP server listens.
type WebConfig struct {
	Address        string     `json:"address"`
	Port           int        `json:"port"`
	ReadBufferSize int        `json:"read-buffer-size,omitempty"`
	TLS            *TLSConfig `json:"tls,omitempty"`
}

type TLSConfig struct {
	CertificateFile string `json:"certificate-file,omitempty"`
	PrivateKeyFile  string `json:"private-key-file,omitempty"`
}

func defaultWebConfig() *WebConfig {
	return &WebConfig{
		Address:        DefaultAddress,
		Port:           DefaultPort,
		ReadBufferSize: DefaultReadBufferSize,
	}
}

// ValidateAndSetDefaults applies defaults and verifies bounds.
func (w *WebConfig) ValidateAndSetDefaults() error {
	if len(w.Address) == 0 {
		w.Address = DefaultAddress
	}
	if w.Port == 0 {
		w.Port = DefaultPort
	} else if w.Port < 0 || w.Port > math.MaxUint16 {
		return fmt.Errorf("invalid port: value should be between 0 and %d", math.MaxUint16)
	}
	if w.ReadBufferSize == 0 {
		w.ReadBufferSize = DefaultReadBufferSize
	} else if w.ReadBufferSize < MinimumReadBufferSize {
		w.ReadBufferSize = MinimumReadBufferSize
	}
	if w.TLS != nil {
		if err := w.TLS.isValid(); err != nil {
			return fmt.Errorf("invalid tls config: %w", err)
		}
	}
	return nil
}

func (w *WebConfig) HasTLS() bool {
	return w.TLS != nil && len(w.TLS.CertificateFile) > 0 && len(w.TLS.PrivateKeyFile) > 0
}

func (w *WebConfig) SocketAddress() string {
	return fmt.Sprintf("%s:%d", w.Address, w.Port)
}

func (t *TLSConfig) isValid() error {
	if len(t.CertificateFile) == 0 || len(t.PrivateKeyFile) == 0 {
		return errors.New("certificate-file and private-key-file must be specified")
	}
	if _, err := tls.LoadX509KeyPair(t.CertificateFile, t.PrivateKeyFile); err != nil {
		return err
	}
	return nil
}
