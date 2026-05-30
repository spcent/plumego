# Release Notes

This document contains release notes for all versions of Markdown Cloud Vault.

## Version 1.0.0

**Release Date**: 2026-05-30

### Overview

Version 1.0.0 is the first stable release of Markdown Cloud Vault, featuring a complete document management system with full-text search, AI-powered features, and comprehensive organization tools.

### Features

#### Core Features

- **Document Management**
  - Create, edit, and delete Markdown documents
  - Version history with diff view
  - Document metadata and tags
  - Collections for organizing documents

- **Import/Export**
  - Import from directories and ZIP archives
  - Automatic duplicate detection
  - Metadata extraction from frontmatter
  - Auto-tagging based on filename and path
  - Export to Markdown, JSON, or CSV

- **Full-Text Search**
  - Powered by SQLite FTS5
  - Search across document content, titles, and tags
  - Advanced search operators (phrase, wildcard, boolean)
  - Search result highlighting
  - Search history and saved searches

- **Organization**
  - Collections for grouping documents
  - Tags for flexible categorization
  - AI-powered topic detection
  - Review queue for unorganized documents
  - Duplicate detection and management

- **AI Features**
  - Document summarization
  - Question answering (Q&A)
  - Prompt extraction
  - Auto-tagging suggestions
  - Topic detection
  - Support for OpenAI, Azure OpenAI, and local LLMs (Ollama)

- **Storage**
  - Local filesystem storage (default)
  - Qiniu cloud storage integration
  - Automatic backup and restore
  - Storage optimization and cleanup

- **Authentication**
  - User authentication with secure passwords
  - Session management with secure cookies
  - Role-based access control (admin, user, readonly)
  - Rate limiting and brute force protection
  - Optional two-factor authentication (2FA)

- **Desktop Application**
  - Native apps for Windows, macOS, and Linux
  - System tray integration
  - Automatic updates
  - Desktop notifications
  - Keyboard shortcuts

- **Server Deployment**
  - Standalone server mode
  - Docker container support
  - Systemd service integration
  - Reverse proxy configuration
  - HTTPS with Let's Encrypt

#### User Interface

- **Web Interface**
  - Modern, responsive design
  - Dark mode support
  - Keyboard shortcuts
  - Drag-and-drop file upload
  - Real-time search
  - Markdown preview with syntax highlighting

- **Desktop Application**
  - Native look and feel
  - System tray integration
  - Automatic updates
  - Desktop notifications

#### Administration

- **System Monitoring**
  - Health checks for all components
  - Performance metrics
  - Storage usage statistics
  - Database integrity checks

- **Diagnostics**
  - Export diagnostic bundles
  - Automatic redaction of sensitive data
  - Log rotation and management
  - Health check reports

- **Backup & Restore**
  - Automatic scheduled backups
  - Manual backup creation
  - Restore from backup
  - Backup verification

### Security

- **Authentication**
  - Secure password hashing (bcrypt)
  - Session management with secure cookies
  - Rate limiting for login attempts
  - Account lockout after failed attempts

- **Authorization**
  - Role-based access control
  - API authentication with bearer tokens
  - CORS configuration

- **Data Protection**
  - HTTPS support with TLS 1.2+
  - Encryption at rest (optional)
  - Backup encryption
  - Automatic redaction of sensitive data

- **Input Validation**
  - SQL injection prevention (parameterized queries)
  - XSS prevention (output escaping)
  - File upload validation
  - Request size limits

### Performance

- **Database**
  - SQLite with WAL mode for better concurrency
  - Automatic indexing and optimization
  - Connection pooling

- **Search**
  - FTS5 full-text search index
  - Search result caching
  - Batch indexing for large imports

- **AI**
  - Async task queue for AI operations
  - Result caching to reduce API calls
  - Rate limiting to control costs

### Configuration

- **Flexible Configuration**
  - TOML configuration file
  - Environment variable overrides
  - Command-line flags
  - Configuration validation

- **Key Configuration Options**
  - Storage provider (local or Qiniu)
  - AI provider and model selection
  - Authentication settings
  - Backup schedule and retention
  - Logging level and rotation

