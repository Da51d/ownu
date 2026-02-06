import { useCallback, useEffect, useState } from 'react'
import { usePlaidLink, PlaidLinkOnSuccess, PlaidLinkOptions } from 'react-plaid-link'
import { createLinkToken, exchangePublicToken, PlaidItem } from '../services/plaid'

interface PlaidLinkProps {
  onSuccess: (item: { itemId: string; accounts: PlaidItem['accounts'] }) => void
  onExit?: () => void
  children?: React.ReactNode
  disabled?: boolean
}

export default function PlaidLinkButton({
  onSuccess,
  onExit,
  children,
  disabled = false,
}: PlaidLinkProps) {
  const [linkToken, setLinkToken] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Fetch link token on mount
  useEffect(() => {
    const fetchLinkToken = async () => {
      try {
        setLoading(true)
        setError(null)
        const token = await createLinkToken()
        setLinkToken(token)
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to initialize bank connection')
      } finally {
        setLoading(false)
      }
    }

    fetchLinkToken()
  }, [])

  // Handle successful link
  const handleSuccess = useCallback<PlaidLinkOnSuccess>(
    async (publicToken, metadata) => {
      try {
        setLoading(true)
        setError(null)

        const institutionId = metadata.institution?.institution_id || ''
        const institutionName = metadata.institution?.name || 'Unknown Bank'

        const response = await exchangePublicToken(
          publicToken,
          institutionId,
          institutionName
        )

        onSuccess({
          itemId: response.item_id,
          accounts: response.accounts,
        })
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to connect bank')
      } finally {
        setLoading(false)
      }
    },
    [onSuccess]
  )

  const config: PlaidLinkOptions = {
    token: linkToken,
    onSuccess: handleSuccess,
    onExit: onExit,
  }

  const { open, ready } = usePlaidLink(config)

  const handleClick = () => {
    if (ready && linkToken) {
      open()
    }
  }

  if (error) {
    return (
      <div style={{ color: 'var(--error)', fontSize: '0.875rem' }}>
        {error}
      </div>
    )
  }

  return (
    <button
      onClick={handleClick}
      disabled={disabled || !ready || loading}
      className="btn btn-primary"
      style={{ padding: '0.5rem 1rem', fontSize: '0.875rem' }}
    >
      {loading ? 'Connecting...' : children || 'Connect Bank'}
    </button>
  )
}
