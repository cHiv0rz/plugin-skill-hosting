<script setup lang="ts">
const codes = [
  { code: '400', label: 'Bad Request',  desc: 'Malformed JSON, invalid slug, body too large, or an explicit validation rule failed.' },
  { code: '401', label: 'Unauthorized', desc: 'Missing or invalid credential. The marketplace and git endpoints add a WWW-Authenticate header so git and curl prompt for credentials.' },
  { code: '403', label: 'Forbidden',    desc: 'Authenticated, but the resource belongs to another user.' },
  { code: '404', label: 'Not Found',    desc: "Plugin, skill, file, or version doesn't exist (or is soft-deleted on a read path)." },
  { code: '409', label: 'Conflict',     desc: 'Unique-key violation: a plugin or skill with that name already exists.' },
  { code: '500', label: 'Server Error', desc: 'Database or server error. Check server logs.' },
  { code: '502', label: 'Bad Gateway',  desc: 'An upstream call (Claude API) failed.' },
]
</script>

<template>
  <section class="dev-section">
    <header class="section-head">
      <h2>Errors</h2>
      <p class="section-lede">
        Errors return an appropriate HTTP status and a JSON body of the form:
      </p>
    </header>

<pre>{ "error": "human-readable message" }</pre>

    <div class="error-grid">
      <div v-for="e in codes" :key="e.code" class="error-row" :class="`error-row--${e.code[0]}xx`">
        <div class="error-code">
          <code>{{ e.code }}</code>
          <span>{{ e.label }}</span>
        </div>
        <p>{{ e.desc }}</p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.error-grid { margin-top: 16px; display: flex; flex-direction: column; gap: 8px; }
.error-row {
  display: grid;
  grid-template-columns: 180px 1fr;
  gap: 16px;
  align-items: start;
  border: 1px solid var(--border-soft);
  border-left-width: 3px;
  padding: 10px 14px;
  background: var(--bg-2);
}
.error-row--4xx { border-left-color: var(--accent); }
.error-row--5xx { border-left-color: var(--rust, #d65a31); }
.error-code { display: flex; flex-direction: column; gap: 2px; }
.error-code code {
  font-family: var(--mono);
  font-size: 18px;
  color: var(--text);
}
.error-code span {
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--text-soft);
}
.error-row p { margin: 4px 0 0; color: var(--text-soft); }

@media (max-width: 600px) {
  .error-row { grid-template-columns: 1fr; }
}
</style>
