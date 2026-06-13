import { createRouter, createWebHashHistory, type RouteRecordRaw } from 'vue-router'

// Hash history is required for Electron's file:// loader — HTML5 history
// API can't push to file:// paths. (Phase 9 documents this.)
const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'list',
    component: () => import('../views/ChapterListView.vue'),
  },
  {
    path: '/chapter/:id',
    name: 'chapter',
    component: () => import('../views/ChapterView.vue'),
    props: true,
  },
  { path: '/:pathMatch(.*)*', redirect: '/' },
]

export const router = createRouter({
  history: createWebHashHistory(),
  routes,
})
