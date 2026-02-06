import { useState } from 'react'
import { Link } from 'react-router-dom'
import { register, isWebAuthnSupported } from '../services/auth'

interface RegisterProps {
  onLogin: (token: string) => void
}

type Step = 'username' | 'recovery' | 'complete'

export default function Register({ onLogin }: RegisterProps) {
  const [step, setStep] = useState<Step>('username')
  const [username, setUsername] = useState('')
  const [recoveryPhrase, setRecoveryPhrase] = useState('')
  const [confirmed, setConfirmed] = useState(false)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [token, setToken] = useState('')

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setLoading(true)

    if (!isWebAuthnSupported()) {
      setError('WebAuthn is not supported in this browser')
      setLoading(false)
      return
    }

    try {
      const result = await register(username)
      setRecoveryPhrase(result.recoveryPhrase)
      setToken(result.response.token)
      setStep('recovery')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Registration failed')
    } finally {
      setLoading(false)
    }
  }

  const handleConfirmRecovery = () => {
    if (confirmed) {
      onLogin(token)
    }
  }

  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      padding: '1rem',
    }}>
      <div className="card" style={{ width: '100%', maxWidth: '480px' }}>
        {step === 'username' && (
          <>
            <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
              <h1 style={{ fontSize: '1.75rem', fontWeight: 'bold', color: 'var(--gray-900)' }}>
                Create your account
              </h1>
              <p style={{ color: 'var(--gray-500)', marginTop: '0.5rem' }}>
                Set up passwordless authentication with a passkey
              </p>
            </div>

            <form onSubmit={handleRegister}>
              <div style={{ marginBottom: '1.5rem' }}>
                <label htmlFor="username" className="label">
                  Username
                </label>
                <input
                  id="username"
                  type="text"
                  className="input"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="Choose a username"
                  required
                  autoComplete="username"
                />
              </div>

              {error && (
                <div className="error" style={{ marginBottom: '1rem' }}>
                  {error}
                </div>
              )}

              <button
                type="submit"
                className="btn btn-primary"
                style={{ width: '100%' }}
                disabled={loading || !username}
              >
                {loading ? 'Creating passkey...' : 'Create Passkey'}
              </button>
            </form>

            <div style={{
              marginTop: '1.5rem',
              textAlign: 'center',
              color: 'var(--gray-500)',
            }}>
              Already have an account?{' '}
              <Link to="/login" style={{ color: 'var(--primary)', textDecoration: 'none' }}>
                Sign in
              </Link>
            </div>
          </>
        )}

        {step === 'recovery' && (
          <>
            <div style={{ textAlign: 'center', marginBottom: '2rem' }}>
              <h1 style={{ fontSize: '1.75rem', fontWeight: 'bold', color: 'var(--gray-900)' }}>
                Save your recovery phrase
              </h1>
              <p style={{ color: 'var(--gray-500)', marginTop: '0.5rem' }}>
                This is the only way to recover your account if you lose access
              </p>
            </div>

            <div style={{
              background: 'var(--warning)',
              color: 'white',
              padding: '0.75rem 1rem',
              borderRadius: '0.5rem',
              marginBottom: '1.5rem',
              fontSize: '0.875rem',
            }}>
              <strong>Important:</strong> Write down these words and store them securely offline.
              Anyone with this phrase can access your account.
            </div>

            <div className="recovery-phrase" style={{ marginBottom: '1.5rem' }}>
              {recoveryPhrase}
            </div>

            <div style={{ marginBottom: '1.5rem' }}>
              <label style={{
                display: 'flex',
                alignItems: 'flex-start',
                gap: '0.75rem',
                cursor: 'pointer',
              }}>
                <input
                  type="checkbox"
                  checked={confirmed}
                  onChange={(e) => setConfirmed(e.target.checked)}
                  style={{ marginTop: '0.25rem' }}
                />
                <span style={{ fontSize: '0.875rem', color: 'var(--gray-700)' }}>
                  I have written down my recovery phrase and stored it in a secure location
                </span>
              </label>
            </div>

            <button
              onClick={handleConfirmRecovery}
              className="btn btn-primary"
              style={{ width: '100%' }}
              disabled={!confirmed}
            >
              Continue to Dashboard
            </button>
          </>
        )}
      </div>
    </div>
  )
}
