<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, errMsg } from '../api'
import ErrorAlert from '../components/ErrorAlert.vue'
import type { UserSummary } from '../types'

const users = ref<UserSummary[]>([])
const loading = ref(true)
const error = ref('')

function fmt(d?: string | null) {
  if (!d) return ''
  return new Date(d).toLocaleString()
}

async function load() {
  loading.value = true
  error.value = ''
  try {
    users.value = await api.listUsers()
  } catch (e: unknown) {
    error.value = errMsg(e)
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <h1>Users</h1>

  <div v-if="loading" class="muted">Loading…</div>
  <ErrorAlert v-else-if="error" :message="error" />
  <div v-else-if="users.length === 0" class="card">
    <p class="muted" style="margin: 0">No users yet.</p>
  </div>
  <table v-else class="card" style="padding: 0">
    <thead>
      <tr>
        <th style="padding-left: 20px">Username</th>
        <th>Email</th>
        <th>Joined</th>
      </tr>
    </thead>
    <tbody>
      <tr v-for="u in users" :key="u.id">
        <td style="padding-left: 20px">{{ u.username }}</td>
        <td class="muted">{{ u.email }}</td>
        <td class="muted" style="white-space: nowrap">
          <small>{{ fmt(u.createdAt) }}</small>
        </td>
      </tr>
    </tbody>
  </table>
</template>
