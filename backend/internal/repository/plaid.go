package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ownu/ownu/internal/models"
)

// PlaidRepository handles Plaid-related database operations
type PlaidRepository struct {
	db *DB
}

// NewPlaidRepository creates a new Plaid repository
func NewPlaidRepository(db *DB) *PlaidRepository {
	return &PlaidRepository{db: db}
}

// CreateItem creates a new Plaid item
func (r *PlaidRepository) CreateItem(ctx context.Context, item *models.PlaidItem) error {
	query := `
		INSERT INTO plaid_items (
			id, user_id, item_id, encrypted_access_token, institution_id,
			institution_name, cursor, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	now := time.Now()
	item.CreatedAt = now
	item.UpdatedAt = now
	item.Status = "active"

	_, err := r.db.Pool.Exec(ctx, query,
		item.ID,
		item.UserID,
		item.ItemID,
		item.EncryptedAccessToken,
		item.InstitutionID,
		item.InstitutionName,
		item.Cursor,
		item.Status,
		item.CreatedAt,
		item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create plaid item: %w", err)
	}

	return nil
}

// GetItemByID retrieves a Plaid item by ID
func (r *PlaidRepository) GetItemByID(ctx context.Context, id uuid.UUID) (*models.PlaidItem, error) {
	query := `
		SELECT id, user_id, item_id, encrypted_access_token, institution_id,
			institution_name, cursor, status, error_code, error_message,
			consent_expires_at, created_at, updated_at
		FROM plaid_items
		WHERE id = $1
	`

	item := &models.PlaidItem{}
	var errorCode, errorMessage sql.NullString
	var consentExpiresAt sql.NullTime

	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&item.ID,
		&item.UserID,
		&item.ItemID,
		&item.EncryptedAccessToken,
		&item.InstitutionID,
		&item.InstitutionName,
		&item.Cursor,
		&item.Status,
		&errorCode,
		&errorMessage,
		&consentExpiresAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get plaid item: %w", err)
	}

	if errorCode.Valid {
		item.ErrorCode = errorCode.String
	}
	if errorMessage.Valid {
		item.ErrorMessage = errorMessage.String
	}
	if consentExpiresAt.Valid {
		item.ConsentExpiresAt = &consentExpiresAt.Time
	}

	return item, nil
}

// GetItemByItemID retrieves a Plaid item by Plaid's item_id
func (r *PlaidRepository) GetItemByItemID(ctx context.Context, itemID string) (*models.PlaidItem, error) {
	query := `
		SELECT id, user_id, item_id, encrypted_access_token, institution_id,
			institution_name, cursor, status, error_code, error_message,
			consent_expires_at, created_at, updated_at
		FROM plaid_items
		WHERE item_id = $1
	`

	item := &models.PlaidItem{}
	var errorCode, errorMessage sql.NullString
	var consentExpiresAt sql.NullTime

	err := r.db.Pool.QueryRow(ctx, query, itemID).Scan(
		&item.ID,
		&item.UserID,
		&item.ItemID,
		&item.EncryptedAccessToken,
		&item.InstitutionID,
		&item.InstitutionName,
		&item.Cursor,
		&item.Status,
		&errorCode,
		&errorMessage,
		&consentExpiresAt,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get plaid item: %w", err)
	}

	if errorCode.Valid {
		item.ErrorCode = errorCode.String
	}
	if errorMessage.Valid {
		item.ErrorMessage = errorMessage.String
	}
	if consentExpiresAt.Valid {
		item.ConsentExpiresAt = &consentExpiresAt.Time
	}

	return item, nil
}

// GetItemsByUserID retrieves all Plaid items for a user
func (r *PlaidRepository) GetItemsByUserID(ctx context.Context, userID uuid.UUID) ([]models.PlaidItem, error) {
	query := `
		SELECT id, user_id, item_id, encrypted_access_token, institution_id,
			institution_name, cursor, status, error_code, error_message,
			consent_expires_at, created_at, updated_at
		FROM plaid_items
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plaid items: %w", err)
	}
	defer rows.Close()

	var items []models.PlaidItem
	for rows.Next() {
		var item models.PlaidItem
		var errorCode, errorMessage sql.NullString
		var consentExpiresAt sql.NullTime

		err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.ItemID,
			&item.EncryptedAccessToken,
			&item.InstitutionID,
			&item.InstitutionName,
			&item.Cursor,
			&item.Status,
			&errorCode,
			&errorMessage,
			&consentExpiresAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plaid item: %w", err)
		}

		if errorCode.Valid {
			item.ErrorCode = errorCode.String
		}
		if errorMessage.Valid {
			item.ErrorMessage = errorMessage.String
		}
		if consentExpiresAt.Valid {
			item.ConsentExpiresAt = &consentExpiresAt.Time
		}

		items = append(items, item)
	}

	return items, nil
}

