package plaid

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/ownu/ownu/internal/config"
	"github.com/plaid/plaid-go/v27/plaid"
)

// Service handles Plaid API interactions
type Service struct {
	client *plaid.APIClient
	config *config.Config
}

// NewService creates a new Plaid service
func NewService(cfg *config.Config) (*Service, error) {
	if cfg.PlaidClientID == "" || cfg.PlaidSecret == "" {
		return nil, fmt.Errorf("plaid credentials not configured")
	}

	// Determine the Plaid environment
	var env plaid.Environment
	switch cfg.PlaidEnv {
	case "sandbox", "development":
		env = plaid.Sandbox
	case "production":
		env = plaid.Production
	default:
		env = plaid.Sandbox
	}

	// Configure the Plaid client
	configuration := plaid.NewConfiguration()
	configuration.AddDefaultHeader("PLAID-CLIENT-ID", cfg.PlaidClientID)
	configuration.AddDefaultHeader("PLAID-SECRET", cfg.PlaidSecret)
	configuration.UseEnvironment(env)

	client := plaid.NewAPIClient(configuration)

	return &Service{
		client: client,
		config: cfg,
	}, nil
}

// IsConfigured returns true if Plaid is configured
func IsConfigured(cfg *config.Config) bool {
	return cfg.PlaidClientID != "" && cfg.PlaidSecret != ""
}

// CreateLinkToken creates a link token for the Plaid Link UI
func (s *Service) CreateLinkToken(ctx context.Context, userID uuid.UUID) (string, error) {
	user := plaid.LinkTokenCreateRequestUser{
		ClientUserId: userID.String(),
	}

	request := plaid.NewLinkTokenCreateRequest(
		"OwnU",
		"en",
		[]plaid.CountryCode{plaid.COUNTRYCODE_US},
		user,
	)

	// Set products - transactions for bank account linking
	request.SetProducts([]plaid.Products{plaid.PRODUCTS_TRANSACTIONS})

	// Set webhook URL if configured
	if s.config.PlaidWebhookURL != "" {
		request.SetWebhook(s.config.PlaidWebhookURL)
	}

	response, _, err := s.client.PlaidApi.LinkTokenCreate(ctx).LinkTokenCreateRequest(*request).Execute()
	if err != nil {
		return "", fmt.Errorf("failed to create link token: %w", err)
	}

	return response.GetLinkToken(), nil
}

// ExchangePublicToken exchanges a public token for an access token
func (s *Service) ExchangePublicToken(ctx context.Context, publicToken string) (string, string, error) {
	request := plaid.NewItemPublicTokenExchangeRequest(publicToken)

	response, _, err := s.client.PlaidApi.ItemPublicTokenExchange(ctx).ItemPublicTokenExchangeRequest(*request).Execute()
	if err != nil {
		return "", "", fmt.Errorf("failed to exchange public token: %w", err)
	}

	return response.GetAccessToken(), response.GetItemId(), nil
}

// GetItem retrieves item information
func (s *Service) GetItem(ctx context.Context, accessToken string) (*plaid.Item, error) {
	request := plaid.NewItemGetRequest(accessToken)

	response, _, err := s.client.PlaidApi.ItemGet(ctx).ItemGetRequest(*request).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	item := response.GetItem()
	return &item, nil
}

// GetAccounts retrieves accounts for an item
func (s *Service) GetAccounts(ctx context.Context, accessToken string) ([]plaid.AccountBase, error) {
	request := plaid.NewAccountsGetRequest(accessToken)

	response, _, err := s.client.PlaidApi.AccountsGet(ctx).AccountsGetRequest(*request).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get accounts: %w", err)
	}

	return response.GetAccounts(), nil
}

// GetInstitution retrieves institution information
func (s *Service) GetInstitution(ctx context.Context, institutionID string) (*plaid.Institution, error) {
	request := plaid.NewInstitutionsGetByIdRequest(
		institutionID,
		[]plaid.CountryCode{plaid.COUNTRYCODE_US},
	)

	response, _, err := s.client.PlaidApi.InstitutionsGetById(ctx).InstitutionsGetByIdRequest(*request).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to get institution: %w", err)
	}

	institution := response.GetInstitution()
	return &institution, nil
}

// SyncTransactionsResult holds the result of a transaction sync
type SyncTransactionsResult struct {
	Added      []plaid.Transaction
	Modified   []plaid.Transaction
	Removed    []plaid.RemovedTransaction
	NextCursor string
	HasMore    bool
}

// SyncTransactions syncs transactions for an item using the cursor
func (s *Service) SyncTransactions(ctx context.Context, accessToken string, cursor string) (*SyncTransactionsResult, error) {
	request := plaid.NewTransactionsSyncRequest(accessToken)
	if cursor != "" {
		request.SetCursor(cursor)
	}

	response, _, err := s.client.PlaidApi.TransactionsSync(ctx).TransactionsSyncRequest(*request).Execute()
	if err != nil {
		return nil, fmt.Errorf("failed to sync transactions: %w", err)
	}

	return &SyncTransactionsResult{
		Added:      response.GetAdded(),
		Modified:   response.GetModified(),
		Removed:    response.GetRemoved(),
		NextCursor: response.GetNextCursor(),
		HasMore:    response.GetHasMore(),
	}, nil
}

// RemoveItem removes an item from Plaid
func (s *Service) RemoveItem(ctx context.Context, accessToken string) error {
	request := plaid.NewItemRemoveRequest(accessToken)

	_, _, err := s.client.PlaidApi.ItemRemove(ctx).ItemRemoveRequest(*request).Execute()
	if err != nil {
		return fmt.Errorf("failed to remove item: %w", err)
	}

	return nil
}
