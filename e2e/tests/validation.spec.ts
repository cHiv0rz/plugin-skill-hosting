import { test, expect } from '@playwright/test'
import { register, createPlugin, type Creds } from './helpers'

const owner: Creds = {
  email: 'validator@e2e.test',
  username: 'e2evalidator',
  password: 'supersecret123',
}

test('creating a plugin with a taken name surfaces a server error', async ({ page }) => {
  await register(page, owner)
  await createPlugin(page, 'taken-name', 'the original')

  // Try to create another plugin with the same (globally-unique) slug.
  await page.goto('/plugins/new')
  const inputs = page.locator('form input')
  await inputs.nth(0).fill('taken-name')
  await inputs.nth(1).fill('a clash')
  await page.getByRole('button', { name: 'Create plugin' }).click()

  // The backend rejects it (409) and the form stays put with an inline error.
  await expect(page.locator('.error')).toBeVisible()
  await expect(page).toHaveURL('/plugins/new')
})
