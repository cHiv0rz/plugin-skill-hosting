# Enterprise-mode walkthrough screenshots

Drop the connect-tab walkthrough screenshots here, named exactly. Steps 1–3 cover
enabling the MCP connector; steps 4–6 cover installing an organization plugin:

- `step-1.png` — open Customize from Chat
- `step-2.png` — open Customize from Cowork
- `step-3.png` — the MCP connector is already there (Connectors)
- `step-4.png` — open the plugin directory (+ → Browse)
- `step-5.png` — add a plugin from "Your organization" (Plugins)
- `step-6.png` — the plugin and its skills are ready

`PluginListView.vue` resolves whatever is present here via `import.meta.glob`, so:
- the build never breaks on a missing file, and
- a screenshot appears on the page as soon as you add it under the matching name.

`.png`, `.jpg`, `.jpeg` and `.webp` are all accepted. Prefer ~1200px-wide PNGs;
they're displayed as small thumbnails and shown full-size in a click-to-enlarge
lightbox.
