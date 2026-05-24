export const adminShellNavGroups = [
  {
    key: 'overview',
    label: '仪表盘',
    items: [
      {
        path: '/dashboard',
        label: '仪表盘',
        title: '系统仪表盘',
        icon: 'DataAnalysis',
        alias: 'dashboard overview home ybp'
      }
    ]
  },
  {
    key: 'media',
    label: '媒体库',
    items: [
      { path: '/videos', label: '视频管理', title: '视频管理', icon: 'Film', alias: 'video videos sj spgl' },
      { path: '/tv-series', label: '电视剧管理', title: '电视剧管理', icon: 'Files', alias: 'tv series dsj' },
      { path: '/actors', label: '演员管理', title: '演员管理', icon: 'Avatar', alias: 'actor actors yy' },
      { path: '/collections', label: '合集管理', title: '合集管理', icon: 'List', alias: 'collection collections hj' },
      { path: '/images', label: '图片管理', title: '图片管理', icon: 'PictureFilled', alias: 'image images tp' },
      { path: '/image-collections', label: '图片合集', title: '图片合集', icon: 'Files', alias: 'image collection tphj' }
    ]
  },
  {
    key: 'ingest',
    label: '录入处理',
    items: [
      { path: '/upload', label: '上传视频', title: '上传中心', icon: 'UploadFilled', alias: 'upload uploader spsc' },
      { path: '/scrape', label: '通用刮削', title: '通用刮削', icon: 'MagicStick', alias: 'scrape tmdb ty' },
      { path: '/av-scrape', label: 'AV 手动刮削', title: 'AV 手动刮削', icon: 'MagicStick', alias: 'av scrape manual' }
    ]
  },
  {
    key: 'service',
    label: '服务',
    items: [
      { path: '/iptv', label: 'IPTV 管理', title: 'IPTV 管理', icon: 'Monitor', alias: 'iptv live' },
      { path: '/tasks', label: '任务监控', title: '任务监控', icon: 'List', alias: 'task tasks jobs rw' }
    ]
  },
  {
    key: 'system',
    label: '系统',
    items: [
      { path: '/users', label: '用户管理', title: '用户管理', icon: 'User', alias: 'user users yh' },
      { path: '/settings', label: '系统设置', title: '系统设置', icon: 'Setting', alias: 'settings system xt' }
    ]
  }
]

export const adminShellNavItems = adminShellNavGroups.flatMap((group) =>
  group.items.map((item) => ({ ...item, groupKey: group.key, groupLabel: group.label }))
)

const PINYIN_INITIAL = {
  仪: 'y',
  表: 'b',
  盘: 'p',
  媒: 'm',
  体: 't',
  库: 'k',
  视: 's',
  频: 'p',
  管: 'g',
  理: 'l',
  电: 'd',
  剧: 'j',
  演: 'y',
  员: 'y',
  合: 'h',
  集: 'j',
  图: 't',
  片: 'p',
  录: 'l',
  入: 'r',
  处: 'c',
  上: 's',
  传: 'c',
  中: 'z',
  心: 'x',
  通: 't',
  用: 'y',
  刮: 'g',
  册: 'c',
  历: 'l',
  史: 's',
  日: 'r',
  志: 'z',
  削: 'x',
  手: 's',
  动: 'd',
  服: 'f',
  务: 'w',
  任: 'r',
  监: 'j',
  控: 'k',
  系: 'x',
  统: 't',
  设: 's',
  置: 'z',
  户: 'h'
}

export function normalizeQuery(query) {
  return String(query || '')
    .trim()
    .toLowerCase()
}

export function toPinyinInitials(label) {
  return [...String(label || '')]
    .map((ch) => PINYIN_INITIAL[ch] || ch.toLowerCase())
    .join('')
}

export function matchPinyinInitial(label, query) {
  const q = normalizeQuery(query)
  if (!q) return true
  return toPinyinInitials(label).includes(q)
}

export function matchMenuItem(query, item) {
  const q = normalizeQuery(query)
  if (!q) return { score: 0, matched: true }
  const path = String(item?.path || '').toLowerCase()
  const label = String(item?.label || '')
  const alias = String(item?.alias || '').toLowerCase()
  if (path.startsWith(q)) return { score: 100, matched: true }
  if (label.includes(q)) return { score: 80, matched: true }
  if (alias.includes(q)) return { score: 60, matched: true }
  if (matchPinyinInitial(label, q)) return { score: 40, matched: true }
  return { score: 0, matched: false }
}

export function searchMenuItems(query, items = adminShellNavItems) {
  return items
    .map((item, index) => ({ ...item, ...matchMenuItem(query, item), index }))
    .filter((item) => item.matched)
    .sort((a, b) => b.score - a.score || a.index - b.index)
}

let commandPaletteOpener = null

export function registerCommandPaletteOpener(opener) {
  commandPaletteOpener = typeof opener === 'function' ? opener : null
}

export function openCommandPalette() {
  commandPaletteOpener?.()
}

export function findAdminNavItemByPath(path) {
  return adminShellNavItems
    .slice()
    .sort((a, b) => b.path.length - a.path.length)
    .find((item) => path === item.path || String(path || '').startsWith(`${item.path}/`))
}
