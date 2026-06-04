import { setupServer } from 'msw/node'

// A single MSW server shared by every test. Individual tests register their own
// handlers with `server.use(...)`; the global setup (see setup.ts) starts it
// once, resets handlers after each test, and closes it at the end. Requests
// that no handler matches throw (onUnhandledRequest: 'error') so a view firing
// an unexpected API call fails loudly instead of silently hanging.
export const server = setupServer()