### Documentation

- **User Guides**
  - Getting Started Guide
  - Installation Guide
  - Desktop Application Guide
  - Server Deployment Guide
  - Storage Configuration Guide
  - Qiniu Setup Guide
  - Import Guide
  - Search Guide
  - Organization Guide
  - AI Features Guide
  - Authentication Guide
  - Backup & Restore Guide
  - Diagnostics Guide
  - Troubleshooting Guide
  - Security Guide
  - Privacy Guide

- **Developer Documentation**
  - Development Guide
  - API documentation
  - Architecture overview
  - Contributing guidelines

### Known Limitations

- **Database**
  - SQLite only (no support for other databases)
  - Single-writer limitation for concurrent writes

- **AI Features**
  - Requires external AI provider or local LLM
  - API costs for cloud providers
  - Quality depends on model selection

- **Storage**
  - Qiniu integration limited to China regions
  - No support for other cloud providers (AWS S3, Azure Blob, etc.)

- **Collaboration**
  - Single-user only (no multi-user collaboration)
  - No real-time editing
  - No sharing or permissions per document

### Upgrade from 0.9.x

**Breaking Changes**: None

**Migration Steps**:
1. Backup your data: `./cloud-vault backup create`
2. Stop the service: `sudo systemctl stop cloud-vault`
3. Update the binary: Download v1.0.0
4. Update configuration: Review new configuration options
5. Start the service: `sudo systemctl start cloud-vault`
6. Verify: `./cloud-vault doctor check`

### Bug Fixes

- Fixed memory leak during large imports
- Fixed search index corruption on crash
- Fixed duplicate detection false positives
- Fixed Qiniu upload timeout issues
- Fixed session expiration not working correctly
- Fixed markdown rendering issues with special characters

### Dependencies

- Go 1.22+
- SQLite 3.35+
- Node.js 18+ (for building frontend)
- React 18
- Vite 5
- Tailwind CSS 3

### Compatibility

- **Operating Systems**
  - Windows 10/11 (64-bit)
  - macOS 12+ (Monterey and later)
  - Linux (Ubuntu 20.04+, Debian 11+, CentOS 8+)

- **Browsers**
  - Chrome 90+
  - Firefox 88+
  - Safari 14+
  - Edge 90+

### Contributors

- Development Team
- Community contributors
- Beta testers

### License

MIT License - See LICENSE file for details

---

## Version 0.9.0 (Beta)

**Release Date**: 2026-05-15

### Overview

Beta release with all core features implemented and ready for testing.

### Features

- Document management with version history
- Full-text search with FTS5
- Import from directories and ZIP files
- Collections and tags
- AI-powered summarization and Q&A
- Local and Qiniu storage
- User authentication
- Desktop application (Windows, macOS, Linux)
- Server deployment mode

### Known Issues

- Memory usage high during large imports
- Search index may corrupt on crash
- Qiniu upload slow for large files
- Session expiration not working correctly

### Feedback

Please report bugs and provide feedback via GitHub issues.

---

## Version 0.8.0 (Alpha)

**Release Date**: 2026-04-30

### Overview

Alpha release with basic functionality for early testing.

### Features

- Basic document management
- Simple search functionality
- Import from directories
- Local storage only
- Basic authentication

### Known Issues

- Many features incomplete
- Performance issues with large datasets
- Limited error handling
- No backup/restore functionality

---

## Version History

| Version | Release Date | Status | Notes |
|---------|--------------|--------|-------|
| 1.0.0 | 2026-05-30 | Stable | First stable release |
| 0.9.0 | 2026-05-15 | Beta | Feature complete |
| 0.8.0 | 2026-04-30 | Alpha | Basic functionality |
| 0.7.0 | 2026-04-15 | Alpha | Early preview |
| 0.6.0 | 2026-04-01 | Alpha | Internal testing |
| 0.5.0 | 2026-03-15 | Alpha | Initial prototype |

---

**Last updated**: 2026-05-30
