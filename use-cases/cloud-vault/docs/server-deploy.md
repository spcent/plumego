# Server Deployment Guide

This guide covers deploying Markdown Cloud Vault in production server environments, including systemd, Docker, and reverse proxy configurations.

## Overview

Cloud Vault can be deployed as:
- **Standalone server**: Direct binary execution
- **Systemd service**: Managed service with auto-restart
- **Docker container**: Isolated, portable deployment
- **Kubernetes**: Orchestrated container deployment (advanced)

## Prerequisites

- Linux server (Ubuntu 20.04+, Debian 11+, CentOS 8+)
- 2 GB RAM minimum (4 GB recommended)
- 10 GB disk space minimum
- Domain name (optional, for HTTPS)
- SSL certificate (optional, for HTTPS)

## Standalone Server Deployment

### Installation

1. **Build binary**:
```bash
git clone <your-plumego-repository-url>
cd plumego/use-cases/cloud-vault
make server-build-v1
cp dist/server/markdown-vault ./cloud-vault
```

If you publish release artifacts, replace this build step with your signed download URL and checksum verification procedure.

2. **Verify installation**:
```bash
./cloud-vault --version
```

3. **Create data directories**:
```bash
mkdir -p data/objects logs backups
chmod 755 data logs backups
```

4. **Create configuration file**:
```bash
cp config.example.toml config.toml
nano config.toml
```

### Basic Configuration

Edit `config.toml` for your environment:

```toml
[server]
addr = ":8080"

[database]
path = "./data/app.db"

[storage]
provider = "local"
root = "./data/objects"

[auth]
enabled = true
bootstrap_admin_enabled = true
bootstrap_admin_username = "admin"
bootstrap_admin_email = "admin@example.com"
bootstrap_admin_password = "ChangeMe123!"

[update]
enabled = false
update_url = ""
check_interval_minutes = 1440
```

### Run Server

```bash
./cloud-vault --config config.toml
```

Access at `http://your-server-ip:8080`

## Systemd Service Deployment

### Create Service User

```bash
sudo useradd -r -s /bin/false -m -d /opt/cloud-vault cloud-vault
```

### Install Application

```bash
sudo mkdir -p /opt/cloud-vault
sudo cp cloud-vault /opt/cloud-vault/
sudo cp config.toml /opt/cloud-vault/
sudo chown -R cloud-vault:cloud-vault /opt/cloud-vault
sudo chmod +x /opt/cloud-vault/cloud-vault
```

### Create Data Directories

```bash
sudo mkdir -p /var/lib/cloud-vault/{data/objects,logs,backups}
sudo chown -R cloud-vault:cloud-vault /var/lib/cloud-vault
```

### Update Configuration

Edit `/opt/cloud-vault/config.toml`:

```toml
[server]
addr = ":8080"

[database]
path = "/var/lib/cloud-vault/data/app.db"

[storage]
provider = "local"
root = "/var/lib/cloud-vault/data/objects"

[logging]
level = "info"
dir = "/var/lib/cloud-vault/logs"
```

### Create Systemd Service

Create `/etc/systemd/system/cloud-vault.service`:

```ini
[Unit]
Description=Markdown Cloud Vault
After=network.target

[Service]
Type=simple
User=cloud-vault
Group=cloud-vault
WorkingDirectory=/opt/cloud-vault
ExecStart=/opt/cloud-vault/cloud-vault --config /opt/cloud-vault/config.toml
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

# Security hardening
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/cloud-vault

[Install]
WantedBy=multi-user.target
```

### Enable and Start Service

```bash
sudo systemctl daemon-reload
sudo systemctl enable cloud-vault
sudo systemctl start cloud-vault
sudo systemctl status cloud-vault
```

### View Logs

```bash
# Service logs
sudo journalctl -u cloud-vault -f

# Application logs
tail -f /var/lib/cloud-vault/logs/app.log
```

### Management Commands

```bash
# Start
sudo systemctl start cloud-vault

# Stop
sudo systemctl stop cloud-vault

# Restart
sudo systemctl restart cloud-vault

# Status
sudo systemctl status cloud-vault

# Logs
sudo journalctl -u cloud-vault -f
```

## Docker Deployment

