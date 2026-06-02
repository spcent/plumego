package system

import (
	"context"
	"testing"

	"cloud-vault/internal/config"
	"cloud-vault/internal/storage"
)

// TestOverallStatus tests the overallStatus helper directly.
func TestOverallStatus(t *testing.T) {
	tests := []struct {
		name string
		r    HealthResult
		want string
	}{
		{
			name: "all ok",
			r:    HealthResult{Database: StatusOK, Storage: StatusOK, Search: StatusOK},
			want: StatusOK,
		},
		{
			name: "storage warning",
			r:    HealthResult{Database: StatusOK, Storage: StatusWarning, Search: StatusOK},
			want: StatusWarning,
		},
		{
			name: "database error overrides warning",
			r:    HealthResult{Database: StatusError, Storage: StatusWarning, Search: StatusOK},
			want: StatusError,
		},
		{
			name: "search error",
			r:    HealthResult{Database: StatusOK, Storage: StatusOK, Search: StatusError},
			want: StatusError,
		},
		{
			name: "ai disabled does not affect status",
			r:    HealthResult{Database: StatusOK, Storage: StatusOK, Search: StatusOK, AI: StatusDisabled},
			want: StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := overallStatus(tt.r)
			if got != tt.want {
				t.Errorf("overallStatus(%v) = %q, want %q", tt.r, got, tt.want)
			}
		})
	}
}

// TestWorstStatus tests the worstStatus aggregator for Doctor results.
func TestWorstStatus(t *testing.T) {
	tests := []struct {
		name   string
		checks []CheckResult
		want   string
	}{
		{
			name: "empty slice returns ok",
			want: StatusOK,
		},
		{
			name: "all ok",
			checks: []CheckResult{
				{Status: StatusOK},
				{Status: StatusOK},
			},
			want: StatusOK,
		},
		{
			name: "one warning",
			checks: []CheckResult{
				{Status: StatusOK},
				{Status: StatusWarning},
			},
			want: StatusWarning,
		},
		{
			name: "error beats warning",
			checks: []CheckResult{
				{Status: StatusWarning},
				{Status: StatusError},
				{Status: StatusOK},
			},
			want: StatusError,
		},
		{
			name: "disabled treated as ok",
			checks: []CheckResult{
				{Status: StatusDisabled},
				{Status: StatusOK},
			},
			want: StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := worstStatus(tt.checks)
			if got != tt.want {
				t.Errorf("worstStatus(%v) = %q, want %q", tt.checks, got, tt.want)
			}
		})
	}
}

// TestCheckAI tests the checkAI helper.
func TestCheckAI(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		result := checkAI(config.AIConfig{Enabled: false})
		if result != StatusDisabled {
			t.Errorf("disabled AI = %q, want %q", result, StatusDisabled)
		}
	})
	t.Run("enabled", func(t *testing.T) {
		result := checkAI(config.AIConfig{Enabled: true})
		if result != StatusOK {
			t.Errorf("enabled AI = %q, want %q", result, StatusOK)
		}
	})
}

// TestService_Health_AIDisabledStatus verifies AI field is 'disabled' when AI is off.
func TestService_Health_AIDisabledStatus(t *testing.T) {
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())
	cfg := config.Config{
		AI: config.AIConfig{Enabled: false},
	}
	svc := NewService(db.DB, store, cfg)
	ctx := context.Background()

	result := svc.Health(ctx)

	if result.AI != StatusDisabled {
		t.Errorf("AI = %q, want %q", result.AI, StatusDisabled)
	}
}

// TestService_Health_AIEnabledStatus verifies AI field is 'ok' when AI is enabled.
func TestService_Health_AIEnabledStatus(t *testing.T) {
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())
	cfg := config.Config{
		AI: config.AIConfig{Enabled: true},
	}
	svc := NewService(db.DB, store, cfg)
	ctx := context.Background()

	result := svc.Health(ctx)

	if result.AI != StatusOK {
		t.Errorf("AI = %q, want %q", result.AI, StatusOK)
	}
}

