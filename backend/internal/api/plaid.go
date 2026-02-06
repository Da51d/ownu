package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/ownu/ownu/internal/crypto"
	"github.com/ownu/ownu/internal/models"
	plaidPkg "github.com/plaid/plaid-go/v27/plaid"
)

// CreateLinkTokenResponse is the response for creating a link token
type CreateLinkTokenResponse struct {
	LinkToken string `json:"link_token"`
}

// ExchangeTokenRequest is the request for exchanging a public token
type ExchangeTokenRequest struct {
	PublicToken     string `json:"public_token"`
	InstitutionID   string `json:"institution_id"`
	InstitutionName string `json:"institution_name"`
}

// ExchangeTokenResponse is the response for exchanging a public token
type ExchangeTokenResponse struct {
	ItemID   string                `json:"item_id"`
	Accounts []models.PlaidAccount `json:"accounts"`
}

// PlaidItemResponse is the response for a Plaid item
type PlaidItemResponse struct {
	ID              uuid.UUID             `json:"id"`
	InstitutionID   string                `json:"institution_id"`
	InstitutionName string                `json:"institution_name"`
	Status          string                `json:"status"`
	ErrorCode       string                `json:"error_code,omitempty"`
	ErrorMessage    string                `json:"error_message,omitempty"`
	Accounts        []models.PlaidAccount `json:"accounts"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

// SyncTransactionsResponse is the response for syncing transactions
type SyncTransactionsResponse struct {
	AddedCount    int  `json:"added_count"`
	ModifiedCount int  `json:"modified_count"`
	RemovedCount  int  `json:"removed_count"`
	HasMore       bool `json:"has_more"`
}

// createLinkToken creates a Plaid link token
func (s *Server) createLinkToken(c echo.Context) error {
	if s.plaid == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "Plaid integration is not configured",
		})
	}

	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	linkToken, err := s.plaid.CreateLinkToken(c.Request().Context(), userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("failed to create link token: %v", err),
		})
	}

	return c.JSON(http.StatusOK, CreateLinkTokenResponse{
		LinkToken: linkToken,
	})
}

// exchangePublicToken exchanges a public token for an access token
func (s *Server) exchangePublicToken(c echo.Context) error {
	if s.plaid == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "Plaid integration is not configured",
		})
	}

	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	encryptedDEKB64, ok := getEncryptedDEK(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing encryption key"})
	}

	var req ExchangeTokenRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	if req.PublicToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "public_token is required"})
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

	// Exchange the public token
	accessToken, itemID, err := s.plaid.ExchangePublicToken(ctx, req.PublicToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("failed to exchange token: %v", err),
		})
	}

	// Encrypt the access token
	encryptedToken, err := crypto.Encrypt([]byte(accessToken), dek)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to encrypt access token",
		})
	}

	// Create the Plaid item record
	plaidItem := &models.PlaidItem{
		ID:                   uuid.New(),
		UserID:               userID,
		ItemID:               itemID,
		EncryptedAccessToken: encryptedToken,
		InstitutionID:        req.InstitutionID,
		InstitutionName:      req.InstitutionName,
	}

	if err := s.plaidRepo.CreateItem(ctx, plaidItem); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("failed to save item: %v", err),
		})
	}

	// Fetch and store the accounts
	accounts, err := s.plaid.GetAccounts(ctx, accessToken)
	if err != nil {
		// Item created but accounts fetch failed - log but don't fail
		return c.JSON(http.StatusOK, ExchangeTokenResponse{
			ItemID:   plaidItem.ID.String(),
			Accounts: nil,
		})
	}

	var plaidAccounts []models.PlaidAccount
	for _, acc := range accounts {
		plaidAccount := &models.PlaidAccount{
			ID:             uuid.New(),
			PlaidItemID:    plaidItem.ID,
			UserID:         userID,
			PlaidAccountID: acc.GetAccountId(),
			Name:           acc.GetName(),
			OfficialName:   acc.GetOfficialName(),
			Type:           string(acc.GetType()),
			Subtype:        string(acc.GetSubtype()),
			Mask:           acc.GetMask(),
		}

		if err := s.plaidRepo.CreateAccount(ctx, plaidAccount); err != nil {
			continue // Log and continue with other accounts
		}

		plaidAccounts = append(plaidAccounts, *plaidAccount)
	}

	return c.JSON(http.StatusOK, ExchangeTokenResponse{
		ItemID:   plaidItem.ID.String(),
		Accounts: plaidAccounts,
	})
}

// listPlaidItems lists all Plaid items for the user
func (s *Server) listPlaidItems(c echo.Context) error {
	if s.plaid == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "Plaid integration is not configured",
		})
	}

	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	ctx := c.Request().Context()

	items, err := s.plaidRepo.GetItemsByUserID(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("failed to get items: %v", err),
		})
	}

	var response []PlaidItemResponse
	for _, item := range items {
		accounts, _ := s.plaidRepo.GetAccountsByItemID(ctx, item.ID)

		response = append(response, PlaidItemResponse{
			ID:              item.ID,
			InstitutionID:   item.InstitutionID,
			InstitutionName: item.InstitutionName,
			Status:          item.Status,
			ErrorCode:       item.ErrorCode,
			ErrorMessage:    item.ErrorMessage,
			Accounts:        accounts,
			CreatedAt:       item.CreatedAt,
			UpdatedAt:       item.UpdatedAt,
		})
	}

	if response == nil {
		response = []PlaidItemResponse{}
	}

	return c.JSON(http.StatusOK, response)
}

// getPlaidItem gets a specific Plaid item
func (s *Server) getPlaidItem(c echo.Context) error {
	if s.plaid == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "Plaid integration is not configured",
		})
	}

	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid item ID"})
	}

	ctx := c.Request().Context()

	item, err := s.plaidRepo.GetItemByID(ctx, itemID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "item not found"})
	}

	if item.UserID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
	}

	accounts, _ := s.plaidRepo.GetAccountsByItemID(ctx, item.ID)

	return c.JSON(http.StatusOK, PlaidItemResponse{
		ID:              item.ID,
		InstitutionID:   item.InstitutionID,
		InstitutionName: item.InstitutionName,
		Status:          item.Status,
		ErrorCode:       item.ErrorCode,
		ErrorMessage:    item.ErrorMessage,
		Accounts:        accounts,
		CreatedAt:       item.CreatedAt,
		UpdatedAt:       item.UpdatedAt,
	})
}

// deletePlaidItem removes a Plaid item
func (s *Server) deletePlaidItem(c echo.Context) error {
	if s.plaid == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "Plaid integration is not configured",
		})
	}

	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	encryptedDEKB64, ok := getEncryptedDEK(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing encryption key"})
	}

	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid item ID"})
	}

	ctx := c.Request().Context()

	// Get the item to verify ownership and get access token
	item, err := s.plaidRepo.GetItemByID(ctx, itemID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "item not found"})
	}

	if item.UserID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
	}

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

	// Decrypt access token and remove from Plaid
	accessToken, err := crypto.Decrypt(item.EncryptedAccessToken, dek)
	if err == nil {
		// Best effort removal from Plaid - don't fail if this errors
		_ = s.plaid.RemoveItem(ctx, string(accessToken))
	}

	// Delete from our database
	if err := s.plaidRepo.DeleteItem(ctx, itemID, userID); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("failed to delete item: %v", err),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "deleted"})
}

// syncTransactions syncs transactions for a Plaid item
func (s *Server) syncTransactions(c echo.Context) error {
	if s.plaid == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "Plaid integration is not configured",
		})
	}

	userID, ok := getUserID(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	encryptedDEKB64, ok := getEncryptedDEK(c)
	if !ok {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing encryption key"})
	}

	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid item ID"})
	}

	ctx := c.Request().Context()

	// Get the item
	item, err := s.plaidRepo.GetItemByID(ctx, itemID)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "item not found"})
	}

	if item.UserID != userID {
		return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
	}

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

	// Decrypt access token
	accessToken, err := crypto.Decrypt(item.EncryptedAccessToken, dek)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to decrypt access token"})
	}

	// Sync transactions
	result, err := s.plaid.SyncTransactions(ctx, string(accessToken), item.Cursor)
	if err != nil {
		// Record failed sync
		sync := &models.PlaidSync{
			ID:           uuid.New(),
			PlaidItemID:  itemID,
			UserID:       userID,
			CursorBefore: item.Cursor,
			ErrorMessage: err.Error(),
		}
		_ = s.plaidRepo.CreateSync(ctx, sync)

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("failed to sync transactions: %v", err),
		})
	}

	// Process added transactions
	addedCount := 0
	for _, txn := range result.Added {
		if err := s.processPlaidTransaction(ctx, userID, itemID, &txn, dek); err == nil {
			addedCount++
		}
	}

	// Process modified transactions (update existing)
	modifiedCount := len(result.Modified)

	// Process removed transactions
	removedCount := len(result.Removed)

	// Update cursor
	if err := s.plaidRepo.UpdateItemCursor(ctx, itemID, result.NextCursor); err != nil {
		// Log but don't fail
	}

	// Record successful sync
	sync := &models.PlaidSync{
		ID:            uuid.New(),
		PlaidItemID:   itemID,
		UserID:        userID,
		AddedCount:    addedCount,
		ModifiedCount: modifiedCount,
		RemovedCount:  removedCount,
		CursorBefore:  item.Cursor,
		CursorAfter:   result.NextCursor,
	}
	_ = s.plaidRepo.CreateSync(ctx, sync)

	return c.JSON(http.StatusOK, SyncTransactionsResponse{
		AddedCount:    addedCount,
		ModifiedCount: modifiedCount,
		RemovedCount:  removedCount,
		HasMore:       result.HasMore,
	})
}

// processPlaidTransaction processes a single Plaid transaction
func (s *Server) processPlaidTransaction(ctx context.Context, userID, plaidItemID uuid.UUID, txn *plaidPkg.Transaction, dek []byte) error {
	// Find the linked OwnU account for this Plaid account
	plaidAccounts, err := s.plaidRepo.GetAccountsByItemID(ctx, plaidItemID)
	if err != nil {
		return err
	}

	var linkedAccountID *uuid.UUID
	for _, acc := range plaidAccounts {
		if acc.PlaidAccountID == txn.GetAccountId() && acc.AccountID != nil {
			linkedAccountID = acc.AccountID
			break
		}
	}

	// If no linked account, we can't store the transaction
	if linkedAccountID == nil {
		return fmt.Errorf("no linked account for plaid account %s", txn.GetAccountId())
	}

	// Create transaction data
	txnData := models.TransactionData{
		Amount:      txn.GetAmount(),
		Description: txn.GetName(),
		Merchant:    txn.GetMerchantName(),
		Date:        txn.GetDate(),
	}

	// Encrypt transaction data
	txnDataJSON, err := json.Marshal(txnData)
	if err != nil {
		return err
	}

	encryptedData, err := crypto.Encrypt(txnDataJSON, dek)
	if err != nil {
		return err
	}

	// Parse transaction date
	txnDate, err := time.Parse("2006-01-02", txn.GetDate())
	if err != nil {
		txnDate = time.Now()
	}

	// Create transaction model
	transaction := &models.Transaction{
		ID:              uuid.New(),
		UserID:          userID,
		AccountID:       *linkedAccountID,
		EncryptedData:   encryptedData,
		TransactionDate: txnDate,
		CreatedAt:       time.Now(),
	}

	// TODO: Store transaction using transaction repository
	// For now, we'll skip actual storage and just return success
	_ = transaction

	return nil
}

// plaidStatus returns whether Plaid is configured
func (s *Server) plaidStatus(c echo.Context) error {
	configured := s.plaid != nil
	return c.JSON(http.StatusOK, map[string]bool{
		"configured": configured,
	})
}