### Single Container

**docker-compose.yml**:

```yaml
version: '3.8'

services:
  cloud-vault:
    image: cloudvault/cloud-vault:1.0.0
    container_name: cloud-vault
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./config.toml:/app/config.toml:ro
      - ./data:/var/lib/cloud-vault/data
      - ./logs:/var/lib/cloud-vault/logs
      - ./backups:/var/lib/cloud-vault/backups
    environment:
      - TZ=UTC
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

**Start**:
```bash
docker-compose up -d
```

**View logs**:
```bash
docker-compose logs -f
```

### Multi-Container with Nginx

**docker-compose.yml**:

```yaml
version: '3.8'

services:
  cloud-vault:
    image: cloudvault/cloud-vault:1.0.0
    restart: unless-stopped
    expose:
      - "8080"
    volumes:
      - ./config.toml:/app/config.toml:ro
      - cloud-vault-data:/var/lib/cloud-vault/data
      - cloud-vault-logs:/var/lib/cloud-vault/logs
    networks:
      - cloud-vault-net

  nginx:
    image: nginx:alpine
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
    depends_on:
      - cloud-vault
    networks:
      - cloud-vault-net

volumes:
  cloud-vault-data:
  cloud-vault-logs:

networks:
  cloud-vault-net:
    driver: bridge
```

**nginx.conf**:

```nginx
events {
    worker_connections 1024;
}

http {
    upstream cloud_vault {
        server cloud-vault:8080;
    }

    server {
        listen 80;
        server_name vault.example.com;
        return 301 https://$server_name$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name vault.example.com;

        ssl_certificate /etc/nginx/ssl/cert.pem;
        ssl_certificate_key /etc/nginx/ssl/key.pem;
        ssl_protocols TLSv1.2 TLSv1.3;

        client_max_body_size 100M;

        location / {
            proxy_pass http://cloud_vault;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            proxy_cache_bypass $http_upgrade;
        }
    }
}
```

## Reverse Proxy Configuration

### Nginx

```nginx
server {
    listen 80;
    server_name vault.example.com;
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl http2;
    server_name vault.example.com;

    ssl_certificate /etc/letsencrypt/live/vault.example.com/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/vault.example.com/privkey.pem;

    client_max_body_size 100M;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }
}
```

### Apache

```apache
<VirtualHost *:80>
    ServerName vault.example.com
    Redirect permanent / https://vault.example.com/
</VirtualHost>

<VirtualHost *:443>
    ServerName vault.example.com

    SSLEngine on
    SSLCertificateFile /etc/letsencrypt/live/vault.example.com/fullchain.pem
    SSLCertificateKeyFile /etc/letsencrypt/live/vault.example.com/privkey.pem

    ProxyPreserveHost On
    ProxyPass / http://127.0.0.1:8080/
    ProxyPassReverse / http://127.0.0.1:8080/

    RequestHeader set X-Forwarded-Proto "https"
    RequestHeader set X-Forwarded-Port "443"
</VirtualHost>
```

### Caddy

```caddyfile
vault.example.com {
    reverse_proxy 127.0.0.1:8080
    
    header {
        X-Frame-Options DENY
        X-Content-Type-Options nosniff
        X-XSS-Protection "1; mode=block"
    }
}
```

## HTTPS with Let's Encrypt

### Certbot with Nginx

```bash
# Install certbot
sudo apt install certbot python3-certbot-nginx

# Obtain certificate
sudo certbot --nginx -d vault.example.com

# Auto-renewal (already configured by certbot)
sudo systemctl status certbot.timer
```

### Certbot with Apache

```bash
sudo apt install certbot python3-certbot-apache
sudo certbot --apache -d vault.example.com
```

### Manual Certificate

```bash
# Generate certificate
certbot certonly --standalone -d vault.example.com

