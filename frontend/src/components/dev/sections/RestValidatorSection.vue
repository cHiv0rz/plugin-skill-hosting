<script setup lang="ts">
import Endpoint from '../Endpoint.vue'
</script>

<template>
  <div class="dev-subsection">
    <header class="section-head">
      <h2>Skill validator</h2>
      <p class="section-lede">Ask Claude to review a draft skill before you save it.</p>
    </header>

    <Endpoint
      method="POST"
      path="/api/skills/validate"
      summary="Run the same review the editor UI uses against an unsaved skill draft."
    >
      <template #request>
<pre>{
  "name":        "summarize-pr",
  "description": "...",
  "body":        "...",
  "files": [                           // optional, paths-only
    { "path": "scripts/run.py", "isBinary": false, "sizeBytes": 1234,
      "updatedAt": "2026-05-10T08:30:00Z" }
  ]
}</pre>
      </template>
      <template #response>
<pre>{
  "summary": "one short sentence verdict",
  "findings": [
    {
      "severity": "problem",     // problem | warning | info
      "title":    "Description is too vague",
      "detail":   "The description doesn't say WHEN to invoke the skill..."
    }
  ],
  "suggestedDescription": "rewritten description, or empty"
}</pre>
      </template>
      <template #notes>
        <p>
          At least one of <code>description</code> or <code>body</code> must be present.
          Requires the server to have <code>ANTHROPIC_API_KEY</code> configured.
        </p>
      </template>
      <template #errors>
        <ul class="dev-list">
          <li><code>400</code> — neither description nor body provided</li>
          <li><code>502</code> — Claude API call failed or returned unparseable output</li>
        </ul>
      </template>
    </Endpoint>
  </div>
</template>
