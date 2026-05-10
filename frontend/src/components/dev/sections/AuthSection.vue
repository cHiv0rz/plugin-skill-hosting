<script setup lang="ts">
import { useApiExamples } from '../useApiExamples'
const { origin, exampleToken } = useApiExamples()
</script>

<template>
  <section class="dev-section">
    <header class="section-head">
      <h2>Authentication</h2>
      <p class="section-lede">
        Almost every endpoint requires a credential. Three forms are accepted; pick whichever fits the caller.
      </p>
    </header>

    <div class="auth-card">
      <span class="auth-badge">1</span>
      <h3>JWT (browser session)</h3>
      <p>
        Issued by <code>POST /api/auth/register</code> and <code>POST /api/auth/login</code>.
        Valid for 30 days. Send it as a Bearer token:
      </p>
      <pre>Authorization: Bearer eyJhbGciOiJIUzI1NiIs...</pre>
      <p class="muted">
        JWTs are recognised by their three dot-separated segments. They're meant for the
        web UI; for scripts, prefer the API token.
      </p>
    </div>

    <div class="auth-card">
      <span class="auth-badge auth-badge--accent">2</span>
      <h3>API token <span class="recommend">recommended for automation</span></h3>
      <p>
        A long-lived opaque token tied to your user. Find it on the home page under
        <em>Advanced: raw API token</em>, or fetch it from <code>GET /api/me</code>.
        Send it the same way as a JWT:
      </p>
      <pre>Authorization: Bearer {{ exampleToken }}</pre>
    </div>

    <div class="auth-card">
      <span class="auth-badge">3</span>
      <h3>HTTP Basic</h3>
      <p>
        Username is ignored; the password must be your API token. This is what
        <code>git clone</code> uses, and it's why the marketplace URL embeds the token
        as <code>https://_:&lt;token&gt;@host/...</code>.
      </p>
      <pre>curl -u _:{{ exampleToken }} {{ origin }}/marketplace.json</pre>
    </div>

    <h3>Regenerating the token</h3>
    <p>
      <code>POST /api/me/token/regenerate</code> issues a new token and invalidates the
      old one. Existing marketplace links and Git remotes will stop working until you
      update them.
    </p>

    <h3>OIDC mode</h3>
    <p>
      When the server is started with <code>AUTH_MODE=oidc</code>, the password endpoints
      are replaced by an OAuth Authorization Code flow:
      <code>GET /api/auth/oidc/login</code> redirects to the IdP and
      <code>GET /api/auth/oidc/callback</code> completes the exchange. The result is the
      same JWT + API-token shape as password mode. Use
      <code>GET /api/auth/config</code> to discover which mode is active.
    </p>
  </section>
</template>

<style scoped>
.auth-card {
  position: relative;
  border: 1px solid var(--border-soft);
  background: var(--bg-2);
  padding: 18px 20px 14px 56px;
  margin: 14px 0;
}
.auth-card h3 {
  margin: 0 0 6px;
  font-size: 15px;
  display: flex;
  align-items: center;
  gap: 10px;
}
.auth-card p { margin: 6px 0; color: var(--text-soft); }
.auth-card pre { margin: 10px 0 6px; }

.auth-badge {
  position: absolute;
  top: 18px;
  left: 18px;
  width: 26px;
  height: 26px;
  border-radius: 999px;
  border: 1px solid var(--border);
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-family: var(--mono);
  font-size: 12px;
  color: var(--text-soft);
}
.auth-badge--accent {
  background: var(--accent);
  color: var(--bg);
  border-color: var(--accent);
}
.recommend {
  font-family: var(--mono);
  font-size: 10px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--accent);
  border: 1px solid rgba(245, 165, 36, 0.45);
  padding: 1px 8px;
  border-radius: 999px;
}
</style>
