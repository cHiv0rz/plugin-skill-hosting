import { describe, it, expect } from 'vitest'
import { render, screen } from '@testing-library/vue'
import userEvent from '@testing-library/user-event'
import PromptDialog from './PromptDialog.vue'
import { usePrompt } from '../composables/usePrompt'

// The composable exposes a singleton `active` ref, so each test needs to
// resolve any pending dialog before the next one renders.

describe('PromptDialog', () => {
  it('does not render when no dialog is active', () => {
    const { answer } = usePrompt()
    answer(null) // clear any leftover state from prior tests
    render(PromptDialog)
    expect(screen.queryByRole('dialog')).toBeNull()
  })

  it('renders the active prompt and resolves with the typed value on confirm', async () => {
    const user = userEvent.setup()
    render(PromptDialog)
    const { prompt } = usePrompt()
    const p = prompt({ message: 'name?', confirmLabel: 'Create' })

    expect(await screen.findByRole('dialog')).toBeInTheDocument()
    expect(screen.getByText('name?')).toBeInTheDocument()

    const input = screen.getByRole('textbox') as HTMLInputElement
    await user.type(input, 'foo.py')
    await user.click(screen.getByRole('button', { name: 'Create' }))
    await expect(p).resolves.toBe('foo.py')
  })

  it('resolves null when the cancel button is clicked', async () => {
    const user = userEvent.setup()
    render(PromptDialog)
    const { prompt } = usePrompt()
    const p = prompt({ message: 'cancel me' })

    await screen.findByRole('dialog')
    await user.click(screen.getByRole('button', { name: 'Cancel' }))
    await expect(p).resolves.toBeNull()
  })

  it('Escape resolves null; Enter resolves the typed value', async () => {
    const user = userEvent.setup()
    render(PromptDialog)
    const { prompt } = usePrompt()

    const cancelled = prompt({ message: 'press esc' })
    await screen.findByRole('dialog')
    await user.keyboard('{Escape}')
    await expect(cancelled).resolves.toBeNull()

    const accepted = prompt({ message: 'press enter', initialValue: 'preset' })
    await screen.findByRole('dialog')
    // initial value is selected, so typing replaces it
    await user.keyboard('typed')
    await user.keyboard('{Enter}')
    await expect(accepted).resolves.toBe('typed')
  })
})
