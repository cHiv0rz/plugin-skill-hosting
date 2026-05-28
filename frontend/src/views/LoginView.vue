<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter, useRoute } from 'vue-router'
import { errMsg } from '../api'

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const auth = useAuthStore()
const router = useRouter()
const route = useRoute()

onMounted(() => { auth.ensureMode() })

async function submit() {
  error.value = ''
  loading.value = true
  try {
    await auth.login(email.value, password.value)
    const dest = typeof route.query.redirect === 'string' ? route.query.redirect : '/'
    router.push(dest)
  } catch (e: unknown) {
    error.value = errMsg(e, 'login failed')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <main class="lv">
    <div class="lv__card">
      <header class="lv__head">
        <span class="lv__brand">plugin / market</span>
        <span class="lv__step">sign in</span>
      </header>

      <p class="lv__intro">access the registry.</p>

      <template v-if="auth.mode === 'oidc'">
        <p class="lv__sso-text">
          this workspace uses single sign-on. continue with your identity provider.
        </p>
        <button type="button" class="lv-btn lv-btn--primary" @click="auth.loginViaOIDC()">
          continue with SSO →
        </button>
      </template>

      <template v-else-if="auth.mode === 'password'">
        <form class="lv__form" @submit.prevent="submit" novalidate>
          <div class="lv-field">
            <label for="lv-email" class="lv-field__label">email</label>
            <input
              id="lv-email"
              v-model="email"
              type="email"
              required
              autocomplete="email"
              spellcheck="false"
              class="lv-field__input"
              placeholder="you@example.com"
            />
          </div>

          <div class="lv-field">
            <label for="lv-pass" class="lv-field__label">password</label>
            <input
              id="lv-pass"
              v-model="password"
              type="password"
              required
              autocomplete="current-password"
              class="lv-field__input"
              placeholder="••••••••"
            />
          </div>

          <div v-if="error" class="lv-error">{{ error }}</div>

          <button type="submit" class="lv-btn lv-btn--primary" :disabled="loading">
            {{ loading ? 'signing in…' : 'sign in' }}
          </button>

          <p class="lv__signup">
            no account?
            <RouterLink to="/register" class="lv__signup-link">create one →</RouterLink>
          </p>
        </form>
      </template>

      <template v-else>
        <p class="lv__loading">loading…</p>
      </template>
    </div>
  </main>
</template>

<style scoped>
.lv {
  min-height: 100vh;
  display: grid;
  place-items: center;
  padding: 48px 20px;
  background: var(--bg);
  color: var(--text);
  font-family: var(--mono);
}
.lv__card {
  width: 100%;
  max-width: 360px;
  padding: 28px 28px 24px;
  background: var(--bg-2);
  border: 1px solid var(--border);
}

/* ─── Header ───────────────────────────────────────────────────── */
.lv__head {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 14px;
  border-bottom: 1px solid var(--border-soft);
}
.lv__brand {
  font-family: var(--mono);
  font-size: 10.5px;
  font-weight: 700;
  letter-spacing: 0.28em;
  text-transform: uppercase;
  color: var(--text);
  flex: 1 1 auto;
}
.lv__step {
  font-family: var(--mono);
  font-size: 10.5px;
  font-weight: 700;
  letter-spacing: 0.28em;
  color: var(--text);
  padding: 3px 8px;
  border: 1px solid var(--accent);
}
.lv__intro {
  margin: 18px 0 22px;
  font-family: var(--mono);
  font-size: 12.5px;
  color: var(--text-soft);
  line-height: 1.55;
}

/* ─── Form ─────────────────────────────────────────────────────── */
.lv__form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}
.lv-field { display: block; }
.lv-field__label {
  display: block;
  margin: 0 0 6px;
  font-family: var(--mono);
  font-size: 10px;
  font-weight: 600;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--text-soft);
}
.lv-field__input {
  width: 100%;
  background: var(--bg);
  color: var(--text);
  border: 1px solid var(--border);
  border-radius: 0;
  padding: 9px 12px;
  font-family: var(--mono);
  font-size: 13px;
  outline: none;
  transition: border-color 0.15s ease;
}
.lv-field__input:focus { border-color: var(--accent); }
.lv-field__input::placeholder { color: var(--muted); }
.lv-field__input:-webkit-autofill {
  -webkit-text-fill-color: var(--text);
  -webkit-box-shadow: 0 0 0 1000px var(--bg) inset;
  transition: background-color 9999s ease;
}

/* ─── Error ────────────────────────────────────────────────────── */
.lv-error {
  padding: 9px 12px;
  font-family: var(--mono);
  font-size: 12px;
  color: #8f3115;
  background: rgba(194, 73, 31, 0.10);
  border-left: 2px solid var(--rust);
}

/* ─── Buttons ──────────────────────────────────────────────────── */
.lv-btn {
  background: transparent;
  color: var(--text);
  border: 1px solid var(--border);
  border-radius: 0;
  padding: 9px 14px;
  font-family: var(--mono);
  font-size: 12px;
  font-weight: 500;
  letter-spacing: 0.04em;
  text-transform: lowercase;
  line-height: 1.5;
  cursor: pointer;
  text-decoration: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: border-color 0.12s ease, color 0.12s ease, background 0.12s ease;
  margin: 0;
}
.lv-btn::before { display: none; content: none; }
.lv-btn:hover {
  background: var(--text);
  color: var(--bg);
  border-color: var(--text);
  transform: none;
}
.lv-btn:active { transform: none; }
.lv-btn:disabled,
.lv-btn:disabled:hover {
  opacity: 0.45;
  cursor: not-allowed;
  color: var(--text-soft);
  border-color: var(--border);
  background: transparent;
}
.lv-btn--primary {
  color: var(--text);
  background: var(--accent);
  border-color: var(--accent);
  font-weight: 700;
  padding: 10px 14px;
}
.lv-btn--primary:hover {
  color: var(--bg);
  background: var(--text);
  border-color: var(--text);
}

/* ─── Footer text ──────────────────────────────────────────────── */
.lv__signup {
  margin: 4px 0 0;
  font-family: var(--mono);
  font-size: 12px;
  color: var(--text-soft);
}
.lv__signup-link {
  color: var(--text);
  border-bottom: 1px solid var(--accent);
  padding-bottom: 1px;
  margin-left: 4px;
  transition: color 0.12s ease;
}
.lv__signup-link:hover { color: var(--accent-2); }

.lv__sso-text {
  margin: 0 0 16px;
  font-family: var(--mono);
  font-size: 12.5px;
  color: var(--text-soft);
  line-height: 1.55;
}
.lv__loading {
  margin: 0;
  font-size: 12.5px;
  color: var(--muted);
}
</style>
