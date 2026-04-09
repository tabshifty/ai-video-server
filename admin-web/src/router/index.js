import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import Login from '../views/Login.vue'
import Dashboard from '../views/Dashboard.vue'
import VideoList from '../views/VideoList.vue'
import VideoUpload from '../views/VideoUpload.vue'
import ScrapePreview from '../views/ScrapePreview.vue'
import UserManage from '../views/UserManage.vue'
import SystemSettings from '../views/SystemSettings.vue'
import TaskMonitor from '../views/TaskMonitor.vue'

const routes = [
  { path: '/login', component: Login, meta: { public: true } },
  { path: '/', redirect: '/dashboard' },
  { path: '/dashboard', component: Dashboard },
  { path: '/videos', component: VideoList },
  { path: '/upload', component: VideoUpload },
  { path: '/scrape', component: ScrapePreview },
  { path: '/users', component: UserManage },
  { path: '/settings', component: SystemSettings },
  { path: '/tasks', component: TaskMonitor }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (to.meta.public) return true
  if (!auth.isAuthed) return '/login'
  if (!auth.isAdmin) {
    auth.logout()
    return '/login'
  }
  return true
})

export default router
