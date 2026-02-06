package config

import (
	"os"
	"testing"
)

func TestLoad(t *testing.T) {
	// Test with default values
	cfg := Load()

	if cfg.ServerPort != 8080 {
		t.Errorf("Load() ServerPort = %d, want 8080", cfg.ServerPort)
	}

	if cfg.WebAuthnRPID != "localhost" {
		t.Errorf("Load() WebAuthnRPID = %s, want localhost", cfg.WebAuthnRPID)
	}

	if cfg.WebAuthnRPName != "OwnU" {
		t.Errorf("Load() WebAuthnRPName = %s, want OwnU", cfg.WebAuthnRPName)
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("JWT_SECRET", "test-secret")
	os.Setenv("WEBAUTHN_RP_ID", "example.com")
	defer func() {
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("JWT_SECRET")
		os.Unsetenv("WEBAUTHN_RP_ID")
	}()

	cfg := Load()

	if cfg.ServerPort != 9090 {
		t.Errorf("Load() ServerPort = %d, want 9090", cfg.ServerPort)
	}

	if cfg.JWTSecret != "test-secret" {
		t.Errorf("Load() JWTSecret = %s, want test-secret", cfg.JWTSecret)
	}

	if cfg.WebAuthnRPID != "example.com" {
		t.Errorf("Load() WebAuthnRPID = %s, want example.com", cfg.WebAuthnRPID)
	}
}

func TestLoadInvalidPort(t *testing.T) {
	os.Setenv("SERVER_PORT", "invalid")
	defer os.Unsetenv("SERVER_PORT")

	cfg := Load()

	// Should fall back to default
	if cfg.ServerPort != 8080 {
		t.Errorf("Load() ServerPort = %d, want 8080 (default)", cfg.ServerPort)
	}
}
