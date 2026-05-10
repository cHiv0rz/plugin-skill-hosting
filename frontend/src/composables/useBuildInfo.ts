import { ref } from 'vue'
import { api } from '../api'
import type { BackendBuildInfo } from '../types'
import { frontendBuildInfo } from '../build-info'

const backend = ref<BackendBuildInfo | null>(null)
let pending: Promise<BackendBuildInfo | null> | null = null

export function useBuildInfo() {
  function load() {
    if (backend.value || pending) return pending
    pending = api.version()
      .then(info => { backend.value = info; return info })
      .catch(() => null)
    return pending
  }
  return { frontend: frontendBuildInfo, backend, load }
}
