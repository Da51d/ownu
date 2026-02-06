package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabaseURL      string
	JWTSecret        string
	ServerPort       int
	WebAuthnRPID     string
	WebAuthnRPOrigin string
	WebAuthnRPName   string
	// Plaid configuration
	PlaidClientID   string
	PlaidSecret     string
	PlaidEnv        string // sandbox, development, or production
	PlaidWebhookURL string
}

func Load() *Config {
	port, err := strconv.Atoi(getEnv("SERVER_PORT", "8080"))
	if err != nil {
		port = 8080
	}

	return &Config{
		DatabaseURL:      getEnv("DATABASE_URL", "postgres://ownu:ownu_dev_password@localhost:5432/ownu?sslmode=disable"),
		JWTSecret:        getEnv("JWT_SECRET", "change_me_in_production"),
		ServerPort:       port,
		WebAuthnRPID:     getEnv("WEBAUTHN_RP_ID", "localhost"),
		WebAuthnRPOrigin: getEnv("WEBAUTHN_RP_ORIGIN", "http://localhost:5173"),
		WebAuthnRPName:   getEnv("WEBAUTHN_RP_NAME", "OwnU"),
		// Plaid configuration
		PlaidClientID:   getEnv("PLAID_CLIENT_ID", ""),
		PlaidSecret:     getEnv("PLAID_SECRET", ""),
		PlaidEnv:        getEnv("PLAID_ENV", "sandbox"),
		PlaidWebhookURL: getEnv("PLAID_WEBHOOK_URL", ""),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
