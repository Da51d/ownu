import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '../../test/test-utils'
import Register from '../Register'

// Mock the auth service
vi.mock('../../services/auth', () => ({
  register: vi.fn(),
  isWebAuthnSupported: vi.fn().mockReturnValue(true),
}))

import { register, isWebAuthnSupported } from '../../services/auth'

describe('Register', () => {
  const mockOnLogin = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders registration form', () => {
    render(<Register onLogin={mockOnLogin} />)

    expect(screen.getByText('Create your account')).toBeInTheDocument()
    expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /create passkey/i })).toBeInTheDocument()
  })

  it('has link to login page', () => {
    render(<Register onLogin={mockOnLogin} />)

    const link = screen.getByRole('link', { name: /sign in/i })
    expect(link).toHaveAttribute('href', '/login')
  })

  it('disables submit button when username is empty', () => {
    render(<Register onLogin={mockOnLogin} />)

    const button = screen.getByRole('button', { name: /create passkey/i })
    expect(button).toBeDisabled()
  })

  it('enables submit button when username is entered', () => {
    render(<Register onLogin={mockOnLogin} />)

    const input = screen.getByLabelText(/username/i)
    fireEvent.change(input, { target: { value: 'testuser' } })

    const button = screen.getByRole('button', { name: /create passkey/i })
    expect(button).not.toBeDisabled()
  })

  it('shows error when WebAuthn is not supported', async () => {
    vi.mocked(isWebAuthnSupported).mockReturnValue(false)

    render(<Register onLogin={mockOnLogin} />)

    const input = screen.getByLabelText(/username/i)
    fireEvent.change(input, { target: { value: 'testuser' } })

    const button = screen.getByRole('button', { name: /create passkey/i })
    fireEvent.click(button)

    await waitFor(() => {
      expect(screen.getByText(/webauthn is not supported/i)).toBeInTheDocument()
    })
  })

  it('shows recovery phrase after successful registration', async () => {
    const mockRecoveryPhrase = 'abandon ability able about above absent absorb abstract absurd abuse access accident'
    vi.mocked(register).mockResolvedValue({
      response: { token: 'mock-token', user: { id: '1', username: 'testuser' } },
      recoveryPhrase: mockRecoveryPhrase,
    })
    vi.mocked(isWebAuthnSupported).mockReturnValue(true)

    render(<Register onLogin={mockOnLogin} />)

    const input = screen.getByLabelText(/username/i)
    fireEvent.change(input, { target: { value: 'testuser' } })

    const button = screen.getByRole('button', { name: /create passkey/i })
    fireEvent.click(button)

    await waitFor(() => {
      expect(screen.getByText('Save your recovery phrase')).toBeInTheDocument()
      expect(screen.getByText(mockRecoveryPhrase)).toBeInTheDocument()
    })
  })

  it('requires confirmation before proceeding after registration', async () => {
    vi.mocked(register).mockResolvedValue({
      response: { token: 'mock-token', user: { id: '1', username: 'testuser' } },
      recoveryPhrase: 'test phrase',
    })
    vi.mocked(isWebAuthnSupported).mockReturnValue(true)

    render(<Register onLogin={mockOnLogin} />)

    const input = screen.getByLabelText(/username/i)
    fireEvent.change(input, { target: { value: 'testuser' } })

    const button = screen.getByRole('button', { name: /create passkey/i })
    fireEvent.click(button)

    await waitFor(() => {
      expect(screen.getByText('Save your recovery phrase')).toBeInTheDocument()
    })

    // Continue button should be disabled until checkbox is checked
    const continueButton = screen.getByRole('button', { name: /continue to dashboard/i })
    expect(continueButton).toBeDisabled()

    // Check the confirmation checkbox
    const checkbox = screen.getByRole('checkbox')
    fireEvent.click(checkbox)

    expect(continueButton).not.toBeDisabled()
  })

  it('calls onLogin when user confirms recovery phrase', async () => {
    const mockToken = 'mock-token'
    vi.mocked(register).mockResolvedValue({
      response: { token: mockToken, user: { id: '1', username: 'testuser' } },
      recoveryPhrase: 'test phrase',
    })
    vi.mocked(isWebAuthnSupported).mockReturnValue(true)

    render(<Register onLogin={mockOnLogin} />)

    const input = screen.getByLabelText(/username/i)
    fireEvent.change(input, { target: { value: 'testuser' } })

    const button = screen.getByRole('button', { name: /create passkey/i })
    fireEvent.click(button)

    await waitFor(() => {
      expect(screen.getByText('Save your recovery phrase')).toBeInTheDocument()
    })

    const checkbox = screen.getByRole('checkbox')
    fireEvent.click(checkbox)

    const continueButton = screen.getByRole('button', { name: /continue to dashboard/i })
    fireEvent.click(continueButton)

    expect(mockOnLogin).toHaveBeenCalledWith(mockToken)
  })

  it('displays error message on registration failure', async () => {
    vi.mocked(register).mockRejectedValue(new Error('Username already exists'))
    vi.mocked(isWebAuthnSupported).mockReturnValue(true)

    render(<Register onLogin={mockOnLogin} />)

    const input = screen.getByLabelText(/username/i)
    fireEvent.change(input, { target: { value: 'testuser' } })

    const button = screen.getByRole('button', { name: /create passkey/i })
    fireEvent.click(button)

    await waitFor(() => {
      expect(screen.getByText('Username already exists')).toBeInTheDocument()
    })

    expect(mockOnLogin).not.toHaveBeenCalled()
  })
})
