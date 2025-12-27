package webhookout

import (
	"net"
	"net/http"
	"time"
)

func NewHTTPClient(timeout time.Duration) *http.Client {
	// Shared transport for performance; tune as needed
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   5 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	return &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}
}
