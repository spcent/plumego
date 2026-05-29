package client

import (
	"net"
	"os"
	"runtime"
	"time"

	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
)

// Ping sends a single ICMP echo and returns whether it succeeded and the
// round-trip duration. Falls back to unprivileged ICMP via UDP on Linux/Darwin
// when the process is not running as root.
//
// Note: kernel must allow unprivileged ICMP (sysctl net.ipv4.ping_group_range)
// for non-root users. On macOS, unprivileged ICMP is enabled by default.
func Ping(address string, config *Config) (bool, time.Duration) {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	network := pingNetwork(config.Network, true)
	dst, err := net.ResolveIPAddr(resolveNetwork(config.Network), address)
	if err != nil {
		return false, 0
	}
	var listenAddr string
	var protocol int
	if dst.IP.To4() != nil {
		listenAddr = "0.0.0.0"
		protocol = 1 // ICMPv4
	} else {
		listenAddr = "::"
		protocol = 58 // ICMPv6
	}
	conn, err := icmp.ListenPacket(network, listenAddr)
	if err != nil && shouldRunPrivileged() {
		// Try privileged path explicitly
		conn, err = icmp.ListenPacket(pingNetwork(config.Network, false), listenAddr)
	}
	if err != nil {
		return false, 0
	}
	defer conn.Close()

	var msg icmp.Message
	if protocol == 1 {
		msg = icmp.Message{
			Type: ipv4.ICMPTypeEcho, Code: 0,
			Body: &icmp.Echo{ID: os.Getpid() & 0xffff, Seq: 1, Data: []byte("guardus")},
		}
	} else {
		msg = icmp.Message{
			Type: ipv6.ICMPTypeEchoRequest, Code: 0,
			Body: &icmp.Echo{ID: os.Getpid() & 0xffff, Seq: 1, Data: []byte("guardus")},
		}
	}
	wb, err := msg.Marshal(nil)
	if err != nil {
		return false, 0
	}
	target := &net.UDPAddr{IP: dst.IP, Zone: dst.Zone}
	start := time.Now()
	if _, err := conn.WriteTo(wb, target); err != nil {
		return false, 0
	}
	if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
		return false, 0
	}
	rb := make([]byte, 1500)
	n, _, err := conn.ReadFrom(rb)
	if err != nil {
		return false, 0
	}
	rtt := time.Since(start)
	rm, err := icmp.ParseMessage(protocol, rb[:n])
	if err != nil {
		return false, 0
	}
	switch rm.Type {
	case ipv4.ICMPTypeEchoReply, ipv6.ICMPTypeEchoReply:
		return true, rtt
	}
	return false, 0
}

// pingNetwork returns the icmp.ListenPacket network string. When unprivileged
// is true and the platform supports it, "udp" sockets are used instead of raw
// IP sockets.
func pingNetwork(net string, unprivileged bool) string {
	is6 := net == "ip6"
	if unprivileged && !shouldRunPrivileged() {
		if is6 {
			return "udp6"
		}
		return "udp4"
	}
	if is6 {
		return "ip6:ipv6-icmp"
	}
	return "ip4:icmp"
}

func resolveNetwork(n string) string {
	switch n {
	case "ip4", "ip6":
		return n
	default:
		return "ip"
	}
}

// shouldRunPrivileged reports whether this process should use raw ICMP sockets.
// Windows always requires privileged mode; otherwise we go privileged only when
// running as root.
func shouldRunPrivileged() bool {
	if runtime.GOOS == "windows" {
		return true
	}
	return os.Geteuid() == 0
}
