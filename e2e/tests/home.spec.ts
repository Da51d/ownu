import { test, expect } from '@playwright/test'

test.describe('Home Page', () => {
  test('redirects to login when not authenticated', async ({ page }) => {
    await page.goto('/')

    // Should redirect to login
    await expect(page).toHaveURL('/login')
  })

  test('login page has correct elements', async ({ page }) => {
    await page.goto('/login')

    // Check for key elements
    await expect(page.getByText('Welcome back')).toBeVisible()
    await expect(page.getByLabel(/username/i)).toBeVisible()
    await expect(page.getByRole('button', { name: /sign in with passkey/i })).toBeVisible()
    await expect(page.getByRole('link', { name: /create one/i })).toBeVisible()
  })

  test('register page has correct elements', async ({ page }) => {
    await page.goto('/register')

    // Check for key elements
    await expect(page.getByText('Create your account')).toBeVisible()
    await expect(page.getByLabel(/username/i)).toBeVisible()
    await expect(page.getByRole('button', { name: /create passkey/i })).toBeVisible()
    await expect(page.getByRole('link', { name: /sign in/i })).toBeVisible()
  })

  test('can navigate between login and register', async ({ page }) => {
    await page.goto('/login')

    // Click register link
    await page.getByRole('link', { name: /create one/i }).click()
    await expect(page).toHaveURL('/register')

    // Click login link
    await page.getByRole('link', { name: /sign in/i }).click()
    await expect(page).toHaveURL('/login')
  })
})

test.describe('API Health Check', () => {
  test('backend health endpoint responds', async ({ request }) => {
    const response = await request.get('/api/v1/health', {
      ignoreHTTPSErrors: true,
    })

    // The health endpoint should work through nginx proxy
    // It may return 200 or other status depending on backend state
    expect(response.status()).toBeLessThan(500)
  })
})
