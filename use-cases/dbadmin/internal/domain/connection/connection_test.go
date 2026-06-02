package connection

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	kvstore "github.com/spcent/plumego/store/kv"
)

func TestSanitizeMongoURI(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty URI",
			input:    "",
			expected: "",
		},
		{
			name:     "URI without credentials",
			input:    "mongodb://localhost:27017",
			expected: "mongodb://localhost:27017",
		},
		{
			name:     "URI with username only",
			input:    "mongodb://admin@localhost:27017",
			expected: "mongodb://admin@localhost:27017",
		},
		{
			name:     "URI with username and password",
			input:    "mongodb://admin:secret123@localhost:27017",
			expected: "mongodb://admin@localhost:27017",
		},
		{
			name:     "URI with complex password",
			input:    "mongodb://user:p@ss:w0rd!@cluster.example.com:27017",
			expected: "mongodb://user@cluster.example.com:27017",
		},
		{
			name:     "mongodb+srv with password",
			input:    "mongodb+srv://admin:password@cluster.mongodb.net",
			expected: "mongodb+srv://admin@cluster.mongodb.net",
		},
		{
			name:     "URI with database and options",
			input:    "mongodb://user:pass@host:27017/mydb?retryWrites=true",
			expected: "mongodb://user@host:27017/mydb?retryWrites=true",
		},
		{
			name:     "invalid URI without scheme",
			input:    "localhost:27017",
			expected: "localhost:27017",
		},
		{
			name:     "URI with special characters in password",
			input:    "mongodb://user:p@ss/w0rd#123@host:27017",
			expected: "mongodb://user@host:27017",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeMongoURI(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeMongoURI(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConnection_Redact(t *testing.T) {
	t.Run("redacts all sensitive fields", func(t *testing.T) {
		c := &Connection{
			ID:       "test-id",
			Name:     "Test Connection",
			Driver:   DriverMongoDB,
			Password: "db-password",
			MongoURI: "mongodb://admin:secret@localhost:27017",
		}

		c.Redact()

		if c.Password != "" {
			t.Errorf("Password should be empty, got %q", c.Password)
		}
		if c.MongoURI != "mongodb://admin@localhost:27017" {
			t.Errorf("MongoURI should be sanitized, got %q", c.MongoURI)
		}
	})

	t.Run("redacts Elasticsearch credentials", func(t *testing.T) {
		c := &Connection{
			ID:         "test-id",
			Name:       "Test ES Connection",
			Driver:     DriverElasticsearch,
			ESNodes:    []string{"http://localhost:9200"},
			ESAPIKey:   "secret-api-key",
			ESPassword: "es-password",
		}

		c.Redact()

		if c.ESAPIKey != "" {
			t.Errorf("ESAPIKey should be empty, got %q", c.ESAPIKey)
		}
		if c.ESPassword != "" {
			t.Errorf("ESPassword should be empty, got %q", c.ESPassword)
		}
	})

	t.Run("preserves non-sensitive fields", func(t *testing.T) {
		c := &Connection{
			ID:       "test-id",
			Name:     "Test Connection",
			Driver:   DriverMySQL,
			Host:     "localhost",
			Port:     3306,
			Database: "mydb",
			Username: "admin",
			Password: "secret",
			Readonly: true,
		}

		c.Redact()

		if c.ID != "test-id" {
			t.Errorf("ID should be preserved")
		}
		if c.Name != "Test Connection" {
			t.Errorf("Name should be preserved")
		}
		if c.Host != "localhost" {
			t.Errorf("Host should be preserved")
		}
		if c.Port != 3306 {
			t.Errorf("Port should be preserved")
		}
		if c.Database != "mydb" {
			t.Errorf("Database should be preserved")
		}
		if c.Username != "admin" {
			t.Errorf("Username should be preserved")
		}
		if !c.Readonly {
			t.Errorf("Readonly should be preserved")
		}
		if c.Password != "" {
			t.Errorf("Password should be redacted")
		}
	})

	t.Run("handles empty connection", func(t *testing.T) {
		c := &Connection{}
		c.Redact()

		if c.Password != "" {
			t.Errorf("Password should be empty")
		}
		if c.ESAPIKey != "" {
			t.Errorf("ESAPIKey should be empty")
		}
		if c.ESPassword != "" {
			t.Errorf("ESPassword should be empty")
		}
		if c.MongoURI != "" {
			t.Errorf("MongoURI should be empty")
		}
	})
}

func TestStoreRejectsPersistentCredentialsWithoutEncryptionKey(t *testing.T) {
	kv, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	store, err := NewStore(kv, "")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	err = store.Create(&Connection{
		Name:         "mysql",
		Driver:       DriverMySQL,
		Host:         "localhost",
		Port:         3306,
		Username:     "root",
		Password:     "secret",
		SavePassword: true,
	})
	if !errors.Is(err, ErrEncryptionRequired) {
		t.Fatalf("Create error = %v, want ErrEncryptionRequired", err)
	}
}

func TestStoreEncryptsAndDecryptsDatasourceCredentials(t *testing.T) {
	kv, err := kvstore.NewKVStore(kvstore.Options{DataDir: t.TempDir()})
	if err != nil {
		t.Fatalf("NewKVStore: %v", err)
	}
	store, err := NewStore(kv, strings.Repeat("1", 64))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	conn := &Connection{
		Name:       "mongo",
		Driver:     DriverMongoDB,
		MongoURI:   "mongodb://user:pass@localhost:27017",
		Readonly:   true,
		ESPassword: "es-pass",
		ESAPIKey:   "api-key",
	}
	if err := store.Create(conn); err != nil {
		t.Fatalf("Create: %v", err)
	}

	raw, err := kv.Get(idKeyPrefix + conn.ID)
	if err != nil {
		t.Fatalf("raw get: %v", err)
	}
	var stored Connection
	if err := json.Unmarshal(raw, &stored); err != nil {
		t.Fatalf("unmarshal raw connection: %v", err)
	}
	for name, value := range map[string]string{
		"mongo_uri":   stored.MongoURI,
		"es_password": stored.ESPassword,
		"es_api_key":  stored.ESAPIKey,
	} {
		for _, secret := range []string{"user:pass", "es-pass", "api-key"} {
			if strings.Contains(value, secret) {
				t.Fatalf("stored %s contains secret %q: %s", name, secret, value)
			}
		}
	}

	got, err := store.Get(conn.ID)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.MongoURI != "mongodb://user:pass@localhost:27017" {
		t.Fatalf("MongoURI = %q, want original", got.MongoURI)
	}
	if got.ESPassword != "es-pass" || got.ESAPIKey != "api-key" {
		t.Fatalf("ES credentials not decrypted: password=%q apiKey=%q", got.ESPassword, got.ESAPIKey)
	}
}
