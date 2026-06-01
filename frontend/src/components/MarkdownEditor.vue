<script setup lang="ts">
import { nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Crepe } from '@milkdown/crepe'
import { replaceAll } from '@milkdown/kit/utils'
import '@milkdown/crepe/theme/common/style.css'
import '@milkdown/crepe/theme/frame.css'

const props = defineProps<{ modelValue: string }>()
const emit = defineEmits<{ 'update:modelValue': [value: string] }>()

type Mode = 'wysiwyg' | 'raw'
const STORAGE_KEY = 'skill-md-editor-mode'

function loadMode(): Mode {
  if (typeof localStorage === 'undefined') return 'wysiwyg'
  const v = localStorage.getItem(STORAGE_KEY)
  return v === 'raw' ? 'raw' : 'wysiwyg'
}

const mode = ref<Mode>(loadMode())
const root = ref<HTMLElement | null>(null)
let crepe: Crepe | null = null
// Tracks the last markdown we've either received from the editor or pushed
// into it, so that v-model round-trips don't fight Crepe's own state.
let internalValue = ''

async function mountCrepe() {
  if (!root.value || crepe) return
  crepe = new Crepe({
    root: root.value,
    defaultValue: props.modelValue,
  })
  await crepe.create()
  internalValue = props.modelValue
  crepe.on((listener) => {
    listener.markdownUpdated((_ctx, markdown) => {
      if (markdown === internalValue) return
      internalValue = markdown
      emit('update:modelValue', markdown)
    })
  })
}

async function unmountCrepe() {
  if (!crepe) return
  const c = crepe
  crepe = null
  await c.destroy()
}

onMounted(() => {
  if (mode.value === 'wysiwyg') mountCrepe()
})

watch(mode, async (next) => {
  if (typeof localStorage !== 'undefined') {
    localStorage.setItem(STORAGE_KEY, next)
  }
  if (next === 'wysiwyg') {
    await nextTick()
    await mountCrepe()
  } else {
    await unmountCrepe()
  }
})

watch(
  () => props.modelValue,
  (next) => {
    if (!crepe) return
    if (next === internalValue) return
    internalValue = next
    crepe.editor.action(replaceAll(next))
  },
)

onBeforeUnmount(unmountCrepe)

function onRawInput(ev: Event) {
  emit('update:modelValue', (ev.target as HTMLTextAreaElement).value)
}
</script>

<template>
  <div class="md-editor">
    <div class="md-editor-toolbar">
      <span class="spacer" />
      <div class="md-mode-toggle" role="tablist" aria-label="Editor mode">
        <button
          type="button"
          class="md-mode-btn"
          :class="{ 'md-mode-btn--active': mode === 'wysiwyg' }"
          role="tab"
          :aria-selected="mode === 'wysiwyg'"
          @click="mode = 'wysiwyg'"
        >Visual</button>
        <button
          type="button"
          class="md-mode-btn"
          :class="{ 'md-mode-btn--active': mode === 'raw' }"
          role="tab"
          :aria-selected="mode === 'raw'"
          @click="mode = 'raw'"
        >Raw</button>
      </div>
    </div>
    <textarea
      v-if="mode === 'raw'"
      class="md-raw"
      :value="modelValue"
      @input="onRawInput"
    />
    <div v-else ref="root" class="md-editor-root" />
  </div>
</template>

<style scoped>
.md-editor {
  border: 1px solid var(--border);
  background: var(--panel-2);
  transition: border-color 0.25s ease;
}
.md-editor:focus-within {
  border-color: var(--accent);
}
.md-editor-toolbar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-bottom: 1px solid var(--border-soft);
  background: var(--panel);
}
.md-mode-toggle {
  display: inline-flex;
  border: 1px solid var(--border);
}
.md-mode-btn {
  background: transparent;
  color: var(--text-soft);
  border: 0;
  border-radius: 0;
  padding: 4px 12px;
  font-family: var(--mono);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  cursor: pointer;
  transition: color 0.15s ease, background 0.15s ease;
}
.md-mode-btn::before { display: none; }
.md-mode-btn:hover {
  color: var(--text);
  background: rgb(var(--text-rgb) / 0.05);
  transform: none;
}
.md-mode-btn + .md-mode-btn {
  border-left: 1px solid var(--border);
}
.md-mode-btn--active,
.md-mode-btn--active:hover {
  color: var(--text);
  background: var(--accent);
}
.md-editor-root {
  min-height: 360px;
}
.md-editor :deep(.milkdown) {
  background: transparent;
  color: var(--text);
}
.md-editor :deep(.ProseMirror) {
  min-height: 320px;
  padding: 14px 16px;
  outline: none;
}
/* Neutralize the app's editorial heading styles inside the editor so
   authored content looks like normal markdown, not section chrome. */
.md-editor :deep(.ProseMirror) :is(h1, h2, h3, h4, h5, h6) {
  font-family: var(--mono);
  font-style: normal;
  font-weight: 700;
  text-transform: none;
  letter-spacing: 0;
  color: var(--text);
  line-height: 1.3;
  margin: 18px 0 8px;
}
.md-editor :deep(.ProseMirror) h1 { font-size: 22px; }
.md-editor :deep(.ProseMirror) h2 { font-size: 18px; }
.md-editor :deep(.ProseMirror) h3 { font-size: 15px; }
.md-editor :deep(.ProseMirror) h4 { font-size: 14px; }
.md-editor :deep(.ProseMirror) h5,
.md-editor :deep(.ProseMirror) h6 {
  font-size: 13px;
  color: var(--text-soft);
}
.md-editor :deep(.ProseMirror) :is(h1, h2, h3, h4, h5, h6):first-child {
  margin-top: 0;
}
.md-editor :deep(.ProseMirror) code {
  background: var(--bg-2);
  border: 1px solid var(--border-soft);
  color: var(--accent-2);
  padding: 1px 6px;
  font-size: 12.5px;
}
.md-editor :deep(.ProseMirror) pre {
  background: var(--bg-2);
  border: 1px solid var(--border-soft);
  padding: 12px 14px;
  margin: 10px 0;
}
.md-editor :deep(.ProseMirror) pre code {
  background: none;
  border: 0;
  padding: 0;
  color: inherit;
}
.md-raw {
  display: block;
  width: 100%;
  background: var(--panel-2);
  color: var(--text);
  border: 0;
  border-radius: 0;
  padding: 14px 16px;
  font-family: var(--mono);
  font-size: 13px;
  line-height: 1.65;
  outline: none;
  resize: vertical;
  min-height: 360px;
}
</style>
