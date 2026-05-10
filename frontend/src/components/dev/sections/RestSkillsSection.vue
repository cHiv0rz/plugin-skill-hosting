<script setup lang="ts">
import Endpoint from '../Endpoint.vue'
import ParamRow from '../ParamRow.vue'
</script>

<template>
  <div class="dev-subsection">
    <header class="section-head">
      <h2>Skill endpoints</h2>
      <p class="section-lede">
        Every skill write also bumps the parent plugin's semver and rewrites its git repo.
      </p>
    </header>

    <aside class="callout">
      <h4>Plugin version bumps</h4>
      <ul class="dev-list">
        <li><strong>First skill in a plugin</strong> — no version bump (the plugin's debut version stays); only <code>updatedAt</code> is refreshed.</li>
        <li><strong>Subsequent create / delete / restore</strong> — bumps the <strong>major</strong>.</li>
        <li><strong>Update body+description</strong> — bumps the <strong>minor</strong> when the body size changes by more than 30%, otherwise the <strong>patch</strong>.</li>
        <li><strong>Skill file upsert/delete</strong> — bumps the <strong>patch</strong>.</li>
      </ul>
    </aside>

    <Endpoint
      method="POST"
      path="/api/plugins/{name}/skills"
      summary="Create a new skill inside a plugin."
    >
      <template #request>
<pre>{
  "name":        "summarize-pr",            // slug, required
  "description": "Summarises a GitHub PR.",  // required, non-empty
  "body":        "# SKILL\n\nSteps..."       // SKILL.md body, no YAML frontmatter
}</pre>
      </template>
      <template #errors>
        <ul class="dev-list">
          <li><code>400</code> — invalid skill name or empty description</li>
          <li><code>404</code> — plugin not found</li>
          <li><code>409</code> — skill with that name already exists</li>
        </ul>
      </template>
    </Endpoint>

    <Endpoint
      method="PUT"
      path="/api/plugins/{name}/skills/{skill}"
      summary="Replace a skill's description and body."
    >
      <template #request>
<pre>{
  "description": "...",
  "body":        "..."
}</pre>
      </template>
      <template #response>
        <p><code>204 No Content</code></p>
      </template>
    </Endpoint>

    <Endpoint
      method="DELETE"
      path="/api/plugins/{name}/skills/{skill}"
      summary="Soft-delete a skill. Bumps the plugin major."
    >
      <template #response>
        <p><code>204 No Content</code></p>
      </template>
    </Endpoint>

    <Endpoint
      method="GET"
      path="/api/plugins/{name}/deleted-skills"
      summary="List soft-deleted skills for a plugin (so you can restore them)."
    />

    <Endpoint
      method="POST"
      path="/api/plugins/{name}/skills/{skill}/restore"
      summary="Un-delete the most-recently-deleted skill of that name. Bumps the plugin major."
    >
      <template #errors>
        <ul class="dev-list">
          <li><code>404</code> — no deleted skill with that name</li>
          <li><code>409</code> — an active skill with that name already exists</li>
        </ul>
      </template>
    </Endpoint>

    <Endpoint
      method="GET"
      path="/api/plugins/{name}/skills/{skill}/versions"
      summary="Return the full edit history for a skill, newest first."
    >
      <template #response>
<pre>[
  {
    "id":          "uuid",
    "skillId":     "uuid",
    "version":     7,
    "action":      "update",   // create | update | delete | restore | revert
    "name":        "summarize-pr",
    "description": "...",
    "body":        "...",      // snapshot at this version
    "editedBy":     "uuid",
    "editedByName": "alice",
    "editedAt":    "2026-05-10T08:30:00Z"
  }
]</pre>
      </template>
    </Endpoint>

    <Endpoint
      method="POST"
      path="/api/plugins/{name}/skills/{skill}/revert/{version}"
      summary="Restore a skill (description, body, and supporting files) to an earlier version."
    >
      <template #params>
        <ParamRow name="version" type="path int" required>
          The version number from the skill's history.
        </ParamRow>
      </template>
      <template #notes>
        <p>
          Acts as both un-delete (if the skill is currently soft-deleted) and content
          rollback in one operation. A new history row of action <code>revert</code> is
          appended.
        </p>
      </template>
    </Endpoint>
  </div>
</template>

<style scoped>
.callout {
  border-left: 3px solid var(--accent);
  background: rgba(245, 165, 36, 0.06);
  padding: 12px 16px;
  margin: 8px 0 20px;
}
.callout h4 {
  margin: 0 0 6px;
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.2em;
  text-transform: uppercase;
  color: var(--accent-2);
}
.callout ul { margin: 4px 0; }
</style>
