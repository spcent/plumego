# API Gateway - Phase 2: Security & Management

Phase 2 adds enterprise-grade security and management features to the API Gateway, building on the observability foundation from Phase 1.

## Overview

Phase 2 introduces:
- **JWT Authentication** - Token-based API authentication
- **Security Headers** - HSTS, CSP, and other security headers
- **TLS/HTTPS** - Encrypted communication
- **Admin API** - Runtime management and monitoring

## Features

### 1. JWT Authentication

Token-based authentication using the Plumego JWT manager.

#### Configuration

```bash
# Enable authentication
AUTH_ENABLED=true

# JWT secret (must be at least 32 characters)
JWT_SECRET=your-secret-key-must-be-at-least-32-characters-long

# Public paths that don't require authentication
AUTH_PUBLIC_PATHS=/health,/status,/metrics

# Token configuration
AUTH_TOKEN_HEADER=Authorization
AUTH_TOKEN_QUERY_PARAM=token

# Token expiration
AUTH_ACCESS_TOKEN_TTL=15    # minutes
AUTH_REFRESH_TOKEN_TTL=7    # days
```

#### How It Works

1. **Public Paths**: Requests to paths in `AUTH_PUBLIC_PATHS` bypass authentication
2. **Token Extraction**: JWT tokens are extracted from the `Authorization` header (Bearer format)
3. **Token Verification**: Tokens are verified using the JWT manager
4. **Claims Injection**: Valid tokens have their claims injected into the request context

#### Middleware Order

JWT authentication middleware runs early in the chain:
```
Security Headers → Access Logging → JWT Auth → Rate Limiting → Timeout → CORS → Routes
```

#### Example Usage

```bash
# Request without authentication (public path)
curl http://localhost:8080/health

# Request with authentication
curl -H "Authorization: Bearer <access_token>" \
     http://localhost:8080/api/v1/users/123
```

#### JWT Manager Details

- Uses in-memory KV store for token state
- Supports token revocation via blacklist
- Automatic key rotation support
- Identity versioning for instant invalidation

### 2. Security Headers

Comprehensive security header policy implementation.

#### Configuration

```bash
# Enable security headers
SECURITY_HEADERS_ENABLED=true

# HSTS (HTTP Strict Transport Security)
SECURITY_HSTS_MAX_AGE=31536000           # 1 year in seconds
SECURITY_HSTS_INCLUDE_SUBDOMAINS=true
SECURITY_HSTS_PRELOAD=false

# Content Security Policy
SECURITY_CSP=default-src 'self'

# Frame Options (DENY, SAMEORIGIN)
SECURITY_X_FRAME_OPTIONS=DENY

# Content Type Options
SECURITY_X_CONTENT_TYPE_OPTIONS=nosniff

# Referrer Policy
SECURITY_REFERRER_POLICY=strict-origin-when-cross-origin
```

#### Headers Applied

| Header | Value | Purpose |
|--------|-------|---------|
| `Strict-Transport-Security` | `max-age=31536000; includeSubDomains` | Force HTTPS |
| `Content-Security-Policy` | `default-src 'self'` | XSS protection |
| `X-Frame-Options` | `DENY` | Clickjacking protection |
| `X-Content-Type-Options` | `nosniff` | MIME sniffing protection |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Referrer control |

#### HSTS Configuration

HTTP Strict Transport Security forces browsers to use HTTPS:

```bash
# Production settings
SECURITY_HSTS_MAX_AGE=31536000        # 1 year
SECURITY_HSTS_INCLUDE_SUBDOMAINS=true # Apply to all subdomains
SECURITY_HSTS_PRELOAD=true            # Submit to HSTS preload list
```

**Warning**: HSTS preload is a one-way operation. Only enable after thoroughly testing HTTPS.

#### CSP Configuration

Content Security Policy prevents XSS and data injection attacks:

```bash
# Restrictive (recommended for APIs)
SECURITY_CSP=default-src 'none'

# Moderate (allow same-origin resources)
SECURITY_CSP=default-src 'self'

# Custom (specific directives)
SECURITY_CSP=default-src 'self'; script-src 'self' 'unsafe-inline'; img-src 'self' data:
```

### 3. TLS/HTTPS Support

Server-side TLS configuration for encrypted communication.

#### Configuration

```bash
# Enable TLS
TLS_ENABLED=true

# Certificate and key files
TLS_CERT_FILE=/path/to/cert.pem
TLS_KEY_FILE=/path/to/key.pem
```

#### Generating Self-Signed Certificates (Development)

```bash
# Generate private key and certificate
openssl req -x509 -newkey rsa:4096 -nodes \
  -keyout key.pem \
  -out cert.pem \
  -days 365 \
  -subj "/CN=localhost"

# Update configuration
TLS_ENABLED=true
TLS_CERT_FILE=./cert.pem
TLS_KEY_FILE=./key.pem
```

