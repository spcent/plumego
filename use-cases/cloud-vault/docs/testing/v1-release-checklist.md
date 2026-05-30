# V1.0 Release Checklist

This checklist verifies all V1.0 features and acceptance criteria before release.

## Phase 1: Version Management

- [ ] **Version package exists**
  - [ ] `internal/version/version.go` contains `BuildInfo` struct
  - [ ] `Version`, `Commit`, `BuildTime`, `Channel` variables defined
  - [ ] `GetBuildInfo()` function returns all fields

- [ ] **Version API endpoint**
  - [ ] `GET /api/v1/system/version` returns version info
  - [ ] Endpoint protected by authentication
  - [ ] Response includes all BuildInfo fields

- [ ] **Build injection**
  - [ ] Makefile includes ldflags for version injection
  - [ ] Built binary reports correct version from ldflags
  - [ ] `make server-build-v1` target works

- [ ] **Tests pass**
  - [ ] `go test ./internal/version/...` passes
  - [ ] Default values tested when ldflags not set

## Phase 2: Update Checker

- [ ] **Update config**
  - [ ] `UpdateConfig` struct in `internal/config/config.go`
  - [ ] `[update]` section in `config.example.toml`
  - [ ] Config validation works

- [ ] **Update package**
  - [ ] `internal/update/model.go` defines `LatestRelease` and `UpdateStatus`
  - [ ] `internal/update/checker.go` implements `CheckForUpdate()`
  - [ ] `CompareVersions()` handles semantic versioning correctly
  - [ ] `SelectAsset()` picks correct download for platform

- [ ] **Update API endpoints**
  - [ ] `GET /api/v1/update/status` returns cached status
  - [ ] `POST /api/v1/update/check` triggers fresh check
  - [ ] Both endpoints protected by authentication

- [ ] **Startup check**
  - [ ] Auto-check on startup when `check_on_startup = true`
  - [ ] Check failure does not block startup
  - [ ] Warning logged on check failure

- [ ] **Tests pass**
  - [ ] `go test ./internal/update/...` passes
  - [ ] Version comparison tests cover edge cases
  - [ ] HTTP fetch tests with httptest.Server

## Phase 3: Diagnostics Export

- [ ] **Diagnostics package**
  - [ ] `internal/diagnostics/model.go` defines `DiagnosticManifest` and `DiagnosticBundle`
  - [ ] `internal/diagnostics/redactor.go` masks sensitive fields
  - [ ] `internal/diagnostics/bundle.go` creates zip with correct contents
  - [ ] `internal/diagnostics/logs.go` reads and redacts recent errors

- [ ] **Bundle contents**
  - [ ] `manifest.json` included with metadata
  - [ ] `system-info.json` included with OS, arch, memory
  - [ ] `build-info.json` included with version info
  - [ ] `config-redacted.json` included with secrets masked
  - [ ] `doctor-summary.json` included with last check results
  - [ ] `recent-errors.log` included with redacted log lines

- [ ] **Bundle exclusions**
  - [ ] `app.db` NOT included (verify zip entries)
  - [ ] `objects/` directory NOT included
  - [ ] Markdown content NOT included (no .md files)
  - [ ] `.env` raw file NOT included
  - [ ] Secrets NOT included (scan for API keys, tokens, passwords)

- [ ] **Diagnostics API endpoints**
  - [ ] `POST /api/v1/system/diagnostics/export` creates bundle
  - [ ] `GET /api/v1/system/diagnostics` lists bundles
  - [ ] `GET /api/v1/system/diagnostics/{name}/download` downloads bundle
  - [ ] All endpoints protected by authentication
  - [ ] Path traversal blocked (name validation)

- [ ] **Tests pass**
  - [ ] `go test ./internal/diagnostics/...` passes
  - [ ] Redactor tests cover all sensitive field types
  - [ ] Bundle contents verified
  - [ ] Exclusions verified
  - [ ] Path traversal rejection tested

## Phase 4: Log Redaction & Rotation

- [ ] **Logging config**
  - [ ] `LoggingConfig` struct in `internal/config/config.go`
  - [ ] `[logging]` section in `config.example.toml`

- [ ] **Log redaction**
  - [ ] `internal/logging/redactor.go` implements `RedactingWriter`
  - [ ] Bearer tokens masked
  - [ ] `api_key=...` masked
  - [ ] `secret_key=...` masked
  - [ ] `password=...` masked
  - [ ] `token=...` masked

- [ ] **Log rotation**
  - [ ] `internal/logging/rotator.go` implements `RotatingFileWriter`
  - [ ] Rotation triggers at `max_size_mb`
  - [ ] Keeps `max_backups` files
  - [ ] Deletes oldest when exceeding limit
  - [ ] Thread-safe with mutex

