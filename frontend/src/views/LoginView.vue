<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useRouter, useRoute } from 'vue-router'
import { errMsg } from '../api'

const email = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)
const focused = ref<'email' | 'password' | null>(null)
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

const feed = [
  { idx: '042', name: 'claude-api',       version: '1.4.2', author: 'anthropic',  tag: 'sdk',       hue: 1 },
  { idx: '041', name: 'frontend-design',  version: '2.0.0', author: 'official',   tag: 'skill',     hue: 2 },
  { idx: '040', name: 'spring-to-go',     version: '0.3.1', author: '@oglimmer',  tag: 'migration', hue: 3 },
  { idx: '039', name: 'docker-e2e',       version: '1.0.0', author: 'official',   tag: 'workflow',  hue: 4 },
  { idx: '038', name: 'version-info',     version: '0.1.4', author: 'community',  tag: 'skill',     hue: 5 },
]

const now = new Date()
const stamp = `${String(now.getUTCFullYear()).slice(-2)}.${String(now.getUTCMonth() + 1).padStart(2, '0')}.${String(now.getUTCDate()).padStart(2, '0')} · ${String(now.getUTCHours()).padStart(2, '0')}:${String(now.getUTCMinutes()).padStart(2, '0')} UTC`
</script>

<template>
  <div class="shell">
    <!-- ─── LEFT: visual stage ─────────────────────────────────────── -->
    <aside class="stage" aria-hidden="true">
      <div class="stage__grid"></div>
      <div class="stage__noise"></div>
      <div class="stage__glow"></div>

      <div class="stage__edge">
        <span>DEVELOPER&nbsp;·&nbsp;PORTAL</span>
        <span class="dot">●</span>
        <span>v0.1 / EDITION&nbsp;01</span>
      </div>

      <header class="stage__head">
        <div class="brandmark">
          <svg viewBox="0 0 32 32" width="22" height="22" aria-hidden="true">
            <path d="M4 8 L16 2 L28 8 L28 24 L16 30 L4 24 Z"
                  fill="none" stroke="currentColor" stroke-width="1.4" />
            <path d="M16 2 L16 30 M4 8 L28 24 M28 8 L4 24"
                  stroke="currentColor" stroke-width="0.6" opacity="0.45" />
          </svg>
          <span>plugin / market</span>
        </div>
        <div class="status">
          <span class="pulse"></span>
          <span>registry online</span>
        </div>
      </header>

      <section class="hero">
        <div class="hero__index">
          <span class="num">01</span>
          <span class="frac">/&nbsp;welcome</span>
        </div>
        <h1 class="hero__title">
          <span class="line">Workflows,</span>
          <span class="line italic">packaged.</span>
          <span class="line">Expertise,</span>
          <span class="line italic last">shared.</span>
        </h1>
        <p class="hero__lede">
          A registry of skills, agents, and hooks for Claude — write a workflow
          once, version it, and let every Claude on your team run it on demand.
        </p>
        <div class="hero__meta">
          <div><em>Issue</em><strong>№ 01</strong></div>
          <div><em>Stamp</em><strong>{{ stamp }}</strong></div>
          <div><em>Region</em><strong>EU&nbsp;/&nbsp;FRA-1</strong></div>
        </div>
      </section>

      <section class="feed" aria-label="Recent registry activity">
        <div class="feed__head">
          <span class="feed__label">▌ recent · pushed</span>
          <span class="feed__pager">5 / 142</span>
        </div>
        <ul class="feed__list">
          <li
            v-for="(p, i) in feed"
            :key="p.name"
            class="card"
            :class="`card--h${p.hue}`"
            :style="{ animationDelay: `${i * 0.6}s` }"
          >
            <div class="card__row">
              <span class="card__idx">{{ p.idx }}</span>
              <span class="card__tag">{{ p.tag }}</span>
            </div>
            <div class="card__name">{{ p.name }}</div>
            <div class="card__row card__foot">
              <span>{{ p.author }}</span>
              <span class="card__ver">v{{ p.version }}</span>
            </div>
            <div class="card__bar"><i></i></div>
          </li>
        </ul>
      </section>

      <div class="ticker">
        <div class="ticker__track">
          <span v-for="n in 2" :key="n">
            ● MARKETPLACE &nbsp;·&nbsp; PLUGINS &nbsp;·&nbsp; SKILLS &nbsp;·&nbsp;
            COMMANDS &nbsp;·&nbsp; AGENTS &nbsp;·&nbsp; HOOKS &nbsp;·&nbsp;
            MCP&nbsp;SERVERS &nbsp;·&nbsp; PUBLISH &nbsp;·&nbsp; INSTALL &nbsp;·&nbsp;
            VERSIONED &nbsp;·&nbsp; OPEN &nbsp;
          </span>
        </div>
      </div>
    </aside>

    <!-- ─── RIGHT: form ────────────────────────────────────────────── -->
    <section class="panel">
      <header class="panel__head">
        <span class="panel__step">auth · session 01</span>
        <span class="panel__locale">EN&nbsp;/&nbsp;UTC</span>
      </header>

      <div class="panel__body">
        <div class="terminal">
          <span class="prompt">$</span>
          <span class="cmd">market auth --resume</span>
          <span class="caret"></span>
        </div>
        <p class="reply">
          <span class="reply__bullet">›</span>
          handshake established · awaiting credentials
        </p>

        <div class="title-row">
          <h2 class="title">Sign&nbsp;in</h2>
          <span class="title__index">— 01.01</span>
        </div>
        <p class="title__sub">
          Welcome back. Pick up where the last commit left off.
        </p>

        <template v-if="auth.mode === 'oidc'">
          <div class="sso-block">
            <p class="sso-text">
              This workspace is sealed by single sign-on. Continue with your
              identity provider to enter the registry.
            </p>
            <button type="button" class="btn-primary" @click="auth.loginViaOIDC()">
              <span>Continue with SSO</span>
              <span class="arrow" aria-hidden="true">→</span>
            </button>
          </div>
        </template>

        <template v-else-if="auth.mode === 'password'">
          <form class="form" @submit.prevent="submit" novalidate>
            <div class="field" :class="{ 'is-focused': focused === 'email', 'is-filled': email }">
              <label for="lf-email">
                <span class="field__num">01</span>
                <span class="field__name">Email&nbsp;address</span>
              </label>
              <input
                id="lf-email"
                v-model="email"
                type="email"
                required
                autocomplete="email"
                spellcheck="false"
                placeholder="you@workshop.dev"
                @focus="focused = 'email'"
                @blur="focused = null"
              />
              <span class="field__rule"><i></i></span>
            </div>

            <div class="field" :class="{ 'is-focused': focused === 'password', 'is-filled': password }">
              <label for="lf-pass">
                <span class="field__num">02</span>
                <span class="field__name">Passphrase</span>
                <a href="#" class="field__aux" @click.prevent>forgot?</a>
              </label>
              <input
                id="lf-pass"
                v-model="password"
                type="password"
                required
                autocomplete="current-password"
                placeholder="••••••••••••"
                @focus="focused = 'password'"
                @blur="focused = null"
              />
              <span class="field__rule"><i></i></span>
            </div>

            <transition name="err">
              <div v-if="error" class="alert" role="alert">
                <span class="alert__mark">!</span>
                <span class="alert__msg">{{ error }}</span>
              </div>
            </transition>

            <button type="submit" class="btn-primary" :disabled="loading">
              <span v-if="!loading">Enter the registry</span>
              <span v-else class="loading">
                <i></i><i></i><i></i>&nbsp;negotiating
              </span>
              <span class="arrow" aria-hidden="true">→</span>
            </button>

            <div class="aside">
              <div class="rule"></div>
              <span>or</span>
              <div class="rule"></div>
            </div>

            <p class="signup">
              No account yet?
              <RouterLink to="/register" class="signup__link">
                Forge a new one<span class="signup__arrow">↗</span>
              </RouterLink>
            </p>
          </form>
        </template>

        <template v-else>
          <p class="muted">Loading…</p>
        </template>
      </div>

      <footer class="panel__foot">
        <span>© plugin / market — {{ new Date().getFullYear() }}</span>
        <span class="spacer"></span>
        <span class="panel__foot-meta">credentials encrypted in transit</span>
      </footer>
    </section>
  </div>
