import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import extras from './extras.js'

const routes = [
  { path: '/login', name: 'login', component: () => import('@/views/Login.vue') },
  { path: '/register', name: 'register', component: () => import('@/views/Register.vue') },
  { path: '/dashboard', name: 'dashboard', component: () => import('@/views/Dashboard.vue') },
  {
    path: '/editor/doc/:id',
    name: 'doc-editor',
    component: () => import('@/views/editor/DocEditor.vue')
  },
  {
    path: '/editor/sheet/:id',
    name: 'sheet-editor',
    component: () => import('@/views/editor/SheetEditor.vue')
  },
  {
    path: '/editor/slide/:id',
    name: 'slide-editor',
    component: () => import('@/views/editor/SlideEditor.vue')
  },
  {
    path: '/viewer/pdf/:id',
    name: 'pdf-viewer',
    component: () => import('@/views/ViewerPdf.vue')
  },
  {
    path: '/admin',
    name: 'admin',
    component: () => import('@/views/Admin.vue'),
    meta: { requireAdmin: true }
  },
  // trash
  {
    path: '/trash',
    name: 'trash',
    component: () => import('@/views/Trash.vue')
  },
  // shared by me
  {
    path: '/shared',
    name: 'shared-by-me',
    component: () => import('@/views/SharedByMe.vue')
  },
  // public share page: no login, guard lets through
  {
    path: '/s/:token',
    name: 'public-share',
    component: () => import('@/views/PublicShare.vue'),
    meta: { public: true }
  },
  // new-module routes (mail/calendar/2FA/search etc.); extras.js maintained elsewhere
  ...extras,
  { path: '/:pathMatch(.*)*', redirect: '/dashboard' }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// route guard: unauthenticated -> login; public pages (meta.public) pass; admin pages -> dashboard for non-admin
router.beforeEach((to) => {
  const auth = useAuthStore()
  // public pages skip auth guard
  if (to.meta.public) {
    // logged-in users hitting login/register go to dashboard; public share stays visible
    if ((to.path === '/login' || to.path === '/register') && auth.isLogin) {
      return '/dashboard'
    }
    return true
  }
  if (to.path !== '/login' && to.path !== '/register' && !auth.isLogin) {
    return '/login'
  }
  if ((to.path === '/login' || to.path === '/register') && auth.isLogin) {
    return '/dashboard'
  }
  if (to.meta.requireAdmin && !auth.isAdmin) {
    return '/dashboard'
  }
  return true
})

export default router
