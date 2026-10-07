// new routes: profile/mail/calendar/search/storage/privacy/terms
// router/index.js already imports and spreads this array; just export the route table.
export default [
  {
    path: '/settings/profile',
    name: 'profile',
    component: () => import('@/views/settings/Profile.vue')
  },
  {
    path: '/mail',
    name: 'mail',
    component: () => import('@/views/Mail.vue')
  },
  {
    path: '/calendar',
    name: 'calendar',
    component: () => import('@/views/Calendar.vue')
  },
  {
    path: '/search',
    name: 'search',
    component: () => import('@/views/SearchResults.vue')
  },
  {
    path: '/admin/storage',
    name: 'admin-storage',
    component: () => import('@/views/admin/StorageSettings.vue'),
    meta: { requireAdmin: true }
  },
  {
    // public page: no login
    path: '/privacy',
    name: 'privacy',
    component: () => import('@/views/Privacy.vue'),
    meta: { public: true }
  },
  {
    path: '/terms',
    name: 'terms',
    component: () => import('@/views/Terms.vue'),
    meta: { public: true }
  }
]