</template>

<style scoped>
/* ─── Scoped tokens ─────────────────────────────────────────────── */
.shell {
  --ink:        #0a0c10;
  --ink-2:      #0e1118;
  --ink-3:      #14181f;
  --line:       #232733;
  --line-soft:  #1a1e27;
  --text:       #ece7d8;
  --text-soft:  #b9b3a3;
  --muted:      #6b7180;
  --amber:      #f5a524;
  --amber-2:    #ffbf52;
  --amber-soft: rgba(245, 165, 36, 0.14);
  --rust:       #d65a31;
  --paper:      #f0e9d6;
  --serif:      'Fraunces Variable', 'Fraunces', 'Times New Roman', serif;
  --mono:       'JetBrains Mono Variable', 'JetBrains Mono', ui-monospace, SFMono-Regular, Menlo, monospace;

  position: fixed;
  inset: 0;
  display: grid;
  grid-template-columns: 1.05fr 0.95fr;
  background: var(--ink);
  color: var(--text);
  font-family: var(--mono);
  font-size: 14px;
  line-height: 1.55;
  overflow: hidden;
}

/* ═══ STAGE (left) ════════════════════════════════════════════════ */
.stage {
  position: relative;
  background:
    radial-gradient(1100px 700px at 18% 28%, rgba(245, 165, 36, 0.10), transparent 60%),
    radial-gradient(900px 600px at 82% 90%, rgba(214, 90, 49, 0.08), transparent 60%),
    linear-gradient(180deg, #0b0e14 0%, #07090d 100%);
  border-right: 1px solid var(--line);
  overflow: hidden;
  isolation: isolate;
}
.stage__grid {
  position: absolute; inset: 0;
  background-image:
    linear-gradient(to right, rgba(255,255,255,0.025) 1px, transparent 1px),
    linear-gradient(to bottom, rgba(255,255,255,0.025) 1px, transparent 1px);
  background-size: 56px 56px;
  mask-image: radial-gradient(120% 80% at 50% 35%, #000 40%, transparent 90%);
  pointer-events: none;
}
.stage__noise {
  position: absolute; inset: 0;
  background-image: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='160' height='160'><filter id='n'><feTurbulence type='fractalNoise' baseFrequency='0.9' numOctaves='2' stitchTiles='stitch'/><feColorMatrix values='0 0 0 0 0  0 0 0 0 0  0 0 0 0 0  0 0 0 0.55 0'/></filter><rect width='100%' height='100%' filter='url(%23n)' opacity='0.5'/></svg>");
  opacity: 0.18;
  mix-blend-mode: overlay;
  pointer-events: none;
}
.stage__glow {
  position: absolute;
  width: 520px; height: 520px;
  left: 60%; top: 12%;
  background: radial-gradient(circle, rgba(245, 165, 36, 0.20), transparent 70%);
  filter: blur(28px);
  animation: drift 18s ease-in-out infinite alternate;
  pointer-events: none;
}
@keyframes drift {
  0%   { transform: translate3d(0,0,0) scale(1); }
  100% { transform: translate3d(-80px, 60px, 0) scale(1.1); }
}

/* edge label */
.stage__edge {
  position: absolute;
  left: 26px; top: 0; bottom: 0;
  display: flex;
  align-items: center;
  gap: 14px;
  transform: rotate(-90deg);
  transform-origin: left top;
  translate: 0 50vh;
  font-size: 10px;
  letter-spacing: 0.32em;
  color: var(--muted);
  text-transform: uppercase;
  white-space: nowrap;
  pointer-events: none;
}
.stage__edge .dot { color: var(--amber); }

.stage__head {
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 28px 44px 28px 64px;
  z-index: 2;
  animation: rise 0.8s 0.05s both ease-out;
}
.brandmark {
  display: inline-flex;
  align-items: center;
  gap: 12px;
  font-family: var(--mono);
  font-weight: 600;
  font-size: 13px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--text);
}
.brandmark svg { color: var(--amber); }
.status {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  font-size: 11px;
  letter-spacing: 0.24em;
  text-transform: uppercase;
  color: var(--text-soft);
}
.pulse {
  width: 8px; height: 8px; border-radius: 50%;
  background: #5fff8f;
  box-shadow: 0 0 0 0 rgba(95, 255, 143, 0.6);
  animation: pulse 2.2s infinite;
}
@keyframes pulse {
  0%   { box-shadow: 0 0 0 0    rgba(95, 255, 143, 0.55); }
  70%  { box-shadow: 0 0 0 12px rgba(95, 255, 143, 0); }
  100% { box-shadow: 0 0 0 0    rgba(95, 255, 143, 0); }
}

/* hero */
.hero {
  position: relative;
  z-index: 2;
  padding: 24px 64px 0 64px;
  max-width: 720px;
}
.hero__index {
  display: inline-flex;
  align-items: baseline;
  gap: 12px;
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.32em;
  text-transform: uppercase;
  color: var(--amber);
  animation: rise 0.7s 0.15s both ease-out;
}
.hero__index .num {
  font-family: var(--serif);
  font-style: italic;
  font-weight: 300;
  font-size: 38px;
  letter-spacing: 0;
  color: var(--text);
  line-height: 1;
  text-transform: none;
}
.hero__index .frac { color: var(--text-soft); }

.hero__title {
  margin: 18px 0 22px;
  font-family: var(--serif);
  font-weight: 380;
  font-size: clamp(48px, 6.6vw, 88px);
  line-height: 0.94;
  letter-spacing: -0.025em;
  color: var(--text);
}
.hero__title .line {
  display: block;
  opacity: 0;
  transform: translateY(18px);
  animation: rise 0.9s both ease-out;
}
.hero__title .line:nth-child(1) { animation-delay: 0.20s; }
.hero__title .line:nth-child(2) { animation-delay: 0.32s; }
.hero__title .line:nth-child(3) { animation-delay: 0.44s; }
.hero__title .line:nth-child(4) { animation-delay: 0.56s; }
.hero__title .italic {
  font-style: italic;
  font-weight: 300;
  color: var(--amber-2);
  margin-left: 0.4em;
}
.hero__title .last {
  position: relative;
}
.hero__title .last::after {
  content: '';
  display: inline-block;
  width: 0.6ch;
  height: 0.78em;
  background: var(--amber);
  margin-left: 0.18em;
  transform: translateY(-0.06em);
  animation: blink 1.05s steps(2, jump-none) infinite;
}
@keyframes blink { 50% { opacity: 0; } }

.hero__lede {
  max-width: 460px;
  margin: 0 0 28px;
  font-family: var(--mono);
  font-size: 13.5px;
  line-height: 1.7;
  color: var(--text-soft);
  animation: rise 0.9s 0.7s both ease-out;
}

.hero__meta {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, max-content));
  gap: 28px 32px;
  padding: 18px 0;
  border-top: 1px solid var(--line-soft);
  border-bottom: 1px solid var(--line-soft);
  animation: rise 0.9s 0.85s both ease-out;
}
.hero__meta > div { display: flex; flex-direction: column; gap: 4px; }
.hero__meta em {
  font-style: normal;
  font-size: 10px;
  letter-spacing: 0.28em;
  text-transform: uppercase;
  color: var(--muted);
}
.hero__meta strong {
  font-family: var(--mono);
  font-weight: 500;
  font-size: 12.5px;
  color: var(--text);
}

