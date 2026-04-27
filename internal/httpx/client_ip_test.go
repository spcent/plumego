package httpx

import (
	"net/http"
	"testing"
)

func TestClientIP(t *testing.T) {
	cases := []struct {
		name       string
		xff        string
		xri        string
		remoteAddr string
		want       string
	}{
		{
			name: "X-Forwarded-For single IP",
			xff:  "203.0.113.1",
			want: "203.0.113.1",
		},
		{
			name: "X-Forwarded-For multi-IP chain picks first",
			xff:  "203.0.113.1, 10.0.0.1, 172.16.0.1",
			want: "203.0.113.1",
		},
		{
			name: "X-Forwarded-For with leading space",
			xff:  " 203.0.113.2 , 10.0.0.1",
			want: "203.0.113.2",
		},
		{
			name: "X-Forwarded-For skips malformed entries",
			xff:  "unknown, 203.0.113.3",
			want: "203.0.113.3",
		},
		{
			name:       "Malformed forwarded headers fall back to RemoteAddr",
			xff:        "unknown",
			xri:        "also-unknown",
			remoteAddr: "192.0.2.9:54321",
			want:       "192.0.2.9",
		},
		{
			name: "X-Real-IP used when XFF absent",
			xri:  "198.51.100.5",
			want: "198.51.100.5",
		},
		{
			name: "IPv6 X-Forwarded-For accepted",
			xff:  "2001:db8::1",
			want: "2001:db8::1",
		},
		{
			name:       "RemoteAddr fallback host:port",
			remoteAddr: "192.0.2.7:54321",
			want:       "192.0.2.7",
		},
		{
			name:       "RemoteAddr fallback bare IP",
			remoteAddr: "192.0.2.8",
			want:       "192.0.2.8",
		},
		{
			name: "nil request returns empty string",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "nil request returns empty string" {
				got := ClientIP(nil)
				if got != "" {
					t.Errorf("ClientIP(nil) = %q, want %q", got, "")
				}
				return
			}

			r := &http.Request{
				Header:     make(http.Header),
				RemoteAddr: tc.remoteAddr,
			}
			if tc.xff != "" {
				r.Header.Set("X-Forwarded-For", tc.xff)
			}
			if tc.xri != "" {
				r.Header.Set("X-Real-IP", tc.xri)
			}

			got := ClientIP(r)
			if got != tc.want {
				t.Errorf("ClientIP() = %q, want %q", got, tc.want)
			}
		})
	}
}