- [ ] **Integration**
  - [ ] RotatingFileWriter created when `logging.dir` set
  - [ ] RedactingWriter wraps when `redact_sensitive = true`
  - [ ] Logger outputs to both console and file

- [ ] **Tests pass**
  - [ ] `go test ./internal/logging/...` passes
  - [ ] Redaction patterns tested
  - [ ] Rotation behavior tested

## Phase 5: Frontend ErrorBoundary & About/Update/Diagnostics UI

- [ ] **ErrorBoundary**
  - [ ] `web/src/components/ErrorBoundary.tsx` exists
  - [ ] Catches errors and displays fallback UI
  - [ ] "Reload App" button works
  - [ ] "Export Diagnostic Bundle" button works
  - [ ] "Open Data Folder" button works (desktop only)
  - [ ] Privacy note displayed
  - [ ] Stack trace shown in dev mode, hidden in production

- [ ] **About section**
  - [ ] Displays app name, version, commit, build time, channel
  - [ ] Displays runtime mode (server/desktop)
  - [ ] Displays data directory (desktop only)
  - [ ] Fetches from `GET /api/v1/system/version`

- [ ] **Update section**
  - [ ] Displays current version
  - [ ] Displays latest version
  - [ ] Shows "Update available" indicator
  - [ ] "Download" button opens download_page_url
  - [ ] "Release Notes" link opens notes_url
  - [ ] "Check for Updates" button triggers check
  - [ ] Fetches from `GET /api/v1/update/status`
  - [ ] Posts to `POST /api/v1/update/check`

- [ ] **Diagnostics section**
  - [ ] "Export Diagnostic Bundle" button works
  - [ ] Lists generated bundles with name, size, created_at
  - [ ] Download button for each bundle
  - [ ] Privacy note displayed
  - [ ] Fetches from `GET /api/v1/system/diagnostics`
  - [ ] Posts to `POST /api/v1/system/diagnostics/export`

- [ ] **API client**
  - [ ] `web/src/api/system.ts` includes all new methods
  - [ ] Error handling for all API calls

- [ ] **i18n keys**
  - [ ] All new UI strings have i18n keys
  - [ ] English, Simplified Chinese, Traditional Chinese translations

- [ ] **Build passes**
  - [ ] `cd web && pnpm build` succeeds
  - [ ] No TypeScript errors
  - [ ] No linting errors

## Phase 6: User Documentation

- [ ] **All 18 guides created**
  - [ ] `docs/getting-started.md`
  - [ ] `docs/installation.md`
  - [ ] `docs/desktop.md`
  - [ ] `docs/server-deploy.md`
  - [ ] `docs/configuration.md`
  - [ ] `docs/storage.md`
  - [ ] `docs/qiniu.md`
  - [ ] `docs/import.md`
  - [ ] `docs/search.md`
  - [ ] `docs/organize.md`
  - [ ] `docs/ai.md`
  - [ ] `docs/auth.md`
  - [ ] `docs/backup-restore.md`
  - [ ] `docs/diagnostics.md`
  - [ ] `docs/troubleshooting.md`
  - [ ] `docs/security.md`
  - [ ] `docs/privacy.md`
  - [ ] `docs/release-notes.md`

- [ ] **Documentation quality**
  - [ ] Each guide starts with "This guide covers:" or overview
  - [ ] Concrete commands included where applicable
  - [ ] Links to related docs included
  - [ ] Tables used for config options
  - [ ] "Known limitations" sections where relevant
  - [ ] Consistent formatting across all guides

- [ ] **Documentation index**
  - [ ] `docs/README.md` links to all guides
  - [ ] Task-based navigation ("I want to...")
  - [ ] Role-based navigation (End Users, Admins, Developers)

## Phase 7: Landing Page

- [ ] **Landing page exists**
  - [ ] `landing/index.html` created
  - [ ] Responsive design (mobile, tablet, desktop)
  - [ ] Modern, professional design

- [ ] **Landing page sections**
  - [ ] Hero section with headline and CTA
  - [ ] Features section (9 features)
  - [ ] Use cases section (4 scenarios)
  - [ ] Tech stack section
  - [ ] CTA section for downloads
  - [ ] Footer with links

- [ ] **Landing page quality**
  - [ ] All links work (internal and external)
  - [ ] Images/icons load correctly
  - [ ] Mobile-responsive tested
  - [ ] Accessibility (alt text, semantic HTML)

## Phase 8: CHANGELOG and Release Notes

- [ ] **CHANGELOG.md**
  - [ ] Created in project root
  - [ ] Follows Keep a Changelog format
  - [ ] V1.0.0 section complete with all features
  - [ ] Security section included
  - [ ] Known limitations listed
  - [ ] Upgrade guide from V0.9.x included

