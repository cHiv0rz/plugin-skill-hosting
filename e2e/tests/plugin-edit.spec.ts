import { test, expect } from '@playwright/test'
import { register, createPlugin, type Creds } from './helpers'

const owner: Creds = {
  email: 'editor@e2e.test',
  username: 'e2eeditor',
  password: 'supersecret123',
}

test('owner can edit a plugin\'s metadata', async ({ page }) => {
  await register(page, owner)
  await createPlugin(page, 'editable-plugin', 'original description')

  // The description shown under the header reflects the stored value.
  await expect(page.locator('.pd-desc')).toHaveText('original description')

  // "edit metadata" auto-switches to the meta tab and reveals the form.
  await page.getByRole('button', { name: 'edit metadata', exact: true }).click()
  const form = page.locator('form.pd-form')
  await expect(form).toBeVisible()
  // First field is the (required) description.
  await form.locator('input').first().fill('updated description')
  await page.getByRole('button', { name: 'save', exact: true }).click()

  // Form closes and the new description is rendered.
  await expect(form).toBeHidden()
  await expect(page.locator('.pd-desc')).toHaveText('updated description')

  // …and it survives a reload (i.e. it was persisted, not just local state).
  await page.reload()
  await expect(page.locator('.pd-desc')).toHaveText('updated description')
})
