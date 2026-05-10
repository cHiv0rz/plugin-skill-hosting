<script setup lang="ts">
import Endpoint from '../Endpoint.vue'
import { useApiExamples } from '../useApiExamples'
const { origin, exampleToken } = useApiExamples()
</script>

<template>
  <div class="dev-subsection">
    <header class="section-head">
      <h2>Account endpoints</h2>
      <p class="section-lede">Inspect the authenticated user and manage the API token.</p>
    </header>

    <Endpoint method="GET" path="/api/me" summary="Return the authenticated user.">
      <template #response>
<pre>{
  "id":        "uuid",
  "email":     "you@example.com",
  "username":  "your-handle",
  "apiToken":  "32-byte hex",
  "createdAt": "2026-05-10T12:00:00Z"
}</pre>
      </template>
      <template #example>
<pre>curl -H "Authorization: Bearer {{ exampleToken }}" \
  {{ origin }}/api/me</pre>
      </template>
    </Endpoint>

    <Endpoint
      method="POST"
      path="/api/me/token/regenerate"
      summary="Roll the API token. The previous token stops working immediately."
    >
      <template #response>
<pre>{ "apiToken": "new 32-byte hex" }</pre>
      </template>
    </Endpoint>

    <Endpoint
      method="GET"
      path="/api/me/deleted-plugins"
      summary="List soft-deleted plugins owned by the caller."
    >
      <template #notes>
        <p>
          Returns the same shape as <code>GET /api/plugins</code>, restricted to rows
          with a non-null <code>deletedAt</code>. Use the restore endpoint to bring one
          back.
        </p>
      </template>
    </Endpoint>

    <Endpoint
      method="GET"
      path="/api/version"
      summary="Build metadata for the running backend."
      :auth="false"
    >
      <template #response>
<pre>{
  "name":      "plugin-skill-hosting-backend",
  "version":   "0.1.11",
  "gitCommit": "d7159cd...",
  "buildTime": "2026-05-09T18:24:00Z"
}</pre>
      </template>
      <template #notes>
        <p>Mostly useful for the UI's version popup and for ops monitoring.</p>
      </template>
    </Endpoint>
  </div>
</template>
