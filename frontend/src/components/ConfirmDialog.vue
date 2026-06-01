<script setup lang="ts">
import { onMounted, onUnmounted, watch, nextTick, ref } from 'vue'
import { useConfirm } from '../composables/useConfirm'

const { active, answer } = useConfirm()
const confirmBtn = ref<HTMLButtonElement | null>(null)

function onKey(e: KeyboardEvent) {
  if (!active.value) return
  if (e.key === 'Escape') {
    e.preventDefault()
    answer(false)
  } else if (e.key === 'Enter') {
    e.preventDefault()
    answer(true)
  }
}

watch(active, async v => {
  if (v) {
    await nextTick()
    confirmBtn.value?.focus()
  }
})

onMounted(() => window.addEventListener('keydown', onKey))
onUnmounted(() => window.removeEventListener('keydown', onKey))
</script>

<template>
  <Teleport to="body">
    <Transition name="confirm">
      <div
        v-if="active"
        class="confirm-backdrop"
        role="dialog"
        aria-modal="true"
        @mousedown.self="answer(false)"
      >
        <div class="confirm-dialog">
          <h3 class="confirm-title">{{ active.title }}</h3>
          <p class="confirm-message">{{ active.message }}</p>
          <div class="confirm-actions">
            <button type="button" class="secondary" @click="answer(false)">
              {{ active.cancelLabel }}
            </button>
            <button
              ref="confirmBtn"
              type="button"
              :class="active.danger ? 'danger' : ''"
              @click="answer(true)"
            >
              {{ active.confirmLabel }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style>
.confirm-backdrop {
  position: fixed;
  inset: 0;
  background: rgb(var(--shadow-rgb) / 0.45);
  -webkit-backdrop-filter: blur(2px);
          backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}
.confirm-dialog {
  position: relative;
  background: linear-gradient(180deg, var(--panel), var(--panel-2));
  border: 1px solid var(--border);
  border-radius: 0;
  padding: 26px 28px 22px;
  width: 100%;
  max-width: 460px;
  box-shadow: 0 24px 80px rgb(var(--shadow-rgb) / 0.28);
}
.confirm-dialog::before {
  content: '';
  position: absolute;
  left: 0; top: 0; bottom: 0;
  width: 2px;
  background: var(--accent);
}
.confirm-title {
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
.confirm-message {
  margin: 0 0 22px;
  color: var(--text-soft);
  font-size: 13.5px;
  line-height: 1.55;
  white-space: pre-wrap;
}
.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

.confirm-enter-active,
.confirm-leave-active { transition: opacity 0.12s ease; }
.confirm-enter-active .confirm-dialog,
.confirm-leave-active .confirm-dialog { transition: transform 0.12s ease; }
.confirm-enter-from,
.confirm-leave-to { opacity: 0; }
.confirm-enter-from .confirm-dialog,
.confirm-leave-to .confirm-dialog { transform: translateY(-6px); }
</style>
