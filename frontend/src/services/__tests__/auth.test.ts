import { describe, it, expect, vi, beforeEach } from 'vitest'
import { isWebAuthnSupported, isPlatformAuthenticatorAvailable } from '../auth'

describe('auth service', () => {
  describe('isWebAuthnSupported', () => {
    it('returns true when PublicKeyCredential is available', () => {
      expect(isWebAuthnSupported()).toBe(true)
    })

    it('returns false when PublicKeyCredential is undefined', () => {
      const original = window.PublicKeyCredential
      // @ts-expect-error - testing undefined case
      window.PublicKeyCredential = undefined

      expect(isWebAuthnSupported()).toBe(false)

      window.PublicKeyCredential = original
    })
  })

  describe('isPlatformAuthenticatorAvailable', () => {
    beforeEach(() => {
      vi.clearAllMocks()
    })

    it('returns true when platform authenticator is available', async () => {
      vi.mocked(window.PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable)
        .mockResolvedValue(true)

      const result = await isPlatformAuthenticatorAvailable()

      expect(result).toBe(true)
    })

    it('returns false when platform authenticator is not available', async () => {
      vi.mocked(window.PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable)
        .mockResolvedValue(false)

      const result = await isPlatformAuthenticatorAvailable()

      expect(result).toBe(false)
    })

    it('returns false when WebAuthn is not supported', async () => {
      const original = window.PublicKeyCredential
      // @ts-expect-error - testing undefined case
      window.PublicKeyCredential = undefined

      const result = await isPlatformAuthenticatorAvailable()

      expect(result).toBe(false)

      window.PublicKeyCredential = original
    })

    it('returns false when check throws an error', async () => {
      vi.mocked(window.PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable)
        .mockRejectedValue(new Error('Not supported'))

      const result = await isPlatformAuthenticatorAvailable()

      expect(result).toBe(false)
    })
  })
})
