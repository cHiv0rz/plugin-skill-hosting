<script setup lang="ts">
import { computed, onMounted, onBeforeUnmount, ref, watch } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter, RouterLink } from 'vue-router'
import Breadcrumbs from './Breadcrumbs.vue'
import ThemeSwitcher from './ThemeSwitcher.vue'

const auth = useAuthStore()
const router = useRouter()

onMounted(() => { auth.ensureMode() })

const isApproved = computed(() => auth.user?.status === 'approved')

// The admin tools (audit, git mirror, user management) are available to any
// admin in every auth mode — even OIDC+hd, where new members are auto-admitted
// but an admin still needs the list to promote/demote other admins. Non-admins
// never see them, so they get a plain static identity label instead of the
// dropdown trigger.
const canManageUsers = computed(() => !!auth.user?.isAdmin)

// Admin dropdown: clicking the username reveals the admin destinations rather
// than crowding the nav bar with a link each.
const menuOpen = ref(false)
const menuRef = ref<HTMLElement | null>(null)

function closeMenu() { menuOpen.value = false }

function onDocPointerDown(e: MouseEvent) {
  if (menuOpen.value && menuRef.value && !menuRef.value.contains(e.target as Node)) {
    closeMenu()
  }
}
function onKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') closeMenu()
}

onMounted(() => {
  document.addEventListener('mousedown', onDocPointerDown)
  document.addEventListener('keydown', onKeydown)
})
onBeforeUnmount(() => {
  document.removeEventListener('mousedown', onDocPointerDown)
  document.removeEventListener('keydown', onKeydown)
})
// Close on navigation so the panel never lingers over a freshly loaded page.
watch(() => router.currentRoute.value.fullPath, closeMenu)

function logout() {
  if (auth.doLogout()) return // full-page redirect already in flight
  router.push('/login')
}
</script>

<template>
  <nav class="top">
    <Breadcrumbs />
    <div class="links">
      <ThemeSwitcher />
      <template v-if="auth.user">
        <template v-if="isApproved">
          <RouterLink to="/plugins/new" class="btn">+ New plugin</RouterLink>
          <div v-if="canManageUsers" ref="menuRef" class="user-menu">
            <button
              type="button"
              class="user-link user-menu__trigger"
              :class="{ 'user-menu__trigger--open': menuOpen }"
              aria-haspopup="true"
              :aria-expanded="menuOpen"
              :title="`Admin menu — signed in as ${auth.user.username}`"
              @click="menuOpen = !menuOpen"
            >
              <span class="user-link-at" aria-hidden="true">@</span>{{ auth.user.username }}
              <span class="user-menu__chev" aria-hidden="true">▾</span>
            </button>
            <Transition name="navmenu">
              <div v-if="menuOpen" class="user-menu__panel" role="menu">
                <RouterLink to="/audit" class="user-menu__item" role="menuitem" @click="closeMenu">
                  Security audit
                </RouterLink>
                <RouterLink to="/external-git" class="user-menu__item" role="menuitem" @click="closeMenu">
                  Git mirror
                </RouterLink>
                <RouterLink to="/users" class="user-menu__item" role="menuitem" @click="closeMenu">
                  Users
                </RouterLink>
              </div>
            </Transition>
          </div>
          <span
            v-else
            class="user-link user-link--static"
            :title="`Signed in as ${auth.user.username}`"
          >
            <span class="user-link-at" aria-hidden="true">@</span>{{ auth.user.username }}
          </span>
        </template>
        <template v-else>
          <span class="user pending-chip" :title="`Signed in as ${auth.user.username} — awaiting approval`">
            <span class="pending-dot" aria-hidden="true"></span>
            {{ auth.user.username }} · pending
          </span>
        </template>
        <button class="secondary" @click="logout">Log out</button>
      </template>
      <template v-else>
        <RouterLink to="/login">Log in</RouterLink>
        <RouterLink v-if="auth.mode !== 'oidc'" to="/register" class="btn">Sign up</RouterLink>
      </template>
    </div>
  </nav>
</template>