# Copy to application directory
sudo cp /etc/letsencrypt/live/vault.example.com/fullchain.pem /opt/cloud-vault/ssl/
sudo cp /etc/letsencrypt/live/vault.example.com/privkey.pem /opt/cloud-vault/ssl/
sudo chown -R cloud-vault:cloud-vault /opt/cloud-vault/ssl
```

Update `config.toml`:

```toml
[server]
addr = ":443"
tls_enabled = true
tls_cert_file = "/opt/cloud-vault/ssl/fullchain.pem"
tls_key_file = "/opt/cloud-vault/ssl/privkey.pem"
```

## Database Configuration

### SQLite (Default)

SQLite is included and requires no additional setup:

```toml
[database]
path = "/var/lib/cloud-vault/data/app.db"
```

**Backup**:
```bash
sqlite3 /var/lib/cloud-vault/data/app.db ".backup '/backups/app-$(date +%Y%m%d).db'"
```

### PostgreSQL (Future)

PostgreSQL support is planned for v2.0.

## Monitoring and Maintenance

### Health Checks

```bash
# Basic health check
curl -f http://localhost:8080/health

# Detailed health check
curl -f http://localhost:8080/api/v1/system/health
```

### Log Rotation

Create `/etc/logrotate.d/cloud-vault`:

```
/var/lib/cloud-vault/logs/*.log {
    daily
    rotate 30
    compress
    delaycompress
    notifempty
    create 0640 cloud-vault cloud-vault
    sharedscripts
    postrotate
        systemctl reload cloud-vault
    endscript
}
```

### Scheduled Backups

Cloud Vault does not include a built-in scheduler. Use an external scheduler that calls the protected backup API with a valid session cookie or service-run script.

Example `/etc/cron.d/cloud-vault-backup`:

```
# Daily backup at 2 AM
0 2 * * * cloud-vault /usr/local/bin/cloud-vault-create-backup
```

The wrapper script should authenticate, store its session cookie securely, and call `POST /api/v1/system/backup`.

### Monitoring with Prometheus

Add to `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'cloud-vault'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
```

## Security Hardening

### Firewall Configuration

```bash
# UFW (Ubuntu)
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp
sudo ufw enable

# firewalld (CentOS/RHEL)
sudo firewall-cmd --permanent --add-service=http
sudo firewall-cmd --permanent --add-service=https
sudo firewall-cmd --reload
```

### Fail2ban

Create `/etc/fail2ban/jail.d/cloud-vault.conf`:

```ini
[cloud-vault]
enabled = true
port = http,https
filter = cloud-vault
logpath = /var/lib/cloud-vault/logs/app.log
maxretry = 5
bantime = 3600
```

Create `/etc/fail2ban/filter.d/cloud-vault.conf`:

```ini
[Definition]
failregex = ^.*authentication.*failed.*client_ip=<HOST>.*$
ignoreregex =
```

Enable:
```bash
sudo systemctl restart fail2ban
```

## Scaling and Performance

### Horizontal Scaling

Cloud Vault is designed for single-instance deployment. For high availability:

1. Use load balancer with sticky sessions
2. Share storage via NFS or cloud storage
3. Use PostgreSQL instead of SQLite (v2.0)

### Performance Tuning

**Increase file descriptors**:
```bash
# /etc/security/limits.conf
cloud-vault soft nofile 65536
cloud-vault hard nofile 65536
```

**Optimize SQLite**:
```toml
[database]
path = "/var/lib/cloud-vault/data/app.db"
cache_size = 10000
journal_mode = "wal"
```

**Enable caching**:
```toml
[search]
index_cache_size = 1000
```

## Troubleshooting

### Service Won't Start

```bash
# Check logs
sudo journalctl -u cloud-vault -n 50

# Check configuration
sudo -u cloud-vault /opt/cloud-vault/cloud-vault --config /opt/cloud-vault/config.toml --validate

# Check permissions
ls -la /var/lib/cloud-vault/
```

### Database Locked

```bash
# Stop service
sudo systemctl stop cloud-vault

# Check for processes
lsof /var/lib/cloud-vault/data/app.db

# Restart service
sudo systemctl start cloud-vault
```

### High Memory Usage

- Check for memory leaks in logs
- Reduce cache sizes in config
- Restart service periodically

## Next Steps

- [Configuration Guide](./configuration.md) - Detailed configuration options
- [Security](./security.md) - Security best practices
- [Backup & Restore](./backup-restore.md) - Protect your data
- [Monitoring](./monitoring.md) - Set up monitoring and alerts

---

**Last updated**: 2026-05-30
