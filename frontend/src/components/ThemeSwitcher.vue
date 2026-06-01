<script setup lang="ts">
import { computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { THEMES } from '../theme'

const auth = useAuthStore()

// Two-way binding to the store: the getter tracks the active theme, the setter
// kicks off the (optimistic) switch. A rejected server save is swallowed here —
// the store has already rolled the value back, so the bound getter snaps the
// <select> to the previous theme on its own.
const current = computed<string>({
  get: () => auth.theme,
  set: (v) => { void auth.setTheme(v).catch(() => {}) },
})
</script>

<template>
  <label class="theme-switcher">
    <span class="theme-switcher__label">Theme</span>
    <select
      v-model="current"
      class="theme-switcher__select"
      aria-label="Theme"
      title="Choose a theme"
    >
      <option v-for="t in THEMES" :key="t.id" :value="t.id">{{ t.label }}</option>
    </select>
  </label>
</template>

<style scoped>
.theme-switcher {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  margin: 0;
}
.theme-switcher__label {
  font-family: var(--mono);
  font-size: 10.5px;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--muted);
}
/* Compact, inline variant of the global <select> — auto width, no full-row
   bottom border block. */
.theme-switcher__select {
  width: auto;
  padding: 4px 4px 5px;
  border-bottom: 1px solid var(--border);
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.06em;
  color: var(--text-soft);
  cursor: pointer;
  transition: color 0.2s ease, border-color 0.2s ease;
}
.theme-switcher__select:hover { color: var(--text); }
.theme-switcher__select:focus { border-bottom-color: var(--accent); }
</style>
