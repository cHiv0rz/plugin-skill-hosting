<script setup lang="ts">
import Endpoint from '../Endpoint.vue'
</script>

<template>
  <div class="dev-subsection">
    <header class="section-head">
      <h2>Auth endpoints</h2>
      <p class="section-lede">Sign in, sign up, and discover server config.</p>
    </header>

    <Endpoint
      method="GET"
      path="/api/auth/config"
      summary="Returns server configuration the login UI needs."
      :auth="false"
    >
      <template #response>
<pre>{
  "mode": "password",
  "marketplaceName": "oglimmer-marketplace",
  "defaultLicense": "MIT"
}</pre>
      </template>
      <template #notes>
        <p><code>mode</code> is either <code>password</code> or <code>oidc</code>.</p>
      </template>
    </Endpoint>

    <Endpoint
      method="POST"
      path="/api/auth/register"
      summary="Create a new account (password mode only)."
      :auth="false"
    >
      <template #request>
<pre>{
  "email":    "you@example.com",
  "username": "your-handle",
  "password": "at-least-8-chars"
}</pre>
      </template>
      <template #response>
<pre>{
  "token": "eyJhbGciOi...",       // JWT, send as Bearer
  "user": {
    "id":       "uuid",
    "email":    "you@example.com",
    "username": "your-handle",
    "apiToken": "32-byte hex",    // permanent API token
    "createdAt": "2026-05-10T12:00:00Z"
  }
}</pre>
      </template>
      <template #errors>
        <ul class="dev-list">
          <li><code>400</code> — invalid email, bad username, or password &lt; 8 chars</li>
          <li><code>409</code> — email or username already taken</li>
        </ul>
      </template>
    </Endpoint>

    <Endpoint
      method="POST"
      path="/api/auth/login"
      summary="Exchange email + password for a JWT (password mode only)."
      :auth="false"
    >
      <template #request>
<pre>{ "email": "you@example.com", "password": "..." }</pre>
      </template>
      <template #response>
        <p>Same shape as <code>/api/auth/register</code>.</p>
      </template>
      <template #errors>
        <ul class="dev-list"><li><code>401</code> — invalid credentials</li></ul>
      </template>
    </Endpoint>

    <Endpoint
      method="GET"
      path="/api/auth/oidc/login"
      summary="Begin the OIDC Authorization Code flow."
      :auth="false"
    >
      <template #notes>
        <p>
          Redirects (<code>302</code>) to the configured IdP. State + nonce are stored in
          short-lived cookies scoped to <code>/api/auth/oidc</code>. Available only when
          <code>AUTH_MODE=oidc</code>.
        </p>
      </template>
    </Endpoint>

    <Endpoint
      method="GET"
      path="/api/auth/oidc/callback"
      summary="OIDC redirect target. Validates the response and finishes login."
      :auth="false"
    >
      <template #notes>
        <p>
          On success it issues the same JWT + API-token pair as the password endpoints
          and redirects the browser back to the SPA.
        </p>
      </template>
    </Endpoint>
  </div>
</template>
