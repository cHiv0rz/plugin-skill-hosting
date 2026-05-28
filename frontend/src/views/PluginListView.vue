<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { errMsg } from '../api'
import ErrorAlert from '../components/ErrorAlert.vue'
import { useAuthStore } from '../stores/auth'
import { usePluginStore } from '../stores/plugins'
import { RouterLink } from 'vue-router'
import { useConfirm } from '../composables/useConfirm'
import { storeToRefs } from 'pinia'

const { confirm } = useConfirm()

const auth = useAuthStore()
const pluginStore = usePluginStore()
const { list: plugins, deleted: deletedPlugins } = storeToRefs(pluginStore)
const loading = ref(true)
const error = ref('')
const tokenError = ref('')
const regenerating = ref(false)
const copied = ref('')
const activeTab = ref<'plugins' | 'connect'>('plugins')

function fmt(d?: string | null) {
  if (!d) return ''
  return new Date(d).toLocaleString()
}

const apiToken = computed(() => auth.user?.apiToken ?? '')

const authedOrigin = computed(() => {
  if (!apiToken.value) return window.location.origin
  const u = new URL(window.location.origin)
  u.username = '_'
  u.password = apiToken.value
  // URL.toString() appends a trailing "/", strip it for clean joining.
  return u.toString().replace(/\/$/, '')
})

const marketplaceUrl = computed(() => `${authedOrigin.value}/marketplace.json`)
const marketplaceCmd = computed(() => `/plugin marketplace add ${marketplaceUrl.value}`)

const mcpUrl = computed(() => `${window.location.origin}/mcp`)
const mcpServerName = computed(() => auth.marketplaceName || 'skill-host')
const mcpAddCmd = computed(() =>
  `claude mcp add --transport http ${mcpServerName.value} ${mcpUrl.value} -H "Authorization: Bearer ${apiToken.value}"`
)
const mcpJsonConfig = computed(() => JSON.stringify({
  mcpServers: {
    [mcpServerName.value]: {
      type: 'http',
      url: mcpUrl.value,
      headers: { Authorization: `Bearer ${apiToken.value}` },
    },
  },
}, null, 2))

let initialLoad = true
async function load() {
  loading.value = true
  error.value = ''
  try {
    await Promise.all([
      pluginStore.loadList(),
      auth.user ? pluginStore.loadDeleted() : Promise.resolve(),
    ])
    if (initialLoad && plugins.value.length === 0) {
      activeTab.value = 'connect'
    }
    initialLoad = false
  } catch (e: unknown) {
    error.value = errMsg(e)
  } finally {
    loading.value = false
  }
}

async function restorePlugin(name: string) {
  try {
    await pluginStore.restorePlugin(name)
  } catch (e: unknown) {
    error.value = errMsg(e)
  }
}

async function regenerate() {
  const ok = await confirm({
    title: 'Regenerate API token',
    message: 'Existing marketplace links will stop working until you update them. Continue?',
    confirmLabel: 'Regenerate',
    danger: true,
  })
  if (!ok) return
  tokenError.value = ''
  regenerating.value = true
  try {
    await auth.regenerateToken()
  } catch (e: unknown) {
    tokenError.value = errMsg(e)
  } finally {
    regenerating.value = false
  }
}

async function copy(text: string, label: string) {
  try {
    await navigator.clipboard.writeText(text)
    copied.value = label
    setTimeout(() => { if (copied.value === label) copied.value = '' }, 1500)
  } catch {}
}

onMounted(load)
</script>