#### Production Certificates

Use Let's Encrypt for production certificates:

```bash
# Install certbot
sudo apt-get install certbot

# Obtain certificate
sudo certbot certonly --standalone -d yourdomain.com

# Configure gateway
TLS_ENABLED=true
TLS_CERT_FILE=/etc/letsencrypt/live/yourdomain.com/fullchain.pem
TLS_KEY_FILE=/etc/letsencrypt/live/yourdomain.com/privkey.pem
```

#### Testing TLS

```bash
# Test HTTPS endpoint
curl -k https://localhost:8080/health

# Check TLS configuration
openssl s_client -connect localhost:8080 -showcerts
```

### 4. Admin API

Runtime management and monitoring endpoints with API key authentication.

#### Configuration

```bash
# Enable admin API
ADMIN_ENABLED=true

# Admin API path prefix
ADMIN_PATH=/admin

# Admin API key (minimum 16 characters)
ADMIN_API_KEY=your-admin-api-key-min-16-chars
```

#### Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/admin/stats` | GET | Gateway statistics and feature status |
| `/admin/health` | GET | Detailed health check information |
| `/admin/config` | GET | Current configuration (sanitized) |
| `/admin/reload` | POST | Hot reload configuration (planned) |

#### Authentication

All admin endpoints require the admin API key in the request:

```bash
# Using X-Admin-API-Key header (recommended)
curl -H "X-Admin-API-Key: your-admin-api-key-min-16-chars" \
     http://localhost:8080/admin/stats

# Using Authorization header
curl -H "Authorization: Bearer your-admin-api-key-min-16-chars" \
     http://localhost:8080/admin/stats
```

#### Example Responses

**GET /admin/stats**
```json
{
  "gateway": {
    "addr": ":8080",
    "debug": false,
    "uptime_seconds": 3600.5
  },
  "services": {
    "user": {
      "enabled": true,
      "targets": 2
    },
    "order": {
      "enabled": true,
      "targets": 2
    },
    "product": {
      "enabled": true,
      "targets": 1
    }
  },
  "features": {
    "auth": false,
    "metrics": true,
    "rate_limit": true,
    "cors": true,
    "security_headers": true,
    "tls": false
  }
}
```

**GET /admin/health**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "checks": {
    "server": {
      "status": "ok"
    }
  }
}
```

**GET /admin/config**
```json
{
  "server": {
    "addr": ":8080",
    "debug": false
  },
  "auth": {
    "enabled": false,
    "token_header": "Authorization",
    "access_token_ttl": 15,
    "refresh_token_ttl": 7,
    "public_paths": ["/health", "/status", "/metrics"]
  },
  "security": {
    "enabled": true,
    "hsts_max_age": 31536000,
    "hsts_include_subdomains": true,
    "hsts_preload": false,
    "content_security_policy": "default-src 'self'",
    "x_frame_options": "DENY",
    "x_content_type_options": "nosniff",
    "referrer_policy": "strict-origin-when-cross-origin"
  },
  "tls": {
    "enabled": false
  },
  "metrics": {...},
  "rate_limit": {...},
  "timeouts": {...},
  "cors": {...}
}
```

Note: Sensitive fields (JWT secret, TLS cert/key paths, admin API key) are not exposed in the config endpoint.

## Architecture

### Middleware Stack

Phase 2 middleware is applied in this order:

```
1. Security Headers    ← First (apply to all responses)
2. Access Logging      ← Log all requests
3. JWT Authentication  ← Verify auth tokens
4. Rate Limiting       ← Enforce rate limits
5. Gateway Timeout     ← Apply timeouts
6. CORS               ← Handle CORS
7. Proxy              ← Forward to backend
```

### Security Flow

```
┌─────────────┐
│   Request   │
└──────┬──────┘
       │
       ▼
┌─────────────────┐
│ Security Headers│ ← Apply security policy
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│   Is Public?    │
└────────┬────────┘
         │
    Yes  │  No
    ┌────┴────┐
    │         ▼
    │   ┌──────────────┐
    │   │ Extract Token│
    │   └──────┬───────┘
    │          │
    │          ▼
    │   ┌──────────────┐
    │   │ Verify Token │
    │   └──────┬───────┘
    │          │
    │     Valid│Invalid
    │          │
    │          ▼
    │   ┌──────────────┐
    │   │   401 Error  │
    │   └──────────────┘
    │
    ▼
┌─────────────────┐
│  Continue Chain │
└─────────────────┘
```

## Deployment

### Development Configuration

```bash
# Minimal Phase 2 features for development
AUTH_ENABLED=false
SECURITY_HEADERS_ENABLED=true
TLS_ENABLED=false
ADMIN_ENABLED=true
ADMIN_API_KEY=dev-admin-key-1234567890
```

### Production Configuration

```bash
# Full Phase 2 features for production
AUTH_ENABLED=true
JWT_SECRET=$(openssl rand -base64 32)

