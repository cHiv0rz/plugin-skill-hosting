<script setup lang="ts">
import Endpoint from '../Endpoint.vue'
import ParamRow from '../ParamRow.vue'
import { useApiExamples } from '../useApiExamples'
const { origin, exampleToken } = useApiExamples()
</script>

<template>
  <div class="dev-subsection">
    <header class="section-head">
      <h2>Plugin endpoints</h2>
      <p class="section-lede">Create, list, soft-delete, and restore plugins.</p>
    </header>

    <Endpoint method="GET" path="/api/plugins" summary="List every active (non-deleted) plugin.">
      <template #response>
<pre>[
  {
    "id":          "uuid",
    "ownerId":     "uuid",
    "ownerName":   "alice",
    "name":        "my-plugin",
    "description": "Short summary",
    "version":     "1.2.0",
    "authorName":  "Alice",
    "authorEmail": "alice@example.com",
    "homepage":    "https://...",
    "license":     "MIT",
    "createdAt":   "2026-04-01T10:00:00Z",
    "updatedAt":   "2026-05-10T08:30:00Z"
  }
]</pre>
      </template>
    </Endpoint>

    <Endpoint
      method="GET"
      path="/api/plugins/{name}"
      summary="Fetch a single plugin and its skills."
    >
      <template #params>
        <ParamRow name="name" type="path string" required>
          The plugin slug.
        </ParamRow>
      </template>
      <template #response>
        <p>
          Same fields as the list endpoint, plus a <code>skills</code> array. Each skill
          row carries <code>id</code>, <code>name</code>, <code>description</code>,
          <code>body</code>, audit columns (<code>createdBy</code> / <code>updatedBy</code>
          with usernames), and timestamps.
        </p>
      </template>
      <template #errors>
        <ul class="dev-list"><li><code>404</code> — plugin not found</li></ul>
      </template>
    </Endpoint>

    <Endpoint
      method="POST"
      path="/api/plugins"
      summary="Create a new plugin owned by the caller."
    >
      <template #request>
<pre>{
  "name":        "my-plugin",          // slug, required
  "description": "Short summary",
  "authorName":  "Alice",
  "authorEmail": "alice@example.com",
  "homepage":    "https://...",
  "license":     "MIT"
}</pre>
      </template>
      <template #notes>
        <p>
          The first plugin you create starts at version <code>0.1.0</code>; every
          subsequent plugin starts at <code>1.0.0</code>. The <code>version</code> field
          in the request is ignored — it is fully managed by the server based on later
          skill activity.
        </p>
        <p>
          On success the plugin's bare git repo is materialized at
          <code>/git/{name}.git</code>.
        </p>
      </template>
      <template #example>
<pre>curl -X POST -H "Authorization: Bearer {{ exampleToken }}" \
  -H "Content-Type: application/json" \
  -d '{"name":"weather-tools","description":"Forecast skills"}' \
  {{ origin }}/api/plugins</pre>
      </template>
      <template #errors>
        <ul class="dev-list">
          <li><code>400</code> — invalid name (not a valid slug)</li>
          <li><code>409</code> — plugin name already taken</li>
        </ul>
      </template>
    </Endpoint>

    <Endpoint
      method="DELETE"
      path="/api/plugins/{name}"
      summary="Soft-delete a plugin you own."
    >
      <template #notes>
        <p>
          The row stays in the database; it disappears from listings, the marketplace
          feed, and the bare repo on disk is removed. Use the restore endpoint to
          recreate it.
        </p>
      </template>
      <template #response>
        <p><code>204 No Content</code></p>
      </template>
      <template #errors>
        <ul class="dev-list">
          <li><code>403</code> — not your plugin</li>
          <li><code>404</code> — plugin not found</li>
        </ul>
      </template>
    </Endpoint>

    <Endpoint
      method="POST"
      path="/api/plugins/{name}/restore"
      summary="Un-delete a plugin and re-materialize its git repo."
    >
      <template #errors>
        <ul class="dev-list">
          <li><code>400</code> — plugin is not deleted</li>
          <li><code>403</code> — not your plugin</li>
          <li><code>404</code> — plugin not found</li>
          <li><code>409</code> — an active plugin with that name already exists</li>
        </ul>
      </template>
    </Endpoint>
  </div>
</template>
