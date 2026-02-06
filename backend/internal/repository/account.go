package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ownu/ownu/internal/models"
)

// AccountRepository handles account database operations
type AccountRepository struct {
	db *DB
}

// NewAccountRepository creates a new account repository
func NewAccountRepository(db *DB) *AccountRepository {
	return &AccountRepository{db: db}
}

// Create inserts a new account into the database
func (r *AccountRepository) Create(ctx context.Context, account *models.Account) error {
	query := `
		INSERT INTO accounts (id, user_id, encrypted_data, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		account.ID,
		account.UserID,
		account.EncryptedData,
		account.CreatedAt,
		account.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}
	return nil
}

// GetByID retrieves an account by ID
func (r *AccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Account, error) {
	query := `
		SELECT id, user_id, encrypted_data, created_at, updated_at
		FROM accounts WHERE id = $1
	`
	account := &models.Account{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&account.ID,
		&account.UserID,
		&account.EncryptedData,
		&account.CreatedAt,
		&account.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get account: %w", err)
	}
	return account, nil
}

// GetByUserID retrieves all accounts for a user
func (r *AccountRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Account, error) {
	query := `
		SELECT id, user_id, encrypted_data, created_at, updated_at
		FROM accounts WHERE user_id = $1
		ORDER BY created_at DESC
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}
	defer rows.Close()

	var accounts []models.Account
	for rows.Next() {
		var account models.Account
		if err := rows.Scan(
			&account.ID,
			&account.UserID,
			&account.EncryptedData,
			&account.CreatedAt,
			&account.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan account: %w", err)
		}
		accounts = append(accounts, account)
	}

	return accounts, nil
}

// Update updates an existing account
func (r *AccountRepository) Update(ctx context.Context, account *models.Account) error {
	query := `
		UPDATE accounts
		SET encrypted_data = $1, updated_at = $2
		WHERE id = $3 AND user_id = $4
	`
	result, err := r.db.Pool.Exec(ctx, query,
		account.EncryptedData,
		time.Now(),
		account.ID,
		account.UserID,
	)
	if err != nil {
		return fmt.Errorf("failed to update account: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// Delete removes an account
func (r *AccountRepository) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	query := `DELETE FROM accounts WHERE id = $1 AND user_id = $2`
	result, err := r.db.Pool.Exec(ctx, query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete account: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