- [ ] **Release notes**
  - [ ] `docs/release-notes.md` created
  - [ ] Mirrors CHANGELOG V1.0.0 section
  - [ ] "Upgrading from V0.9" section included
  - [ ] Breaking changes documented (none for V1.0)

## Phase 9: Release Script and Makefile

- [ ] **Release script**
  - [ ] `scripts/release-v1.sh` created and executable
  - [ ] Runs tests before building
  - [ ] Builds frontend
  - [ ] Builds server binary with ldflags
  - [ ] Packages server distribution
  - [ ] Builds desktop app (if Wails available)
  - [ ] Generates `checksums.txt`
  - [ ] Generates `latest.json`
  - [ ] Copies release notes

- [ ] **Makefile targets**
  - [ ] `make release-v1` target added
  - [ ] `make release-v1-test` target added
  - [ ] `make server-build-v1` target added
  - [ ] `.PHONY` declarations updated

- [ ] **Release build works**
  - [ ] `make release-v1 VERSION=1.0.0` succeeds
  - [ ] `dist/` directory created with all artifacts
  - [ ] `dist/checksums.txt` contains SHA256 hashes
  - [ ] `dist/latest.json` valid JSON with correct structure
  - [ ] Server binary runs and reports correct version

## Phase 10: README Rewrite

- [ ] **User-facing top section**
  - [ ] Headline and badges
  - [ ] Download links (desktop and server)
  - [ ] Features list
  - [ ] Quick start guide (desktop)
  - [ ] Quick start guide (server)
  - [ ] Configuration overview
  - [ ] Importing markdown overview
  - [ ] Search & organize overview
  - [ ] AI features overview
  - [ ] Backup & restore overview
  - [ ] Diagnostics overview
  - [ ] Privacy & security overview
  - [ ] Development overview
  - [ ] License
  - [ ] Current limitations
  - [ ] Links to docs

- [ ] **Developer documentation preserved**
  - [ ] Technology stack section
  - [ ] Local development section
  - [ ] API reference section
  - [ ] All V0.x sections preserved

## Phase 11: Final Verification

- [ ] **All tests pass**
  - [ ] `go test ./...` passes
  - [ ] `cd web && pnpm build` succeeds
  - [ ] `make test` passes

- [ ] **Release build verification**
  - [ ] `make clean` succeeds
  - [ ] `make release-v1 VERSION=1.0.0` succeeds
  - [ ] `make test` passes
  - [ ] `cd web && pnpm build` succeeds

- [ ] **Server smoke test**
  - [ ] `./dist/server/markdown-vault --addr :8081` starts
  - [ ] `curl localhost:8081/api/v1/health` returns 200
  - [ ] `curl localhost:8081/api/v1/system/version` returns 401 (auth on)
  - [ ] Setup via UI, login
  - [ ] `curl -b cookie.txt localhost:8081/api/v1/system/version` returns 200 with version info
  - [ ] `curl -X POST -b cookie.txt localhost:8081/api/v1/update/check` returns 200 with update status
  - [ ] `curl -X POST -b cookie.txt localhost:8081/api/v1/system/diagnostics/export` returns 200 with bundle name
  - [ ] Download bundle, verify contents (no app.db, no objects, no secrets)
  - [ ] Server stops cleanly

- [ ] **Desktop smoke test (if Wails available)**
  - [ ] `./dist/desktop/cloud-vault-desktop` starts
  - [ ] Tray icon appears
  - [ ] About shows version
  - [ ] Check for Updates works
  - [ ] Export Diagnostics works
  - [ ] Open Data Folder works
  - [ ] App closes cleanly

- [ ] **Documentation review**
  - [ ] All 18 user guides reviewed
  - [ ] Landing page reviewed
  - [ ] CHANGELOG reviewed
  - [ ] Release notes reviewed
  - [ ] README reviewed

- [ ] **Acceptance criteria**
  - [ ] All 24 acceptance criteria from user spec pass
  - [ ] No breaking changes to V0.1–V0.9 APIs
  - [ ] No new business features added
  - [ ] Auto-update is check-only, no silent installation
  - [ ] Diagnostics are user-initiated, never auto-uploaded
  - [ ] Diagnostic bundle excludes sensitive data
  - [ ] Logs redact secrets
  - [ ] Server and desktop builds both work
  - [ ] Installer signing limitations documented

## Sign-off

- [ ] All phases complete
- [ ] All tests pass
- [ ] Smoke tests pass
- [ ] Documentation reviewed
- [ ] Ready for release

**Release date**: ___________

**Signed off by**: ___________
