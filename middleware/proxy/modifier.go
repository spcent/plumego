package proxy

import (
	"net"
	"net/http"
	"strings"
)

// RequestModifier is a function that modifies an outgoing request
type RequestModifier func(*http.Request) error

// ResponseModifier is a function that modifies an incoming response
type ResponseModifier func(*http.Response) error

// Hop-by-hop headers that should not be proxied
// See: https://www.rfc-editor.org/rfc/rfc2616#section-13.5.1
var hopByHopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te",
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

// AddForwardedHeaders adds X-Forwarded-* headers to the request
func AddForwardedHeaders() RequestModifier {
	return func(r *http.Request) error {
		// X-Forwarded-For
		clientIP := r.RemoteAddr
		if host, _, err := net.SplitHostPort(clientIP); err == nil {
			clientIP = host
		}

		if prior := r.Header.Get("X-Forwarded-For"); prior != "" {
			clientIP = prior + ", " + clientIP
		}
		r.Header.Set("X-Forwarded-For", clientIP)

		// X-Real-IP (if not already set)
		if r.Header.Get("X-Real-IP") == "" {
			r.Header.Set("X-Real-IP", clientIP)
		}

		// X-Forwarded-Proto
		proto := "http"
		if r.TLS != nil {
			proto = "https"
		}
		r.Header.Set("X-Forwarded-Proto", proto)

		// X-Forwarded-Host
		if r.Header.Get("X-Forwarded-Host") == "" {
			r.Header.Set("X-Forwarded-Host", r.Host)
		}

		return nil
	}
}

// RemoveHopByHopHeaders removes hop-by-hop headers from the request
func RemoveHopByHopHeaders() RequestModifier {
	return func(r *http.Request) error {
		for _, h := range hopByHopHeaders {
			r.Header.Del(h)
		}

		// Also remove any headers listed in Connection header
		if c := r.Header.Get("Connection"); c != "" {
			for _, f := range strings.Split(c, ",") {
				if f = strings.TrimSpace(f); f != "" {
					r.Header.Del(f)
				}
			}
		}

		return nil
	}
}

// AddHeader adds a header to the request
func AddHeader(key, value string) RequestModifier {
	return func(r *http.Request) error {
		r.Header.Add(key, value)
		return nil
	}
}

// SetHeader sets a header on the request
func SetHeader(key, value string) RequestModifier {
	return func(r *http.Request) error {
		r.Header.Set(key, value)
		return nil
	}
}

// DelHeader removes a header from the request
func DelHeader(key string) RequestModifier {
	return func(r *http.Request) error {
		r.Header.Del(key)
		return nil
	}
}

// AddResponseHeader adds a header to the response
func AddResponseHeader(key, value string) ResponseModifier {
	return func(resp *http.Response) error {
		resp.Header.Add(key, value)
		return nil
	}
}

// SetResponseHeader sets a header on the response
func SetResponseHeader(key, value string) ResponseModifier {
	return func(resp *http.Response) error {
		resp.Header.Set(key, value)
		return nil
	}
}

// DelResponseHeader removes a header from the response
func DelResponseHeader(key string) ResponseModifier {
	return func(resp *http.Response) error {
		resp.Header.Del(key)
		return nil
	}
}

// ChainRequestModifiers chains multiple request modifiers
func ChainRequestModifiers(modifiers ...RequestModifier) RequestModifier {
	return func(r *http.Request) error {
		for _, modifier := range modifiers {
			if err := modifier(r); err != nil {
				return err
			}
		}
		return nil
	}
}

// ChainResponseModifiers chains multiple response modifiers
func ChainResponseModifiers(modifiers ...ResponseModifier) ResponseModifier {
	return func(resp *http.Response) error {
		for _, modifier := range modifiers {
			if err := modifier(resp); err != nil {
				return err
			}
		}
		return nil
	}
}
