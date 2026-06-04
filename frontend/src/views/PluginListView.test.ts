import { describe, it, expect, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/vue'
import { createPinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'
import { http, HttpResponse } from 'msw'
import { server } from '../test/server'
import { makeAuthConfig, makePlugin } from '../test/factories'
import PluginListView from './PluginListView.vue'

async function mountList() {
  const pinia = createPinia()
  const router = createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: PluginListView },
      { path: '/plugins/new', component: { template: '<div>new</div>' } },
      { path: '/plugins/:name', component: { template: '<div>detail</div>' } },
    ],
  })
  await router.push('/')
  await router.isReady()
  render(PluginListView, { global: { plugins: [pinia, router] } })
}

describe('PluginListView', () => {
  beforeEach(() => {
    // load() resolves the auth mode alongside the plugin list.
    server.use(http.get('/api/auth/config', () => HttpResponse.json(makeAuthConfig())))
  })

  it('renders the empty state when there are no plugins', async () => {
    server.use(http.get('/api/plugins', () => HttpResponse.json([])))
    await mountList()
    expect(await screen.findByText('no plugins yet')).toBeInTheDocument()
  })

  it('lists plugins returned by the API', async () => {
    server.use(
      http.get('/api/plugins', () =>
        HttpResponse.json([
          makePlugin('alpha', { description: 'first plugin' }),
          makePlugin('beta', { description: 'second plugin' }),
        ]),
      ),
    )
    await mountList()

    // Names render as links to their detail pages.
    expect(await screen.findByRole('link', { name: 'alpha' })).toBeInTheDocument()
    expect(screen.getByRole('link', { name: 'beta' })).toBeInTheDocument()
    expect(screen.getByText('first plugin')).toBeInTheDocument()
    // The plugins tab count reflects the loaded set.
    expect(screen.getByRole('tab', { name: /plugins/ })).toHaveTextContent('[2]')
  })
})
