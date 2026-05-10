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
  if (auth.token && !auth.user?.apiToken) {
    auth.refreshUser().catch(() => { /* token may be invalid; let route guards handle */ })
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
