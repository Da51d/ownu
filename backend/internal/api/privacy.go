package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/ownu/ownu/internal/crypto"
)

// DataExportResponse contains all user data for export
type DataExportResponse struct {
	ExportedAt   time.Time                `json:"exported_at"`
	User         UserExport               `json:"user"`
	Accounts     []AccountExport          `json:"accounts"`
	Transactions []TransactionExport      `json:"transactions"`
	PlaidItems   []PlaidItemExport        `json:"plaid_items,omitempty"`
}

// UserExport is the user data for export
type UserExport struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// AccountExport is account data for export
type AccountExport struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Type        string    `json:"type"`
	Institution string    `json:"institution"`
	CreatedAt   time.Time `json:"created_at"`
}

// TransactionExport is transaction data for export
type TransactionExport struct {
	ID          string    `json:"id"`
	AccountID   string    `json:"account_id"`
	Date        string    `json:"date"`
	Amount      float64   `json:"amount"`
	Description string    `json:"description"`
	Merchant    string    `json:"merchant,omitempty"`
	Category    string    `json:"category,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// PlaidItemExport is Plaid connection data for export
type PlaidItemExport struct {
	ID              string    `json:"id"`
	InstitutionName string    `json:"institution_name"`
	Status          string    `json:"status"`
	AccountCount    int       `json:"account_count"`
	CreatedAt       time.Time `json:"created_at"`
}

// exportData exports all user data (GDPR Article 20 - Right to data portability)
func (s *Server) exportData(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	encryptedDEKB64, ok := getEncryptedDEK(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing encryption key"})
	}

	ctx := c.Request().Context()

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get user"})
	}

	// Decrypt the DEK
	dek, err := s.decryptDEK(encryptedDEKB64, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to decrypt key"})
	}

	export := DataExportResponse{
		ExportedAt: time.Now().UTC(),
		User: UserExport{
			ID:        user.ID.String(),
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
		},
		Accounts:     make([]AccountExport, 0),
		Transactions: make([]TransactionExport, 0),
		PlaidItems:   make([]PlaidItemExport, 0),
	}

	// Export accounts
	accounts, err := s.accountRepo.GetByUserID(ctx, userID)
	if err == nil {
		for _, account := range accounts {
			data, err := s.decryptAccountData(account.EncryptedData, dek)
			if err != nil {
				continue
			}
			export.Accounts = append(export.Accounts, AccountExport{
				ID:          account.ID.String(),
				Name:        data.Name,
				Type:        data.Type,
				Institution: data.Institution,
				CreatedAt:   account.CreatedAt,
			})
		}
	}

	// Export Plaid items (if configured)
	if s.plaidRepo != nil {
		items, err := s.plaidRepo.GetItemsByUserID(ctx, userID)
		if err == nil {
			for _, item := range items {
				accounts, _ := s.plaidRepo.GetAccountsByItemID(ctx, item.ID)
				export.PlaidItems = append(export.PlaidItems, PlaidItemExport{
					ID:              item.ID.String(),
					InstitutionName: item.InstitutionName,
					Status:          item.Status,
					AccountCount:    len(accounts),
					CreatedAt:       item.CreatedAt,
				})
			}
		}
	}

	// TODO: Export transactions when transaction repository is implemented

	// Set headers for file download
	c.Response().Header().Set("Content-Disposition", "attachment; filename=ownu-export.json")
	c.Response().Header().Set("Content-Type", "application/json")

	return c.JSON(http.StatusOK, export)
}

// DeleteAccountRequest is the request to delete user account
type DeleteAccountRequest struct {
	Confirmation string `json:"confirmation"` // Must be "DELETE MY ACCOUNT"
}

// deleteUserAccount permanently deletes all user data (GDPR Article 17 - Right to erasure)
func (s *Server) deleteUserAccount(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	var req DeleteAccountRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	// Require explicit confirmation
	if req.Confirmation != "DELETE MY ACCOUNT" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "confirmation required: send {\"confirmation\": \"DELETE MY ACCOUNT\"}",
		})
	}

	ctx := c.Request().Context()

	// Get user first to get DEK for Plaid cleanup
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get user"})
	}

	// If Plaid is configured, disconnect all items first
	if s.plaid != nil && s.plaidRepo != nil {
		encryptedDEKB64, ok := getEncryptedDEK(c)
		if ok {
			dek, err := s.decryptDEK(encryptedDEKB64, user)
			if err == nil {
				items, _ := s.plaidRepo.GetItemsByUserID(ctx, userID)
				for _, item := range items {
					accessToken, err := crypto.Decrypt(item.EncryptedAccessToken, dek)
					if err == nil {
						// Best effort - don't fail if Plaid removal fails
						_ = s.plaid.RemoveItem(ctx, string(accessToken))
					}
				}
			}
		}
	}

	// Delete user (cascades to all related data due to ON DELETE CASCADE)
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete account"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status":  "deleted",
		"message": "Your account and all associated data have been permanently deleted",
	})
}

// PrivacySettingsResponse contains user privacy settings
type PrivacySettingsResponse struct {
	DataRetentionDays int  `json:"data_retention_days"`
	AuditLogsEnabled  bool `json:"audit_logs_enabled"`
}

// getPrivacySettings returns current privacy settings
func (s *Server) getPrivacySettings(c echo.Context) error {
	_, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	// For now, return defaults (could be made configurable per-user)
	return c.JSON(http.StatusOK, PrivacySettingsResponse{
		DataRetentionDays: 0, // 0 = indefinite
		AuditLogsEnabled:  true,
	})
}

// ConsentStatus tracks user consent for various data processing activities
type ConsentStatus struct {
	PlaidDataSharing bool      `json:"plaid_data_sharing"`
	ConsentedAt      time.Time `json:"consented_at,omitempty"`
}

// getConsentStatus returns current consent status
func (s *Server) getConsentStatus(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	ctx := c.Request().Context()

	// Check if user has any Plaid connections (implies consent)
	hasPlaidConsent := false
	if s.plaidRepo != nil {
		items, err := s.plaidRepo.GetItemsByUserID(ctx, userID)
		if err == nil && len(items) > 0 {
			hasPlaidConsent = true
		}
	}

	return c.JSON(http.StatusOK, ConsentStatus{
		PlaidDataSharing: hasPlaidConsent,
	})
}

// ExportFormat specifies the export format
type ExportFormat string

const (
	ExportFormatJSON ExportFormat = "json"
	ExportFormatCSV  ExportFormat = "csv"
)

// exportTransactionsCSV exports transactions in CSV format
func (s *Server) exportTransactionsCSV(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	encryptedDEKB64, ok := getEncryptedDEK(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing encryption key"})
	}

	ctx := c.Request().Context()

	// Get user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to get user"})
	}

	// Decrypt the DEK
	_, err = s.decryptDEK(encryptedDEKB64, user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to decrypt key"})
	}

	// TODO: Implement CSV export when transaction repository is ready
	// For now, return a placeholder

	csv := "date,amount,description,merchant,category,account\n"
	csv += "# No transactions yet\n"

	c.Response().Header().Set("Content-Disposition", "attachment; filename=transactions.csv")
	c.Response().Header().Set("Content-Type", "text/csv")

	return c.String(http.StatusOK, csv)
}

// Helper to convert any struct to map for safe JSON (no sensitive fields)
func safeJSON(v interface{}) map[string]interface{} {
	data, _ := json.Marshal(v)
	var result map[string]interface{}
	json.Unmarshal(data, &result)

	// Remove any potentially sensitive fields
	sensitiveFields := []string{"encrypted_dek", "dek_salt", "recovery_phrase_hash", "encrypted_data", "encrypted_access_token"}
	for _, field := range sensitiveFields {
		delete(result, field)
	}

	return result
}
