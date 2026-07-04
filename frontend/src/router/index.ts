import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', redirect: '/tools/json-formatter' },
    {
      path: '/tools/json-formatter',
      name: 'json-formatter',
      component: () => import('@/views/JsonFormatterView.vue'),
      meta: { title: 'JSON Formatter' },
    },
    {
      path: '/tools/hash-generator',
      name: 'hash-generator',
      component: () => import('@/views/HashGeneratorView.vue'),
      meta: { title: 'Hash Generator' },
    },
    {
      path: '/tools/qr-generator',
      name: 'qr-generator',
      component: () => import('@/views/QrGeneratorView.vue'),
      meta: { title: 'QR Generator' },
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

router.afterEach((to) => {
  const title = to.meta.title as string | undefined
  document.title = title ? `${title} · All-in-One Tools` : 'All-in-One Tools'
})

export default router
