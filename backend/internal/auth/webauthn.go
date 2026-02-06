package auth

import (
	"fmt"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/ownu/ownu/internal/config"
)

// WebAuthnService handles WebAuthn operations
type WebAuthnService struct {
	webauthn *webauthn.WebAuthn
}

// NewWebAuthnService creates a new WebAuthn service
func NewWebAuthnService(cfg *config.Config) (*WebAuthnService, error) {
	wconfig := &webauthn.Config{
		RPDisplayName: cfg.WebAuthnRPName,
		RPID:          cfg.WebAuthnRPID,
		RPOrigins:     []string{cfg.WebAuthnRPOrigin},
	}

	w, err := webauthn.New(wconfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create webauthn: %w", err)
	}

	return &WebAuthnService{webauthn: w}, nil
}

// WebAuthn returns the underlying webauthn instance
func (s *WebAuthnService) WebAuthn() *webauthn.WebAuthn {
	return s.webauthn
}
