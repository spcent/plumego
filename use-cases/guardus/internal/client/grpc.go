package client

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

// PerformGRPCHealthCheck dials a gRPC target and runs the standard Health/Check RPC.
func PerformGRPCHealthCheck(address string, useTLS bool, cfg *Config) (bool, string, error, time.Duration) {
	if cfg == nil {
		cfg = GetDefaultConfig()
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	var opts []grpc.DialOption
	if useTLS {
		tlsCfg := &tls.Config{InsecureSkipVerify: cfg.Insecure}
		if cfg.HasTLSConfig() && cfg.TLS.isValid() == nil {
			tlsCfg = configureTLS(tlsCfg, *cfg.TLS)
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	opts = append(opts, grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
		return cfg.dialContext(ctx, "tcp", addr)
	}))

	start := time.Now()
	conn, err := grpc.DialContext(ctx, address, opts...) //nolint:staticcheck // DialContext is still the supported entrypoint for one-shot health checks.
	if err != nil {
		return false, "", err, time.Since(start)
	}
	defer conn.Close()
	resp, err := healthpb.NewHealthClient(conn).Check(ctx, &healthpb.HealthCheckRequest{Service: ""})
	if err != nil {
		return false, "", err, time.Since(start)
	}
	return true, resp.GetStatus().String(), nil, time.Since(start)
}
