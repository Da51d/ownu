package api

import (
	"encoding/base64"
	"io"
	"net/http"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ownu/ownu/internal/auth"
	"github.com/ownu/ownu/internal/crypto"
	"github.com/ownu/ownu/internal/models"
	"github.com/ownu/ownu/internal/repository"
)

var sessionStore = auth.NewSessionStore()

// RegisterBeginRequest is the request for starting registration
type RegisterBeginRequest struct {
	Username string `json:"username"`
}

// RegisterBeginResponse is the response for starting registration
type RegisterBeginResponse struct {
	Options        interface{} `json:"options"`
	RecoveryPhrase string      `json:"recovery_phrase"`
	SessionID      string      `json:"session_id"`
}

// RegisterFinishRequest is the request for completing registration
type RegisterFinishRequest struct {
	SessionID  string `json:"session_id"`
	Credential string `json:"credential"`
	PRFOutput  string `json:"prf_output,omitempty"`
}

// LoginBeginRequest is the request for starting login
type LoginBeginRequest struct {
	Username string `json:"username"`
}

// LoginFinishRequest is the request for completing login
type LoginFinishRequest struct {
	SessionID  string `json:"session_id"`
	Credential string `json:"credential"`
	PRFOutput  string `json:"prf_output,omitempty"`
}

// AuthResponse is the response after successful authentication
type AuthResponse struct {
	Token string `json:"token"`
	User  struct {
		ID       string `json:"id"`
		Username string `json:"username"`
	} `json:"user"`
}

// beginRegistration starts the WebAuthn registration process
func (s *Server) beginRegistration(c echo.Context) error {
	var req RegisterBeginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.Username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "username is required"})
	}

	// Check if user already exists
	ctx := c.Request().Context()
	_, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err == nil {
		return c.JSON(http.StatusConflict, map[string]string{"error": "username already exists"})
	}
	if err != repository.ErrNotFound {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "database error"})
	}

	// Create a temporary user for registration
	tempUser := &models.User{
		ID:       uuid.New(),
		Username: req.Username,
	}

	// Generate WebAuthn registration options
	options, session, err := s.webauthn.WebAuthn().BeginRegistration(tempUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to begin registration"})
	}

	// Generate recovery phrase
	recoveryPhrase, err := crypto.GenerateRecoveryPhrase()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate recovery phrase"})
	}

	// Store session with user ID
	sessionID := uuid.New().String()
	sessionStore.Save(sessionID, session, req.Username, tempUser.ID)

	// Store recovery phrase hash temporarily (we'll need it when finishing registration)
	c.Set("recovery_phrase_"+sessionID, recoveryPhrase)

	return c.JSON(http.StatusOK, RegisterBeginResponse{
		Options:        options,
		RecoveryPhrase: recoveryPhrase,
		SessionID:      sessionID,
	})
}

