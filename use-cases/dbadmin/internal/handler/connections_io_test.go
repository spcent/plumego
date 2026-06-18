package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	kvstore "github.com/spcent/plumego/store/kv"

	"dbadmin/internal/dbmanager"
	"dbadmin/internal/domain/connection"
	"dbadmin/internal/esmanager"
	"dbadmin/internal/mongomanager"
	"dbadmin/internal/redismanager"
)

func newTestConnectionHandler(t *testing.T, encryptionKey string) ConnectionHandler {
	t.Helper()
	kv, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("create kv store: %v", err)
	}
	store, err := connection.NewStore(kv, encryptionKey)
	if err != nil {
		t.Fatalf("create connection store: %v", err)
	}
	return ConnectionHandler{
		Connections:  store,
		Manager:      dbmanager.NewManager(),
		RedisManager: redismanager.NewManager(),
		MongoManager: mongomanager.NewManager(),
		ESManager:    esmanager.NewManager(),
		Logger:       testLogger{},
	}
}

// TestConnectionExportRedactsSecrets verifies that Export never returns
// plaintext or encrypted-at-rest secret fields, matching the same redaction
// applied to List/Get API responses.
func TestConnectionExportRedactsSecrets(t *testing.T) {
	h := newTestConnectionHandler(t, strings.Repeat("1", 64))

	if err := h.Connections.Create(&connection.Connection{
		Name:         "mysql-prod",
		Driver:       connection.DriverMySQL,
		Host:         "localhost",
		Port:         3306,
		Username:     "root",
		Password:     "super-secret",
		SavePassword: true,
	}); err != nil {
		t.Fatalf("create connection: %v", err)
	}
	if err := h.Connections.Create(&connection.Connection{
		Name:       "mongo-prod",
		Driver:     connection.DriverMongoDB,
		MongoURI:   "mongodb://user:pass@localhost:27017",
		ESPassword: "es-pass",
		ESAPIKey:   "api-key",
	}); err != nil {
		t.Fatalf("create connection: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/connections/export", nil)
	rec := httptest.NewRecorder()
	h.Export(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data connectionExportDocument `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Data.Connections) != 2 {
		t.Fatalf("got %d connections, want 2", len(body.Data.Connections))
	}
	for _, c := range body.Data.Connections {
		if c.Password != "" {
			t.Errorf("connection %q: Password not redacted: %q", c.Name, c.Password)
		}
		if c.ESPassword != "" {
			t.Errorf("connection %q: ESPassword not redacted: %q", c.Name, c.ESPassword)
		}
		if c.ESAPIKey != "" {
			t.Errorf("connection %q: ESAPIKey not redacted: %q", c.Name, c.ESAPIKey)
		}
		if strings.Contains(c.MongoURI, "pass") {
			t.Errorf("connection %q: MongoURI still contains credentials: %q", c.Name, c.MongoURI)
		}
	}

	// Raw response body must not leak the plaintext secrets anywhere, not
	// just on the typed fields above.
	if strings.Contains(rec.Body.String(), "super-secret") {
		t.Error("export response body contains plaintext password")
	}
	if strings.Contains(rec.Body.String(), "es-pass") || strings.Contains(rec.Body.String(), "api-key") {
		t.Error("export response body contains plaintext Elasticsearch secrets")
	}
}

// TestConnectionImportPartialSuccess verifies that Import processes each
// entry independently: a bad entry is reported in the result list without
// blocking the valid entries in the same batch.
func TestConnectionImportPartialSuccess(t *testing.T) {
	h := newTestConnectionHandler(t, "")

	doc := connectionExportDocument{
		FormatVersion: connectionExportFormatVersion,
		Connections: []*connection.Connection{
			{Name: "valid-mysql", Driver: connection.DriverMySQL, Host: "localhost", Port: 3306},
			{Name: "invalid-missing-driver-fields", Driver: connection.DriverSQLite}, // missing file_path
			{Name: "valid-sqlite", Driver: connection.DriverSQLite, FilePath: "/tmp/test.db"},
		},
	}
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal doc: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/connections/import", bytes.NewReader(data))
	rec := httptest.NewRecorder()
	h.Import(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data connectionImportSummary `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.Imported != 2 {
		t.Errorf("Imported = %d, want 2", body.Data.Imported)
	}
	if body.Data.Failed != 1 {
		t.Errorf("Failed = %d, want 1", body.Data.Failed)
	}
	if len(body.Data.Results) != 3 {
		t.Fatalf("got %d results, want 3", len(body.Data.Results))
	}
	if body.Data.Results[1].Success {
		t.Error("invalid entry should not be marked success")
	}
	if body.Data.Results[1].Error == "" {
		t.Error("invalid entry should carry an error message")
	}

	conns, err := h.Connections.List()
	if err != nil {
		t.Fatalf("list connections: %v", err)
	}
	if len(conns) != 2 {
		t.Fatalf("store has %d connections, want 2", len(conns))
	}
}

// TestConnectionImportDoesNotRequirePassword confirms that an imported
// connection without a password (the normal case, since exports are
// redacted) still validates and saves successfully.
func TestConnectionImportDoesNotRequirePassword(t *testing.T) {
	h := newTestConnectionHandler(t, "")

	doc := connectionExportDocument{
		Connections: []*connection.Connection{
			{Name: "no-password", Driver: connection.DriverMySQL, Host: "localhost", Port: 3306, Username: "root"},
		},
	}
	data, _ := json.Marshal(doc)

	req := httptest.NewRequest(http.MethodPost, "/api/connections/import", bytes.NewReader(data))
	rec := httptest.NewRecorder()
	h.Import(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}
	var body struct {
		Data connectionImportSummary `json:"data"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Data.Imported != 1 || body.Data.Failed != 0 {
		t.Fatalf("got imported=%d failed=%d, want imported=1 failed=0", body.Data.Imported, body.Data.Failed)
	}
}

// TestConnectionImportIgnoresSuppliedSecretsAndID verifies that even if an
// import document contains a password/ID (e.g. a hand-edited file), Import
// strips them rather than persisting attacker- or stale-supplied secrets.
func TestConnectionImportIgnoresSuppliedSecretsAndID(t *testing.T) {
	h := newTestConnectionHandler(t, strings.Repeat("2", 64))

	doc := connectionExportDocument{
		Connections: []*connection.Connection{
			{
				ID:           "should-be-ignored",
				Name:         "smuggled-secret",
				Driver:       connection.DriverMySQL,
				Host:         "localhost",
				Port:         3306,
				Password:     "smuggled-password",
				SavePassword: true,
			},
		},
	}
	data, _ := json.Marshal(doc)

	req := httptest.NewRequest(http.MethodPost, "/api/connections/import", bytes.NewReader(data))
	rec := httptest.NewRecorder()
	h.Import(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200; body=%s", rec.Code, rec.Body.String())
	}

	conns, err := h.Connections.List()
	if err != nil {
		t.Fatalf("list connections: %v", err)
	}
	if len(conns) != 1 {
		t.Fatalf("got %d connections, want 1", len(conns))
	}
	if conns[0].ID == "should-be-ignored" {
		t.Error("imported connection kept the supplied ID instead of generating a new one")
	}
	got, err := h.Connections.Get(conns[0].ID)
	if err != nil {
		t.Fatalf("get connection: %v", err)
	}
	if got.Password != "" {
		t.Errorf("imported connection persisted a supplied password: %q", got.Password)
	}
}

// TestConnectionImportEmptyBatch verifies an empty connections list is
// rejected with a clear error rather than silently succeeding.
func TestConnectionImportEmptyBatch(t *testing.T) {
	h := newTestConnectionHandler(t, "")

	data, _ := json.Marshal(connectionExportDocument{})
	req := httptest.NewRequest(http.MethodPost, "/api/connections/import", bytes.NewReader(data))
	rec := httptest.NewRecorder()
	h.Import(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}

// TestConnectionImportAllInvalidReturns400 verifies that a batch where every
// entry fails validation reports failure via the HTTP status, not just the
// per-entry result list.
func TestConnectionImportAllInvalidReturns400(t *testing.T) {
	h := newTestConnectionHandler(t, "")

	doc := connectionExportDocument{
		Connections: []*connection.Connection{
			{Name: "missing-host", Driver: connection.DriverMySQL},
		},
	}
	data, _ := json.Marshal(doc)
	req := httptest.NewRequest(http.MethodPost, "/api/connections/import", bytes.NewReader(data))
	rec := httptest.NewRecorder()
	h.Import(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400; body=%s", rec.Code, rec.Body.String())
	}
}
