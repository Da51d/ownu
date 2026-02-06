package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB wraps the PostgreSQL connection pool
type DB struct {
	Pool *pgxpool.Pool
}

// NewDB creates a new database connection pool
func NewDB(ctx context.Context, databaseURL string) (*DB, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

// Close closes the database connection pool
func (db *DB) Close() {
	db.Pool.Close()
}

// RunMigrations executes database migrations
func (db *DB) RunMigrations(ctx context.Context) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			username VARCHAR(255) UNIQUE NOT NULL,
			encrypted_dek BYTEA NOT NULL,
			dek_salt BYTEA NOT NULL,
			recovery_phrase_hash BYTEA NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS credentials (
			id BYTEA PRIMARY KEY,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			public_key BYTEA NOT NULL,
			attestation_type VARCHAR(255) NOT NULL DEFAULT '',
			sign_count INTEGER DEFAULT 0,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS categories (
			id UUID PRIMARY KEY,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			encrypted_name BYTEA NOT NULL,
			parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
			is_system BOOLEAN DEFAULT FALSE
		)`,
		`CREATE TABLE IF NOT EXISTS accounts (
			id UUID PRIMARY KEY,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			encrypted_data BYTEA NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS transactions (
			id UUID PRIMARY KEY,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			account_id UUID REFERENCES accounts(id) ON DELETE CASCADE,
			encrypted_data BYTEA NOT NULL,
			category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
			transaction_date DATE NOT NULL,
			created_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_transactions_user_date ON transactions(user_id, transaction_date)`,
		`CREATE INDEX IF NOT EXISTS idx_accounts_user ON accounts(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_credentials_user ON credentials(user_id)`,
		// Plaid integration tables
		`CREATE TABLE IF NOT EXISTS plaid_items (
			id UUID PRIMARY KEY,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			item_id VARCHAR(255) UNIQUE NOT NULL,
			encrypted_access_token BYTEA NOT NULL,
			institution_id VARCHAR(255),
			institution_name VARCHAR(255),
			cursor VARCHAR(255),
			status VARCHAR(50) DEFAULT 'active',
			error_code VARCHAR(100),
			error_message TEXT,
			consent_expires_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS plaid_accounts (
			id UUID PRIMARY KEY,
			plaid_item_id UUID REFERENCES plaid_items(id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,
			plaid_account_id VARCHAR(255) NOT NULL,
			name VARCHAR(255),
			official_name VARCHAR(255),
			type VARCHAR(50),
			subtype VARCHAR(50),
			mask VARCHAR(10),
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(plaid_item_id, plaid_account_id)
		)`,
		`CREATE TABLE IF NOT EXISTS plaid_syncs (
			id UUID PRIMARY KEY,
			plaid_item_id UUID REFERENCES plaid_items(id) ON DELETE CASCADE,
			user_id UUID REFERENCES users(id) ON DELETE CASCADE,
			added_count INTEGER DEFAULT 0,
			modified_count INTEGER DEFAULT 0,
			removed_count INTEGER DEFAULT 0,
			cursor_before VARCHAR(255),
			cursor_after VARCHAR(255),
			error_message TEXT,
			synced_at TIMESTAMPTZ DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_plaid_items_user ON plaid_items(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_plaid_accounts_user ON plaid_accounts(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_plaid_syncs_item ON plaid_syncs(plaid_item_id)`,
	}

	for _, migration := range migrations {
		if _, err := db.Pool.Exec(ctx, migration); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}
