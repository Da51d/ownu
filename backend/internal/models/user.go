package models

import (
	"time"

	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

// User represents a user in the system
type User struct {
	ID                 uuid.UUID    `json:"id"`
	Username           string       `json:"username"`
	EncryptedDEK       []byte       `json:"-"`
	DEKSalt            []byte       `json:"-"`
	RecoveryPhraseHash []byte       `json:"-"`
	CreatedAt          time.Time    `json:"created_at"`
	UpdatedAt          time.Time    `json:"updated_at"`
	Credentials        []Credential `json:"-"`
}

// WebAuthnID returns the user's ID as bytes for WebAuthn
func (u *User) WebAuthnID() []byte {
	return u.ID[:]
}

// WebAuthnName returns the user's username
func (u *User) WebAuthnName() string {
	return u.Username
}

// WebAuthnDisplayName returns the user's display name
func (u *User) WebAuthnDisplayName() string {
	return u.Username
}

// WebAuthnIcon returns the user's icon URL (deprecated but required by interface)
func (u *User) WebAuthnIcon() string {
	return ""
}

// WebAuthnCredentials returns the user's WebAuthn credentials
func (u *User) WebAuthnCredentials() []webauthn.Credential {
	creds := make([]webauthn.Credential, len(u.Credentials))
	for i, c := range u.Credentials {
		creds[i] = webauthn.Credential{
			ID:              c.ID,
			PublicKey:       c.PublicKey,
			AttestationType: c.AttestationType,
			Authenticator: webauthn.Authenticator{
				SignCount: c.SignCount,
			},
		}
	}
	return creds
}

// Credential represents a WebAuthn credential
type Credential struct {
	ID              []byte    `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	PublicKey       []byte    `json:"-"`
	AttestationType string    `json:"-"`
	SignCount       uint32    `json:"-"`
	CreatedAt       time.Time `json:"created_at"`
}

// Account represents a financial account
type Account struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	EncryptedData []byte    `json:"-"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// AccountData represents decrypted account information
type AccountData struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Institution string `json:"institution"`
}

// Transaction represents a financial transaction
type Transaction struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	AccountID       uuid.UUID `json:"account_id"`
	EncryptedData   []byte    `json:"-"`
	CategoryID      *uuid.UUID `json:"category_id,omitempty"`
	TransactionDate time.Time `json:"transaction_date"`
	CreatedAt       time.Time `json:"created_at"`
}

// TransactionData represents decrypted transaction information
type TransactionData struct {
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
	Merchant    string  `json:"merchant"`
	Date        string  `json:"date"`
}

// Category represents a transaction category
type Category struct {
	ID            uuid.UUID  `json:"id"`
	UserID        uuid.UUID  `json:"user_id"`
	EncryptedName []byte     `json:"-"`
	ParentID      *uuid.UUID `json:"parent_id,omitempty"`
	IsSystem      bool       `json:"is_system"`
}

// PlaidItem represents a linked Plaid bank connection
type PlaidItem struct {
	ID                   uuid.UUID  `json:"id"`
	UserID               uuid.UUID  `json:"user_id"`
	ItemID               string     `json:"-"`
	EncryptedAccessToken []byte     `json:"-"`
	InstitutionID        string     `json:"institution_id"`
	InstitutionName      string     `json:"institution_name"`
	Cursor               string     `json:"-"`
	Status               string     `json:"status"`
	ErrorCode            string     `json:"error_code,omitempty"`
	ErrorMessage         string     `json:"error_message,omitempty"`
	ConsentExpiresAt     *time.Time `json:"consent_expires_at,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
}

// PlaidAccount represents an account from a Plaid item
type PlaidAccount struct {
	ID             uuid.UUID  `json:"id"`
	PlaidItemID    uuid.UUID  `json:"plaid_item_id"`
	UserID         uuid.UUID  `json:"user_id"`
	AccountID      *uuid.UUID `json:"account_id,omitempty"` // Link to our account
	PlaidAccountID string     `json:"-"`
	Name           string     `json:"name"`
	OfficialName   string     `json:"official_name,omitempty"`
	Type           string     `json:"type"`
	Subtype        string     `json:"subtype"`
	Mask           string     `json:"mask"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// PlaidSync represents a sync history record
type PlaidSync struct {
	ID            uuid.UUID `json:"id"`
	PlaidItemID   uuid.UUID `json:"plaid_item_id"`
	UserID        uuid.UUID `json:"user_id"`
	AddedCount    int       `json:"added_count"`
	ModifiedCount int       `json:"modified_count"`
	RemovedCount  int       `json:"removed_count"`
	CursorBefore  string    `json:"-"`
	CursorAfter   string    `json:"-"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	SyncedAt      time.Time `json:"synced_at"`
}