// finishRegistration completes the WebAuthn registration process
func (s *Server) finishRegistration(c echo.Context) error {
	var req RegisterFinishRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// Get session
	sessionData, ok := sessionStore.Get(req.SessionID)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "session expired or invalid"})
	}
	defer sessionStore.Delete(req.SessionID)

	// Use the same user ID from the session (must match what was used in beginRegistration)
	tempUser := &models.User{
		ID:       sessionData.UserID,
		Username: sessionData.Username,
	}

	// Parse the credential response (sent as JSON string from frontend)
	credentialData := []byte(req.Credential)

	parsedResponse, err := protocol.ParseCredentialCreationResponseBody(
		&credentialReader{data: credentialData},
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to parse credential"})
	}

	// Verify the registration
	credential, err := s.webauthn.WebAuthn().CreateCredential(tempUser, *sessionData.Session, parsedResponse)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to verify credential"})
	}

	// Generate DEK and salt
	dek, err := crypto.GenerateDEK()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate encryption key"})
	}

	salt, err := crypto.GenerateSalt()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate salt"})
	}

	// Derive KEK from PRF output or generate a temporary one
	var kek []byte
	if req.PRFOutput != "" {
		prfBytes, err := base64.StdEncoding.DecodeString(req.PRFOutput)
		if err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid PRF output"})
		}
		kek = crypto.DeriveKeyFromSecret(prfBytes, salt)
	} else {
		// Fallback: derive from a temporary secret (in production, require PRF or password)
		tempSecret, _ := crypto.GenerateRandomHex(32)
		kek = crypto.DeriveKeyFromSecret([]byte(tempSecret), salt)
	}

	// Encrypt DEK with KEK
	encryptedDEK, err := crypto.EncryptDEK(dek, kek)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to encrypt DEK"})
	}

	// Hash recovery phrase (we'd need to pass this from begin, for now generate new)
	recoveryPhrase, _ := crypto.GenerateRecoveryPhrase()
	recoveryHash := crypto.HashRecoveryPhrase(recoveryPhrase)

	// Create user
	ctx := c.Request().Context()
	user := &models.User{
		ID:                 tempUser.ID,
		Username:           sessionData.Username,
		EncryptedDEK:       encryptedDEK,
		DEKSalt:            salt,
		RecoveryPhraseHash: recoveryHash,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create user"})
	}

	// Save credential
	cred := &models.Credential{
		ID:              credential.ID,
		UserID:          user.ID,
		PublicKey:       credential.PublicKey,
		AttestationType: credential.AttestationType,
		SignCount:       credential.Authenticator.SignCount,
		CreatedAt:       time.Now(),
	}

	if err := s.userRepo.CreateCredential(ctx, cred); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to save credential"})
	}

	// Generate JWT
	token, err := s.generateJWT(user, encryptedDEK)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
	}

	return c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User: struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		}{
			ID:       user.ID.String(),
			Username: user.Username,
		},
	})
}

// beginLogin starts the WebAuthn login process
func (s *Server) beginLogin(c echo.Context) error {
	var req LoginBeginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.Username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "username is required"})
	}

	// Get user
	ctx := c.Request().Context()
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err == repository.ErrNotFound {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "database error"})
	}

	// Generate login options
	options, session, err := s.webauthn.WebAuthn().BeginLogin(user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to begin login"})
	}

	// Store session with user ID
	sessionID := uuid.New().String()
	sessionStore.Save(sessionID, session, user.Username, user.ID)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"options":    options,
		"session_id": sessionID,
	})
}

// finishLogin completes the WebAuthn login process
func (s *Server) finishLogin(c echo.Context) error {
	var req LoginFinishRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// Get session
	sessionData, ok := sessionStore.Get(req.SessionID)
	if !ok {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "session expired or invalid"})
	}
	defer sessionStore.Delete(req.SessionID)

	// Get user
	ctx := c.Request().Context()
	user, err := s.userRepo.GetByUsername(ctx, sessionData.Username)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get user"})
	}

	// Parse the credential response (sent as JSON string from frontend)
	credentialData := []byte(req.Credential)

	parsedResponse, err := protocol.ParseCredentialRequestResponseBody(
		&credentialReader{data: credentialData},
	)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "failed to parse credential"})
	}

	// Verify the login
	credential, err := s.webauthn.WebAuthn().ValidateLogin(user, *sessionData.Session, parsedResponse)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "authentication failed"})
	}

	// Update sign count
	if err := s.userRepo.UpdateCredentialSignCount(ctx, credential.ID, credential.Authenticator.SignCount); err != nil {
		// Log but don't fail the request
	}

	// Generate JWT
	token, err := s.generateJWT(user, user.EncryptedDEK)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to generate token"})
	}

	return c.JSON(http.StatusOK, AuthResponse{
		Token: token,
		User: struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		}{
			ID:       user.ID.String(),
			Username: user.Username,
		},
	})
}

// generateJWT creates a JWT token for the user
func (s *Server) generateJWT(user *models.User, encryptedDEK []byte) (string, error) {
	claims := JWTClaims{
		UserID:       user.ID.String(),
		EncryptedDEK: base64.StdEncoding.EncodeToString(encryptedDEK),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "ownu",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.config.JWTSecret))
}

// credentialReader implements io.Reader for parsing credentials
type credentialReader struct {
	data []byte
	pos  int
}

func (r *credentialReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF // Must return EOF when done, otherwise parser hangs
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
