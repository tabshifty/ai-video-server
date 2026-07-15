import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const layout = readFileSync(new URL('./Layout.vue', import.meta.url), 'utf8')
const commandPaletteHelper = readFileSync(new URL('./base/commandPalette.helpers.js', import.meta.url), 'utf8')

function findBlock(source, pattern) {
  const match = source.match(pattern)
  expect(match).not.toBeNull()
  return match?.[0] || ''
}

function findFunctionBlock(name) {
  return findBlock(layout, new RegExp(`function ${name}\\([^)]*\\) \\{[\\s\\S]*?^\\}`, 'm'))
}

describe('Layout shell', () => {
  it('uses grouped navigation and the command palette shell', () => {
    expect(layout).toContain('分组')
    expect(commandPaletteHelper).toContain('媒体库')
    expect(commandPaletteHelper).toContain('录入处理')
    expect(commandPaletteHelper).toContain('工具箱')
    expect(commandPaletteHelper).toContain('系统')
    expect(layout).toMatch(/CommandPalette|command-palette/i)
  })

  it('persists the sidebar collapse preference and profile chip', () => {
    expect(layout).toContain('admin-sidebar-collapsed')
    expect(layout).toMatch(/profile/i)
  })

  it('keeps a visible expand affordance in the collapsed sidebar', () => {
    expect(layout).toContain('.admin-shell.is-collapsed .brand-block {')
    expect(layout).toContain('flex-direction: column;')
    expect(layout).toContain('.admin-shell.is-collapsed .collapse-button {')
    expect(layout).toContain('display: inline-flex;')
    expect(layout).toContain('width: var(--space-8);')
  })

  it('keeps collapsed group toggles operable without rendering their text', () => {
    expect(layout).toContain('<span v-if="!isSidebarCollapsed">{{ group.label }}</span>')
    expect(layout).toContain(':aria-label="`${group.label}分组')
    expect(layout).toContain('.admin-shell.is-collapsed button.nav-group__label {')
  })

  it('renders the Precision Ops workspace header contract', () => {
    expect(layout).toContain('admin-recent-routes-v1')
    expect(layout).toContain('admin-nav-groups-v1')
    expect(layout).toContain('name="header-actions"')
    expect(layout).toContain('matchedNavItem.value?.groupLabel')
    expect(layout).toContain('recentNavItems')
    expect(layout).toContain('toggleGroup(group.key)')
    expect(layout).toContain('aria-expanded')
    expect(layout).toContain('var(--admin-header-height)')
  })

  it('keeps the migration compatibility boundary', () => {
    expect(layout).toContain('const showShellPageHeader = computed(() => !route.meta?.hideShellPageHeader)')
    const shellHeader = findBlock(layout, /<header class="shell-header">[\s\S]*?<\/header>/)
    const identity = findBlock(shellHeader, /<div v-if="showShellPageHeader" class="workspace-identity">[\s\S]*?<\/div>/)
    const actions = findBlock(shellHeader, /<div v-if="\$slots\['header-actions'\]" class="shell-header__actions">[\s\S]*?<\/div>/)

    expect(identity).not.toContain('header-actions')
    expect(actions).not.toContain('showShellPageHeader')
  })

  it('guards each shell preference reader with its own fallback', () => {
    const recentReader = findFunctionBlock('readStoredRecentRoutes')
    const expandedReader = findFunctionBlock('readStoredExpandedGroups')

    expect(recentReader).toContain('parseRecentRoutes(window.localStorage.getItem(RECENT_ROUTES_KEY), validNavPaths)')
    expect(recentReader).toMatch(/try \{[\s\S]*?catch \(_\) \{\s*return \[\]\s*\}/)
    expect(expandedReader).toContain('parseExpandedGroupKeys(window.localStorage.getItem(NAV_GROUPS_KEY), validGroupKeys)')
    expect(expandedReader).toMatch(/try \{[\s\S]*?catch \(_\) \{\s*return \[\.\.\.validGroupKeys\]\s*\}/)
  })

  it('updates every shell preference inside the route watcher', () => {
    const routeWatch = findBlock(layout, /watch\(\s*\(\) => route\.fullPath,[\s\S]*?^\)/m)

    expect(routeWatch).toContain('const path = route.path')
    expect(routeWatch).toContain('mobileNavVisible.value = false')
    expect(routeWatch).toContain('pushRecentRoute(recentRoutePaths.value, path, validNavPaths, 3)')
    expect(routeWatch).toContain('ensureActiveGroup(expandedGroupKeys.value, activeGroupKey.value, validGroupKeys)')
    expect(routeWatch).toContain('persistShellPreference(RECENT_ROUTES_KEY')
    expect(routeWatch).toContain('persistShellPreference(NAV_GROUPS_KEY')
    expect(routeWatch.match(/persistShellPreference\(/g) || []).toHaveLength(2)
    expect(routeWatch).toContain('{ immediate: true }')
  })

  it('names exactly the two desktop links whose text disappears when collapsed', () => {
    const desktopNav = findBlock(layout, /<nav class="nav-groups"[\s\S]*?<\/nav>/)
    const drawerNav = findBlock(layout, /<nav class="drawer-nav"[\s\S]*?<\/nav>/)
    const desktopLinks = desktopNav.match(/<RouterLink\b[\s\S]*?<\/RouterLink>/g) || []

    expect(desktopLinks).toHaveLength(2)
    for (const link of desktopLinks) {
      expect(link).toContain(':aria-label="item.label"')
    }
    expect(layout.match(/:aria-label="item\.label"/g) || []).toHaveLength(2)
    expect(drawerNav).not.toContain(':aria-label="item.label"')
  })

  it('uses shared expanded groups and precise responsive workspace spacing', () => {
    expect(layout.match(/v-for="group in navGroups"/g) || []).toHaveLength(2)
    expect(layout.match(/v-show="isGroupExpanded\(group.key\)"/g) || []).toHaveLength(2)
    expect(layout).toContain('height: var(--admin-header-height);')
    expect(layout).toContain('padding: var(--space-5);')
    expect(layout).toContain('@media (max-width: 63.9375rem)')
    expect(layout).toContain('padding: var(--space-4);')
    expect(layout).toContain('@media (max-width: 47.9375rem)')
    expect(layout).toContain('padding: var(--space-3);')
    expect(layout).toContain('min-height: 44px;')
    expect(layout).not.toContain('letter-spacing: 0.08em;')
  })

  it('removes the legacy rose gradient and admin subtitle copy', () => {
    expect(layout).not.toContain('#881337')
    expect(layout).not.toContain('#7f1d1d')
    expect(layout).not.toContain('linear-gradient(180deg, #881337')
    expect(layout).not.toContain("'管理员工作台'")
  })
})
