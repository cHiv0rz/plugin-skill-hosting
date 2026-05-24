import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from './stores/auth'

declare module 'vue-router' {
  interface RouteMeta {
    requiresAuth?: boolean
    hideChrome?: boolean
  }
}

export const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: () => import('./views/PluginListView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/login',
      component: () => import('./views/LoginView.vue'),
      meta: { hideChrome: true },
    },
    {
      path: '/register',
      component: () => import('./views/RegisterView.vue'),
    },
    {
      path: '/auth/callback',
      component: () => import('./views/OIDCCallbackView.vue'),
    },
    {
      path: '/developers',
      component: () => import('./views/DevelopersView.vue'),
    },
    {
      path: '/users',
      component: () => import('./views/UsersView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/pending',
      component: () => import('./views/PendingView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/plugins/new',
      component: () => import('./views/NewPluginView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/plugins/:name',
      component: () => import('./views/PluginDetailView.vue'),
      props: true,
      meta: { requiresAuth: true },
    },
    {
      path: '/plugins/:name/skills/new',
      component: () => import('./views/SkillEditView.vue'),
      props: route => ({ pluginName: route.params.name, skillName: null }),
      meta: { requiresAuth: true },
    },
    {
      path: '/plugins/:name/skills/:skillName/edit',
      component: () => import('./views/SkillEditView.vue'),
      props: route => ({
        pluginName: route.params.name,
        skillName: route.params.skillName,
      }),
      meta: { requiresAuth: true },
    },
  ],
})

router.beforeEach(async (to) => {
  if (to.meta.requiresAuth) {
    const auth = useAuthStore()
    if (!auth.token) {
      return { path: '/login', query: { redirect: to.fullPath } }
    }
    // Pending / rejected users are confined to /pending until an existing
    // user approves them (or until they log out from there).
    const status = auth.user?.status
    if (status && status !== 'approved' && to.path !== '/pending') {
      return { path: '/pending' }
    }
    if (status === 'approved' && to.path === '/pending') {
      return { path: '/' }
    }
    // /users is admin-only, and additionally hidden in OIDC + Google Workspace
    // mode where membership is managed by the upstream `hd` filter. ensureMode
    // is awaited because direct deep-links can arrive before auth config has
    // been fetched and userApprovalRequired would otherwise be unset.
    if (to.path === '/users') {
      await auth.ensureMode()
      const userMgmtOn = auth.mode !== 'oidc' || auth.userApprovalRequired
      if (!auth.user?.isAdmin || !userMgmtOn) {
        return { path: '/' }
      }
    }
  }
})
