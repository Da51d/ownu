import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '../../test/test-utils'
import Login from '../Login'

// Mock the auth service
vi.mock('../../services/auth', () => ({
  login: vi.fn(),
  isWebAuthnSupported: vi.fn().mockReturnValue(true),
}))

import { login, isWebAuthnSupported } from '../../services/auth'

describe('Login', () => {
  const mockOnLogin = vi.fn()

  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders login form', () => {
    render(<Login onLogin={mockOnLogin} />)

    expect(screen.getByText('Welcome back')).toBeInTheDocument()
    expect(screen.getByLabelText(/username/i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /sign in with passkey/i })).toBeInTheDocument()
  })

  it('has link to registration page', () => {
    render(<Login onLogin={mockOnLogin} />)

    const link = screen.getByRole('link', { name: /create one/i })
    expect(link).toHaveAttribute('href', '/register')
  })

  it('disables submit button when username is empty', () => {
    render(<Login onLogin={mockOnLogin} />)

    const button = screen.getByRole('button', { name: /sign in with passkey/i })
    expect(button).toBeDisabled()
  })

  it('enables submit button when username is entered', () => {
    render(<Login onLogin={mockOnLogin} />)

    const input = screen.getByLabelText(/username/i)
    fireEvent.change(input, { target: { value: 'testuser' } })

    const button = screen.getByRole('button', { name: /sign in with passkey/i })
    expect(button).not.toBeDisabled()
  })

  it('shows error when WebAuthn is not supported', async () => {
    vi.mocked(isWebAuthnSupported).mockReturnValue(false)

    render(<Login onLogin={mockOnLogin} />)

    const input = screen.getByLabelText(/username/i)
    fireEvent.change(input, { target: { value: 'testuser' } })

    const button = screen.getByRole('button', { name: /sign in with passkey/i })
    fireEvent.click(button)

    await waitFor(() => {
      expect(screen.getByText(/webauthn is not supported/i)).toBeInTheDocument()
    })
  })

  it('calls login and onLogin on successful authentication', async () => {
    const mockToken = 'mock-jwt-token'
    vi.mocked(login).mockResolvedValue({ token: mockToken, user: { id: '1', username: 'testuser' } })
    vi.mocked(isWebAuthnSupported).mockReturnValue(true)

    render(<Login onLogin={mockOnLogin} />)

    const input = screen.getByLabelText(/username/i)
    fireEvent.change(input, { target: { value: 'testuser' } })

    const button = screen.getByRole('button', { name: /sign in with passkey/i })
    fireEvent.click(button)

    await waitFor(() => {
      expect(login).toHaveBeenCalledWith('testuser')
      expect(mockOnLogin).toHaveBeenCalledWith(mockToken)
    })
  })

  it('displays error message on login failure', async () => {
    vi.mocked(login).mockRejectedValue(new Error('Authentication failed'))
    vi.mocked(isWebAuthnSupported).mockReturnValue(true)

    render(<Login onLogin={mockOnLogin} />)

    const input = screen.getByLabelText(/username/i)
    fireEvent.change(input, { target: { value: 'testuser' } })

    const button = screen.getByRole('button', { name: /sign in with passkey/i })
    fireEvent.click(button)

    await waitFor(() => {
      expect(screen.getByText('Authentication failed')).toBeInTheDocument()
    })

    expect(mockOnLogin).not.toHaveBeenCalled()
  })

  it('shows loading state during authentication', async () => {
    vi.mocked(login).mockImplementation(() => new Promise(() => {})) // Never resolves
    vi.mocked(isWebAuthnSupported).mockReturnValue(true)

    render(<Login onLogin={mockOnLogin} />)

    const input = screen.getByLabelText(/username/i)
    fireEvent.change(input, { target: { value: 'testuser' } })

    const button = screen.getByRole('button', { name: /sign in with passkey/i })
    fireEvent.click(button)

    await waitFor(() => {
      expect(screen.getByText(/authenticating/i)).toBeInTheDocument()
    })
  })
})