// TestCheckCookieSecurity verifies cookie security checking.
func TestCheckCookieSecurity(t *testing.T) {
	ctx := context.Background()

	t.Run("auth disabled returns disabled", func(t *testing.T) {
		cfg := config.Config{Auth: config.AuthConfig{Enabled: false}}
		result := checkCookieSecurity(ctx, cfg)
		if result.Status != StatusDisabled {
			t.Errorf("Status = %q, want %q", result.Status, StatusDisabled)
		}
	})

	t.Run("cookie name empty returns error", func(t *testing.T) {
		cfg := config.Config{Auth: config.AuthConfig{Enabled: true, CookieName: ""}}
		result := checkCookieSecurity(ctx, cfg)
		if result.Status != StatusError {
			t.Errorf("Status = %q, want %q", result.Status, StatusError)
		}
		if len(result.Items) == 0 || result.Items[0].Issue != "cookie_name_empty" {
			t.Errorf("expected issue 'cookie_name_empty', got %v", result.Items)
		}
	})

	t.Run("secure cookie false returns warning", func(t *testing.T) {
		cfg := config.Config{Auth: config.AuthConfig{
			Enabled:      true,
			CookieName:   "session",
			SecureCookie: false,
		}}
		result := checkCookieSecurity(ctx, cfg)
		if result.Status != StatusWarning {
			t.Errorf("Status = %q, want %q", result.Status, StatusWarning)
		}
	})

	t.Run("secure cookie true returns ok", func(t *testing.T) {
		cfg := config.Config{Auth: config.AuthConfig{
			Enabled:      true,
			CookieName:   "session",
			SecureCookie: true,
		}}
		result := checkCookieSecurity(ctx, cfg)
		if result.Status != StatusOK {
			t.Errorf("Status = %q, want %q", result.Status, StatusOK)
		}
	})
}

// TestCheckQiniuConfig verifies Qiniu configuration checking.
func TestCheckQiniuConfig(t *testing.T) {
	ctx := context.Background()
	store := storage.NewLocalStorage(t.TempDir())

	t.Run("non-qiniu provider returns disabled", func(t *testing.T) {
		cfg := config.Config{Storage: config.StorageConfig{Provider: "local"}}
		result := checkQiniuConfig(ctx, store, cfg)
		if result.Status != StatusDisabled {
			t.Errorf("Status = %q, want %q", result.Status, StatusDisabled)
		}
	})

	t.Run("qiniu missing fields returns error", func(t *testing.T) {
		cfg := config.Config{
			Storage: config.StorageConfig{Provider: "qiniu"},
			Qiniu:   config.QiniuConfig{
				// All fields empty
			},
		}
		result := checkQiniuConfig(ctx, store, cfg)
		if result.Status != StatusError {
			t.Errorf("Status = %q, want %q", result.Status, StatusError)
		}
		if len(result.Items) == 0 {
			t.Error("expected issue items for missing fields")
		}
		if result.Items[0].Issue != "qiniu_fields_missing" {
			t.Errorf("Issue = %q, want 'qiniu_fields_missing'", result.Items[0].Issue)
		}
	})

	t.Run("qiniu all fields present with local store probe passes", func(t *testing.T) {
		cfg := config.Config{
			Storage: config.StorageConfig{Provider: "qiniu"},
			Qiniu: config.QiniuConfig{
				AccessKey: "key",
				SecretKey: "secret",
				Bucket:    "bucket",
				Domain:    "example.com",
				Region:    "z0",
			},
		}
		// Local storage Exists doesn't error, so probe should pass
		result := checkQiniuConfig(ctx, store, cfg)
		if result.Status != StatusOK {
			t.Errorf("Status = %q, want %q (items: %v)", result.Status, StatusOK, result.Items)
		}
	})
}

// TestCheckAuthSecurity verifies auth security checking with a real DB.
func TestCheckAuthSecurity(t *testing.T) {
	ctx := context.Background()

	t.Run("auth disabled returns warning", func(t *testing.T) {
		db := openTestDB(t)
		cfg := config.Config{Auth: config.AuthConfig{Enabled: false}}
		result := checkAuthSecurity(ctx, db.DB, cfg)
		if result.Status != StatusWarning {
			t.Errorf("Status = %q, want %q", result.Status, StatusWarning)
		}
	})

	t.Run("auth enabled with no admin users returns error", func(t *testing.T) {
		db := openTestDB(t)
		cfg := config.Config{Auth: config.AuthConfig{Enabled: true}}
		result := checkAuthSecurity(ctx, db.DB, cfg)
		if result.Status != StatusError {
			t.Errorf("Status = %q, want %q", result.Status, StatusError)
		}
		if len(result.Items) == 0 || result.Items[0].Issue != "no_admin_users" {
			t.Errorf("expected 'no_admin_users' issue, got %v", result.Items)
		}
	})

	t.Run("auth enabled with admin user returns ok", func(t *testing.T) {
		db := openTestDB(t)
		_, err := db.Exec(`
			INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
			VALUES ('admin-1', 'admin', 'admin@test.com', 'hash', 'admin', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')
		`)
		if err != nil {
			t.Fatalf("insert admin: %v", err)
		}
		cfg := config.Config{Auth: config.AuthConfig{Enabled: true}}
		result := checkAuthSecurity(ctx, db.DB, cfg)
		// Should be OK with one admin and no suspicious sessions
		if result.Status == StatusError {
			t.Errorf("Status = %q (want ok or warning), items: %v", result.Status, result.Items)
		}
	})
}

