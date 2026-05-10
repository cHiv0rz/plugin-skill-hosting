import { computed } from 'vue'
import { useAuthStore } from '../../stores/auth'

export function useApiExamples() {
  const auth = useAuthStore()
  const origin = computed(() => window.location.origin)
  const apiToken = computed(() => auth.user?.apiToken ?? '')
  const exampleToken = computed(() => apiToken.value || '<your-api-token>')
  const hostNoScheme = computed(() => origin.value.replace(/^https?:\/\//, ''))
  const authedOrigin = computed(() => {
    const t = apiToken.value || 'TOKEN'
    return origin.value.replace(/^(https?:\/\/)/, `$1_:${t}@`)
  })
  return { origin, apiToken, exampleToken, hostNoScheme, authedOrigin }
}
