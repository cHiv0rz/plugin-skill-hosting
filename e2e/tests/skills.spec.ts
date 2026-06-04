import { test, expect } from '@playwright/test'
import { register, createPlugin, type Creds } from './helpers'

const owner: Creds = {
  email: 'skiller@e2e.test',
  username: 'e2eskiller',
  password: 'supersecret123',
}

test('create a skill in a plugin, then edit its description', async ({ page }) => {
  await register(page, owner)
  await createPlugin(page, 'skill-host-plugin', 'holds a skill')

  const skillName = 'my-first-skill'

  // --- create the skill ---
  await page.getByRole('link', { name: '+ new skill' }).click()
  await expect(page).toHaveURL(`/plugins/skill-host-plugin/skills/new`)
  await page.getByPlaceholder('my-skill-slug').fill(skillName)
  // The body editor ships a sensible default, so description is the only field
  // we need to set for a valid skill.
  await page.getByPlaceholder(/One sentence/).fill('does a useful thing')
  await page.getByRole('button', { name: 'create', exact: true }).click()

  // Back on the plugin detail page, the skill shows up in the skills table.
  await expect(page).toHaveURL(`/plugins/skill-host-plugin`)
  await expect(page.getByRole('link', { name: skillName })).toBeVisible()
  await expect(page.locator('.pd-table__desc')).toHaveText('does a useful thing')

  // --- edit the skill's description ---
  await page.getByRole('link', { name: skillName }).click()
  await expect(page).toHaveURL(`/plugins/skill-host-plugin/skills/${skillName}/edit`)
  const description = page.getByPlaceholder(/One sentence/)
  await description.fill('does an even more useful thing')
  await page.getByRole('button', { name: 'save', exact: true }).click()

  // Redirects back to the plugin; the table reflects the new description.
  await expect(page).toHaveURL(`/plugins/skill-host-plugin`)
  await expect(page.locator('.pd-table__desc')).toHaveText('does an even more useful thing')
})