<template>
  <div class="pl">
    <!-- Tabs (no italic title — breadcrumbs handle naming) -->
    <nav class="pl-tabs" role="tablist">
      <button
        type="button"
        class="pl-tab"
        role="tab"
        :class="{ 'pl-tab--active': activeTab === 'plugins' }"
        :aria-selected="activeTab === 'plugins'"
        @click="activeTab = 'plugins'"
      >
        plugins
        <span class="pl-tab__count">[{{ plugins.length }}]</span>
      </button>
      <button
        type="button"
        class="pl-tab"
        role="tab"
        :class="{ 'pl-tab--active': activeTab === 'connect' }"
        :aria-selected="activeTab === 'connect'"
        @click="activeTab = 'connect'"
      >connect</button>
    </nav>

    <!-- PLUGINS tab -->
    <section v-show="activeTab === 'plugins'" role="tabpanel">
      <p v-if="loading" class="pl-loading">loading…</p>
      <ErrorAlert v-else-if="error" :message="error" />

      <div v-else-if="plugins.length === 0" class="pl-empty">
        <p class="pl-empty__line">
          <span class="pl-empty__prompt">$</span>
          no plugins yet
        </p>
        <div class="pl-empty__actions">
          <RouterLink to="/plugins/new" class="pl-btn pl-btn--primary">+ create your first plugin</RouterLink>
          <button type="button" class="pl-btn" @click="activeTab = 'connect'">connect to claude code →</button>
        </div>
      </div>

      <table v-else class="pl-table">
        <thead>
          <tr>
            <th>name</th>
            <th>description</th>
            <th>owner</th>
            <th>ver</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="p in plugins" :key="p.id">
            <td class="pl-table__name">
              <RouterLink :to="`/plugins/${p.name}`">{{ p.name }}</RouterLink>
            </td>
            <td class="pl-table__desc">{{ p.description }}</td>
            <td class="pl-table__owner">{{ p.ownerName }}</td>
            <td class="pl-table__ver"><span class="pl-ver">{{ p.version }}</span></td>
          </tr>
        </tbody>
      </table>

      <details v-if="deletedPlugins.length > 0" class="pl-disclosure">
        <summary class="pl-disclosure__head">
          <span class="pl-disclosure__toggle" aria-hidden="true"></span>
          <span class="pl-disclosure__title">deleted plugins</span>
          <span class="pl-disclosure__count">{{ deletedPlugins.length }}</span>
          <span class="spacer"></span>
          <span class="pl-disclosure__hint" aria-hidden="true">
            <span class="pl-disclosure__hint-open">expand</span>
            <span class="pl-disclosure__hint-close">collapse</span>
            <span class="pl-disclosure__chev">▸</span>
          </span>
        </summary>
        <p class="pl-disclosure__note">
          soft-deleted · restore to put back in the marketplace.
        </p>
        <table class="pl-table pl-table--nested">
          <thead>
            <tr>
              <th>name</th>
              <th>description</th>
              <th>deleted</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="p in deletedPlugins" :key="p.id">
              <td>{{ p.name }}</td>
              <td>{{ p.description }}</td>
              <td class="pl-table__when">{{ fmt(p.deletedAt) }}</td>
              <td class="pl-table__act">
                <button type="button" class="pl-btn" @click="restorePlugin(p.name)">restore</button>
              </td>
            </tr>
          </tbody>
        </table>
      </details>
    </section>

    <!-- CONNECT tab -->
    <section v-show="activeTab === 'connect'" role="tabpanel">
      <div class="pl-block">
        <header class="pl-block__head">
          <span class="pl-block__title">marketplace install</span>
        </header>
        <p class="pl-block__body">
          the command below contains your personal API token. keep it secret.
        </p>
        <div class="pl-code">
          <pre>{{ marketplaceCmd }}</pre>
        </div>
        <div class="pl-code-actions">
          <button type="button" class="pl-btn" @click="copy(marketplaceCmd, 'cmd')">
            {{ copied === 'cmd' ? '✓ copied' : 'copy command' }}
          </button>
          <button type="button" class="pl-btn" @click="copy(marketplaceUrl, 'url')">
            {{ copied === 'url' ? '✓ copied' : 'copy url' }}
          </button>
        </div>

        <details class="pl-disclosure pl-disclosure--inset">
          <summary class="pl-disclosure__head">
            <span class="pl-disclosure__toggle" aria-hidden="true"></span>
            <span class="pl-disclosure__title">raw api token</span>
            <span class="spacer"></span>
            <span class="pl-disclosure__hint" aria-hidden="true">
              <span class="pl-disclosure__hint-open">expand</span>
              <span class="pl-disclosure__hint-close">collapse</span>
              <span class="pl-disclosure__chev">▸</span>
            </span>
          </summary>
          <div class="pl-token">
            <input
              type="text"
              :value="apiToken"
              readonly
              class="pl-token__input"
            />
            <button type="button" class="pl-btn" @click="copy(apiToken, 'token')">
              {{ copied === 'token' ? '✓ copied' : 'copy' }}
            </button>
            <button type="button" class="pl-btn pl-btn--danger" :disabled="regenerating" @click="regenerate">
              {{ regenerating ? 'regenerating…' : 'regenerate' }}
            </button>
          </div>
          <ErrorAlert :message="tokenError" />
        </details>
      </div>

      <div class="pl-block">
        <header class="pl-block__head">
          <span class="pl-block__title">mcp server</span>
        </header>
        <p class="pl-block__body">
          lets claude (or any MCP-aware client) read plugins and create / modify skills directly.
          plugins are read-only over MCP — nothing can be deleted.
        </p>
        <div class="pl-tools">
          <span class="pl-tools__label">tools</span>
          <code>list_plugins</code>
          <code>get_plugin</code>
          <code>get_skill</code>
          <code>create_skill</code>
          <code>update_skill</code>
          <code>list_skill_files</code>
          <code>get_skill_file</code>
          <code>upsert_skill_file</code>
        </div>

        <p class="pl-block__sub">claude code · one-line install</p>
        <div class="pl-code">
          <pre>{{ mcpAddCmd }}</pre>
        </div>
        <div class="pl-code-actions">
          <button type="button" class="pl-btn" @click="copy(mcpAddCmd, 'mcpCmd')">
            {{ copied === 'mcpCmd' ? '✓ copied' : 'copy command' }}
          </button>
          <button type="button" class="pl-btn" @click="copy(mcpUrl, 'mcpUrl')">
            {{ copied === 'mcpUrl' ? '✓ copied' : 'copy url' }}
          </button>
        </div>

        <details class="pl-disclosure pl-disclosure--inset">
          <summary class="pl-disclosure__head">
            <span class="pl-disclosure__toggle" aria-hidden="true"></span>
            <span class="pl-disclosure__title">json config</span>
            <span class="pl-disclosure__count">claude desktop & other MCP clients</span>
            <span class="spacer"></span>
            <span class="pl-disclosure__hint" aria-hidden="true">
              <span class="pl-disclosure__hint-open">expand</span>
              <span class="pl-disclosure__hint-close">collapse</span>
              <span class="pl-disclosure__chev">▸</span>
            </span>
          </summary>
          <p class="pl-block__body pl-block__body--inset">
            paste under <code>mcpServers</code> in your client's MCP config.
          </p>
          <div class="pl-code">
            <pre>{{ mcpJsonConfig }}</pre>
          </div>
          <div class="pl-code-actions">
            <button type="button" class="pl-btn" @click="copy(mcpJsonConfig, 'mcpJson')">
              {{ copied === 'mcpJson' ? '✓ copied' : 'copy json' }}
            </button>
          </div>
        </details>
      </div>
    </section>
  </div>
