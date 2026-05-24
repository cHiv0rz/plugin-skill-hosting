<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter, RouterLink } from 'vue-router'
import Breadcrumbs from './Breadcrumbs.vue'

const auth = useAuthStore()
const router = useRouter()

onMounted(() => { auth.ensureMode() })

const isApproved = computed(() => auth.user?.status === 'approved')

// User-management UI is hidden in OIDC+hd mode (the workspace filter handles
// membership) and for non-admin users (only admins can list/approve/reject).
const canManageUsers = computed(() =>
  !!auth.user?.isAdmin && (auth.mode !== 'oidc' || auth.userApprovalRequired)
)

function logout() {
  if (auth.doLogout()) return // full-page redirect already in flight
  router.push('/login')
}
</script>

<template>
  <nav class="top">
    <Breadcrumbs />
    <div class="links">
      <template v-if="auth.user">
        <template v-if="isApproved">
          <RouterLink to="/plugins/new" class="btn">+ New plugin</RouterLink>
          <RouterLink
            v-if="canManageUsers"
            to="/users"
            class="user-link"
            :title="`Manage users — signed in as ${auth.user.username}`"
          >
            <span class="user-link-at" aria-hidden="true">@</span>{{ auth.user.username }}
          </RouterLink>
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
