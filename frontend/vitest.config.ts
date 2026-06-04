import { defineConfig } from 'vitest/config'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  test: {
    globals: true,
    environment: 'jsdom',
    setupFiles: ['./src/test/setup.ts'],
    css: false,
    coverage: {
      provider: 'v8',
      // text → CI log, text-summary → the one-line total, html → uploaded
      // artifact, json-summary → a machine-readable total for any later gate.
      reporter: ['text-summary', 'text', 'html', 'json-summary'],
      reportsDirectory: './coverage',
      include: ['src/**/*.{ts,vue}'],
      // Exclude things with no meaningful logic to cover: the tests and their
      // fixtures, type-only modules, the app entrypoint, and generated build info.
      exclude: [
        'src/**/*.test.ts',
        'src/test/**',
        'src/types.ts',
        'src/main.ts',
        'src/build-info.ts',
        'src/**/*.d.ts',
      ],
    },
  },
})
