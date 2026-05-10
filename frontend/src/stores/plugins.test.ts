import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('../api', () => ({
  api: {
    listPlugins: vi.fn(),
    listDeletedPlugins: vi.fn(),
    getPlugin: vi.fn(),
    createPlugin: vi.fn(),
    deletePlugin: vi.fn(),
    restorePlugin: vi.fn(),
  },
}))

import { api } from '../api'
import { usePluginStore } from './plugins'
import type { Plugin } from '../types'

function makePlugin(name: string, overrides: Partial<Plugin> = {}): Plugin {
  return {
    id: `p-${name}`,
    ownerId: 'u1',
    name,
    description: `desc for ${name}`,
    version: '0.1.0',
    authorName: '',
    authorEmail: '',
    homepage: '',
    license: 'MIT',
    createdAt: '',
    updatedAt: '',
    ...overrides,
  }
}

describe('plugin store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('loadList populates list from API', async () => {
    const p = makePlugin('a')
    vi.mocked(api.listPlugins).mockResolvedValue([p])
    const s = usePluginStore()
    await s.loadList()
    expect(s.list).toEqual([p])
  })

  it('loadDeleted populates deleted from API', async () => {
    const p = makePlugin('a', { deletedAt: '2026-01-01' })
    vi.mocked(api.listDeletedPlugins).mockResolvedValue([p])
    const s = usePluginStore()
    await s.loadDeleted()
    expect(s.deleted).toEqual([p])
  })

  it('loadPlugin sets current and returns it', async () => {
    const p = makePlugin('a')
    vi.mocked(api.getPlugin).mockResolvedValue(p)
    const s = usePluginStore()
    const result = await s.loadPlugin('a')
    expect(s.current).toEqual(p)
    expect(result).toEqual(p)
  })

  it('createPlugin appends to list', async () => {
    const a = makePlugin('a')
    const b = makePlugin('b')
    vi.mocked(api.listPlugins).mockResolvedValue([a])
    vi.mocked(api.createPlugin).mockResolvedValue(b)
    const s = usePluginStore()
    await s.loadList()
    await s.createPlugin({ name: 'b' })
    expect(s.list.map(p => p.name)).toEqual(['a', 'b'])
  })

  it('deletePlugin removes from list and clears current if matching', async () => {
    const a = makePlugin('a')
    const b = makePlugin('b')
    vi.mocked(api.listPlugins).mockResolvedValue([a, b])
    vi.mocked(api.getPlugin).mockResolvedValue(a)
    vi.mocked(api.deletePlugin).mockResolvedValue(undefined)
    const s = usePluginStore()
    await s.loadList()
    await s.loadPlugin('a')
    await s.deletePlugin('a')
    expect(s.list.map(p => p.name)).toEqual(['b'])
    expect(s.current).toBeNull()
  })

  it('deletePlugin keeps current if it does not match', async () => {
    const a = makePlugin('a')
    const b = makePlugin('b')
    vi.mocked(api.listPlugins).mockResolvedValue([a, b])
    vi.mocked(api.getPlugin).mockResolvedValue(b)
    vi.mocked(api.deletePlugin).mockResolvedValue(undefined)
    const s = usePluginStore()
    await s.loadList()
    await s.loadPlugin('b')
    await s.deletePlugin('a')
    expect(s.current).toEqual(b)
  })

  it('restorePlugin moves a plugin from deleted to list', async () => {
    const a = makePlugin('a', { deletedAt: '2026-01-01' })
    const restored = makePlugin('a')
    vi.mocked(api.listDeletedPlugins).mockResolvedValue([a])
    vi.mocked(api.restorePlugin).mockResolvedValue(restored)
    const s = usePluginStore()
    await s.loadDeleted()
    await s.restorePlugin('a')
    expect(s.deleted).toEqual([])
    expect(s.list).toEqual([restored])
  })

  it('refreshCurrent re-fetches the named plugin', async () => {
    const v1 = makePlugin('a', { version: '0.1.0' })
    const v2 = makePlugin('a', { version: '0.2.0' })
    vi.mocked(api.getPlugin).mockResolvedValueOnce(v1).mockResolvedValueOnce(v2)
    const s = usePluginStore()
    await s.loadPlugin('a')
    await s.refreshCurrent()
    expect(s.current?.version).toBe('0.2.0')
  })

  it('refreshCurrent is a no-op when current is null', async () => {
    const s = usePluginStore()
    const r = await s.refreshCurrent()
    expect(r).toBeNull()
    expect(api.getPlugin).not.toHaveBeenCalled()
  })
})
