package connection

import (
	"testing"
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
