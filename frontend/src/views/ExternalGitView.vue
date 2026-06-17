<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, errMsg, errStatus } from '../api'
import ErrorAlert from '../components/ErrorAlert.vue'
import { useConfirm } from '../composables/useConfirm'
import type { ExternalSyncStatus, ReconcileReport } from '../types'

const { confirm } = useConfirm()

const status = ref<ExternalSyncStatus | null>(null)
const report = ref<ReconcileReport | null>(null)
const loading = ref(false)
const reconciling = ref(false)
const error = ref('')
// Set when the backend reports external sync isn't configured (503), so we can
// explain the env setup instead of showing a bare error.
const notConfigured = ref(false)

const drift = computed(() => {
  if (!status.value) return 0
  return status.value.missing.length + status.value.outOfDate.length + status.value.extra.length
})

// A browsable link for http(s) remotes (drop the trailing .git so it opens the
// repo page rather than the git endpoint). SSH/other remotes render as plain
// text. The URL is already credential-scrubbed by the backend.
const repoHref = computed(() => {
  const u = status.value?.remoteUrl ?? ''
  if (!/^https?:\/\//i.test(u)) return null
  return u.replace(/\.git$/, '')
})

async function check() {
  loading.value = true
  error.value = ''
  report.value = null
  notConfigured.value = false
  try {
    status.value = await api.externalGitStatus()
  } catch (e: unknown) {
    if (errStatus(e) === 503) {
      notConfigured.value = true
    } else {
      error.value = errMsg(e)
    }
    status.value = null
  } finally {
    loading.value = false
  }
}

async function reconcile() {
  if (!status.value) return
  const { missing, outOfDate, extra } = status.value
  const ok = await confirm({
    title: 'Bring mirror in sync?',
    message:
      `This will push ${missing.length + outOfDate.length} plugin(s) to the external git ` +
      `mirror and remove ${extra.length} stale plugin folder(s) from it. The marketplace ` +
      `database is the source of truth — any manual edits in the mirror will be overwritten.`,
    confirmLabel: 'Sync now',
    danger: extra.length > 0,
  })
  if (!ok) return
  reconciling.value = true
  error.value = ''
  try {
    report.value = await api.externalGitReconcile()
    // Re-check so the status reflects the post-reconcile state.
    await check()
  } catch (e: unknown) {
    error.value = errMsg(e)
  } finally {
    reconciling.value = false
  }
}

const reportErrors = computed(() =>
  report.value?.errors ? Object.entries(report.value.errors) : [],
)

onMounted(check)
</script>

<template>
  <h1>Git mirror</h1>

  <p class="muted intro">
    The marketplace mirrors every plugin to an external git repository on each
    create, update, and delete. Use this page to verify the mirror matches the
    database and to repair it if a push was missed.
  </p>

  <div v-if="notConfigured" class="card warn">
    External git sync is <strong>not configured</strong>. Set
    <code>EXTERNAL_GIT_REMOTE_URL</code> (and <code>EXTERNAL_GIT_TOKEN</code>) to
    enable the mirror.
  </div>

  <template v-else>
    <div v-if="status?.remoteUrl" class="card repo">
      <span class="repo__label">Remote</span>
      <a
        v-if="repoHref"
        :href="repoHref"
        target="_blank"
        rel="noopener noreferrer"
        class="repo__url"
      >{{ status.remoteUrl }}</a>
      <code v-else class="repo__url">{{ status.remoteUrl }}</code>
      <span v-if="status.branch" class="repo__branch">branch <code>{{ status.branch }}</code></span>
    </div>

    <div class="toolbar">
      <button type="button" class="secondary" :disabled="loading || reconciling" @click="check">
        {{ loading ? 'Checking…' : 'Re-check' }}
      </button>
      <button
        type="button"
        class="secondary"
        :disabled="loading || reconciling || !status || status.inSync"
        @click="reconcile"
      >
        {{ reconciling ? 'Syncing…' : 'Bring in sync' }}
      </button>
    </div>

    <ErrorAlert :message="error" />

    <div v-if="loading && !status" class="muted">Checking mirror against the database…</div>

    <template v-else-if="status">
      <div v-if="status.inSync" class="card ok">
        <strong>In sync.</strong> Every active plugin matches the external mirror.
      </div>

      <template v-else>
        <div class="card drift-summary">
          The mirror is <strong>out of sync</strong> — {{ drift }} difference(s) found.
          Use <em>Bring in sync</em> to reconcile.
        </div>

        <section v-if="status.missing.length" class="section">
          <h2 class="section-title">
            Missing from mirror
            <span class="chip chip--warn">{{ status.missing.length }}</span>
          </h2>
          <p class="muted hint">In the database but absent from the mirror — will be pushed.</p>
          <ul class="plugin-list">
            <li v-for="p in status.missing" :key="p"><code>{{ p }}</code></li>
          </ul>
        </section>

        <section v-if="status.outOfDate.length" class="section">
          <h2 class="section-title">
            Out of date
            <span class="chip chip--warn">{{ status.outOfDate.length }}</span>
          </h2>
          <p class="muted hint">Present in the mirror but content differs — will be re-pushed.</p>
          <ul class="plugin-list">
            <li v-for="p in status.outOfDate" :key="p"><code>{{ p }}</code></li>
          </ul>
        </section>

        <section v-if="status.extra.length" class="section">
          <h2 class="section-title">
            Stale in mirror
            <span class="chip chip--flag">{{ status.extra.length }}</span>
          </h2>
          <p class="muted hint">In the mirror with no matching active plugin — will be removed.</p>
          <ul class="plugin-list">
            <li v-for="p in status.extra" :key="p"><code>{{ p }}</code></li>
          </ul>
        </section>
      </template>
    </template>

    <section v-if="report" class="section">
      <h2 class="section-title">Last sync</h2>
      <div class="card">
        <p v-if="report.pushed.length" style="margin: 0 0 8px">
          Pushed: <code v-for="p in report.pushed" :key="p" class="inline-code">{{ p }}</code>
        </p>
        <p v-if="report.removed.length" style="margin: 0 0 8px">
          Removed: <code v-for="p in report.removed" :key="p" class="inline-code">{{ p }}</code>
        </p>
        <p v-if="!report.pushed.length && !report.removed.length && !reportErrors.length"
           class="muted" style="margin: 0">
          Nothing to do — mirror was already in sync.
        </p>
        <ul v-if="reportErrors.length" class="errlist">
          <li v-for="[name, msg] in reportErrors" :key="name">
            <code>{{ name }}</code> — <span class="muted">{{ msg }}</span>
          </li>
        </ul>
      </div>
    </section>
  </template>
</template>

<style scoped>
.intro {
  margin: -14px 0 20px;
  max-width: 70ch;
}
.warn {
  margin-bottom: 16px;
  border-color: rgb(var(--accent-rgb) / 0.5);
}
.ok {
  border-color: rgb(var(--success-rgb, var(--accent-rgb)) / 0.5);
}
.drift-summary {
  margin-bottom: 20px;
  border-color: rgb(var(--accent-rgb) / 0.5);
}
.repo {
  display: flex;
  align-items: baseline;
  flex-wrap: wrap;
  gap: 8px 14px;
  margin-bottom: 16px;
}
.repo__label {
  font-family: var(--mono);
  font-size: 10.5px;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--muted);
}
.repo__url {
  font-family: var(--mono);
  font-size: 13px;
  color: var(--text);
  word-break: break-all;
}
a.repo__url {
  border-bottom: 1px solid var(--accent);
  text-decoration: none;
  transition: color 0.12s ease;
}
a.repo__url:hover {
  color: var(--accent);
}
.repo__branch {
  font-family: var(--mono);
  font-size: 12px;
  color: var(--muted);
  margin-left: auto;
}
.repo__branch code {
  color: var(--text-soft);
}
.toolbar {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}
.section {
  margin-bottom: 28px;
}
.section-title {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  margin: 0 0 8px;
  font-family: var(--mono);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--text-soft);
}
.hint {
  margin: 0 0 10px;
  font-size: 13px;
}
.chip {
  display: inline-grid;
  place-items: center;
  min-width: 22px;
  padding: 0 8px;
  height: 20px;
  font-family: var(--mono);
  font-size: 10.5px;
  font-weight: 600;
  letter-spacing: 0.12em;
  color: var(--muted);
  border: 1px solid var(--border);
  border-radius: 999px;
}
.chip--warn {
  color: var(--accent);
  border-color: rgb(var(--accent-rgb) / 0.5);
}
.chip--flag {
  color: var(--sev-critical);
  border-color: rgb(var(--sev-critical-rgb) / 0.5);
}
.plugin-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.plugin-list code {
  font-family: var(--mono);
  font-size: 13px;
}
.inline-code {
  font-family: var(--mono);
  font-size: 12.5px;
  margin-right: 8px;
}
.errlist {
  margin: 8px 0 0;
  padding-left: 18px;
  font-size: 13px;
}
.errlist li {
  margin-bottom: 6px;
}
</style>
