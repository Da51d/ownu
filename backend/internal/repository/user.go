package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/ownu/ownu/internal/models"
)

var ErrNotFound = errors.New("not found")

// UserRepository handles user database operations
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, username, encrypted_dek, dek_salt, recovery_phrase_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		user.ID,
		user.Username,
		user.EncryptedDEK,
		user.DEKSalt,
		user.RecoveryPhraseHash,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	query := `
		SELECT id, username, encrypted_dek, dek_salt, recovery_phrase_hash, created_at, updated_at
		FROM users WHERE id = $1
	`
	user := &models.User{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.EncryptedDEK,
		&user.DEKSalt,
		&user.RecoveryPhraseHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Load credentials
	creds, err := r.GetCredentialsByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Credentials = creds

	return user, nil
}

// GetByUsername retrieves a user by username
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	query := `
		SELECT id, username, encrypted_dek, dek_salt, recovery_phrase_hash, created_at, updated_at
		FROM users WHERE username = $1
	`
	user := &models.User{}
	err := r.db.Pool.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.EncryptedDEK,
		&user.DEKSalt,
		&user.RecoveryPhraseHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Load credentials
	creds, err := r.GetCredentialsByUserID(ctx, user.ID)
	if err != nil {
		return nil, err
	}
	user.Credentials = creds

	return user, nil
}

// CreateCredential stores a new WebAuthn credential
func (r *UserRepository) CreateCredential(ctx context.Context, cred *models.Credential) error {
	query := `
		INSERT INTO credentials (id, user_id, public_key, attestation_type, sign_count, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.db.Pool.Exec(ctx, query,
		cred.ID,
		cred.UserID,
		cred.PublicKey,
		cred.AttestationType,
		cred.SignCount,
		cred.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}
	return nil
}

// GetCredentialsByUserID retrieves all credentials for a user
func (r *UserRepository) GetCredentialsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Credential, error) {
	query := `
		SELECT id, user_id, public_key, attestation_type, sign_count, created_at
		FROM credentials WHERE user_id = $1
	`
	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get credentials: %w", err)
	}
	defer rows.Close()

	var creds []models.Credential
	for rows.Next() {
		var cred models.Credential
		if err := rows.Scan(
			&cred.ID,
			&cred.UserID,
			&cred.PublicKey,
			&cred.AttestationType,
			&cred.SignCount,
			&cred.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan credential: %w", err)
		}
		creds = append(creds, cred)
	}

	return creds, nil
}

// UpdateCredentialSignCount updates the sign count for a credential
func (r *UserRepository) UpdateCredentialSignCount(ctx context.Context, credID []byte, signCount uint32) error {
	query := `UPDATE credentials SET sign_count = $1 WHERE id = $2`
	_, err := r.db.Pool.Exec(ctx, query, signCount, credID)
	if err != nil {
		return fmt.Errorf("failed to update sign count: %w", err)
	}
	return nil
}

// Delete permanently removes a user and all associated data
// Due to ON DELETE CASCADE, this will also delete:
// - credentials
// - accounts
// - transactions
// - categories
// - plaid_items, plaid_accounts, plaid_syncs
func (r *UserRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.Pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