SECURITY_HEADERS_ENABLED=true
SECURITY_HSTS_MAX_AGE=31536000
SECURITY_HSTS_INCLUDE_SUBDOMAINS=true
SECURITY_HSTS_PRELOAD=false
SECURITY_CSP=default-src 'self'
SECURITY_X_FRAME_OPTIONS=DENY
SECURITY_X_CONTENT_TYPE_OPTIONS=nosniff
SECURITY_REFERRER_POLICY=strict-origin-when-cross-origin

TLS_ENABLED=true
TLS_CERT_FILE=/etc/letsencrypt/live/yourdomain.com/fullchain.pem
TLS_KEY_FILE=/etc/letsencrypt/live/yourdomain.com/privkey.pem

ADMIN_ENABLED=true
ADMIN_PATH=/admin
ADMIN_API_KEY=$(openssl rand -hex 32)
```

### Docker Deployment

```dockerfile
# Dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /build
COPY . .
RUN go build -o gateway ./examples/api-gateway/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /build/gateway .

# Copy TLS certificates if using TLS
# COPY certs/ /app/certs/

EXPOSE 8080
EXPOSE 8443

CMD ["./gateway"]
```

```yaml
# docker-compose.yml
version: '3.8'

services:
  gateway:
    build: .
    ports:
      - "8080:8080"
      - "8443:8443"
    environment:
      GATEWAY_ADDR: ":8080"

      # Phase 2 features
      AUTH_ENABLED: "true"
      JWT_SECRET: "${JWT_SECRET}"
      AUTH_PUBLIC_PATHS: "/health,/status,/metrics"

      SECURITY_HEADERS_ENABLED: "true"
      SECURITY_HSTS_MAX_AGE: "31536000"
      SECURITY_CSP: "default-src 'self'"

      TLS_ENABLED: "true"
      TLS_CERT_FILE: "/app/certs/cert.pem"
      TLS_KEY_FILE: "/app/certs/key.pem"

      ADMIN_ENABLED: "true"
      ADMIN_API_KEY: "${ADMIN_API_KEY}"

      # Backend services
      USER_SERVICE_TARGETS: "http://user-service:8081"
      ORDER_SERVICE_TARGETS: "http://order-service:9001"
      PRODUCT_SERVICE_TARGETS: "http://product-service:7001"
    volumes:
      - ./certs:/app/certs:ro
    restart: unless-stopped
```

### Kubernetes Deployment

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gateway-config
data:
  AUTH_PUBLIC_PATHS: "/health,/status,/metrics"
  SECURITY_HEADERS_ENABLED: "true"
  SECURITY_HSTS_MAX_AGE: "31536000"
  SECURITY_CSP: "default-src 'self'"
  ADMIN_ENABLED: "true"
  ADMIN_PATH: "/admin"

---
apiVersion: v1
kind: Secret
metadata:
  name: gateway-secrets
type: Opaque
stringData:
  JWT_SECRET: "your-jwt-secret-here"
  ADMIN_API_KEY: "your-admin-api-key-here"

---
apiVersion: v1
kind: Secret
metadata:
  name: gateway-tls
type: kubernetes.io/tls
data:
  tls.crt: <base64-encoded-cert>
  tls.key: <base64-encoded-key>

---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
spec:
  replicas: 3
  selector:
    matchLabels:
      app: api-gateway
  template:
    metadata:
      labels:
        app: api-gateway
    spec:
      containers:
      - name: gateway
        image: plumego-gateway:latest
        ports:
        - containerPort: 8080
          name: http
        - containerPort: 8443
          name: https
        env:
        - name: GATEWAY_ADDR
          value: ":8080"
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: gateway-secrets
              key: JWT_SECRET
        - name: ADMIN_API_KEY
          valueFrom:
            secretKeyRef:
              name: gateway-secrets
              key: ADMIN_API_KEY
        - name: TLS_ENABLED
          value: "true"
        - name: TLS_CERT_FILE
          value: "/etc/tls/tls.crt"
        - name: TLS_KEY_FILE
          value: "/etc/tls/tls.key"
        envFrom:
        - configMapRef:
            name: gateway-config
        volumeMounts:
        - name: tls
          mountPath: /etc/tls
          readOnly: true
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
      volumes:
      - name: tls
        secret:
          secretName: gateway-tls

---
apiVersion: v1
kind: Service
metadata:
  name: api-gateway
spec:
  type: LoadBalancer
  selector:
    app: api-gateway
  ports:
  - name: http
    port: 80
    targetPort: 8080
  - name: https
    port: 443
    targetPort: 8443
```

## Security Best Practices

### 1. JWT Secrets

