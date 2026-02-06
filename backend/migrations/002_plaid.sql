-- Plaid integration tables

-- Plaid Items (linked bank connections)
CREATE TABLE IF NOT EXISTS plaid_items (
    id UUID PRIMARY KEY,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    item_id VARCHAR(255) UNIQUE NOT NULL,             -- Plaid's item_id
    encrypted_access_token BYTEA NOT NULL,            -- Encrypted access token
    institution_id VARCHAR(255),                      -- Plaid institution ID
    institution_name VARCHAR(255),                    -- Institution display name
    cursor VARCHAR(255),                              -- Transactions sync cursor
    status VARCHAR(50) DEFAULT 'active',              -- active, error, pending_expiration
    error_code VARCHAR(100),                          -- Last error code if any
    error_message TEXT,                               -- Last error message
    consent_expires_at TIMESTAMPTZ,                   -- When user consent expires
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Plaid Accounts (accounts within a plaid item)
CREATE TABLE IF NOT EXISTS plaid_accounts (
    id UUID PRIMARY KEY,
    plaid_item_id UUID REFERENCES plaid_items(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    account_id UUID REFERENCES accounts(id) ON DELETE SET NULL,  -- Link to our account
    plaid_account_id VARCHAR(255) NOT NULL,          -- Plaid's account_id
    name VARCHAR(255),                               -- Account name from Plaid
    official_name VARCHAR(255),                      -- Official account name
    type VARCHAR(50),                                -- depository, credit, loan, investment
    subtype VARCHAR(50),                             -- checking, savings, credit card, etc.
    mask VARCHAR(10),                                -- Last 4 digits
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(plaid_item_id, plaid_account_id)
);

-- Plaid sync history (for debugging and monitoring)
CREATE TABLE IF NOT EXISTS plaid_syncs (
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
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_plaid_items_user ON plaid_items(user_id);
CREATE INDEX IF NOT EXISTS idx_plaid_items_item_id ON plaid_items(item_id);
CREATE INDEX IF NOT EXISTS idx_plaid_accounts_user ON plaid_accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_plaid_accounts_item ON plaid_accounts(plaid_item_id);
CREATE INDEX IF NOT EXISTS idx_plaid_syncs_item ON plaid_syncs(plaid_item_id);
CREATE INDEX IF NOT EXISTS idx_plaid_syncs_user ON plaid_syncs(user_id);
