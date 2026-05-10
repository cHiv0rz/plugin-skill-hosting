<script setup lang="ts">
import Endpoint from '../Endpoint.vue'
import { useApiExamples } from '../useApiExamples'
const { origin, authedOrigin } = useApiExamples()
</script>

<template>
  <section class="dev-section">
    <header class="section-head">
      <h2>Marketplace feed</h2>
      <p class="section-lede">
        The machine-readable plugin index Claude Code consumes when you add this server
        as a marketplace.
      </p>
    </header>

    <Endpoint
      method="GET"
      path="/marketplace.json"
      summary="Returns the marketplace document. Authenticated like every other endpoint."
    >
      <template #notes>
        <p>
          Accepts both <code>Bearer</code> and HTTP Basic so <code>git</code> and
          <code>curl</code> can both fetch it. Each plugin's <code>source.url</code>
          embeds <em>your</em> API token as
          <code>https://_:&lt;token&gt;@host/git/&lt;name&gt;.git</code>, so the URL works
          as-is for cloning. Responses are served <code>Cache-Control: no-store</code>.
        </p>
      </template>
      <template #response>
<pre>{
  "name":  "oglimmer-marketplace",
  "owner": { "name": "...", "url": "{{ origin }}" },
  "plugins": [
    {
      "name":        "my-plugin",
      "description": "...",
      "version":     "1.2.0",
      "author":      { "name": "Alice", "email": "alice@example.com" },
      "homepage":    "https://...",
      "license":     "MIT",
      "source":      {
        "source": "url",
        "url":    "{{ authedOrigin }}/git/my-plugin.git"
      }
    }
  ]
}</pre>
      </template>
      <template #example>
<pre># Add the marketplace to Claude Code:
/plugin marketplace add {{ authedOrigin }}/marketplace.json</pre>
      </template>
    </Endpoint>
  </section>
</template>