- Use strong, randomly generated secrets (minimum 32 characters)
- Rotate secrets regularly
- Never commit secrets to version control
- Use environment variables or secret management systems

```bash
# Generate strong JWT secret
JWT_SECRET=$(openssl rand -base64 32)
```

### 2. Admin API Keys

- Use long, random API keys (minimum 16 characters recommended)
- Restrict admin API access to trusted networks
- Log all admin API access
- Rotate API keys periodically

```bash
# Generate strong admin API key
ADMIN_API_KEY=$(openssl rand -hex 32)
```

### 3. TLS Configuration

- Use strong cipher suites
- Enable only TLS 1.2 and 1.3
- Use certificates from trusted CAs
- Implement certificate renewal automation

### 4. Security Headers

- Start with restrictive policies and relax as needed
- Test CSP thoroughly before production deployment
- Enable HSTS preload only after thorough testing
- Monitor for security header violations

### 5. Rate Limiting

- Combine with authentication for better protection
- Set appropriate limits based on expected traffic
- Monitor rate limit violations
- Consider IP-based rate limiting for unauthenticated endpoints

## Monitoring

### Admin API Metrics

Monitor gateway health using the admin API:

```bash
#!/bin/bash
# Monitor gateway health
while true; do
  curl -s -H "X-Admin-API-Key: $ADMIN_API_KEY" \
       http://localhost:8080/admin/stats | jq .
  sleep 30
done
```

### Security Event Logging

Monitor security-related events:

```bash
# Watch for authentication failures
grep "Invalid token" gateway.log | tail -f

# Watch for rate limit violations
grep "Rate limit exceeded" gateway.log | tail -f

# Watch for admin API access
grep "admin" gateway.log | tail -f
```

### Prometheus Metrics

Phase 2 adds security-related metrics:

```promql
# Authentication failure rate
rate(http_requests_total{status="401"}[5m])

# Admin API access rate
rate(http_requests_total{path=~"/admin/.*"}[5m])

# TLS handshake errors (if available)
rate(tls_handshake_errors_total[5m])
```

## Troubleshooting

### JWT Authentication Issues

**Problem**: 401 Unauthorized errors

**Solutions**:
1. Verify JWT secret is correct
2. Check token expiration
3. Ensure token format is correct (Bearer <token>)
4. Verify issuer and audience match configuration
5. Check if path requires authentication

```bash
# Test JWT token
curl -v -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/users

# Decode JWT to check claims
echo "<token>" | cut -d. -f2 | base64 -d | jq .
```

### Security Headers Not Applied

**Problem**: Security headers missing from responses

**Solutions**:
1. Verify `SECURITY_HEADERS_ENABLED=true`
2. Check middleware order (security headers should be first)
3. Ensure response is not being cached
4. Check for conflicting headers from backend services

```bash
# Test security headers
curl -I http://localhost:8080/api/v1/users

# Should see:
# Strict-Transport-Security: max-age=31536000
# Content-Security-Policy: default-src 'self'
# X-Frame-Options: DENY
# X-Content-Type-Options: nosniff
```

### TLS Certificate Issues

**Problem**: TLS handshake failures

**Solutions**:
1. Verify certificate and key files exist and are readable
2. Check certificate validity dates
3. Ensure certificate matches domain
4. Verify certificate chain is complete

```bash
# Check certificate
openssl x509 -in cert.pem -text -noout

# Test TLS connection
openssl s_client -connect localhost:8443 -servername localhost

# Verify certificate chain
openssl verify -CAfile ca.pem cert.pem
```

### Admin API Access Denied

**Problem**: 401 errors when accessing admin endpoints

**Solutions**:
1. Verify admin API key is correct
2. Check header format (X-Admin-API-Key or Authorization)
3. Ensure `ADMIN_ENABLED=true`
4. Verify admin path matches configuration

```bash
# Test admin API
curl -v \
  -H "X-Admin-API-Key: your-admin-api-key" \
  http://localhost:8080/admin/stats
```

## Next Steps

### Phase 3 (Planned)

Future enhancements:
- Hot configuration reload (complete `/admin/reload` implementation)
- Distributed tracing with OpenTelemetry
- Circuit breakers for backend services
- Request/response transformations
- API versioning support
- GraphQL gateway support

### Integration with Phase 1

Phase 2 builds on Phase 1 features:
- JWT authentication logged via access logging
- Security headers included in Prometheus metrics
- Admin API provides access to Phase 1 metrics

## Conclusion

Phase 2 transforms the API Gateway into an enterprise-grade solution with:
- ✅ Secure authentication via JWT
- ✅ Comprehensive security headers
- ✅ TLS/HTTPS support
- ✅ Runtime management via Admin API
- ✅ Production-ready configuration
- ✅ Kubernetes deployment support

Combined with Phase 1's observability features, the gateway now provides complete production-ready functionality for API management.