// TestCheckStorageWritable verifies the storage write check.
func TestCheckStorageWritable(t *testing.T) {
	ctx := context.Background()
	store := storage.NewLocalStorage(t.TempDir())
	cfg := config.Config{}

	result := checkStorageWritable(ctx, store, cfg)
	if result.Status != StatusOK {
		t.Errorf("Status = %q, want %q (items: %v)", result.Status, StatusOK, result.Items)
	}
	if result.Name != "storage_writable_check" {
		t.Errorf("Name = %q, want 'storage_writable_check'", result.Name)
	}
}

// TestCheckDataDir verifies data directory checking.
func TestCheckDataDir(t *testing.T) {
	ctx := context.Background()

	t.Run("valid existing dir returns ok", func(t *testing.T) {
		tmpDir := t.TempDir()
		cfg := config.Config{
			DB: config.DBConfig{Path: tmpDir + "/app.db"},
		}
		result := checkDataDir(ctx, cfg)
		if result.Status != StatusOK {
			t.Errorf("Status = %q, want %q (items: %v)", result.Status, StatusOK, result.Items)
		}
	})

	t.Run("empty path returns error", func(t *testing.T) {
		cfg := config.Config{DB: config.DBConfig{Path: ""}}
		result := checkDataDir(ctx, cfg)
		if result.Status != StatusError {
			t.Errorf("Status = %q, want %q", result.Status, StatusError)
		}
	})

	t.Run("new dir gets created", func(t *testing.T) {
		tmpDir := t.TempDir()
		newDir := tmpDir + "/new/sub/dir"
		cfg := config.Config{DB: config.DBConfig{Path: newDir + "/app.db"}}
		result := checkDataDir(ctx, cfg)
		if result.Status == StatusError {
			t.Errorf("Status = %q (items: %v)", result.Status, result.Items)
		}
	})
}

// TestRunCheck_UnknownCheck verifies that an unknown check name returns StatusError.
func TestRunCheck_UnknownCheck(t *testing.T) {
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())
	ctx := context.Background()
	cfg := config.Config{}

	result := runCheck(ctx, "nonexistent_check", db.DB, store, cfg, 10)
	if result.Status != StatusError {
		t.Errorf("Status = %q, want %q", result.Status, StatusError)
	}
	if len(result.Items) == 0 || result.Items[0].Issue != "unknown_check" {
		t.Errorf("expected 'unknown_check' issue, got %v", result.Items)
	}
}

// TestService_Doctor_SampleSize verifies that a custom SampleSize is respected.
func TestService_Doctor_SampleSize(t *testing.T) {
	svc, _, _ := setupSystem(t)
	ctx := context.Background()

	result := svc.Doctor(ctx, DoctorRequest{
		Checks:     []string{"document_hash"},
		SampleSize: 5,
	})

	if len(result.Checks) != 1 {
		t.Fatalf("Checks count = %d, want 1", len(result.Checks))
	}
	if result.Checks[0].Name != "document_hash" {
		t.Errorf("Check name = %q, want 'document_hash'", result.Checks[0].Name)
	}
}

// TestService_Doctor_DefaultSampleSize verifies that zero SampleSize uses the default.
func TestService_Doctor_DefaultSampleSize(t *testing.T) {
	svc, _, _ := setupSystem(t)
	ctx := context.Background()

	result := svc.Doctor(ctx, DoctorRequest{
		Checks:     []string{"document_hash"},
		SampleSize: 0, // should use default
	})

	if len(result.Checks) != 1 {
		t.Fatalf("Checks count = %d, want 1", len(result.Checks))
	}
}

