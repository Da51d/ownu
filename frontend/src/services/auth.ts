import {
  startRegistration,
  startAuthentication,
} from '@simplewebauthn/browser'
import { api } from './api'
import type { AuthResponse } from '../types'

// Backend returns WebAuthn options in go-webauthn format
interface WebAuthnOptions {
  publicKey: {
    challenge: string
    rp: { name: string; id: string }
    user: { id: string; name: string; displayName: string }
    pubKeyCredParams: Array<{ type: string; alg: number }>
    timeout?: number
    attestation?: string
    authenticatorSelection?: {
      authenticatorAttachment?: string
      requireResidentKey?: boolean
      residentKey?: string
      userVerification?: string
    }
    allowCredentials?: Array<{
      type: string
      id: string
      transports?: string[]
    }>
    rpId?: string
    userVerification?: string
  }
}

interface RegisterBeginResponse {
  options: WebAuthnOptions
  recovery_phrase: string
  session_id: string
}

interface LoginBeginResponse {
  options: WebAuthnOptions
  session_id: string
}

export async function beginRegistration(username: string): Promise<RegisterBeginResponse> {
  return api.post<RegisterBeginResponse>('/auth/register/begin', { username })
}

export async function finishRegistration(
  sessionId: string,
  credential: unknown
): Promise<AuthResponse> {
  return api.post<AuthResponse>('/auth/register/finish', {
    session_id: sessionId,
    credential: JSON.stringify(credential),
  })
}

export async function beginLogin(username: string): Promise<LoginBeginResponse> {
  return api.post<LoginBeginResponse>('/auth/login/begin', { username })
}

export async function finishLogin(
  sessionId: string,
  credential: unknown
): Promise<AuthResponse> {
  return api.post<AuthResponse>('/auth/login/finish', {
    session_id: sessionId,
    credential: JSON.stringify(credential),
  })
}

export async function register(username: string): Promise<{
  response: AuthResponse
  recoveryPhrase: string
}> {
  // Step 1: Begin registration
  const beginResponse = await beginRegistration(username)

  // Step 2: Create credential using WebAuthn
  // Pass the publicKey options directly to simplewebauthn
  // @ts-expect-error - go-webauthn format is compatible but types differ slightly
  const credential = await startRegistration(beginResponse.options.publicKey)

  // Step 3: Finish registration
  const response = await finishRegistration(beginResponse.session_id, credential)

  return {
    response,
    recoveryPhrase: beginResponse.recovery_phrase,
  }
}

export async function login(username: string): Promise<AuthResponse> {
  // Step 1: Begin login
  const beginResponse = await beginLogin(username)

  // Step 2: Get credential using WebAuthn
  // Pass the publicKey options directly to simplewebauthn
  // @ts-expect-error - go-webauthn format is compatible but types differ slightly
  const credential = await startAuthentication(beginResponse.options.publicKey)

  // Step 3: Finish login
  return finishLogin(beginResponse.session_id, credential)
}

export function isWebAuthnSupported(): boolean {
  return (
    window.PublicKeyCredential !== undefined &&
    typeof window.PublicKeyCredential === 'function'
  )
}

export async function isPlatformAuthenticatorAvailable(): Promise<boolean> {
  if (!isWebAuthnSupported()) {
    return false
  }
  try {
    return await PublicKeyCredential.isUserVerifyingPlatformAuthenticatorAvailable()
  } catch {
    return false
  }
}
