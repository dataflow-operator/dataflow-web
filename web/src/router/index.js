import { createRouter, createWebHistory } from 'vue-router'
import { i18n } from '../i18n'

const routes = [
  {
    path: '/',
    name: 'dashboard',
    component: () => import('../views/DashboardView.vue'),
    meta: { titleKey: 'nav.dashboard' },
  },
  {
    path: '/manifests',
    name: 'manifests',
    component: () => import('../views/ManifestsView.vue'),
    meta: { titleKey: 'nav.manifests' },
  },
  {
    path: '/logs',
    name: 'logs',
    component: () => import('../views/LogsView.vue'),
    meta: { titleKey: 'nav.logs' },
  },
  {
    path: '/metrics',
    name: 'metrics',
    component: () => import('../views/MetricsView.vue'),
    meta: { titleKey: 'nav.metrics' },
  },
  {
    path: '/events',
    name: 'events',
    component: () => import('../views/EventsView.vue'),
    meta: { titleKey: 'nav.events' },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.afterEach((to) => {
  const titleKey = to.meta.titleKey
  const appTitle = i18n.global.t('app.title')
  document.title = titleKey ? `${i18n.global.t(titleKey)} — ${appTitle}` : appTitle
})

export default router