</template>

<style scoped>
.pl {
  margin-top: -16px;
}

/* ─── Tabs ─────────────────────────────────────────────────────── */
.pl-tabs {
  display: flex;
  gap: 0;
  margin: 0 0 24px;
  border-bottom: 1px solid var(--border);
}
.pl-tab {
  background: transparent;
  color: var(--text-soft);
  border: 0;
  border-bottom: 2px solid transparent;
  border-radius: 0;
  padding: 12px 18px;
  margin-bottom: -1px;
  font-family: var(--mono);
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 0.02em;
  text-transform: none;
  line-height: 1.4;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  transition: color 0.12s ease, border-color 0.12s ease;
}
.pl-tab::before { display: none; content: none; }
.pl-tab:hover { color: var(--text); transform: none; background: transparent; }
.pl-tab--active {
  color: var(--text);
  border-bottom-color: var(--accent);
}
.pl-tab__count {
  font-size: 10.5px;
  color: var(--muted);
  letter-spacing: 0;
}

.pl-loading {
  font-family: var(--mono);
  font-size: 12.5px;
  color: var(--muted);
  margin: 0;
}

/* ─── Empty state ──────────────────────────────────────────────── */
.pl-empty {
  padding: 24px;
  border: 1px dashed var(--border);
  background: var(--bg-2);
}
.pl-empty__line {
  margin: 0 0 16px;
  font-family: var(--mono);
  font-size: 14px;
  color: var(--text);
  letter-spacing: 0.02em;
}
.pl-empty__prompt {
  color: var(--accent);
  margin-right: 8px;
  font-weight: 700;
}
.pl-empty__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

