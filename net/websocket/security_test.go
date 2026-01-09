package websocket

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/spcent/plumego/security/password"
)

func TestValidateSecurityConfig(t *testing.T) {
	tests := []struct {
		name    string
		cfg     SecurityConfig
		wantErr bool
	}{
		{
			name: "valid config with 32-byte secret",
			cfg: SecurityConfig{
				JWTSecret:               make([]byte, 32),
				MinJWTSecretLength:      32,
				EnforcePasswordStrength: true,
			},
			wantErr: false,
		},
		{
			name: "invalid config - secret too short",
			cfg: SecurityConfig{
				JWTSecret:               make([]byte, 16),
				MinJWTSecretLength:      32,
				EnforcePasswordStrength: true,
			},
			wantErr: true,
		},
		{
			name: "invalid config - weak secret pattern",
			cfg: SecurityConfig{
				JWTSecret:               []byte("secret12345678901234567890123456"), // 32 bytes with weak pattern
				MinJWTSecretLength:      32,
				EnforcePasswordStrength: true,
			},
			wantErr: false, // Should warn but not fail
		},
		{
			name: "empty secret",
			cfg: SecurityConfig{
				JWTSecret:               []byte{},
				MinJWTSecretLength:      32,
				EnforcePasswordStrength: true,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSecurityConfig(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSecurityConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateWebSocketKey(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid key",
			key:     "dGhlIHNhbXBsZSBub25jZQ==",
			wantErr: false,
		},
		{
			name:    "empty key",
			key:     "",
			wantErr: true,
		},
		{
			name:    "invalid base64",
			key:     "not-valid-base64!!!",
			wantErr: true,
		},
		{
			name:    "wrong length (not 16 bytes)",
			key:     base64.StdEncoding.EncodeToString([]byte("short")),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWebSocketKey(tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateWebSocketKey() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRoomPassword(t *testing.T) {
	config := password.DefaultPasswordStrengthConfig()

	tests := []struct {
		name    string
		pwd     string
		config  password.PasswordStrengthConfig
		enforce bool
		wantErr bool
	}{
		{
			name:    "valid strong password",
			pwd:     "StrongP@ssw0rd",
			config:  config,
			enforce: true,
			wantErr: false,
		},
		{
			name:    "empty password",
			pwd:     "",
			config:  config,
			enforce: true,
			wantErr: true,
		},
		{
			name:    "weak password with enforcement",
			pwd:     "weak",
			config:  config,
			enforce: true,
			wantErr: true,
		},
		{
			name:    "weak password without enforcement",
			pwd:     "weak",
			config:  config,
			enforce: false,
			wantErr: false, // Should warn but not fail
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateRoomPassword(tt.pwd, tt.config, tt.enforce)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRoomPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateSecureSecret(t *testing.T) {
	secret, err := GenerateSecureSecret(32)
	if err != nil {
		t.Fatalf("GenerateSecureSecret() error = %v", err)
	}
	if len(secret) != 32 {
		t.Errorf("GenerateSecureSecret() length = %d, want 32", len(secret))
	}

	// Verify it's random
	secret2, _ := GenerateSecureSecret(32)
	if string(secret) == string(secret2) {
		t.Error("GenerateSecureSecret() generated same secret twice")
	}

	// Test too short
	_, err = GenerateSecureSecret(16)
	if err == nil {
		t.Error("GenerateSecureSecret() should fail with length < 32")
	}
}

func TestSecureRoomAuth(t *testing.T) {
	secret := make([]byte, 32)
	rand.Read(secret)

	cfg := SecurityConfig{
		JWTSecret:               secret,
		MinJWTSecretLength:      32,
		EnforcePasswordStrength: true,
		RoomPasswordConfig:      password.DefaultPasswordStrengthConfig(),
		EnableDebugLogging:      false,
	}

	auth, err := NewSecureRoomAuth(secret, cfg)
	if err != nil {
		t.Fatalf("NewSecureRoomAuth() error = %v", err)
	}

	// Test setting valid password
	err = auth.SetRoomPassword("test", "StrongP@ssw0rd")
	if err != nil {
		t.Errorf("SetRoomPassword() with valid password failed: %v", err)
	}

	// Test setting weak password
	err = auth.SetRoomPassword("test2", "weak")
	if err == nil {
		t.Error("SetRoomPassword() should fail with weak password")
	}

	// Test JWT verification
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"user1","exp":` +
		fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()) + `}`))

	// Create signature
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(header + "." + payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	token := header + "." + payload + "." + sig

	claims, err := auth.VerifyJWT(token)
	if err != nil {
		t.Errorf("VerifyJWT() failed: %v", err)
	}
	if claims["sub"] != "user1" {
		t.Errorf("VerifyJWT() returned wrong claims: %v", claims)
	}

	// Test metrics
	metrics := GetSecurityMetrics()
	if metrics.SuccessfulAuthentications == 0 {
		t.Error("Metrics should track successful authentications")
	}
}

func TestSecurityMetrics(t *testing.T) {
	ResetSecurityMetrics()

	metrics := GetSecurityMetrics()
	if metrics.InvalidJWTSecrets != 0 || metrics.WeakRoomPasswords != 0 {
		t.Error("ResetSecurityMetrics() failed")
	}

	// Simulate some events
	securityMetrics.InvalidJWTSecrets = 5
	securityMetrics.WeakRoomPasswords = 3

	metrics = GetSecurityMetrics()
	if metrics.InvalidJWTSecrets != 5 || metrics.WeakRoomPasswords != 3 {
		t.Error("GetSecurityMetrics() returned wrong values")
	}
}

func TestHubSecurityIntegration(t *testing.T) {
	// Test that hub with security config works
	cfg := HubConfig{
		WorkerCount:           2,
		JobQueueSize:          10,
		MaxConnections:        10,
		MaxRoomConnections:    5,
		EnableDebugLogging:    false,
		EnableMetrics:         true,
		RejectOnQueueFull:     true,
		EnableSecurityMetrics: true,
	}

	hub := NewHubWithConfig(cfg)
	defer hub.Stop()

	// Verify config was applied
	if hub.config.EnableSecurityMetrics != true {
		t.Error("Hub config not properly applied")
	}

	// Test metrics collection
	metrics := hub.Metrics()
	if metrics.MaxConnections != 10 {
		t.Errorf("Expected max connections 10, got %d", metrics.MaxConnections)
	}
}

func TestHubBroadcastWithSecurity(t *testing.T) {
	cfg := HubConfig{
		WorkerCount:           2,
		JobQueueSize:          2, // Small queue to test overflow
		MaxConnections:        10,
		EnableDebugLogging:    true,
		EnableMetrics:         true,
		RejectOnQueueFull:     true,
		EnableSecurityMetrics: true,
	}

	hub := NewHubWithConfig(cfg)
	defer hub.Stop()

	// Create mock connections
	conn1, _ := createMockConnection(t)
	conn2, _ := createMockConnection(t)
	defer conn1.Close()
	defer conn2.Close()

	// Join room
	hub.Join("test", conn1)
	hub.Join("test", conn2)

	// Broadcast multiple messages to fill queue
	for i := 0; i < 10; i++ {
		hub.BroadcastRoom("test", OpcodeText, []byte("test message"))
	}

	// Give workers time to process
	time.Sleep(100 * time.Millisecond)

	// Check that metrics were updated
	metrics := GetSecurityMetrics()
	// Some messages might have been dropped due to queue full
	if metrics.BroadcastQueueFull > 0 {
		t.Logf("Broadcast queue full events: %d", metrics.BroadcastQueueFull)
	}
}

func TestHubConnectionLimitsSecurity(t *testing.T) {
	cfg := HubConfig{
		WorkerCount:           2,
		JobQueueSize:          10,
		MaxConnections:        2,
		MaxRoomConnections:    1,
		EnableMetrics:         true,
		EnableSecurityMetrics: true,
	}

	hub := NewHubWithConfig(cfg)
	defer hub.Stop()

	conn1, _ := createMockConnection(t)
	conn2, _ := createMockConnection(t)
	conn3, _ := createMockConnection(t)
	defer conn1.Close()
	defer conn2.Close()
	defer conn3.Close()

	// First connection should succeed
	if err := hub.TryJoin("room1", conn1); err != nil {
		t.Errorf("First connection failed: %v", err)
	}

	// Second connection to same room should fail (room limit)
	if err := hub.TryJoin("room1", conn2); err != ErrRoomFull {
		t.Errorf("Expected ErrRoomFull, got %v", err)
	}

	// Second connection to different room should succeed
	if err := hub.TryJoin("room2", conn2); err != nil {
		t.Errorf("Second connection to different room failed: %v", err)
	}

	// Third connection should fail (hub limit)
	if err := hub.TryJoin("room3", conn3); err != ErrHubFull {
		t.Errorf("Expected ErrHubFull, got %v", err)
	}

	// Check metrics
	metrics := hub.Metrics()
	if metrics.RejectedTotal != 2 {
		t.Errorf("Expected 2 rejections, got %d", metrics.RejectedTotal)
	}
}

func TestSecurityEventLogging(t *testing.T) {
	cfg := SecurityConfig{
		JWTSecret:          make([]byte, 32),
		MinJWTSecretLength: 32,
		EnableDebugLogging: true,
		EnableMetrics:      true,
		RejectOnQueueFull:  true,
	}

	// Test event logging
	details := map[string]any{
		"test":  "value",
		"count": 42,
	}

	// This should not panic
	LogSecurityEvent("test_event", details, cfg)
}
