import { ref } from 'vue'

export interface PromptOptions {
  title?: string
  message: string
  placeholder?: string
  initialValue?: string
  confirmLabel?: string
  cancelLabel?: string
}

interface ActivePrompt extends Required<PromptOptions> {
  resolve: (v: string | null) => void
}

const active = ref<ActivePrompt | null>(null)

export function usePrompt() {
  function prompt(opts: PromptOptions): Promise<string | null> {
    return new Promise(resolve => {
      active.value = {
        title: opts.title ?? 'Enter a value',
        message: opts.message,
        placeholder: opts.placeholder ?? '',
        initialValue: opts.initialValue ?? '',
        confirmLabel: opts.confirmLabel ?? 'OK',
        cancelLabel: opts.cancelLabel ?? 'Cancel',
        resolve,
      }
    })
  }

  function answer(value: string | null) {
    if (!active.value) return
    active.value.resolve(value)
    active.value = null
  }

  return { active, prompt, answer }
}