/* ─── Flat buttons (matches skill edit) ────────────────────────── */
.pl-btn {
  background: transparent;
  color: var(--text);
  border: 1px solid var(--border);
  border-radius: 0;
  padding: 6px 12px;
  margin: 0;
  font-family: var(--mono);
  font-size: 11.5px;
  font-weight: 500;
  letter-spacing: 0.02em;
  text-transform: lowercase;
  line-height: 1.5;
  cursor: pointer;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: border-color 0.12s ease, color 0.12s ease, background 0.12s ease;
}
.pl-btn::before { display: none; content: none; }
.pl-btn:hover {
  background: transparent;
  color: var(--accent);
  border-color: var(--accent);
  transform: none;
}
.pl-btn:active { transform: none; }
.pl-btn:disabled,
.pl-btn:disabled:hover {
  opacity: 0.35;
  cursor: not-allowed;
  color: var(--text-soft);
  border-color: var(--border);
}
.pl-btn--primary {
  color: var(--text);
  background: var(--accent);
  border-color: var(--accent);
  font-weight: 700;
}
.pl-btn--primary:hover {
  color: var(--bg);
  background: var(--text);
  border-color: var(--text);
}
.pl-btn--danger {
  color: var(--rust);
  border-color: rgba(214, 90, 49, 0.5);
}
.pl-btn--danger:hover {
  color: var(--text);
  background: var(--rust);
  border-color: var(--rust);
}

/* ─── Plugin table ─────────────────────────────────────────────── */
.pl-table {
  width: 100%;
  border-collapse: collapse;
  border: 1px solid var(--border);
  background: var(--bg-2);
  margin: 0 0 24px;
  font-family: var(--mono);
}
.pl-table th {
  text-align: left;
  padding: 9px 14px;
  font-family: var(--mono);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--muted);
  border-bottom: 1px solid var(--border);
  background: var(--bg);
}
.pl-table td {
  padding: 11px 14px;
  border-bottom: 1px solid var(--border-soft);
  font-size: 13px;
  color: var(--text);
  vertical-align: top;
}
.pl-table tbody tr:last-child td { border-bottom: 0; }
.pl-table tbody tr {
  transition: background 0.12s ease;
}
.pl-table tbody tr:hover {
  background: rgba(245, 165, 36, 0.04);
}
.pl-table__name a {
  color: var(--text);
  border-bottom: 1px solid var(--accent);
  padding-bottom: 1px;
  font-weight: 600;
  transition: color 0.12s ease, border-color 0.12s ease;
}
.pl-table__name a:hover {
  color: var(--accent);
}
.pl-table__desc { color: var(--text-soft); }
.pl-table__owner {
  color: var(--muted);
  font-size: 12px;
  white-space: nowrap;
}
.pl-table__ver { width: 1%; white-space: nowrap; }
.pl-table__when {
  color: var(--muted);
  font-size: 11.5px;
  white-space: nowrap;
}
.pl-table__act { text-align: right; width: 1%; white-space: nowrap; }

.pl-ver {
  display: inline-block;
  padding: 2px 9px;
  font-family: var(--mono);
  font-size: 10.5px;
  font-weight: 600;
  letter-spacing: 0.06em;
  color: var(--accent-2);
  border: 1px solid var(--border);
  background: var(--bg);
}

.pl-table--nested {
  margin: 0;
}

/* ─── Connect blocks ───────────────────────────────────────────── */
.pl-block {
  margin: 0 0 28px;
  padding: 0 0 0 16px;
  border-left: 2px solid var(--border);
}
.pl-block__head {
  margin-bottom: 8px;
}
.pl-block__title {
  font-family: var(--mono);
  font-size: 10.5px;
  font-weight: 700;
  letter-spacing: 0.28em;
  text-transform: uppercase;
  color: var(--accent);
}
.pl-block__body {
  margin: 0 0 12px;
  font-size: 12.5px;
  color: var(--text-soft);
  line-height: 1.55;
}
.pl-block__body code {
  font-size: 11.5px;
}
.pl-block__body--inset { margin: 10px 0 12px; }
.pl-block__sub {
  margin: 16px 0 6px;
  font-family: var(--mono);
  font-size: 10.5px;
  font-weight: 600;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--text-soft);
}

.pl-tools {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
  margin: 0 0 16px;
  padding: 10px 12px;
  border: 1px solid var(--border);
  background: var(--bg-2);
}
.pl-tools__label {
  font-family: var(--mono);
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--muted);
  margin-right: 4px;
}
.pl-tools code {
  font-size: 11.5px;
  padding: 2px 7px;
}

