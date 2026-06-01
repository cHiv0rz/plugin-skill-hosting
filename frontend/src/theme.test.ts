import { describe, it, expect, beforeEach } from 'vitest'
import {
  THEMES,
  DEFAULT_THEME,
  isValidTheme,
  normalizeTheme,
  applyTheme,
  getStoredTheme,
  setStoredTheme,
} from './theme'

describe('theme module', () => {
  beforeEach(() => {
    localStorage.clear()
    delete document.documentElement.dataset.theme
  })

  it('exposes the default among the registered themes', () => {
    expect(THEMES.map(t => t.id)).toContain(DEFAULT_THEME)
  })

  it('isValidTheme accepts known ids and rejects everything else', () => {
    expect(isValidTheme('dark')).toBe(true)
    expect(isValidTheme('light')).toBe(true)
    expect(isValidTheme('nope')).toBe(false)
    expect(isValidTheme('')).toBe(false)
    expect(isValidTheme(undefined)).toBe(false)
    expect(isValidTheme(42)).toBe(false)
  })

  it('normalizeTheme passes through valid ids and falls back otherwise', () => {
    expect(normalizeTheme('sepia')).toBe('sepia')
    expect(normalizeTheme('garbage')).toBe(DEFAULT_THEME)
    expect(normalizeTheme(null)).toBe(DEFAULT_THEME)
  })

  it('applyTheme writes a normalized data-theme onto <html>', () => {
    applyTheme('midnight')
    expect(document.documentElement.dataset.theme).toBe('midnight')
    applyTheme('bogus')
    expect(document.documentElement.dataset.theme).toBe(DEFAULT_THEME)
  })

  it('getStoredTheme returns the default when nothing is stored', () => {
    expect(getStoredTheme()).toBe(DEFAULT_THEME)
  })

  it('round-trips a stored theme and sanitizes a bad one', () => {
    setStoredTheme('contrast')
    expect(getStoredTheme()).toBe('contrast')
    localStorage.setItem('theme', 'tampered')
    expect(getStoredTheme()).toBe(DEFAULT_THEME)
  })
})
