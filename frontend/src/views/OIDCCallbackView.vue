<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter } from 'vue-router'
import { errMsg } from '../api'
import type { User } from '../types'
import ErrorAlert from '../components/ErrorAlert.vue'

const error = ref('')
const auth = useAuthStore()
const router = useRouter()

function decodeUser(b64: string): User {
  const norm = b64.replace(/-/g, '+').replace(/_/g, '/')
  const pad = norm.length % 4 === 0 ? norm : norm + '='.repeat(4 - (norm.length % 4))
  return JSON.parse(atob(pad))
}

onMounted(() => {
  const hash = window.location.hash.startsWith('#')
    ? window.location.hash.slice(1)
    : window.location.hash
  const params = new URLSearchParams(hash)

  const errParam = params.get('error')
  if (errParam) {
    error.value = errParam
    return
  }
  const token = params.get('token')
  const userParam = params.get('user')
  if (!token || !userParam) {
    error.value = 'missing token or user'
    return
  }
  try {
    auth.setSession(token, decodeUser(userParam))
  } catch (e: unknown) {
    error.value = 'failed to parse session: ' + errMsg(e)
    return
  }
  // Wipe the hash from the URL before navigating away.
  history.replaceState(null, '', window.location.pathname)
  router.replace('/')
})
</script>

<template>
  <div class="card" style="max-width: 420px">
    <template v-if="error">
      <h1>Sign-in failed</h1>
      <ErrorAlert :message="error" />
      <p style="margin-top: 16px">
        <RouterLink to="/login">Try again</RouterLink>
      </p>
    </template>
    <template v-else>
      <p class="muted">Signing you in…</p>
    </template>
  </div>
</template>
