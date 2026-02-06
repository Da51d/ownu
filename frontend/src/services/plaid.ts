import { api } from './api'

export interface PlaidAccount {
  id: string
  plaid_item_id: string
  user_id: string
  account_id?: string
  name: string
  official_name?: string
  type: string
  subtype: string
  mask: string
  created_at: string
  updated_at: string
}

export interface PlaidItem {
  id: string
  institution_id: string
  institution_name: string
  status: string
  error_code?: string
  error_message?: string
  accounts: PlaidAccount[]
  created_at: string
  updated_at: string
}

export interface PlaidStatus {
  configured: boolean
}

export interface LinkTokenResponse {
  link_token: string
}

export interface ExchangeTokenResponse {
  item_id: string
  accounts: PlaidAccount[]
}

export interface SyncTransactionsResponse {
  added_count: number
  modified_count: number
  removed_count: number
  has_more: boolean
}

// Check if Plaid is configured on the server
export async function getPlaidStatus(): Promise<PlaidStatus> {
  return api.get<PlaidStatus>('/plaid/status')
}

// Create a link token for Plaid Link UI
export async function createLinkToken(): Promise<string> {
  const response = await api.post<LinkTokenResponse>('/plaid/link-token')
  return response.link_token
}

// Exchange public token after successful link
export async function exchangePublicToken(
  publicToken: string,
  institutionId: string,
  institutionName: string
): Promise<ExchangeTokenResponse> {
  return api.post<ExchangeTokenResponse>('/plaid/exchange-token', {
    public_token: publicToken,
    institution_id: institutionId,
    institution_name: institutionName,
  })
}

// Get all Plaid items for the user
export async function getPlaidItems(): Promise<PlaidItem[]> {
  return api.get<PlaidItem[]>('/plaid/items')
}

// Get a specific Plaid item
export async function getPlaidItem(itemId: string): Promise<PlaidItem> {
  return api.get<PlaidItem>(`/plaid/items/${itemId}`)
}

// Delete a Plaid item (disconnect bank)
export async function deletePlaidItem(itemId: string): Promise<void> {
  await api.delete(`/plaid/items/${itemId}`)
}

// Sync transactions for a Plaid item
export async function syncTransactions(itemId: string): Promise<SyncTransactionsResponse> {
  return api.post<SyncTransactionsResponse>(`/plaid/items/${itemId}/sync`)
}
