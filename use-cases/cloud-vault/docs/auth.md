# Authentication Guide

This guide covers authentication and user management in Markdown Cloud Vault.

## Overview

Cloud Vault provides:
- **User authentication**: Secure login with passwords
- **Session management**: Persistent sessions with secure cookies
- **Access control**: Role-based permissions
- **Rate limiting**: Protection against brute force attacks

## Configuration

### Enable Authentication

```toml
[auth]
enabled = true
session_secret = "${SESSION_SECRET}"
session_ttl = 86400  # 24 hours
```

### Session Configuration

```toml
[auth]
session_ttl = 86400  # Session duration (seconds)
cookie_secure = true  # HTTPS only
cookie_httponly = true  # No JavaScript access
cookie_samesite = "strict"  # CSRF protection
```

### Rate Limiting

```toml
[auth.rate_limit]
enabled = true
max_attempts = 5
window_seconds = 300  # 5 minutes
lockout_seconds = 900  # 15 minutes
```

## User Management

### Create User

**Via CLI**:
```bash
./cloud-vault user create \
  --username admin \
  --email admin@example.com \
  --password
```

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "email": "admin@example.com",
    "password": "secure-password"
  }'
```

### Change Password

**Via Web Interface**:
1. Navigate to **Settings** → **Account**
2. Click **Change Password**
3. Enter current and new password
4. Click **Save**

**Via CLI**:
```bash
./cloud-vault user change-password \
  --username admin
```

### Reset Password

**Via CLI**:
```bash
./cloud-vault user reset-password \
  --username admin \
  --new-password
```

### List Users

**Via CLI**:
```bash
./cloud-vault user list
```

**Via API**:
```bash
curl http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer $TOKEN"
```

### Delete User

**Via CLI**:
```bash
./cloud-vault user delete \
  --username olduser
```

**Via API**:
```bash
curl -X DELETE http://localhost:8080/api/v1/users/123 \
  -H "Authorization: Bearer $TOKEN"
```

## Roles and Permissions

### Built-in Roles

- **admin**: Full access to all features
- **user**: Standard access to documents and collections
- **readonly**: Read-only access

### Assign Role

**Via CLI**:
```bash
./cloud-vault user set-role \
  --username admin \
  --role admin
```

**Via API**:
```bash
curl -X PUT http://localhost:8080/api/v1/users/123/role \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "role": "admin"
  }'
```

## Session Management

### View Active Sessions

**Via Web Interface**:
1. Navigate to **Settings** → **Security**
2. View active sessions
3. Revoke suspicious sessions

**Via API**:
```bash
curl http://localhost:8080/api/v1/sessions \
  -H "Authorization: Bearer $TOKEN"
```

### Logout

**Via Web Interface**:
1. Click user menu
2. Click **Logout**

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout \
  -H "Authorization: Bearer $TOKEN"
```

### Logout All Sessions

**Via Web Interface**:
1. Navigate to **Settings** → **Security**
2. Click **Logout All Sessions**

**Via API**:
```bash
curl -X POST http://localhost:8080/api/v1/auth/logout-all \
  -H "Authorization: Bearer $TOKEN"
```

## Security Features

### Password Requirements

Configure password policy:
```toml
[auth.password]
min_length = 12
require_uppercase = true
require_lowercase = true
require_numbers = true
require_special = true
```

### Two-Factor Authentication (2FA)

**Enable 2FA**:
1. Navigate to **Settings** → **Security**
2. Click **Enable 2FA**
3. Scan QR code with authenticator app
4. Enter verification code
5. Save backup codes

**Configuration**:
```toml
[auth]
two_factor_enabled = true
two_factor_required = false  # Optional for all users
```

### Login Notifications

```toml
[auth]
notify_new_login = true
notify_email = true
```

## API Authentication

### Bearer Token

```bash
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/documents
```

### Get Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password"
  }'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "expires_at": "2026-05-31T02:00:00Z"
}
```

### Refresh Token

```bash
curl -X POST http://localhost:8080/api/v1/auth/refresh \
  -H "Authorization: Bearer $TOKEN"
```

## Security Best Practices

### 1. Strong Passwords

Use password manager with:
- Minimum 12 characters
- Mix of upper/lowercase
- Numbers and special characters
- Unique for each service

### 2. Enable 2FA

Two-factor authentication adds security:
- Use authenticator app (not SMS)
- Save backup codes securely
- Test 2FA before enabling requirement

### 3. Regular Session Review

Check active sessions:
1. Navigate to **Settings** → **Security**
2. Review active sessions
3. Revoke unknown sessions

### 4. Secure Session Secret

Generate strong session secret:
```bash
openssl rand -hex 32
```

Store in environment variable:
```bash
export SESSION_SECRET="your-secret-here"
```

### 5. HTTPS in Production

```toml
[server]
tls_enabled = true
tls_cert_file = "/path/to/cert.pem"
tls_key_file = "/path/to/key.pem"

[auth]
cookie_secure = true
```

### 6. Rate Limiting

Protect against brute force:
```toml
[auth.rate_limit]
enabled = true
max_attempts = 5
window_seconds = 300
lockout_seconds = 900
```

### 7. Audit Logs

Monitor authentication events:
```toml
[logging]
level = "info"
audit_auth = true
```

## Troubleshooting

### Login Fails

**Error**: `Invalid credentials`
- Check username and password
- Verify account is not locked
- Reset password if needed

**Error**: `Account locked`
- Wait for lockout period (15 minutes)
- Or unlock via CLI:
```bash
./cloud-vault user unlock --username admin
```

### Session Expired

**Error**: `Session expired`
- Login again
- Check session TTL configuration
- Verify session secret hasn't changed

### 2FA Issues

**Lost authenticator app**:
1. Use backup codes to login
2. Disable 2FA: `./cloud-vault user disable-2fa --username admin`
3. Re-enable 2FA with new authenticator

**Backup codes not working**:
- Check if codes were already used
- Each backup code is single-use
- Contact admin to reset 2FA

## Next Steps

- [Security](./security.md) - Security best practices
- [Configuration](./configuration.md) - Authentication settings
- [Server Deployment](./server-deploy.md) - Production authentication

---

**Last updated**: 2026-05-30