/* feed */
.feed {
  position: absolute;
  right: -44px; top: 96px;
  width: 360px;
  z-index: 2;
  transform: rotate(-2deg);
  animation: rise 1s 0.4s both ease-out;
  pointer-events: none;
}
.feed__head {
  display: flex; justify-content: space-between; align-items: center;
  margin: 0 22px 14px;
  font-size: 10px;
  letter-spacing: 0.28em;
  text-transform: uppercase;
}
.feed__label { color: var(--amber); }
.feed__pager { color: var(--muted); }
.feed__list {
  list-style: none;
  margin: 0; padding: 0;
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.card {
  position: relative;
  padding: 14px 16px 14px;
  background: linear-gradient(180deg, rgba(20, 24, 31, 0.85), rgba(14, 17, 24, 0.85));
  border: 1px solid var(--line);
  border-radius: 2px;
  backdrop-filter: blur(2px);
  font-family: var(--mono);
  color: var(--text);
  box-shadow:
    0 1px 0 rgba(255,255,255,0.03) inset,
    0 12px 30px -16px rgba(0,0,0,0.65);
  animation: float 9s ease-in-out infinite;
}
.card--h1 { transform: translateX(-30px); }
.card--h2 { transform: translateX(8px); }
.card--h3 { transform: translateX(-14px); }
.card--h4 { transform: translateX(18px); }
.card--h5 { transform: translateX(-22px); }
@keyframes float {
  0%, 100% { translate: 0 0; }
  50%      { translate: 0 -10px; }
}
.card__row {
  display: flex; justify-content: space-between; align-items: center;
  font-size: 10.5px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--muted);
}
.card__idx {
  font-family: var(--serif);
  font-style: italic;
  font-weight: 300;
  font-size: 16px;
  letter-spacing: 0;
  text-transform: none;
  color: var(--text-soft);
}
.card__tag {
  padding: 2px 8px;
  border: 1px solid var(--line);
  border-radius: 999px;
  color: var(--amber);
  font-size: 9.5px;
}
.card__name {
  margin: 8px 0 10px;
  font-size: 16px;
  font-weight: 600;
  letter-spacing: -0.01em;
  color: var(--text);
}
.card__foot { color: var(--text-soft); font-size: 11px; letter-spacing: 0.06em; text-transform: none; }
.card__ver { color: var(--amber-2); }
.card__bar {
  position: relative;
  margin-top: 12px;
  height: 2px;
  background: var(--line-soft);
  overflow: hidden;
}
.card__bar i {
  position: absolute; left: 0; top: 0; bottom: 0;
  width: 38%;
  background: linear-gradient(90deg, transparent, var(--amber), transparent);
  animation: shimmer 3.6s ease-in-out infinite;
}
@keyframes shimmer {
  0%   { transform: translateX(-110%); }
  100% { transform: translateX(280%); }
}