/* ─── Code block ───────────────────────────────────────────────── */
.pl-code {
  margin: 0 0 8px;
}
.pl-code pre {
  margin: 0;
  padding: 12px 14px 12px 22px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-left: 2px solid var(--accent);
  border-radius: 0;
  font-family: var(--mono);
  font-size: 12.5px;
  line-height: 1.55;
  color: var(--text);
  white-space: pre-wrap;
  word-break: break-all;
}
.pl-code pre::before { content: none; }
.pl-code-actions {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

/* ─── Token row ────────────────────────────────────────────────── */
.pl-token {
  display: flex;
  gap: 6px;
  align-items: stretch;
  flex-wrap: wrap;
  margin: 10px 0 4px;
}
.pl-token__input {
  flex: 1 1 280px;
  min-width: 0;
  background: var(--bg);
  color: var(--text);
  border: 1px solid var(--border);
  border-bottom: 1px solid var(--border);
  border-radius: 0;
  padding: 6px 10px;
  font-family: var(--mono);
  font-size: 12px;
  letter-spacing: 0;
  outline: none;
}
.pl-token__input:focus { border-color: var(--accent); }

/* ─── Disclosure bar (matches skill edit audit/history) ────────── */
.pl-disclosure {
  margin-top: 14px;
}
.pl-disclosure__head {
  cursor: pointer;
  list-style: none;
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 9px 12px;
  border: 1px solid var(--border);
  background: var(--bg-2);
  font-family: var(--mono);
  font-size: 10.5px;
  font-weight: 700;
  letter-spacing: 0.28em;
  text-transform: uppercase;
  color: var(--text-soft);
  transition: color 0.15s ease, border-color 0.15s ease, background 0.15s ease;
  user-select: none;
}
.pl-disclosure__head::-webkit-details-marker { display: none; }
.pl-disclosure__toggle {
  display: inline-grid;
  place-items: center;
  width: 18px;
  height: 18px;
  border: 1px solid var(--border);
  color: var(--accent);
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0;
  line-height: 1;
  flex: 0 0 auto;
  transition: border-color 0.15s ease;
}
.pl-disclosure:not([open]) > .pl-disclosure__head .pl-disclosure__toggle::before { content: '+'; }
.pl-disclosure[open] > .pl-disclosure__head .pl-disclosure__toggle::before { content: '−'; }
.pl-disclosure__title { letter-spacing: inherit; flex: 0 0 auto; }
.pl-disclosure__count {
  font-family: var(--mono);
  font-size: 10.5px;
  letter-spacing: 0.08em;
  text-transform: lowercase;
  color: var(--muted);
  font-weight: 500;
}
.pl-disclosure__hint {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  font-family: var(--mono);
  font-size: 10px;
  font-weight: 500;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--muted);
}
.pl-disclosure__hint-open,
.pl-disclosure__hint-close { display: none; }
.pl-disclosure:not([open]) > .pl-disclosure__head .pl-disclosure__hint-open { display: inline; }
.pl-disclosure[open] > .pl-disclosure__head .pl-disclosure__hint-close { display: inline; }
.pl-disclosure__chev {
  display: inline-block;
  color: var(--accent);
  font-size: 12px;
  transition: transform 0.18s ease;
  letter-spacing: 0;
}
.pl-disclosure[open] > .pl-disclosure__head .pl-disclosure__chev { transform: rotate(90deg); }
.pl-disclosure__head:hover {
  color: var(--text);
  border-color: var(--accent);
  background: rgba(245, 165, 36, 0.04);
}
.pl-disclosure__head:hover .pl-disclosure__toggle { border-color: var(--accent); }
.pl-disclosure__head:hover .pl-disclosure__hint { color: var(--text-soft); }
.pl-disclosure[open] > .pl-disclosure__head {
  color: var(--text);
  border-bottom-color: var(--accent);
  margin-bottom: 12px;
}
.pl-disclosure__note {
  margin: 0 0 10px;
  font-size: 11.5px;
  color: var(--muted);
}
.pl-disclosure--inset {
  margin-top: 18px;
}

@media (max-width: 720px) {
  .pl-tab { padding: 10px 12px; }
  .pl-block { padding-left: 12px; }
  .pl-code pre { padding: 10px 12px 10px 18px; font-size: 12px; }
}
</style>
