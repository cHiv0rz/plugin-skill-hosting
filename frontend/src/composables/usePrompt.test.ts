import { describe, it, expect } from 'vitest'
import { usePrompt } from './usePrompt'

describe('usePrompt', () => {
  it('exposes prompt/answer that resolves a promise with the entered value', async () => {
    const { active, prompt, answer } = usePrompt()
    const p = prompt({ message: 'name?' })
    expect(active.value?.message).toBe('name?')
    answer('hello')
    await expect(p).resolves.toBe('hello')
    expect(active.value).toBeNull()
  })

  it('resolves null when cancelled', async () => {
    const { prompt, answer } = usePrompt()
    const p = prompt({ message: 'm' })
    answer(null)
    await expect(p).resolves.toBeNull()
  })

  it('fills defaults for optional fields', () => {
    const { active, prompt, answer } = usePrompt()
    prompt({ message: 'm' })
    expect(active.value).toMatchObject({
      title: 'Enter a value',
      message: 'm',
      placeholder: '',
      initialValue: '',
      confirmLabel: 'OK',
      cancelLabel: 'Cancel',
    })
    answer(null)
  })

  it('answer is a no-op when nothing is pending', () => {
    const { answer } = usePrompt()
    expect(() => answer('x')).not.toThrow()
  })

  it('shares state across hook calls (singleton store)', async () => {
    const a = usePrompt()
    const b = usePrompt()
    const p = a.prompt({ message: 'shared' })
    expect(b.active.value?.message).toBe('shared')
    b.answer('ok')
    await expect(p).resolves.toBe('ok')
  })
})
