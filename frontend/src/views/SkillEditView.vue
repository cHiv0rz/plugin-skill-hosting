<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import { api, type SkillVersion } from '../api'
import { useConfirm } from '../composables/useConfirm'

const { confirm } = useConfirm()

const props = defineProps<{
  pluginName: string
  skillName: string | null
}>()

const router = useRouter()
const isEdit = computed(() => !!props.skillName)
const name = ref('')
const description = ref('')
const body = ref(defaultBody())
const error = ref('')
const loading = ref(false)
const versions = ref<SkillVersion[]>([])
const versionsError = ref('')
const audit = ref<{
  createdByName?: string
  createdAt?: string
  updatedByName?: string
  updatedAt?: string
}>({})

function defaultBody() {
  return `## Instructions

Describe what the skill does, step by step.
`
}

function fmt(d?: string | null) {
  if (!d) return ''
  return new Date(d).toLocaleString()
}

async function load() {
  if (!isEdit.value) return
  try {
    const p = await api.getPlugin(props.pluginName)
    const s = p.skills?.find(s => s.name === props.skillName)
    if (!s) {
      error.value = 'skill not found'
      return
    }
    name.value = s.name
    description.value = s.description
    body.value = s.body
    audit.value = {
      createdByName: s.createdByName,
      createdAt: s.createdAt,
      updatedByName: s.updatedByName,
      updatedAt: s.updatedAt,
    }
    await loadVersions()
  } catch (e: any) {
    error.value = e.message
  }
}

async function loadVersions() {
  if (!props.skillName) return
  versionsError.value = ''
  try {
    versions.value = await api.skillVersions(props.pluginName, props.skillName)
  } catch (e: any) {
    versionsError.value = e.message
  }
}

async function submit() {
  error.value = ''
  loading.value = true
  try {
    if (isEdit.value) {
      await api.updateSkill(props.pluginName, props.skillName!, {
        description: description.value,
        body: body.value,
      })
    } else {
      await api.createSkill(props.pluginName, {
        name: name.value,
        description: description.value,
        body: body.value,
      })
    }
    router.push(`/plugins/${props.pluginName}`)
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function revert(version: number) {
  if (!props.skillName) return
  const ok = await confirm({
    title: `Revert to version ${version}`,
    message: 'This restores the description and body from that version and creates a new history entry. Continue?',
    confirmLabel: 'Revert',
  })
  if (!ok) return
  try {
    const s = await api.revertSkill(props.pluginName, props.skillName, version)
    description.value = s.description
    body.value = s.body
    audit.value = {
      createdByName: s.createdByName,
      createdAt: s.createdAt,
      updatedByName: s.updatedByName,
      updatedAt: s.updatedAt,
    }
    await loadVersions()
  } catch (e: any) {
    error.value = e.message
  }
}

onMounted(load)
</script>

<template>
  <h1>{{ isEdit ? `Edit skill: ${skillName}` : 'New skill' }}</h1>
  <div class="card">
    <form @submit.prevent="submit">
      <label>Skill name (slug, lowercase, [a-z0-9-])</label>
      <input
        v-model="name"
        :disabled="isEdit"
        required
        pattern="[a-z0-9][a-z0-9-]{1,62}[a-z0-9]"
      />

      <label>Description (used by Claude to decide when to use this skill)</label>
      <input v-model="description" required />

      <label>Body (Markdown — becomes the contents of SKILL.md after the frontmatter)</label>
      <textarea v-model="body" />

      <div v-if="error" class="error">{{ error }}</div>
      <div class="row" style="margin-top: 16px">
        <button type="submit" :disabled="loading">
          {{ loading ? 'Saving…' : isEdit ? 'Save' : 'Create skill' }}
        </button>
        <RouterLink :to="`/plugins/${pluginName}`" class="btn secondary">Cancel</RouterLink>
      </div>
    </form>
  </div>

  <div v-if="isEdit" class="card">
    <h2 style="margin-top: 0">Audit</h2>
    <table>
      <tbody>
        <tr>
          <th>Created</th>
          <td>{{ audit.createdByName || '—' }} · {{ fmt(audit.createdAt) }}</td>
        </tr>
        <tr>
          <th>Last edited</th>
          <td>{{ audit.updatedByName || '—' }} · {{ fmt(audit.updatedAt) }}</td>
        </tr>
      </tbody>
    </table>
  </div>

  <div v-if="isEdit" class="card">
    <h2 style="margin-top: 0">Edit history</h2>
    <p v-if="versionsError" class="error">{{ versionsError }}</p>
    <p v-else-if="versions.length === 0" class="muted">No history yet.</p>
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
              @click="revert(v.version)"
            >Revert</button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
