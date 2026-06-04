import { test, expect } from '@playwright/test'
import { register, type Creds } from './helpers'

const user: Creds = {
  email: 'themer@e2e.test',
  username: 'e2ethemer',
  password: 'supersecret123',
}

test('theme selection applies and persists across a reload', async ({ page }) => {
  await register(page, user)

  const html = page.locator('html')
  // Default theme is light (no/`light` data-theme).
  const themePicker = page.getByRole('combobox', { name: 'Theme' })

  await themePicker.selectOption({ label: 'Dark' })
  await expect(html).toHaveAttribute('data-theme', 'dark')

  // The choice is persisted to the user's account (and localStorage), so it
  // survives a full reload — not just an in-memory toggle.
  await page.reload()
  await expect(html).toHaveAttribute('data-theme', 'dark')
})
