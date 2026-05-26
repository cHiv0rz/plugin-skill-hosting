<script setup lang="ts">
import { useApiExamples } from '../useApiExamples'
const { origin, exampleToken } = useApiExamples()
</script>

<template>
  <section class="dev-section">
    <header class="section-head">
      <h2>CLI: import-plugin</h2>
      <p class="section-lede">
        A small Go command-line tool that uploads an on-disk plugin directory
        into a running marketplace via the REST API. Useful for one-off
        migrations from a local checkout, an old marketplace instance, or any
        directory that follows the plugin layout below.
      </p>
    </header>

    <h3>Build</h3>
    <pre>cd backend
go build -o import-plugin ./cmd/import-plugin</pre>
    <p class="muted">
      The binary is statically linked and has no runtime dependencies.
    </p>

    <h3>Input layout</h3>
    <p>
      The directory you pass must contain a plugin manifest at
      <code>.claude-plugin/plugin.json</code>. Skills, if any, must live under
      <code>skills/&lt;name&gt;/</code> with a <code>SKILL.md</code> at the
      root of each skill folder. This is the same layout the server
      materialises into git, so the output of <code>git clone</code> of one
      marketplace can be fed directly to the importer of another.
    </p>
<pre>my-plugin/
├── .claude-plugin/
│   └── plugin.json
└── skills/
    ├── foo/
    │   ├── SKILL.md
    │   └── scripts/run.sh
    └── bar/
        └── SKILL.md</pre>

    <h3>Configuration</h3>
    <table class="dev-table">
      <thead>
        <tr><th>Flag</th><th>Env var</th><th>Purpose</th></tr>
      </thead>
      <tbody>
        <tr>
          <td><code>--url</code></td>
          <td><code>MARKETPLACE_URL</code></td>
          <td>Marketplace base URL, e.g. <code>{{ origin }}</code>.</td>
        </tr>
        <tr>
          <td><code>--token</code></td>
          <td><code>MARKETPLACE_TOKEN</code></td>
          <td>Your API token (shown on the home page).</td>
        </tr>
      </tbody>
    </table>
    <p class="muted">Flags take precedence over env vars.</p>

    <h3>Example</h3>
<pre>MARKETPLACE_URL={{ origin }} \
MARKETPLACE_TOKEN={{ exampleToken }} \
./import-plugin ./my-plugin

Creating plugin "my-plugin"...
Importing skill "foo"...
Importing skill "bar"...
Done. Imported 2 skill(s).</pre>

    <h3>What happens server-side</h3>
    <ul class="dev-list">
      <li>
        <strong>Plugin</strong> — <code>POST /api/plugins</code> with the
        manifest's name, description, author, homepage, and license. The
        version is assigned by the server.
      </li>
      <li>
        <strong>Each skill</strong> — the skill folder is zipped in memory and
        sent to <code>POST /api/plugins/{name}/skills/import</code>. The server
        parses <code>SKILL.md</code>, validates every supporting-file path, and
        inserts the skill plus its files in a single transaction. Same code
        path the web UI's "Import skill" button uses.
      </li>
      <li>
        After all skills land, the plugin's bare git repo is re-materialised
        and the marketplace feed picks up the new entry on the next request.
      </li>
    </ul>

    <h3>On conflict</h3>
    <p>
      The tool is intentionally not idempotent: if a plugin with the same name
      already exists, the create call fails with
      <code>409 plugin name already taken</code> and the importer aborts
      before touching any skills. Delete the existing plugin (or rename the
      incoming one) and re-run.
    </p>
  </section>
</template>

<style scoped>
.dev-section h3 {
  margin: 22px 0 8px;
  font-size: 14px;
  font-family: var(--mono);
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--accent-2);
}
.dev-table th, .dev-table td {
  padding: 6px 12px 6px 0;
  border-bottom: 1px solid var(--border-soft);
  font-size: 13px;
  color: var(--text-soft);
}
.dev-table th {
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: var(--accent-2);
}
.muted { color: var(--muted); font-size: 12px; margin: 4px 0 12px; }
</style>
