import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { resolveRouterHistoryBase } from './historyBase'
import Login from '../views/Login.vue'
import Dashboard from '../views/Dashboard.vue'
import VideoList from '../views/VideoList.vue'
import VideoUpload from '../views/VideoUpload.vue'
import TvSeriesManage from '../views/TvSeriesManage.vue'
import IPTVManage from '../views/IPTVManage.vue'
import ScrapePreview from '../views/ScrapePreview.vue'
import AVManualScrape from '../views/AVManualScrape.vue'
import UserManage from '../views/UserManage.vue'
import ActorManage from '../views/ActorManage.vue'
import CollectionManage from '../views/CollectionManage.vue'
import ImageManage from '../views/ImageManage.vue'
import ImageCollectionManage from '../views/ImageCollectionManage.vue'
import SystemSettings from '../views/SystemSettings.vue'
import TaskMonitor from '../views/TaskMonitor.vue'
import Toolbox from '../views/Toolbox.vue'
import ToolboxEd2k from '../views/ToolboxEd2k.vue'
import ToolboxImageWorkbench from '../views/ToolboxImageWorkbench.vue'

const routes = [
  { path: '/login', component: Login, meta: { public: true } },
  { path: '/', redirect: '/dashboard' },
  { path: '/dashboard', component: Dashboard, meta: { hideShellPageHeader: true } },
  { path: '/videos', component: VideoList, meta: { hideShellPageHeader: true } },
  { path: '/tv-series', component: TvSeriesManage, meta: { hideShellPageHeader: true } },
  { path: '/iptv', component: IPTVManage, meta: { hideShellPageHeader: true } },
  { path: '/upload', component: VideoUpload, meta: { hideShellPageHeader: true } },
  { path: '/scrape', component: ScrapePreview, meta: { hideShellPageHeader: true } },
  { path: '/av-scrape', component: AVManualScrape, meta: { hideShellPageHeader: true } },
  { path: '/actors', component: ActorManage, meta: { hideShellPageHeader: true } },
  { path: '/collections', component: CollectionManage, meta: { hideShellPageHeader: true } },
  { path: '/images', component: ImageManage, meta: { hideShellPageHeader: true } },
  { path: '/image-collections', component: ImageCollectionManage, meta: { hideShellPageHeader: true } },
  { path: '/users', component: UserManage, meta: { hideShellPageHeader: true } },
  { path: '/toolbox', component: Toolbox, meta: { hideShellPageHeader: true } },
  { path: '/toolbox/ed2k', component: ToolboxEd2k },
  { path: '/toolbox/image-workbench', component: ToolboxImageWorkbench },
  { path: '/settings', component: SystemSettings, meta: { hideShellPageHeader: true } },
  { path: '/tasks', component: TaskMonitor, meta: { hideShellPageHeader: true } }
]

const router = createRouter({
  history: createWebHistory(resolveRouterHistoryBase(import.meta.env.BASE_URL)),
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
