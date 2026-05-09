import { ref } from 'vue'

export interface ConfirmOptions {
  title?: string
  message: string
  confirmLabel?: string
  cancelLabel?: string
  danger?: boolean
}

interface ActiveConfirm extends Required<Omit<ConfirmOptions, 'danger'>> {
  danger: boolean
  resolve: (v: boolean) => void
}

const active = ref<ActiveConfirm | null>(null)

export function useConfirm() {
  function confirm(opts: ConfirmOptions): Promise<boolean> {
    return new Promise(resolve => {
      active.value = {
        title: opts.title ?? 'Are you sure?',
        message: opts.message,
        confirmLabel: opts.confirmLabel ?? 'Confirm',
        cancelLabel: opts.cancelLabel ?? 'Cancel',
        danger: !!opts.danger,
        resolve,
      }
    })
  }

  function answer(value: boolean) {
    if (!active.value) return
    active.value.resolve(value)
    active.value = null
  }

  return { active, confirm, answer }
}
