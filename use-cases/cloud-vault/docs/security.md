# Security Guide

This guide covers security best practices and features for Markdown Cloud Vault.

## Overview

Cloud Vault provides multiple security features:
- **Authentication**: Secure user login and session management
- **Authorization**: Role-based access control
- **Encryption**: Data protection in transit and at rest
- **Rate limiting**: Protection against brute force attacks
- **Input validation**: Protection against injection attacks

## Authentication Security

### Strong Passwords

Configure password requirements:
```toml
[auth.password]
min_length = 12
require_uppercase = true
require_lowercase = true
require_numbers = true
require_special = true
```

### Session Security

```toml
[auth]
session_ttl = 86400  # 24 hours
cookie_secure = true  # HTTPS only
cookie_httponly = true  # No JavaScript access
cookie_samesite = "strict"  # CSRF protection
```

### Rate Limiting

Protect against brute force:
```toml
[auth.rate_limit]
enabled = true
max_attempts = 5
window_seconds = 300  # 5 minutes
lockout_seconds = 900  # 15 minutes
```

### Two-Factor Authentication

Enable 2FA for all users:
```toml
[auth]
two_factor_enabled = true
two_factor_required = true
```

## Network Security

### HTTPS Configuration

Enable HTTPS in production:
```toml
[server]
tls_enabled = true
tls_cert_file = "/path/to/cert.pem"
tls_key_file = "/path/to/key.pem"
tls_min_version = "1.2"
```

### Firewall Configuration

```bash
# Allow only necessary ports
sudo ufw allow 443/tcp  # HTTPS
sudo ufw allow 22/tcp   # SSH
sudo ufw enable
```

### Reverse Proxy

Use reverse proxy for additional security:
```nginx
# Nginx configuration
server {
    listen 443 ssl http2;
    
    # SSL configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers HIGH:!aNULL:!MD5;
    
    # Security headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "strict-origin-when-cross-origin" always;
    
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

## Data Security

### Encryption at Rest

Encrypt sensitive data:
```toml
[security]
encryption_enabled = true
encryption_key = "${ENCRYPTION_KEY}"
```

Generate encryption key:
```bash
openssl rand -hex 32
```

### Backup Encryption

Encrypt backups:
```toml
[backup]
encryption = true
encryption_key = "${BACKUP_ENCRYPTION_KEY}"
```

### Database Security

```toml
[database]
# Use dedicated database user
username = "cloud_vault"
password = "${DB_PASSWORD}"

# Restrict file permissions
file_mode = "0600"
```

## API Security

### CORS Configuration

```toml
[cors]
enabled = true
allowed_origins = ["https://vault.yourdomain.com"]
allowed_methods = ["GET", "POST", "PUT", "DELETE"]
allow_credentials = true
```

### Rate Limiting

```toml
[rate_limit]
enabled = true
requests_per_minute = 60
burst = 10
```

### API Key Security

```bash
# Store API keys in environment variables
export OPENAI_API_KEY="sk-..."
export QINIU_ACCESS_KEY="..."
export QINIU_SECRET_KEY="..."

# Never commit keys to version control
echo "*.key" >> .gitignore
echo ".env" >> .gitignore
```

## Input Validation

### File Upload Validation

```toml
[import]
max_file_size = 10485760  # 10 MB
allowed_extensions = [".md", ".markdown", ".txt"]
scan_malware = true
```

### SQL Injection Prevention

Cloud Vault uses parameterized queries:
```go
// Safe: parameterized query
stmt, _ := db.Prepare("SELECT * FROM documents WHERE id = ?")
rows, _ := stmt.Query(docID)

// Unsafe: string concatenation (never do this)
// query := "SELECT * FROM documents WHERE id = " + docID
```

### XSS Prevention

```toml
[security]
sanitize_html = true
escape_output = true
content_security_policy = "default-src 'self'"
```

## Security Headers

Configure security headers:
```toml
[security.headers]
strict_transport_security = "max-age=31536000; includeSubDomains"
x_frame_options = "SAMEORIGIN"
x_content_type_options = "nosniff"
x_xss_protection = "1; mode=block"
referrer_policy = "strict-origin-when-cross-origin"
content_security_policy = "default-src 'self'; script-src 'self' 'unsafe-inline'"
```

## Monitoring and Auditing

### Security Logs

```toml
[logging]
level = "info"
audit_enabled = true
audit_events = ["login", "logout", "permission_denied", "config_change"]
```

### Failed Login Monitoring

```bash
# Monitor failed logins
./cloud-vault logs tail | grep "login failed"

# Check for brute force attempts
./cloud-vault logs tail | grep "rate limit exceeded"
```

### Access Logs

```toml
[logging]
access_log_enabled = true
access_log_format = "combined"
```

## Vulnerability Management

### Keep Software Updated

```bash
# Check for updates
./cloud-vault update check

# Apply updates
./cloud-vault update apply
```

### Dependency Updates

```bash
# Update Go dependencies
go get -u ./...

# Update system packages
sudo apt update && sudo apt upgrade
```

### Security Scanning

```bash
# Scan for vulnerabilities
gosec ./...

# Check dependencies
govulncheck ./...
```

## Security Checklist

### Pre-Deployment

- [ ] Enable HTTPS
- [ ] Configure firewall
- [ ] Set strong passwords
- [ ] Enable 2FA
- [ ] Configure rate limiting
- [ ] Set up backups
- [ ] Enable logging
- [ ] Review CORS settings
- [ ] Configure security headers
- [ ] Test authentication

### Production

- [ ] Use dedicated database user
- [ ] Encrypt sensitive data
- [ ] Enable backup encryption
- [ ] Monitor security logs
- [ ] Set up alerts
- [ ] Regular security audits
- [ ] Keep software updated
- [ ] Review access logs
- [ ] Test disaster recovery

## Incident Response

### Security Breach

If you suspect a security breach:

1. **Isolate affected systems**:
```bash
sudo systemctl stop cloud-vault
```

2. **Preserve evidence**:
```bash
tar -czf incident-$(date +%Y%m%d-%H%M%S).tar.gz \
  /var/log/cloud-vault/ \
  /var/lib/cloud-vault/
```

3. **Reset credentials**:
```bash
./cloud-vault user reset-password --username admin
./cloud-vault auth invalidate-all-sessions
```

4. **Review logs**:
```bash
./cloud-vault logs view /var/log/cloud-vault/app.log | grep -E "(ERROR|WARN)"
```

5. **Restore from backup**:
```bash
./cloud-vault backup restore --backup clean-backup.zip
```

6. **Notify users** (if data compromised)

## Compliance

### GDPR

- **Data minimization**: Collect only necessary data
- **Right to erasure**: Delete user data on request
- **Data portability**: Export data in standard formats
- **Consent**: Clear privacy policy

### SOC 2

- **Access control**: Role-based permissions
- **Audit logs**: Track all access
- **Encryption**: Protect data in transit and at rest
- **Monitoring**: Continuous security monitoring

## Security Resources

### Documentation

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [NIST Cybersecurity Framework](https://www.nist.gov/cyberframework)
- [Cloud Security Alliance](https://cloudsecurityalliance.org/)

### Tools

- **Vulnerability scanning**: gosec, govulncheck
- **Penetration testing**: OWASP ZAP, Burp Suite
- **Log analysis**: ELK Stack, Splunk
- **Monitoring**: Prometheus, Grafana

## Next Steps

- [Authentication](./auth.md) - Authentication features
- [Privacy](./privacy.md) - Privacy considerations
- [Diagnostics](./diagnostics.md) - Security diagnostics

---

**Last updated**: 2026-05-30
