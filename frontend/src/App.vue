<script setup lang="ts">
import { computed, onMounted } from 'vue'
import NavBar from './components/NavBar.vue'
import SiteFooter from './components/Footer.vue'
import ConfirmDialog from './components/ConfirmDialog.vue'
import PromptDialog from './components/PromptDialog.vue'
import { useAuthStore } from './stores/auth'
import { useRoute } from 'vue-router'

const auth = useAuthStore()
const route = useRoute()
const hideChrome = computed(() => Boolean(route.meta?.hideChrome))

onMounted(() => {
  // Refresh /api/me once at startup so the navbar and any public (non-guarded)
  // view reflect the current server-side user instead of stale localStorage.
  // Protected routes are already refreshed by the route guard (which awaits the
  // same shared promise) before they render, and a 401 clears the session.
  auth.ensureFreshUser()
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
