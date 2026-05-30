import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('../api', async (importOriginal) => ({
  // Keep the real isJwtExpired so store init exercises the actual expiry logic;
  // only the network-touching `api` object is stubbed out.
  ...(await importOriginal<typeof import('../api')>()),
  api: {
    login: vi.fn(),
    register: vi.fn(),
    me: vi.fn(),
    regenerateToken: vi.fn(),
    revokeSessions: vi.fn(),
    authConfig: vi.fn(),
  },
}))

// makeJwt builds a structurally valid HS256-style JWT with the given exp
// (seconds since epoch); only the payload is decoded client-side, so the
// header/signature can be placeholders.
function makeJwt(expSeconds: number): string {
  const b64 = (o: unknown) =>
    btoa(JSON.stringify(o)).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '')
  return `${b64({ alg: 'HS256', typ: 'JWT' })}.${b64({ sub: 'u1', exp: expSeconds })}.sig`
}

import { api } from '../api'
import { useAuthStore } from './auth'

const fakeUser = {
  id: 'u1',
  email: 'a@b.c',
  username: 'alice',
  apiToken: 'tok',
  status: 'approved' as const,
  isAdmin: false,
}

describe('auth store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('rehydrates user/token from localStorage on init', () => {
    localStorage.setItem('token', 't0')
    localStorage.setItem('user', JSON.stringify(fakeUser))
    const s = useAuthStore()
    expect(s.token).toBe('t0')
    expect(s.user).toEqual(fakeUser)
  })

  it('keeps a still-valid JWT session on init', () => {
    const tok = makeJwt(Math.floor(Date.now() / 1000) + 3600)
    localStorage.setItem('token', tok)
    localStorage.setItem('user', JSON.stringify(fakeUser))
    const s = useAuthStore()
    expect(s.token).toBe(tok)
    expect(s.user).toEqual(fakeUser)
  })

  it('drops an expired JWT session (and cached user) on init', () => {
    localStorage.setItem('token', makeJwt(Math.floor(Date.now() / 1000) - 3600))
    localStorage.setItem('user', JSON.stringify(fakeUser))
    const s = useAuthStore()
    expect(s.token).toBeNull()
    expect(s.user).toBeNull()
    expect(localStorage.getItem('token')).toBeNull()
    expect(localStorage.getItem('user')).toBeNull()
  })

  it('returns null user when localStorage holds garbage JSON', () => {
    localStorage.setItem('user', '{not json')
    const s = useAuthStore()
    expect(s.user).toBeNull()
  })

  it('login persists token and user', async () => {
    vi.mocked(api.login).mockResolvedValue({ token: 'tok123', user: fakeUser })
    const s = useAuthStore()
    await s.login('a@b.c', 'pw')
    expect(s.token).toBe('tok123')
    expect(s.user).toEqual(fakeUser)
    expect(localStorage.getItem('token')).toBe('tok123')
    expect(JSON.parse(localStorage.getItem('user')!)).toEqual(fakeUser)
  })

  it('logout clears state and storage', () => {
    localStorage.setItem('token', 't')
    localStorage.setItem('user', JSON.stringify(fakeUser))
    const s = useAuthStore()
    s.logout()
    expect(s.token).toBeNull()
    expect(s.user).toBeNull()
    expect(localStorage.getItem('token')).toBeNull()
    expect(localStorage.getItem('user')).toBeNull()
  })

  it('doLogout stays local for password mode', async () => {
    vi.mocked(api.authConfig).mockResolvedValue({
      mode: 'password',
      marketplaceName: 'mp',
      defaultLicense: 'MIT',
      userApprovalRequired: false,
    })
    const s = useAuthStore()
    await s.ensureMode()
    localStorage.setItem('token', 't')
    localStorage.setItem('user', JSON.stringify(fakeUser))
    expect(s.doLogout()).toBe(false)
    expect(s.token).toBeNull()
    expect(localStorage.getItem('user')).toBeNull()
  })

  it('doLogout stays local for corporate OIDC (domain-restricted)', async () => {
    vi.mocked(api.authConfig).mockResolvedValue({
      mode: 'oidc',
      marketplaceName: 'mp',
      defaultLicense: 'MIT',
      userApprovalRequired: false,
    })
    const s = useAuthStore()
    await s.ensureMode()
    expect(s.doLogout()).toBe(false)
  })

  it('doLogout kicks off RP-initiated logout for open OIDC', async () => {
    vi.mocked(api.authConfig).mockResolvedValue({
      mode: 'oidc',
      marketplaceName: 'mp',
      defaultLicense: 'MIT',
      userApprovalRequired: true,
    })
    // jsdom's window.location is read-only by default; assign via Object.defineProperty.
    const setHref = vi.fn()
    Object.defineProperty(window, 'location', {
      configurable: true,
      value: {
        get href() { return '' },
        set href(v: string) { setHref(v) },
      },
    })

    const s = useAuthStore()
    await s.ensureMode()
    localStorage.setItem('user', JSON.stringify(fakeUser))
    expect(s.doLogout()).toBe(true)
    expect(setHref).toHaveBeenCalledWith('/api/auth/oidc/logout')
    // Local state still cleared, so the in-flight redirect lands on a clean SPA.
    expect(s.token).toBeNull()
    expect(localStorage.getItem('user')).toBeNull()
  })

  it('ensureMode caches the auth config response', async () => {
    vi.mocked(api.authConfig).mockResolvedValue({
      mode: 'password',
      marketplaceName: 'mp',
      defaultLicense: 'MIT',
      userApprovalRequired: false,
    })
    const s = useAuthStore()
    const m1 = await s.ensureMode()
    const m2 = await s.ensureMode()
    expect(m1).toBe('password')
    expect(m2).toBe('password')
    expect(api.authConfig).toHaveBeenCalledTimes(1)
    expect(s.marketplaceName).toBe('mp')
  })

  it('regenerateToken updates user.apiToken and storage', async () => {
    vi.mocked(api.regenerateToken).mockResolvedValue({ apiToken: 'NEW' })
    localStorage.setItem('user', JSON.stringify(fakeUser))
    const s = useAuthStore()
    const tok = await s.regenerateToken()
    expect(tok).toBe('NEW')
    expect(s.user?.apiToken).toBe('NEW')
    expect(JSON.parse(localStorage.getItem('user')!).apiToken).toBe('NEW')
  })

  it('signOutEverywhere revokes server-side then clears local state (password mode)', async () => {
    vi.mocked(api.authConfig).mockResolvedValue({
      mode: 'password',
      marketplaceName: 'mp',
      defaultLicense: 'MIT',
      userApprovalRequired: false,
    })
    vi.mocked(api.revokeSessions).mockResolvedValue()
    localStorage.setItem('token', 't')
    localStorage.setItem('user', JSON.stringify(fakeUser))
    const s = useAuthStore()
    await s.ensureMode()
    const redirecting = await s.signOutEverywhere()
    expect(api.revokeSessions).toHaveBeenCalledOnce()
    expect(redirecting).toBe(false) // password mode stays local, no OIDC redirect
    expect(s.token).toBeNull()
    expect(s.user).toBeNull()
    expect(localStorage.getItem('token')).toBeNull()
  })

  it('refreshUser is a no-op without a token', async () => {
    const s = useAuthStore()
    await s.refreshUser()
    expect(api.me).not.toHaveBeenCalled()
  })
})
