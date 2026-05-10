<script setup lang="ts">
import { ref, watch } from 'vue'
import { api, errMsg } from '../api'
import type { SkillVersion } from '../types'
import ErrorAlert from './ErrorAlert.vue'

const props = defineProps<{
  pluginName: string
  skillName: string | null
}>()

const emit = defineEmits<{
  revert: [version: number]
}>()

const versions = ref<SkillVersion[]>([])
const versionsError = ref('')

async function reload() {
  if (!props.skillName) {
    versions.value = []
    versionsError.value = ''
    return
  }
  versionsError.value = ''
  try {
    versions.value = await api.skillVersions(props.pluginName, props.skillName)
  } catch (e: unknown) {
    versionsError.value = errMsg(e)
  }
}

defineExpose({ reload })

watch(
  () => [props.pluginName, props.skillName] as const,
  reload,
  { immediate: true },
)

function fmt(d?: string | null) {
  if (!d) return ''
  return new Date(d).toLocaleString()
}
</script>

<template>
  <div class="card">
    <h2 style="margin-top: 0">Edit history</h2>
    <ErrorAlert :message="versionsError" />
    <p v-if="!versionsError && versions.length === 0" class="muted">No history yet.</p>
    <table v-else>
      <thead>
        <tr>
          <th>Version</th>
          <th>Action</th>
          <th>By</th>
          <th>When</th>
          <th>Description</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="v in versions" :key="v.id">
          <td>{{ v.version }}</td>
          <td><span class="badge">{{ v.action }}</span></td>
          <td>{{ v.editedByName || '—' }}</td>
          <td class="muted" style="white-space: nowrap">{{ fmt(v.editedAt) }}</td>
          <td>{{ v.description }}</td>
          <td style="text-align: right">
            <button
              v-if="v.action !== 'delete'"
              class="secondary"
              type="button"
              @click="emit('revert', v.version)"
            >Revert</button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
