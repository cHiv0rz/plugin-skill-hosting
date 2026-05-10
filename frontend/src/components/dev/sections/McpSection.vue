<script setup lang="ts">
import { useApiExamples } from '../useApiExamples'
const { origin, exampleToken } = useApiExamples()

const tools = [
  { name: 'list_plugins', desc: 'List all active plugins (name, description, version, owner, updatedAt).', mode: 'read' },
  { name: 'get_plugin', desc: "Read a plugin's metadata and the list of its skills (names + descriptions, no bodies).", mode: 'read' },
  { name: 'get_skill', desc: "Read a skill's description, SKILL.md body, and the list of its supporting files.", mode: 'read' },
  { name: 'create_skill', desc: 'Add a new skill to a plugin. Bumps the plugin version and rewrites the git repo.', mode: 'write' },
  { name: 'update_skill', desc: "Replace a skill's description and body. Bumps the plugin version.", mode: 'write' },
  { name: 'list_skill_files', desc: 'List supporting files attached to a skill (paths + sizes, no content).', mode: 'read' },
  { name: 'get_skill_file', desc: 'Read one supporting file. Binary files are returned as base64 (isBinary=true).', mode: 'read' },
  { name: 'upsert_skill_file', desc: 'Write a supporting file under scripts/, references/, or assets/. Bumps the plugin patch version.', mode: 'write' },
]
</script>

<template>
  <section class="dev-section">
    <header class="section-head">
      <h2>MCP server</h2>
      <p class="section-lede">
        Speaks the
        <a href="https://modelcontextprotocol.io" target="_blank" rel="noopener">Model Context Protocol</a>
        over Streamable HTTP at <code>/mcp</code>. Plugins are read-only here — nothing
        can be deleted from this surface.
      </p>
    </header>

    <h3>Connect from Claude Code</h3>
<pre>claude mcp add --transport http skill-host {{ origin }}/mcp \
  -H "Authorization: Bearer {{ exampleToken }}"</pre>

    <h3>JSON config (Claude Desktop and other MCP clients)</h3>
<pre>{
  "mcpServers": {
    "skill-host": {
      "type": "http",
      "url":  "{{ origin }}/mcp",
      "headers": { "Authorization": "Bearer {{ exampleToken }}" }
    }
  }
}</pre>

    <h3>Tools</h3>
    <div class="tool-grid">
      <div v-for="t in tools" :key="t.name" class="tool" :class="`tool--${t.mode}`">
        <div class="tool-head">
          <code>{{ t.name }}</code>
          <span class="tool-mode">{{ t.mode }}</span>
        </div>
        <p>{{ t.desc }}</p>
      </div>
    </div>
  </section>
</template>

<style scoped>
.tool-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 12px;
  margin-top: 12px;
}
.tool {
  border: 1px solid var(--border-soft);
  border-left-width: 3px;
  padding: 12px 14px;
  background: var(--bg-2);
}
.tool--read  { border-left-color: #5ea0ff; }
.tool--write { border-left-color: var(--accent); }

.tool-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 6px;
}
.tool-head code {
  font-family: var(--mono);
  font-size: 13px;
  color: var(--text);
}
.tool-mode {
  font-family: var(--mono);
  font-size: 10px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--text-soft);
  border: 1px solid var(--border);
  padding: 1px 6px;
  border-radius: 999px;
}
.tool p { margin: 0; font-size: 13px; color: var(--text-soft); line-height: 1.45; }
</style>
