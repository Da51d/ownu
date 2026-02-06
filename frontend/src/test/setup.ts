import '@testing-library/jest-dom'
import { vi } from 'vitest'

// Mock window.PublicKeyCredential for WebAuthn tests
Object.defineProperty(window, 'PublicKeyCredential', {
  value: class MockPublicKeyCredential {
    static isUserVerifyingPlatformAuthenticatorAvailable = vi.fn().mockResolvedValue(true)
    static isConditionalMediationAvailable = vi.fn().mockResolvedValue(false)
  },
  writable: true,
})

// Mock navigator.credentials
Object.defineProperty(navigator, 'credentials', {
  value: {
    create: vi.fn(),
    get: vi.fn(),
  },
  writable: true,
})

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
}
Object.defineProperty(window, 'localStorage', { value: localStorageMock })

// Mock fetch
global.fetch = vi.fn()

// Reset mocks before each test
beforeEach(() => {
  vi.clearAllMocks()
  localStorageMock.getItem.mockReturnValue(null)
})
