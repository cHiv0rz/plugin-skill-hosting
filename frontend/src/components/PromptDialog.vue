<script setup lang="ts">
import { onMounted, onUnmounted, watch, nextTick, ref } from 'vue'
import { usePrompt } from '../composables/usePrompt'

const { active, answer } = usePrompt()
const input = ref<HTMLInputElement | null>(null)
const value = ref('')

function submit() {
  if (!active.value) return
  answer(value.value)
}

function cancel() {
  answer(null)
}

function onKey(e: KeyboardEvent) {
  if (!active.value) return
  if (e.key === 'Escape') {
    e.preventDefault()
    cancel()
  } else if (e.key === 'Enter') {
    e.preventDefault()
    submit()
  }
}

watch(active, async v => {
  if (v) {
    value.value = v.initialValue
    await nextTick()
    input.value?.focus()
    input.value?.select()
  }
})

onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))
</script>

<template>
  <Teleport to="body">
    <Transition name="prompt">
      <div
        v-if="active"
        class="prompt-backdrop"
        role="dialog"
        aria-modal="true"
        @mousedown.self="cancel"
      >
        <div class="prompt-dialog">
          <h3 class="prompt-title">{{ active.title }}</h3>
          <p class="prompt-message">{{ active.message }}</p>
          <input
            ref="input"
            v-model="value"
            type="text"
            class="prompt-input"
            :placeholder="active.placeholder"
            autocomplete="off"
            spellcheck="false"
          />
          <div class="prompt-actions">
            <button type="button" class="secondary" @click="cancel">
              {{ active.cancelLabel }}
            </button>
            <button type="button" @click="submit">
              {{ active.confirmLabel }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style>
.prompt-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.65);
  -webkit-backdrop-filter: blur(2px);
          backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}
.prompt-dialog {
  position: relative;
  background: linear-gradient(180deg, rgba(20, 24, 31, 0.95), rgba(14, 17, 24, 0.95));
  border: 1px solid var(--border);
  border-radius: 0;
  padding: 26px 28px 22px;
  width: 100%;
  max-width: 460px;
  box-shadow: 0 24px 80px rgba(0, 0, 0, 0.65);
}
.prompt-dialog::before {
  content: '';
  position: absolute;
  left: 0; top: 0; bottom: 0;
  width: 2px;
  background: var(--accent);
}
.prompt-title {
  margin: 0 0 10px;
  font-family: 'Fraunces Variable', 'Fraunces', Georgia, serif;
  font-style: italic;
  font-weight: 380;
  font-size: 22px;
  line-height: 1.2;
  letter-spacing: -0.01em;
  text-transform: none;
  color: var(--text);
}
.prompt-message {
  margin: 0 0 12px;
  color: var(--text-soft);
  font-size: 13.5px;
  line-height: 1.55;
  white-space: pre-wrap;
}
.prompt-input {
  margin-bottom: 22px;
}
.prompt-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.prompt-enter-active,
.prompt-leave-active { transition: opacity 0.12s ease; }
.prompt-enter-active .prompt-dialog,
.prompt-leave-active .prompt-dialog { transition: transform 0.12s ease; }
.prompt-enter-from,
.prompt-leave-to { opacity: 0; }
.prompt-enter-from .prompt-dialog,
.prompt-leave-to .prompt-dialog { transform: translateY(-6px); }
</style>
