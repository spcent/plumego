# Getting Started

Welcome to Markdown Cloud Vault! This guide will help you set up and start using your vault within minutes.

## Overview

Markdown Cloud Vault is a self-hosted knowledge management system that helps you:
- Import and organize Markdown documents from various sources
- Search your entire vault instantly with full-text search
- Use AI to summarize, analyze, and extract insights from your documents
- Manage document versions and track changes over time
- Sync with cloud storage providers like Qiniu

## Quick Start

### Step 1: Install Cloud Vault

Choose your preferred installation method:

**Desktop App (Recommended for personal use)**
- Download the installer for your platform (Windows, macOS, or Linux)
- Run the installer and follow the setup wizard
- See [Installation Guide](./installation.md) for detailed instructions

**Server Deployment (For teams and advanced users)**
- Download the server binary or build from source
- Configure using `config.toml`
- Run as a service or Docker container
- See [Server Deployment](./server-deploy.md) for production setup

### Step 2: First Launch

**Desktop App:**
1. Launch the application from your Applications folder or Start menu
2. The app will open with a welcome screen
3. Click "Get Started" to create your first admin account

**Server:**
1. Start the server: `./cloud-vault` or `docker compose up`
2. Open your browser to `http://localhost:8080`
3. Complete the setup wizard to create an admin account

### Step 3: Configure Storage

By default, Cloud Vault stores documents locally in `./data/objects`. For cloud storage:

1. Go to **Settings** → **Configuration**
2. Set `storage_provider` to `qiniu`
3. Enter your Qiniu credentials (Access Key, Secret Key, Bucket)
4. Save and restart the application

See [Storage Configuration](./storage.md) and [Qiniu Setup](./qiniu.md) for details.

### Step 4: Import Your First Documents

1. Navigate to **Import** in the sidebar
2. Click "Select Directory" and choose a folder containing Markdown files
3. Review the scan preview showing files to be imported
4. Click "Start Import" to begin the import process
5. Monitor progress in the import job list

See [Import Guide](./import.md) for advanced options like duplicate detection and tag extraction.

### Step 5: Search and Organize

Once imported, your documents are ready to search:

1. Use the **Search** page to find documents by content, title, or tags
2. Create **Collections** to group related documents
3. Use **Organize** tools to detect duplicates and suggest tags
4. Browse your vault in the **Vault** page

See [Search Guide](./search.md) and [Organization Guide](./organize.md) for more features.

### Step 6: Enable AI Features (Optional)

To use AI-powered features like summarization and Q&A:

1. Go to **Settings** → **Configuration**
2. Set `ai_enabled` to `true`
3. Configure your AI provider (OpenAI, Ollama, etc.)
4. Save and restart

See [AI Features](./ai.md) for supported providers and usage examples.

## Key Concepts

### Documents
Markdown files stored in your vault with metadata like tags, collections, and versions.

### Collections
Named groups of documents for organizing related content (e.g., "Project Alpha", "Research Papers").

### Tags
Labels attached to documents for flexible categorization and filtering.

### Import Jobs
Background tasks that process and import Markdown files from directories or archives.

### Index
Full-text search index that enables instant searching across all document content.

## Next Steps

- Read the [Configuration Guide](./configuration.md) to customize your vault
- Explore [Backup & Restore](./backup-restore.md) to protect your data
- Check [Security & Privacy](./security.md) for best practices
- Join the community for support and feature requests

## Need Help?

- **Troubleshooting**: See [Troubleshooting Guide](./troubleshooting.md) for common issues
- **Diagnostics**: Export a [diagnostic bundle](./diagnostics.md) to share with support
- **Documentation**: Browse all guides in the [docs folder](./README.md)

---

**Last updated**: 2026-05-30
