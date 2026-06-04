import { describe, it, expect, beforeEach } from 'vitest'
import { render, screen, fireEvent, waitFor } from '@testing-library/vue'
import { createPinia } from 'pinia'
import { createRouter, createMemoryHistory, type Router } from 'vue-router'
import { http, HttpResponse } from 'msw'
import { server } from '../test/server'
import { makeAuthConfig, makeUser, makeJwt } from '../test/factories'
import LoginView from './LoginView.vue'

async function mountLogin(): Promise<Router> {
  const pinia = createPinia()
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div>home</div>' } },
      { path: '/login', component: LoginView },
      { path: '/register', component: { template: '<div>register</div>' } },
    ],
  })
  await router.push('/login')
  await router.isReady()
  render(LoginView, { global: { plugins: [pinia, router] } })
  return router
}

describe('LoginView', () => {
  beforeEach(() => {
    localStorage.clear()
    // The view resolves the auth mode on mount; password mode reveals the form.
    server.use(http.get('/api/auth/config', () => HttpResponse.json(makeAuthConfig())))
  })

  it('signs in with valid credentials and navigates home', async () => {
    const token = makeJwt(3600)
    let loginBody: unknown
    server.use(
      http.post('/api/auth/login', async ({ request }) => {
        loginBody = await request.json()
        return HttpResponse.json({ token, user: makeUser() })
      }),
    )

    const router = await mountLogin()
    // Wait for password mode to render the form.
    const email = await screen.findByLabelText('email')
    await fireEvent.update(email, 'alice@example.com')
    await fireEvent.update(screen.getByLabelText('password'), 'hunter2pass')
    await fireEvent.click(screen.getByRole('button', { name: 'sign in' }))

    await waitFor(() => expect(router.currentRoute.value.path).toBe('/'))
    expect(loginBody).toEqual({ email: 'alice@example.com', password: 'hunter2pass' })
    expect(localStorage.getItem('token')).toBe(token)
  })

  it('shows an inline error and stays on the page when login fails', async () => {
    server.use(
      http.post('/api/auth/login', () =>
        HttpResponse.json({ error: 'invalid credentials' }, { status: 401 }),
      ),
    )

    const router = await mountLogin()
    await fireEvent.update(await screen.findByLabelText('email'), 'alice@example.com')
    await fireEvent.update(screen.getByLabelText('password'), 'wrongpass1')
    await fireEvent.click(screen.getByRole('button', { name: 'sign in' }))

    expect(await screen.findByText('invalid credentials')).toBeInTheDocument()
    expect(router.currentRoute.value.path).toBe('/login')
    expect(localStorage.getItem('token')).toBeNull()
  })
})
