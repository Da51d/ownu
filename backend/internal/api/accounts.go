package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ownu/ownu/internal/crypto"
	"github.com/ownu/ownu/internal/models"
	"github.com/ownu/ownu/internal/repository"
)

// CreateAccountRequest is the request body for creating an account
type CreateAccountRequest struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Institution string `json:"institution"`
}

// AccountResponse is the response for account operations
type AccountResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Institution string `json:"institution"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// listAccounts returns all accounts for the authenticated user
func (s *Server) listAccounts(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	encryptedDEKB64, ok := getEncryptedDEK(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing encryption key"})
	}

	// Get user to retrieve DEK decryption info
	ctx := c.Request().Context()
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get user"})
	}

	// Decrypt the DEK
	dek, err := s.decryptDEK(encryptedDEKB64, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to decrypt key"})
	}

	// Get accounts
	accounts, err := s.accountRepo.GetByUserID(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get accounts"})
	}

	// Decrypt account data
	responses := make([]AccountResponse, 0, len(accounts))
	for _, account := range accounts {
		data, err := s.decryptAccountData(account.EncryptedData, dek)
		if err != nil {
			continue // Skip accounts that can't be decrypted
		}
		responses = append(responses, AccountResponse{
			ID:          account.ID.String(),
			Name:        data.Name,
			Type:        data.Type,
			Institution: data.Institution,
			CreatedAt:   account.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   account.UpdatedAt.Format(time.RFC3339),
		})
	}

	return c.JSON(http.StatusOK, responses)
}

// createAccount creates a new account for the authenticated user
func (s *Server) createAccount(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	encryptedDEKB64, ok := getEncryptedDEK(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing encryption key"})
	}

	var req CreateAccountRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
	}

	// Get user to retrieve DEK decryption info
	ctx := c.Request().Context()
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get user"})
	}

	// Decrypt the DEK
	dek, err := s.decryptDEK(encryptedDEKB64, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to decrypt key"})
	}

	// Create account data
	accountData := models.AccountData{
		Name:        req.Name,
		Type:        req.Type,
		Institution: req.Institution,
	}

	// Encrypt account data
	encryptedData, err := s.encryptAccountData(accountData, dek)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to encrypt data"})
	}

	// Create account
	now := time.Now()
	account := &models.Account{
		ID:            uuid.New(),
		UserID:        userID,
		EncryptedData: encryptedData,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.accountRepo.Create(ctx, account); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create account"})
	}

	return c.JSON(http.StatusCreated, AccountResponse{
		ID:          account.ID.String(),
		Name:        accountData.Name,
		Type:        accountData.Type,
		Institution: accountData.Institution,
		CreatedAt:   account.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   account.UpdatedAt.Format(time.RFC3339),
	})
}

// getAccount returns a specific account
func (s *Server) getAccount(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	encryptedDEKB64, ok := getEncryptedDEK(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing encryption key"})
	}

	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid account ID"})
	}

	ctx := c.Request().Context()

	// Get user to retrieve DEK decryption info
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get user"})
	}

	// Decrypt the DEK
	dek, err := s.decryptDEK(encryptedDEKB64, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to decrypt key"})
	}

	// Get account
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err == repository.ErrNotFound {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "account not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get account"})
	}

	// Verify ownership
	if account.UserID != userID {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "account not found"})
	}

	// Decrypt account data
	data, err := s.decryptAccountData(account.EncryptedData, dek)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to decrypt data"})
	}

	return c.JSON(http.StatusOK, AccountResponse{
		ID:          account.ID.String(),
		Name:        data.Name,
		Type:        data.Type,
		Institution: data.Institution,
		CreatedAt:   account.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   account.UpdatedAt.Format(time.RFC3339),
	})
}

// updateAccount updates an existing account
func (s *Server) updateAccount(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	encryptedDEKB64, ok := getEncryptedDEK(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing encryption key"})
	}

	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid account ID"})
	}

	var req CreateAccountRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "name is required"})
	}

	ctx := c.Request().Context()

	// Get user to retrieve DEK decryption info
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get user"})
	}

	// Decrypt the DEK
	dek, err := s.decryptDEK(encryptedDEKB64, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to decrypt key"})
	}

	// Get existing account to verify ownership
	account, err := s.accountRepo.GetByID(ctx, accountID)
	if err == repository.ErrNotFound {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "account not found"})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get account"})
	}

	if account.UserID != userID {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "account not found"})
	}

	// Create new account data
	accountData := models.AccountData{
		Name:        req.Name,
		Type:        req.Type,
		Institution: req.Institution,
	}

	// Encrypt account data
	encryptedData, err := s.encryptAccountData(accountData, dek)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to encrypt data"})
	}

	// Update account
	account.EncryptedData = encryptedData
	account.UpdatedAt = time.Now()

	if err := s.accountRepo.Update(ctx, account); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update account"})
	}

	return c.JSON(http.StatusOK, AccountResponse{
		ID:          account.ID.String(),
		Name:        accountData.Name,
		Type:        accountData.Type,
		Institution: accountData.Institution,
		CreatedAt:   account.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   account.UpdatedAt.Format(time.RFC3339),
	})
}

// deleteAccount removes an account
func (s *Server) deleteAccount(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	accountID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid account ID"})
	}

	ctx := c.Request().Context()

	if err := s.accountRepo.Delete(ctx, accountID, userID); err == repository.ErrNotFound {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "account not found"})
	} else if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete account"})
	}

	return c.NoContent(http.StatusNoContent)
}

// Helper functions for encryption/decryption

func (s *Server) decryptDEK(encryptedDEKB64 string, user *models.User) ([]byte, error) {
	encryptedDEK, err := base64.StdEncoding.DecodeString(encryptedDEKB64)
	if err != nil {
		return nil, err
	}

	// For now, we derive the KEK from a placeholder. In production, this would
	// use the PRF output or password-derived key stored in the session.
	// The KEK should be derived the same way it was during registration.
	kek := crypto.DeriveKeyFromSecret([]byte("placeholder-secret"), user.DEKSalt)

	return crypto.DecryptDEK(encryptedDEK, kek)
}

func (s *Server) encryptAccountData(data models.AccountData, dek []byte) ([]byte, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return crypto.Encrypt(jsonData, dek)
}

func (s *Server) decryptAccountData(encryptedData []byte, dek []byte) (*models.AccountData, error) {
	jsonData, err := crypto.Decrypt(encryptedData, dek)
	if err != nil {
		return nil, err
	}
	var data models.AccountData
	if err := json.Unmarshal(jsonData, &data); err != nil {
		return nil, err
	}
	return &data, nil
}