// UpdateItemCursor updates the sync cursor for a Plaid item
func (r *PlaidRepository) UpdateItemCursor(ctx context.Context, id uuid.UUID, cursor string) error {
	query := `
		UPDATE plaid_items
		SET cursor = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Pool.Exec(ctx, query, cursor, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update plaid item cursor: %w", err)
	}

	return nil
}

// UpdateItemStatus updates the status and error info for a Plaid item
func (r *PlaidRepository) UpdateItemStatus(ctx context.Context, id uuid.UUID, status, errorCode, errorMessage string) error {
	query := `
		UPDATE plaid_items
		SET status = $1, error_code = $2, error_message = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := r.db.Pool.Exec(ctx, query, status, errorCode, errorMessage, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update plaid item status: %w", err)
	}

	return nil
}

// DeleteItem deletes a Plaid item
func (r *PlaidRepository) DeleteItem(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM plaid_items WHERE id = $1 AND user_id = $2`

	result, err := r.db.Pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete plaid item: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("plaid item not found")
	}

	return nil
}

// CreateAccount creates a new Plaid account
func (r *PlaidRepository) CreateAccount(ctx context.Context, account *models.PlaidAccount) error {
	query := `
		INSERT INTO plaid_accounts (
			id, plaid_item_id, user_id, account_id, plaid_account_id,
			name, official_name, type, subtype, mask, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		ON CONFLICT (plaid_item_id, plaid_account_id) DO UPDATE
		SET name = EXCLUDED.name, official_name = EXCLUDED.official_name,
			type = EXCLUDED.type, subtype = EXCLUDED.subtype, updated_at = EXCLUDED.updated_at
	`

	now := time.Now()
	account.CreatedAt = now
	account.UpdatedAt = now

	_, err := r.db.Pool.Exec(ctx, query,
		account.ID,
		account.PlaidItemID,
		account.UserID,
		account.AccountID,
		account.PlaidAccountID,
		account.Name,
		account.OfficialName,
		account.Type,
		account.Subtype,
		account.Mask,
		account.CreatedAt,
		account.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create plaid account: %w", err)
	}

	return nil
}

// GetAccountsByItemID retrieves all Plaid accounts for an item
func (r *PlaidRepository) GetAccountsByItemID(ctx context.Context, plaidItemID uuid.UUID) ([]models.PlaidAccount, error) {
	query := `
		SELECT id, plaid_item_id, user_id, account_id, plaid_account_id,
			name, official_name, type, subtype, mask, created_at, updated_at
		FROM plaid_accounts
		WHERE plaid_item_id = $1
		ORDER BY name
	`

	rows, err := r.db.Pool.Query(ctx, query, plaidItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plaid accounts: %w", err)
	}
	defer rows.Close()

	var accounts []models.PlaidAccount
	for rows.Next() {
		var account models.PlaidAccount
		var accountID *uuid.UUID
		var officialName sql.NullString

		err := rows.Scan(
			&account.ID,
			&account.PlaidItemID,
			&account.UserID,
			&accountID,
			&account.PlaidAccountID,
			&account.Name,
			&officialName,
			&account.Type,
			&account.Subtype,
			&account.Mask,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plaid account: %w", err)
		}

		account.AccountID = accountID
		if officialName.Valid {
			account.OfficialName = officialName.String
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

// GetAccountsByUserID retrieves all Plaid accounts for a user
func (r *PlaidRepository) GetAccountsByUserID(ctx context.Context, userID uuid.UUID) ([]models.PlaidAccount, error) {
	query := `
		SELECT id, plaid_item_id, user_id, account_id, plaid_account_id,
			name, official_name, type, subtype, mask, created_at, updated_at
		FROM plaid_accounts
		WHERE user_id = $1
		ORDER BY name
	`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get plaid accounts: %w", err)
	}
	defer rows.Close()

	var accounts []models.PlaidAccount
	for rows.Next() {
		var account models.PlaidAccount
		var accountID *uuid.UUID
		var officialName sql.NullString

		err := rows.Scan(
			&account.ID,
			&account.PlaidItemID,
			&account.UserID,
			&accountID,
			&account.PlaidAccountID,
			&account.Name,
			&officialName,
			&account.Type,
			&account.Subtype,
			&account.Mask,
			&account.CreatedAt,
			&account.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plaid account: %w", err)
		}

		account.AccountID = accountID
		if officialName.Valid {
			account.OfficialName = officialName.String
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

// LinkAccountToOwnUAccount links a Plaid account to an OwnU account
func (r *PlaidRepository) LinkAccountToOwnUAccount(ctx context.Context, plaidAccountID, ownuAccountID uuid.UUID) error {
	query := `
		UPDATE plaid_accounts
		SET account_id = $1, updated_at = $2
		WHERE id = $3
	`

	_, err := r.db.Pool.Exec(ctx, query, ownuAccountID, time.Now(), plaidAccountID)
	if err != nil {
		return fmt.Errorf("failed to link plaid account: %w", err)
	}

	return nil
}

// CreateSync creates a sync history record
func (r *PlaidRepository) CreateSync(ctx context.Context, sync *models.PlaidSync) error {
	query := `
		INSERT INTO plaid_syncs (
			id, plaid_item_id, user_id, added_count, modified_count,
			removed_count, cursor_before, cursor_after, error_message, synced_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	sync.SyncedAt = time.Now()

	_, err := r.db.Pool.Exec(ctx, query,
		sync.ID,
		sync.PlaidItemID,
		sync.UserID,
		sync.AddedCount,
		sync.ModifiedCount,
		sync.RemovedCount,
		sync.CursorBefore,
		sync.CursorAfter,
		sync.ErrorMessage,
		sync.SyncedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create plaid sync: %w", err)
	}

	return nil
}

// GetSyncsByItemID retrieves sync history for an item
func (r *PlaidRepository) GetSyncsByItemID(ctx context.Context, plaidItemID uuid.UUID, limit int) ([]models.PlaidSync, error) {
	query := `
		SELECT id, plaid_item_id, user_id, added_count, modified_count,
			removed_count, cursor_before, cursor_after, error_message, synced_at
		FROM plaid_syncs
		WHERE plaid_item_id = $1
		ORDER BY synced_at DESC
		LIMIT $2
	`

	rows, err := r.db.Pool.Query(ctx, query, plaidItemID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get plaid syncs: %w", err)
	}
	defer rows.Close()

	var syncs []models.PlaidSync
	for rows.Next() {
		var sync models.PlaidSync
		var errorMessage sql.NullString

		err := rows.Scan(
			&sync.ID,
			&sync.PlaidItemID,
			&sync.UserID,
			&sync.AddedCount,
			&sync.ModifiedCount,
			&sync.RemovedCount,
			&sync.CursorBefore,
			&sync.CursorAfter,
			&errorMessage,
			&sync.SyncedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan plaid sync: %w", err)
		}

		if errorMessage.Valid {
			sync.ErrorMessage = errorMessage.String
		}

		syncs = append(syncs, sync)
	}

	return syncs, nil
}