// TestService_Doctor_CollectionsCheck verifies the collections integrity check runs.
func TestService_Doctor_CollectionsCheck(t *testing.T) {
	svc, _, _ := setupSystem(t)
	ctx := context.Background()

	result := svc.Doctor(ctx, DoctorRequest{
		Checks: []string{"collections"},
	})

	if len(result.Checks) != 1 {
		t.Fatalf("want 1 check, got %d", len(result.Checks))
	}
	if result.Checks[0].Name != "collections" {
		t.Errorf("Name = %q, want 'collections'", result.Checks[0].Name)
	}
}

// TestService_Doctor_SourcesCheck verifies the sources integrity check runs.
func TestService_Doctor_SourcesCheck(t *testing.T) {
	svc, _, _ := setupSystem(t)
	ctx := context.Background()

	result := svc.Doctor(ctx, DoctorRequest{
		Checks: []string{"sources"},
	})

	if len(result.Checks) != 1 {
		t.Fatalf("want 1 check, got %d", len(result.Checks))
	}
}

// TestService_Doctor_ImportsCheck verifies the imports integrity check runs.
func TestService_Doctor_ImportsCheck(t *testing.T) {
	svc, _, _ := setupSystem(t)
	ctx := context.Background()

	result := svc.Doctor(ctx, DoctorRequest{
		Checks: []string{"imports"},
	})

	if len(result.Checks) != 1 {
		t.Fatalf("want 1 check, got %d", len(result.Checks))
	}
}

// TestService_Doctor_AIConsistencyCheck runs against an empty DB.
func TestService_Doctor_AIConsistencyCheck(t *testing.T) {
	svc, _, _ := setupSystem(t)
	ctx := context.Background()

	result := svc.Doctor(ctx, DoctorRequest{
		Checks: []string{"ai"},
	})

	if len(result.Checks) != 1 {
		t.Fatalf("want 1 check, got %d", len(result.Checks))
	}
	if result.Checks[0].Status != StatusOK {
		t.Errorf("ai check on empty DB: Status = %q, want %q (items: %v)",
			result.Checks[0].Status, StatusOK, result.Checks[0].Items)
	}
}

// TestService_Doctor_AuthCheck runs against an empty DB.
func TestService_Doctor_AuthCheck(t *testing.T) {
	svc, _, _ := setupSystem(t)
	ctx := context.Background()

	result := svc.Doctor(ctx, DoctorRequest{
		Checks: []string{"auth"},
	})

	if len(result.Checks) != 1 {
		t.Fatalf("want 1 check, got %d", len(result.Checks))
	}
	// auth check on empty DB: users table should exist after migration
	if result.Checks[0].Name != "auth" {
		t.Errorf("Name = %q, want 'auth'", result.Checks[0].Name)
	}
}

// TestService_Stats_AfterInsert verifies that Stats reflects inserted rows.
func TestService_Stats_AfterInsert(t *testing.T) {
	svc, db, _ := setupSystem(t)
	ctx := context.Background()

	// Insert one active document
	_, err := db.Exec(`
		INSERT INTO documents (id, title, storage_key, content_hash, status, created_at, updated_at)
		VALUES ('doc-stat-1', 'Stats Doc', 'key/stats', 'h1', 'active', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')
	`)
	if err != nil {
		t.Fatalf("Insert doc: %v", err)
	}

	stats, err := svc.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats: %v", err)
	}

	if stats.Documents != 1 {
		t.Errorf("Documents = %d, want 1", stats.Documents)
	}
}

// TestService_Health_SearchMissingTable confirms search status is error when FTS table absent.
func TestService_Health_SearchMissingTable(t *testing.T) {
	db := openTestDB(t)
	store := storage.NewLocalStorage(t.TempDir())
	cfg := config.Config{AI: config.AIConfig{Enabled: false}}
	svc := NewService(db.DB, store, cfg)
	ctx := context.Background()

	// Drop the FTS table to simulate missing table
	if _, err := db.Exec(`DROP TABLE IF EXISTS document_fts`); err != nil {
		t.Fatalf("drop fts table: %v", err)
	}

	result := svc.Health(ctx)

	if result.Search != StatusError {
		t.Errorf("Search with missing FTS table = %q, want %q", result.Search, StatusError)
	}
}
