import '@testing-library/jest-dom/vitest'
import { afterAll, afterEach, beforeAll, beforeEach } from 'vitest'
import { cleanup } from '@testing-library/vue'
import { server } from './server'

// jsdom 25 + Node 22+ leaves `localStorage` as `undefined` because Node's
// built-in localStorage stub shadows jsdom's. Provide a simple in-memory
// implementation so the app code that touches localStorage works under test.
function createMemoryStorage(): Storage {
  let store: Record<string, string> = {}
  return {
    get length() {
      return Object.keys(store).length
    },
    clear() { store = {} },
    getItem(k) { return Object.prototype.hasOwnProperty.call(store, k) ? store[k] : null },
    key(i) { return Object.keys(store)[i] ?? null },
    removeItem(k) { delete store[k] },
    setItem(k, v) { store[k] = String(v) },
  }
}

const memoryStorage = createMemoryStorage()
Object.defineProperty(globalThis, 'localStorage', {
  configurable: true,
  value: memoryStorage,
})
if (typeof window !== 'undefined') {
  Object.defineProperty(window, 'localStorage', {
    configurable: true,
    value: memoryStorage,
  })
}

// MSW lifecycle: start once, reset registered handlers between tests so they
// don't leak, and close at the end. Tests that never make a request are
// unaffected; those that do register handlers via `server.use(...)`.
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }))
afterAll(() => server.close())

beforeEach(() => {
  memoryStorage.clear()
})

afterEach(() => {
  cleanup()
  server.resetHandlers()
})
