import { createRouter, createWebHistory, type RouteLocationNormalized } from 'vue-router'
import HomeView from './views/HomeView.vue'
import { adminToken } from './composables/useAdminSession'

export type NavSection = 'articles' | 'topics' | 'archive' | 'about'

declare module 'vue-router' {
  interface RouteMeta {
    section?: NavSection
    requiresAdmin?: boolean
  }
}

const findHashTarget = (to: RouteLocationNormalized) => to.hash && document.getElementById(to.hash.slice(1))

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/', name: 'home', component: HomeView, meta: { section: 'articles' } },
    {
      path: '/articles/:slug',
      name: 'article',
      component: () => import('./views/ArticleView.vue'),
      meta: { section: 'articles' },
    },
    { path: '/archive', name: 'archive', component: () => import('./views/ArchiveView.vue'), meta: { section: 'archive' } },
    { path: '/topics', name: 'topics', component: () => import('./views/TopicsView.vue'), meta: { section: 'topics' } },
    {
      path: '/topics/:topic',
      name: 'topic',
      component: () => import('./views/TopicView.vue'),
      meta: { section: 'topics' },
    },
    { path: '/about', name: 'about', component: () => import('./views/AboutView.vue'), meta: { section: 'about' } },
    {
      path: '/admin',
      component: () => import('./views/admin/AdminShell.vue'),
      children: [
        { path: 'login', name: 'admin-login', component: () => import('./views/admin/AdminLoginView.vue') },
        { path: '', redirect: { name: 'admin-articles' } },
        { path: 'articles', name: 'admin-articles', component: () => import('./views/admin/AdminArticlesView.vue'), meta: { requiresAdmin: true } },
        { path: 'articles/new', name: 'admin-article-new', component: () => import('./views/admin/AdminArticleFormView.vue'), meta: { requiresAdmin: true } },
        { path: 'articles/:slug/edit', name: 'admin-article-edit', component: () => import('./views/admin/AdminArticleFormView.vue'), meta: { requiresAdmin: true } },
        { path: 'topics', name: 'admin-topics', component: () => import('./views/admin/AdminTopicsView.vue'), meta: { requiresAdmin: true } },
      ],
    },
    { path: '/:pathMatch(.*)*', name: 'not-found', component: () => import('./views/NotFoundView.vue') },
  ],
  scrollBehavior(to, from, savedPosition) {
    if (savedPosition) return savedPosition
    if (findHashTarget(to)) return { el: to.hash, behavior: 'smooth' }
    if (to.path === from.path) return false
    // `html { scroll-behavior: smooth }` would otherwise animate every page change.
    return { top: 0, behavior: 'instant' }
  },
})

router.beforeEach((to) => {
  const needsAdmin = to.matched.some((record) => record.meta.requiresAdmin)
  if (needsAdmin && !adminToken()) {
    return { name: 'admin-login', query: { redirect: to.fullPath } }
  }
  if (to.name === 'admin-login' && adminToken()) {
    return { name: 'admin-articles' }
  }
  return true
})
