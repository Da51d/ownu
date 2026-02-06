import { test, expect } from '@playwright/test'

test.describe('Authentication Flow', () => {
  test.describe('Registration', () => {
    test('shows error for empty username', async ({ page }) => {
      await page.goto('/register')

      const submitButton = page.getByRole('button', { name: /create passkey/i })

      // Button should be disabled when username is empty
      await expect(submitButton).toBeDisabled()
    })

    test('enables submit when username entered', async ({ page }) => {
      await page.goto('/register')

      await page.getByLabel(/username/i).fill('testuser')

      const submitButton = page.getByRole('button', { name: /create passkey/i })
      await expect(submitButton).toBeEnabled()
    })

    // Note: Full WebAuthn flow requires browser support and can't be fully tested
    // in automated E2E tests without mocking. These tests verify the UI flow.
  })

  test.describe('Login', () => {
    test('shows error for empty username', async ({ page }) => {
      await page.goto('/login')

      const submitButton = page.getByRole('button', { name: /sign in with passkey/i })

      // Button should be disabled when username is empty
      await expect(submitButton).toBeDisabled()
    })

    test('enables submit when username entered', async ({ page }) => {
      await page.goto('/login')

      await page.getByLabel(/username/i).fill('testuser')

      const submitButton = page.getByRole('button', { name: /sign in with passkey/i })
      await expect(submitButton).toBeEnabled()
    })
  })
})

test.describe('Protected Routes', () => {
  test('dashboard redirects to login when not authenticated', async ({ page }) => {
    await page.goto('/dashboard')

    // Should redirect to login
    await expect(page).toHaveURL('/login')
  })

  test('accessing protected API without token returns 401', async ({ request }) => {
    const response = await request.get('/api/v1/accounts', {
      ignoreHTTPSErrors: true,
    })

    expect(response.status()).toBe(401)
  })
})
