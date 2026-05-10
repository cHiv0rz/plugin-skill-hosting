<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { RouterLink } from 'vue-router'
import { useBuildInfo } from '../composables/useBuildInfo'

const { frontend, backend, load } = useBuildInfo()

onMounted(() => { load() })

const year = new Date().getFullYear()

const frontendLine = computed(() =>
  `v${frontend.version} · ${frontend.gitCommit} · ${frontend.buildTime}`,
)
const backendLine = computed(() => {
  if (!backend.value) return '…'
  return `v${backend.value.version} · ${backend.value.gitCommit} · ${backend.value.buildTime}`
})
</script>

<template>
  <footer class="site-footer">
    <div class="footer-inner">
      <div class="footer-meta">
        <p class="copyright">
          © {{ year }} Oli Zimpasser ·
          <a href="https://github.com/oglimmer/plugin-skill-hosting/blob/master/LICENSE"
             target="_blank" rel="noopener">MIT License</a>
        </p>
        <nav class="footer-links">
          <RouterLink to="/developers">
            Developer Portal
            <svg class="link-icon" viewBox="0 0 16 16" aria-hidden="true" focusable="false">
              <path d="M6 3h7v7" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M13 3 6.5 9.5" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M11 9v3.5A1.5 1.5 0 0 1 9.5 14h-6A1.5 1.5 0 0 1 2 12.5v-6A1.5 1.5 0 0 1 3.5 5H7" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
          </RouterLink>
        </nav>
      </div>
      <dl class="footer-versions" aria-label="Build information">
        <div class="version-row">
          <dt>Frontend</dt>
          <dd>{{ frontendLine }}</dd>
        </div>
        <div class="version-row">
          <dt>Backend</dt>
          <dd>{{ backendLine }}</dd>
        </div>
      </dl>
    </div>
  </footer>
</template>

<style scoped>
.site-footer {
  margin-top: 64px;
  padding: 28px 32px 36px;
  border-top: 1px solid var(--border);
  background: rgba(10, 12, 16, 0.6);
  color: var(--text-soft);
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.04em;
}

.footer-inner {
  max-width: 1080px;
  margin: 0 auto;
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 32px;
  flex-wrap: wrap;
}

.footer-meta {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}
.copyright {
  margin: 0;
  font-size: 11.5px;
  color: var(--text-soft);
}
.copyright a {
  color: var(--text-soft);
  border-bottom: 1px solid var(--border);
  transition: color 0.2s ease, border-color 0.2s ease;
}
.copyright a:hover { color: var(--accent); border-bottom-color: var(--accent); }

.footer-links {
  display: flex;
  gap: 18px;
}
.footer-links a {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--text-soft);
  border-bottom: 1px solid transparent;
  padding-bottom: 2px;
  transition: color 0.2s ease, border-color 0.2s ease;
}
.footer-links .link-icon {
  width: 11px;
  height: 11px;
  flex-shrink: 0;
}
.footer-links a:hover {
  color: var(--text);
  border-bottom-color: var(--accent);
}

.footer-versions {
  margin: 0;
  display: grid;
  grid-template-columns: auto;
  gap: 4px;
  text-align: right;
  font-size: 10.5px;
  color: var(--muted);
  letter-spacing: 0.04em;
}
.version-row {
  display: flex;
  gap: 10px;
  justify-content: flex-end;
  align-items: baseline;
}
.version-row dt {
  font-weight: 600;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--text-soft);
  font-size: 10px;
  min-width: 64px;
  text-align: right;
}
.version-row dd {
  margin: 0;
  font-family: var(--mono);
  color: var(--muted);
  word-break: break-all;
}

@media (max-width: 720px) {
  .site-footer { padding: 24px 18px 32px; }
  .footer-inner { flex-direction: column; gap: 18px; }
  .footer-versions { text-align: left; }
  .version-row { justify-content: flex-start; }
  .version-row dt { text-align: left; min-width: 0; }
}
</style>
