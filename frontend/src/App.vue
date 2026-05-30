<script setup lang="ts">
import { computed, onMounted } from 'vue'
import NavBar from './components/NavBar.vue'
import SiteFooter from './components/Footer.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'
import PromptDialog from './components/PromptDialog.vue'
import { useAuthStore } from './stores/auth'
import { useRoute, useRouter } from 'vue-router'
import { errStatus } from './api'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const hideChrome = computed(() => Boolean(route.meta?.hideChrome))

onMounted(() => {
  if (auth.token && !auth.user?.apiToken) {
    auth.refreshUser().catch((e: unknown) => {
      // A 401 here means the stored token is expired, revoked (e.g. the user
      // hit "sign out everywhere" on another device), or otherwise invalid —
      // clear it and, if this page needs auth, route to login. Other failures
      // (offline, 5xx) are left for the route guards / views to surface.
      if (errStatus(e) === 401) {
        auth.logout()
        if (route.meta?.requiresAuth) {
          router.replace({ path: '/login', query: { redirect: route.fullPath } })
        }
      }
    })
  }
})
</script>

<template>
  <template v-if="hideChrome">
    <RouterView />
  </template>
  <template v-else>
    <div class="app-shell">
      <NavBar />
      <main>
        <RouterView />
      </main>
      <SiteFooter />
    </div>
  </template>
  <ConfirmDialog />
  <PromptDialog />
</template>

<style scoped>
.app-shell {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
}
.app-shell main {
  flex: 1;
}
</style>
