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
    {
      path: '/tools/pdf',
      name: 'pdf-toolkit',
      component: () => import('@/views/PdfToolsView.vue'),
      meta: { title: 'PDF Toolkit' },
    },
    {
      path: '/tools/excel',
      name: 'excel-toolkit',
      component: () => import('@/views/ExcelToolsView.vue'),
      meta: { title: 'Excel Toolkit' },
    },
    {
      path: '/tools/markdown',
      name: 'markdown-toolkit',
      component: () => import('@/views/MarkdownToolsView.vue'),
      meta: { title: 'Markdown Toolkit' },
    },
    {
      path: '/tools/file-security',
      name: 'file-security',
      component: () => import('@/views/FileSecurityView.vue'),
      meta: { title: 'File Encryption' },
    },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

router.afterEach((to) => {
  const title = to.meta.title as string | undefined
  document.title = title ? `${title} · All-in-One Tools` : 'All-in-One Tools'
})

export default router
