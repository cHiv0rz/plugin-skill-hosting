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
  background: rgba(0, 0, 0, 0.55);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 20px;
}
.confirm-dialog {
  background: var(--panel);
  border: 1px solid var(--border);
  border-radius: 10px;
  padding: 22px 22px 18px;
  width: 100%;
  max-width: 440px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
}
.confirm-title {
  margin: 0 0 8px;
  font-size: 16px;
  font-weight: 600;
}
.confirm-message {
  margin: 0 0 18px;
  color: var(--muted);
  font-size: 14px;
  line-height: 1.45;
  white-space: pre-wrap;
}
.confirm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
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