/* ticker */
.ticker {
  position: absolute;
  left: 0; right: 0; bottom: 0;
  z-index: 2;
  border-top: 1px solid var(--line);
  background: linear-gradient(180deg, transparent, rgba(0,0,0,0.55));
  overflow: hidden;
  padding: 14px 0;
}
.ticker__track {
  display: inline-flex;
  white-space: nowrap;
  animation: scroll 38s linear infinite;
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.34em;
  text-transform: uppercase;
  color: var(--text-soft);
}
.ticker__track > span { padding-right: 28px; }
@keyframes scroll {
  from { transform: translateX(0); }
  to   { transform: translateX(-50%); }
}

/* ═══ PANEL (right) ═══════════════════════════════════════════════ */
.panel {
  position: relative;
  display: grid;
  grid-template-rows: auto 1fr auto;
  background:
    radial-gradient(900px 600px at 100% 0%, rgba(245, 165, 36, 0.05), transparent 60%),
    linear-gradient(180deg, #0e1118 0%, #0a0c10 100%);
}
.panel::before {
  content: '';
  position: absolute; inset: 24px;
  border: 1px solid var(--line-soft);
  pointer-events: none;
}
.panel__head {
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 28px 56px 12px;
  font-size: 11px;
  letter-spacing: 0.24em;
  text-transform: uppercase;
  color: var(--muted);
}
.panel__step {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: var(--text-soft);
}
.panel__step::before {
  content: '';
  width: 6px; height: 6px;
  background: var(--amber);
  border-radius: 50%;
}
.panel__locale { font-variant-numeric: tabular-nums; }

.panel__body {
  position: relative;
  padding: 16px 56px 0;
  max-width: 540px;
  width: 100%;
  margin: 0 auto;
  align-self: center;
}

.terminal {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  padding: 6px 14px;
  background: rgba(245, 165, 36, 0.06);
  border: 1px solid rgba(245, 165, 36, 0.22);
  border-radius: 2px;
  font-size: 12px;
  color: var(--text);
  animation: rise 0.7s 0.15s both ease-out;
}
.terminal .prompt { color: var(--amber); font-weight: 600; }
.terminal .cmd {
  font-family: var(--mono);
  white-space: nowrap;
  overflow: hidden;
  display: inline-block;
  animation: type 1.4s steps(20, end) 0.25s both;
}
@keyframes type { from { width: 0; } to { width: 18ch; } }
.terminal .caret {
  display: inline-block;
  width: 7px; height: 14px;
  background: var(--amber);
  animation: blink 1s steps(2, jump-none) infinite;
}
.reply {
  margin: 12px 0 0;
  font-size: 12px;
  color: var(--text-soft);
  animation: rise 0.7s 0.45s both ease-out;
}
.reply__bullet { color: var(--amber); margin-right: 6px; }

.title-row {
  display: flex;
  align-items: baseline;
  gap: 14px;
  margin-top: 38px;
  animation: rise 0.7s 0.55s both ease-out;
}
.title {
  margin: 0;
  font-family: var(--serif);
  font-style: italic;
  font-weight: 300;
  font-size: clamp(40px, 5vw, 56px);
  letter-spacing: -0.02em;
  line-height: 1;
  color: var(--text);
}
.title__index {
  font-family: var(--mono);
  font-size: 11px;
  letter-spacing: 0.28em;
  text-transform: uppercase;
  color: var(--muted);
}
.title__sub {
  margin: 12px 0 36px;
  font-size: 13.5px;
  color: var(--text-soft);
  max-width: 380px;
  animation: rise 0.7s 0.65s both ease-out;
}

/* form */
.form { display: flex; flex-direction: column; gap: 22px; }

.field {
  position: relative;
  animation: rise 0.7s both ease-out;
}
.field:nth-of-type(1) { animation-delay: 0.75s; }
.field:nth-of-type(2) { animation-delay: 0.85s; }
.field label {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin: 0 0 8px;
  font-family: var(--mono);
  font-size: 10.5px;
  font-weight: 500;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: var(--text-soft);
}
.field__num { color: var(--amber); }
.field__name { flex: 1; }
.field__aux {
  margin-left: auto;
  font-size: 10.5px;
  color: var(--muted);
  text-decoration: none;
  text-transform: lowercase;
  letter-spacing: 0.06em;
  transition: color 0.25s;
}
.field__aux:hover { color: var(--amber); }

.field input {
  width: 100%;
  margin: 0;
  padding: 10px 0 12px;
  background: transparent;
  border: 0;
  border-bottom: 1px solid var(--line);
  border-radius: 0;
  color: var(--text);
  font-family: var(--mono);
  font-size: 16px;
  letter-spacing: 0.01em;
  outline: none;
  transition: border-color 0.25s ease, color 0.25s ease;
}
.field input::placeholder { color: #3a4050; }
.field input:focus { border-bottom-color: transparent; }
.field input:-webkit-autofill {
  -webkit-text-fill-color: var(--text);
  -webkit-box-shadow: 0 0 0 1000px transparent inset;
  transition: background-color 9999s ease;
}
.field__rule {
  position: absolute; left: 0; right: 0; bottom: 0;
  height: 1px;
  pointer-events: none;
}
.field__rule i {
  display: block;
  height: 2px;
  width: 0;
  background: linear-gradient(90deg, var(--amber), var(--rust));
  transition: width 0.45s cubic-bezier(0.2, 0.8, 0.2, 1);
}
.field.is-focused .field__rule i,
.field.is-filled  .field__rule i { width: 100%; }

/* alert */
.alert {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  background: rgba(214, 90, 49, 0.10);
  border-left: 2px solid var(--rust);
  font-size: 13px;
  color: #f1c1ae;
}
.alert__mark {
  width: 22px; height: 22px;
  display: grid; place-items: center;
  background: var(--rust);
  color: var(--ink);
  border-radius: 999px;
  font-weight: 700;
  font-size: 12px;
}
.err-enter-active, .err-leave-active { transition: all 0.25s ease; }
.err-enter-from, .err-leave-to { opacity: 0; transform: translateY(-6px); }

/* primary button */
.btn-primary {
  position: relative;
  display: inline-flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  padding: 16px 22px;
  background: var(--text);
  color: var(--ink);
  border: 0;
  border-radius: 0;
  font-family: var(--mono);
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  cursor: pointer;
  overflow: hidden;
  transition: color 0.35s ease, transform 0.25s ease;
  animation: rise 0.7s 0.95s both ease-out;
}
.btn-primary::before {
  content: '';
  position: absolute; inset: 0;
  background: var(--amber);
  transform: translateX(-101%);
  transition: transform 0.55s cubic-bezier(0.2, 0.8, 0.2, 1);
  z-index: 0;
}
.btn-primary > * { position: relative; z-index: 1; }
.btn-primary:hover::before { transform: translateX(0); }
.btn-primary:hover { color: var(--ink); transform: translateY(-1px); }
.btn-primary:active { transform: translateY(0); }
.btn-primary:disabled { opacity: 0.55; cursor: progress; }
.btn-primary .arrow {
  font-family: var(--mono);
  font-size: 18px;
  transition: transform 0.35s ease;
}
.btn-primary:hover .arrow { transform: translateX(6px); }

.loading { display: inline-flex; align-items: center; gap: 6px; }
.loading i {
  width: 4px; height: 4px; border-radius: 50%;
  background: currentColor;
  animation: bounce 1.1s ease-in-out infinite;
}
.loading i:nth-child(2) { animation-delay: 0.15s; }
.loading i:nth-child(3) { animation-delay: 0.30s; }
@keyframes bounce {
  0%, 100% { transform: translateY(0); opacity: 0.6; }
  50%      { transform: translateY(-4px); opacity: 1; }
}

.aside {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: 12px;
  margin-top: 6px;
  font-size: 10.5px;
  letter-spacing: 0.32em;
  text-transform: uppercase;
  color: var(--muted);
  animation: rise 0.7s 1.05s both ease-out;
}
.rule { height: 1px; background: var(--line-soft); }

.signup {
  margin: 0;
  font-size: 13px;
  color: var(--text-soft);
  animation: rise 0.7s 1.15s both ease-out;
}
.signup__link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  margin-left: 4px;
  color: var(--text);
  text-decoration: none;
  border-bottom: 1px solid var(--amber);
  padding-bottom: 1px;
  transition: color 0.25s ease, border-color 0.25s ease;
}
.signup__link:hover { color: var(--amber); }
.signup__arrow {
  display: inline-block;
  transition: transform 0.25s ease;
}
.signup__link:hover .signup__arrow { transform: translate(2px, -2px); }

.sso-block {
  display: flex; flex-direction: column; gap: 22px;
  margin-top: 8px;
}
.sso-text {
  margin: 0;
  font-size: 14px;
  color: var(--text-soft);
  max-width: 420px;
}

.panel__foot {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 22px 56px 32px;
  font-size: 10.5px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--muted);
}
.panel__foot .spacer { flex: 1; }
.panel__foot-meta { color: var(--text-soft); }

.muted { color: var(--muted); font-family: var(--mono); }

/* ═══ Shared rise animation ══════════════════════════════════════ */
@keyframes rise {
  from { opacity: 0; transform: translateY(14px); }
  to   { opacity: 1; transform: translateY(0); }
}

/* ═══ Responsive ═════════════════════════════════════════════════ */
@media (max-width: 1100px) {
  .feed { display: none; }
  .hero { padding-right: 44px; }
}
@media (max-width: 860px) {
  .shell { grid-template-columns: 1fr; overflow-y: auto; position: absolute; }
  .stage { display: none; }
  .panel::before { inset: 12px; }
  .panel__head, .panel__body, .panel__foot { padding-left: 24px; padding-right: 24px; }
}
@media (prefers-reduced-motion: reduce) {
  *, *::before, *::after {
    animation: none !important;
    transition: none !important;
  }
}
</style>
