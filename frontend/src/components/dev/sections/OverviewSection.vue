<script setup lang="ts">
import { useApiExamples } from '../useApiExamples'
const { origin } = useApiExamples()
</script>

<template>
  <section class="dev-section">
    <header class="section-head">
      <h2>Overview</h2>
      <p class="section-lede">
        Four surfaces talk to the marketplace — pick the one that fits your client.
      </p>
    </header>

    <div class="surface-grid">
      <div class="surface">
        <h3>REST API</h3>
        <code>/api/*</code>
        <p>Used by the web UI and any programmatic client. JSON in, JSON out.</p>
      </div>
      <div class="surface">
        <h3>Marketplace feed</h3>
        <code>/marketplace.json</code>
        <p>Machine-readable plugin index Claude Code consumes when you add this server as a marketplace.</p>
      </div>
      <div class="surface">
        <h3>Git</h3>
        <code>/git/*</code>
        <p>Every plugin is a real bare repository served over Smart HTTP, ready for <code>git clone</code>.</p>
      </div>
      <div class="surface">
        <h3>MCP</h3>
        <code>/mcp</code>
        <p>A Model Context Protocol server so Claude (or any MCP client) can list plugins and modify skills as tool calls.</p>
      </div>
    </div>

    <h3>Base URL</h3>
    <pre>{{ origin }}</pre>
    <p class="muted">All paths in this document are relative to the base URL above.</p>

    <h3>Conventions</h3>
    <ul class="dev-list">
      <li>Request and response bodies are JSON encoded as UTF-8.</li>
      <li>Names — for plugins and skills — are lowercase slugs matching <code>^[a-z0-9][a-z0-9-]{1,62}[a-z0-9]$</code> (3–64 characters).</li>
      <li>Timestamps are RFC 3339 / ISO 8601 in UTC.</li>
      <li>Unless noted otherwise, success returns <code>200</code> with a JSON body, or <code>204 No Content</code> for write operations that don't need to echo state.</li>
    </ul>
  </section>
</template>

<style scoped>
.surface-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
  gap: 14px;
  margin: 16px 0 24px;
}
.surface {
  border: 1px solid var(--border-soft);
  padding: 14px 16px;
  background: var(--bg-2);
}
.surface h3 {
  margin: 0 0 4px;
  font-size: 14px;
}
.surface code {
  display: inline-block;
  font-family: var(--mono);
  font-size: 12px;
  color: var(--accent-2);
  margin-bottom: 6px;
}
.surface p {
  margin: 0;
  color: var(--text-soft);
  font-size: 13px;
  line-height: 1.45;
}
</style>
