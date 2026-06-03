# Backup and Restore Guide

This guide documents the backup and restore behavior implemented in the source tree.

## What Is Included

Manual backups include:

- `app.db` — SQLite database snapshot created with `VACUUM INTO`
- `objects/` — local object storage files when `storage.provider = "local"`
- `manifest.json` — backup metadata and format version

When `storage.provider = "qiniu"`, backups include database metadata and the manifest. Qiniu bucket objects are not downloaded into the backup zip; use Qiniu's lifecycle, replication, or export tooling for object-level recovery.

## Create A Backup

Use the web UI: **Settings** -> **Backup** -> **Create Backup**.

Or call the protected API:

```bash
curl -X POST http://localhost:8080/api/v1/system/backup \
  -H "Cookie: session=<your-session-cookie>"
```

Backups are written to the configured backup directory as:

```text
backups/cloud-vault-backup-YYYYMMDD-HHMMSS.zip
```

## List, Download, And Delete

```bash
curl http://localhost:8080/api/v1/system/backups \
  -H "Cookie: session=<your-session-cookie>"

curl -O http://localhost:8080/api/v1/system/backups/cloud-vault-backup-20260603-120000.zip/download \
  -H "Cookie: session=<your-session-cookie>"

curl -X DELETE http://localhost:8080/api/v1/system/backups/cloud-vault-backup-20260603-120000.zip \
  -H "Cookie: session=<your-session-cookie>"
```

## Restore

The recommended production restore path is offline:

```bash
# Stop the server first.
./cloud-vault restore --file backups/cloud-vault-backup-20260603-120000.zip --data-dir ./data
```

The restore command validates the zip manifest, extracts the database and local objects to temporary paths, then renames them into place. After restore, start the server and run the doctor checks.

The API restore path is available for controlled local use and requires explicit confirmation:

```bash
curl -X POST http://localhost:8080/api/v1/system/restore \
  -H "Content-Type: application/json" \
  -H "Cookie: session=<your-session-cookie>" \
  -d '{"backup_name":"cloud-vault-backup-20260603-120000.zip","confirm":"RESTORE"}'
```

Do not restore while users or background writers are active. Concurrent writes during restore are undefined.

## Verification

After creating a backup:

```bash
unzip -l backups/cloud-vault-backup-20260603-120000.zip
```

After restoring:

```bash
./cloud-vault --config config.toml
```

Then open the web UI and run the system doctor checks from Settings.

## Operational Notes

- Backups are full snapshots, not incremental backups.
- Scheduling and rotation are not built in; use cron, systemd timers, or your platform scheduler.
- Backup zip files are not encrypted at rest; encrypt them before storing offsite.
- Qiniu bucket contents need a separate backup strategy.
- Keep at least one tested backup outside the Cloud Vault data directory.

Encrypt an offsite copy with GPG:

```bash
gpg -c backups/cloud-vault-backup-20260603-120000.zip
```

## Troubleshooting

`permission denied`: verify the server user can write to the backup directory and the data directory.

`no space left on device`: move older backups off the server or delete backups through the UI/API.

`invalid backup manifest`: the file is not a Cloud Vault backup or is corrupted. Try an earlier backup.

`restore succeeded but objects are missing`: check whether the original system used Qiniu storage. Qiniu objects are restored with Qiniu tooling, not the Cloud Vault backup zip.
