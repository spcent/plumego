package client

import (
	"encoding/hex"
	"fmt"
	"net"
	"strings"

	"github.com/miekg/dns"
)

const dnsPort = 53

// QueryDNS sends a DNS query and returns the response.
func QueryDNS(queryType, queryName, server string) (connected bool, dnsRcode string, body []byte, err error) {
	if !strings.Contains(server, ":") {
		server = fmt.Sprintf("%s:%d", server, dnsPort)
	}
	queryTypeAsUint16 := dns.StringToType[queryType]
	if queryTypeAsUint16 == dns.TypePTR &&
		!strings.HasSuffix(queryName, ".in-addr.arpa.") &&
		!strings.HasSuffix(queryName, ".ip6.arpa.") {
		rev, convErr := reverseNameForIP(queryName)
		if convErr != nil {
			return false, "", nil, convErr
		}
		queryName = rev
	}
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(queryName, queryTypeAsUint16)
	r, _, err := c.Exchange(m, server)
	if err != nil {
		return false, "", nil, err
	}
	connected = true
	dnsRcode = dns.RcodeToString[r.Rcode]
	for _, rr := range r.Answer {
		switch rr.Header().Rrtype {
		case dns.TypeA:
			if a, ok := rr.(*dns.A); ok {
				body = []byte(a.A.String())
			}
		case dns.TypeAAAA:
			if aaaa, ok := rr.(*dns.AAAA); ok {
				body = []byte(aaaa.AAAA.String())
			}
		case dns.TypeCNAME:
			if cname, ok := rr.(*dns.CNAME); ok {
				body = []byte(cname.Target)
			}
		case dns.TypeMX:
			if mx, ok := rr.(*dns.MX); ok {
				body = []byte(mx.Mx)
			}
		case dns.TypeNS:
			if ns, ok := rr.(*dns.NS); ok {
				body = []byte(ns.Ns)
			}
		case dns.TypePTR:
			if ptr, ok := rr.(*dns.PTR); ok {
				body = []byte(ptr.Ptr)
			}
		case dns.TypeSRV:
			if srv, ok := rr.(*dns.SRV); ok {
				body = []byte(fmt.Sprintf("%s:%d", srv.Target, srv.Port))
			}
		default:
			body = []byte("query type is not supported yet")
		}
	}
	return connected, dnsRcode, body, nil
}

func reverseNameForIP(ipStr string) (string, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return "", fmt.Errorf("invalid IP: %s", ipStr)
	}
	if ipv4 := ip.To4(); ipv4 != nil {
		parts := strings.Split(ipv4.String(), ".")
		for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
			parts[i], parts[j] = parts[j], parts[i]
		}
		return strings.Join(parts, ".") + ".in-addr.arpa.", nil
	}
	ip = ip.To16()
	hexStr := hex.EncodeToString(ip)
	nibbles := strings.Split(hexStr, "")
	for i, j := 0, len(nibbles)-1; i < j; i, j = i+1, j-1 {
		nibbles[i], nibbles[j] = nibbles[j], nibbles[i]
	}
	return strings.Join(nibbles, ".") + ".ip6.arpa.", nil
}
