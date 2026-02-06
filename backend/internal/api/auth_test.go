package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/ownu/ownu/internal/auth"
	"github.com/ownu/ownu/internal/config"
	"github.com/ownu/ownu/internal/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(t *testing.T) (*Server, func()) {
	cfg := &config.Config{
		DatabaseURL:      "postgres://ownu:ownu_dev_password@localhost:5432/ownu_test?sslmode=disable",
		JWTSecret:        "test-secret-key-for-testing-only",
		WebAuthnRPID:     "localhost",
		WebAuthnRPOrigin: "https://localhost",
		WebAuthnRPName:   "OwnU Test",
	}

	// Try to connect to test database
	db, err := repository.NewDB(cfg.DatabaseURL)
	if err != nil {
		t.Skipf("Skipping integration test: database not available: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	plaidRepo := repository.NewPlaidRepository(db)
	webauthnService, err := auth.NewWebAuthnService(cfg.WebAuthnRPID, cfg.WebAuthnRPOrigin, cfg.WebAuthnRPName)
	if err != nil {
		t.Fatalf("Failed to create WebAuthn service: %v", err)
	}

	server := NewServer(cfg, userRepo, accountRepo, plaidRepo, webauthnService, nil)

	cleanup := func() {
		db.Close()
	}

	return server, cleanup
}

func TestHealthCheck(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	cfg := &config.Config{
		JWTSecret:        "test-secret",
		WebAuthnRPID:     "localhost",
		WebAuthnRPOrigin: "https://localhost",
		WebAuthnRPName:   "Test",
	}

	// Create minimal server for health check (doesn't need DB)
	webauthnService, _ := auth.NewWebAuthnService(cfg.WebAuthnRPID, cfg.WebAuthnRPOrigin, cfg.WebAuthnRPName)
	server := &Server{
		echo:     e,
		config:   cfg,
		webauthn: webauthnService,
	}

	err := server.healthCheck(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

func TestRegisterBegin_EmptyUsername(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	e := echo.New()
	body := `{"username":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register/begin", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := server.beginRegistration(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "username is required", response["error"])
}

func TestRegisterBegin_ValidUsername(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	e := echo.New()
	body := `{"username":"testuser_register"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register/begin", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := server.beginRegistration(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response RegisterBeginResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotEmpty(t, response.SessionID)
	assert.NotEmpty(t, response.RecoveryPhrase)
	assert.NotNil(t, response.Options)
}

func TestLoginBegin_UserNotFound(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	e := echo.New()
	body := `{"username":"nonexistent_user_12345"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/begin", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := server.beginLogin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "user not found", response["error"])
}

func TestLoginBegin_EmptyUsername(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	e := echo.New()
	body := `{"username":""}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/begin", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := server.beginLogin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestRegisterFinish_InvalidSession(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	e := echo.New()
	body := `{"session_id":"invalid-session-id","credential":"{}"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register/finish", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := server.finishRegistration(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "session expired or invalid", response["error"])
}

func TestLoginFinish_InvalidSession(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	e := echo.New()
	body := `{"session_id":"invalid-session-id","credential":"{}"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login/finish", bytes.NewBufferString(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := server.finishLogin(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]string
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "session expired or invalid", response["error"])
}
