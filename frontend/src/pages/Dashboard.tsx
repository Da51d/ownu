import { useState, useEffect, useCallback } from 'react'
import { api } from '../services/api'
import {
  getPlaidStatus,
  getPlaidItems,
  deletePlaidItem,
  syncTransactions,
  PlaidItem,
  PlaidStatus,
} from '../services/plaid'
import PlaidLinkButton from '../components/PlaidLink'

interface DashboardProps {
  onLogout: () => void
}

interface Account {
  id: string
  name: string
  type: string
  institution: string
  created_at: string
  updated_at: string
}

interface AccountFormData {
  name: string
  type: string
  institution: string
}

export default function Dashboard({ onLogout }: DashboardProps) {
  const [accounts, setAccounts] = useState<Account[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [showAddForm, setShowAddForm] = useState(false)
  const [formData, setFormData] = useState<AccountFormData>({
    name: '',
    type: 'checking',
    institution: '',
  })
  const [formError, setFormError] = useState('')
  const [formLoading, setFormLoading] = useState(false)

  // Plaid state
  const [plaidStatus, setPlaidStatus] = useState<PlaidStatus | null>(null)
  const [plaidItems, setPlaidItems] = useState<PlaidItem[]>([])
  const [syncingItem, setSyncingItem] = useState<string | null>(null)

  const fetchAccounts = useCallback(async () => {
    try {
      setLoading(true)
      const data = await api.get<Account[]>('/accounts')
      setAccounts(data || [])
      setError('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load accounts')
    } finally {
      setLoading(false)
    }
  }, [])

  const fetchPlaidData = useCallback(async () => {
    try {
      const status = await getPlaidStatus()
      setPlaidStatus(status)
      if (status.configured) {
        const items = await getPlaidItems()
        setPlaidItems(items || [])
      }
    } catch (err) {
      // Plaid fetch errors are not critical
      console.error('Failed to fetch Plaid data:', err)
    }
  }, [])

  useEffect(() => {
    fetchAccounts()
    fetchPlaidData()
  }, [fetchAccounts, fetchPlaidData])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormError('')
    setFormLoading(true)

    try {
      const newAccount = await api.post<Account>('/accounts', formData)
      setAccounts((prev) => [newAccount, ...prev])
      setShowAddForm(false)
      setFormData({ name: '', type: 'checking', institution: '' })
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to create account')
    } finally {
      setFormLoading(false)
    }
  }

  const handleDelete = async (accountId: string) => {
    if (!confirm('Are you sure you want to delete this account?')) {
      return
    }

    try {
      await api.delete(`/accounts/${accountId}`)
      setAccounts((prev) => prev.filter((a) => a.id !== accountId))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to delete account')
    }
  }

  const handlePlaidSuccess = (data: { itemId: string; accounts: PlaidItem['accounts'] }) => {
    // Refresh the Plaid items list
    fetchPlaidData()
  }

  const handleDisconnectBank = async (itemId: string) => {
    if (!confirm('Are you sure you want to disconnect this bank?')) {
      return
    }
    try {
      await deletePlaidItem(itemId)
      setPlaidItems((prev) => prev.filter((item) => item.id !== itemId))
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to disconnect bank')
    }
  }

  const handleSyncTransactions = async (itemId: string) => {
    try {
      setSyncingItem(itemId)
      const result = await syncTransactions(itemId)
      alert(`Synced: ${result.added_count} added, ${result.modified_count} modified, ${result.removed_count} removed`)
      if (result.has_more) {
        // Continue syncing if there's more
        await handleSyncTransactions(itemId)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to sync transactions')
    } finally {
      setSyncingItem(null)
    }
  }

  const accountTypes = [
    { value: 'checking', label: 'Checking' },
    { value: 'savings', label: 'Savings' },
    { value: 'credit', label: 'Credit Card' },
    { value: 'investment', label: 'Investment' },
    { value: 'loan', label: 'Loan' },
    { value: 'other', label: 'Other' },
  ]

  return (
    <div style={{ minHeight: '100vh' }}>
      {/* Header */}
      <header style={{
        background: 'white',
        borderBottom: '1px solid var(--gray-200)',
        padding: '1rem',
      }}>
        <div className="container" style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
        }}>
          <h1 style={{ fontSize: '1.5rem', fontWeight: 'bold', color: 'var(--gray-900)' }}>
            OwnU
          </h1>
          <button
            onClick={onLogout}
            style={{
              background: 'none',
              border: 'none',
              color: 'var(--gray-500)',
              cursor: 'pointer',
              fontSize: '0.875rem',
            }}
          >
            Sign out
          </button>
        </div>
      </header>

      {/* Main content */}
      <main className="container" style={{ padding: '2rem 1rem' }}>
        {error && (
          <div className="error" style={{ marginBottom: '1rem' }}>
            {error}
          </div>
        )}

        {loading ? (
          <div style={{ textAlign: 'center', color: 'var(--gray-500)' }}>
            Loading...
          </div>
        ) : (
          <>
            {/* Welcome section */}
            <div className="card" style={{ marginBottom: '1.5rem' }}>
              <h2 style={{ fontSize: '1.25rem', fontWeight: '600', marginBottom: '0.5rem' }}>
                Welcome to OwnU
              </h2>
              <p style={{ color: 'var(--gray-500)' }}>
                Your privacy-first personal finance tracker. All your data is encrypted and stored securely.
              </p>
            </div>

            {/* Add Account Form Modal */}
            {showAddForm && (
              <div style={{
                position: 'fixed',
                top: 0,
                left: 0,
                right: 0,
                bottom: 0,
                background: 'rgba(0, 0, 0, 0.5)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                zIndex: 1000,
              }}>
                <div className="card" style={{
                  width: '100%',
                  maxWidth: '400px',
                  margin: '1rem',
                }}>
                  <h2 style={{ fontSize: '1.25rem', fontWeight: '600', marginBottom: '1rem' }}>
                    Add Account
                  </h2>

                  <form onSubmit={handleSubmit}>
                    <div style={{ marginBottom: '1rem' }}>
                      <label htmlFor="name" className="label">
                        Account Name
                      </label>
                      <input
                        id="name"
                        type="text"
                        className="input"
                        value={formData.name}
                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                        placeholder="e.g., Main Checking"
                        required
                      />
                    </div>

                    <div style={{ marginBottom: '1rem' }}>
                      <label htmlFor="type" className="label">
                        Account Type
                      </label>
                      <select
                        id="type"
                        className="input"
                        value={formData.type}
                        onChange={(e) => setFormData({ ...formData, type: e.target.value })}
                      >
                        {accountTypes.map((type) => (
                          <option key={type.value} value={type.value}>
                            {type.label}
                          </option>
                        ))}
                      </select>
                    </div>

                    <div style={{ marginBottom: '1rem' }}>
                      <label htmlFor="institution" className="label">
                        Institution
                      </label>
                      <input
                        id="institution"
                        type="text"
                        className="input"
                        value={formData.institution}
                        onChange={(e) => setFormData({ ...formData, institution: e.target.value })}
                        placeholder="e.g., Chase Bank"
                      />
                    </div>

                    {formError && (
                      <div className="error" style={{ marginBottom: '1rem' }}>
                        {formError}
                      </div>
                    )}

                    <div style={{ display: 'flex', gap: '0.5rem' }}>
                      <button
                        type="button"
                        className="btn"
                        style={{
                          flex: 1,
                          background: 'var(--gray-100)',
                          color: 'var(--gray-700)',
                        }}
                        onClick={() => {
                          setShowAddForm(false)
                          setFormError('')
                          setFormData({ name: '', type: 'checking', institution: '' })
                        }}
                      >
                        Cancel
                      </button>
                      <button
                        type="submit"
                        className="btn btn-primary"
                        style={{ flex: 1 }}
                        disabled={formLoading || !formData.name}
                      >
                        {formLoading ? 'Adding...' : 'Add Account'}
                      </button>
                    </div>
                  </form>
                </div>
              </div>
            )}

            {/* Connected Banks section (Plaid) */}
            {plaidStatus?.configured && (
              <div className="card" style={{ marginBottom: '1.5rem' }}>
                <div style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  marginBottom: '1rem',
                }}>
                  <h2 style={{ fontSize: '1.125rem', fontWeight: '600' }}>
                    Connected Banks
                  </h2>
                  <PlaidLinkButton onSuccess={handlePlaidSuccess}>
                    Connect Bank
                  </PlaidLinkButton>
                </div>

                {plaidItems.length === 0 ? (
                  <div style={{
                    textAlign: 'center',
                    padding: '2rem',
                    color: 'var(--gray-500)',
                  }}>
                    <p>No banks connected yet.</p>
                    <p style={{ fontSize: '0.875rem', marginTop: '0.5rem' }}>
                      Connect your bank to automatically import transactions.
                    </p>
                  </div>
                ) : (
                  <div>
                    {plaidItems.map((item) => (
                      <div
                        key={item.id}
                        style={{
                          padding: '1rem',
                          borderBottom: '1px solid var(--gray-200)',
                        }}
                      >
                        <div style={{
                          display: 'flex',
                          justifyContent: 'space-between',
                          alignItems: 'center',
                          marginBottom: '0.5rem',
                        }}>
                          <div>
                            <div style={{ fontWeight: '500' }}>{item.institution_name}</div>
                            <div style={{ fontSize: '0.875rem', color: 'var(--gray-500)' }}>
                              {item.accounts?.length || 0} account{item.accounts?.length !== 1 ? 's' : ''} linked
                              {item.status !== 'active' && (
                                <span style={{ color: 'var(--error)', marginLeft: '0.5rem' }}>
                                  ({item.status})
                                </span>
                              )}
                            </div>
                          </div>
                          <div style={{ display: 'flex', gap: '0.5rem' }}>
                            <button
                              onClick={() => handleSyncTransactions(item.id)}
                              disabled={syncingItem === item.id}
                              style={{
                                background: 'none',
                                border: '1px solid var(--gray-300)',
                                borderRadius: '4px',
                                color: 'var(--gray-700)',
                                cursor: syncingItem === item.id ? 'wait' : 'pointer',
                                fontSize: '0.875rem',
                                padding: '0.25rem 0.5rem',
                              }}
                            >
                              {syncingItem === item.id ? 'Syncing...' : 'Sync'}
                            </button>
                            <button
                              onClick={() => handleDisconnectBank(item.id)}
                              style={{
                                background: 'none',
                                border: 'none',
                                color: 'var(--error)',
                                cursor: 'pointer',
                                fontSize: '0.875rem',
                                padding: '0.25rem 0.5rem',
                              }}
                            >
                              Disconnect
                            </button>
                          </div>
                        </div>
                        {/* Show linked accounts */}
                        {item.accounts && item.accounts.length > 0 && (
                          <div style={{ marginTop: '0.5rem', paddingLeft: '1rem' }}>
                            {item.accounts.map((acc) => (
                              <div
                                key={acc.id}
                                style={{
                                  fontSize: '0.875rem',
                                  color: 'var(--gray-600)',
                                  padding: '0.25rem 0',
                                }}
                              >
                                {acc.name} ({acc.subtype}) ****{acc.mask}
                              </div>
                            ))}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                )}
              </div>
            )}

            {/* Accounts section */}
            <div className="card">
              <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginBottom: '1rem',
              }}>
                <h2 style={{ fontSize: '1.125rem', fontWeight: '600' }}>
                  Your Accounts
                </h2>
                <button
                  className="btn btn-primary"
                  style={{ padding: '0.5rem 1rem', fontSize: '0.875rem' }}
                  onClick={() => setShowAddForm(true)}
                >
                  Add Account
                </button>
              </div>

              {accounts.length === 0 ? (
                <div style={{
                  textAlign: 'center',
                  padding: '2rem',
                  color: 'var(--gray-500)',
                }}>
                  <p>No accounts yet.</p>
                  <p style={{ fontSize: '0.875rem', marginTop: '0.5rem' }}>
                    Add your first account to start tracking your finances.
                  </p>
                </div>
              ) : (
                <div>
                  {accounts.map((account) => (
                    <div
                      key={account.id}
                      style={{
                        padding: '1rem',
                        borderBottom: '1px solid var(--gray-200)',
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                      }}
                    >
                      <div>
                        <div style={{ fontWeight: '500' }}>{account.name}</div>
                        <div style={{ fontSize: '0.875rem', color: 'var(--gray-500)' }}>
                          {account.institution ? `${account.institution} - ` : ''}
                          {accountTypes.find((t) => t.value === account.type)?.label || account.type}
                        </div>
                      </div>
                      <button
                        onClick={() => handleDelete(account.id)}
                        style={{
                          background: 'none',
                          border: 'none',
                          color: 'var(--error)',
                          cursor: 'pointer',
                          fontSize: '0.875rem',
                          padding: '0.5rem',
                        }}
                      >
                        Delete
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </>
        )}
      </main>
    </div>
  )
}
