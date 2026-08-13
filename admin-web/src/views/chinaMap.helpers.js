export const ROOT_REGION_CODE = 'CN'

function normalizeText(value) {
  return String(value ?? '')
    .trim()
    .toLocaleLowerCase('zh-CN')
    .replace(/\s+/g, ' ')
}

function regionChildren(region) {
  return Array.isArray(region?.child_region_codes) ? region.child_region_codes : []
}

function buildPath(byCode, regionCode) {
  const path = []
  const visited = new Set()
  let current = byCode.get(regionCode)

  while (current && !visited.has(current.region_code)) {
    path.unshift(current)
    visited.add(current.region_code)
    current = current.parent_region_code ? byCode.get(current.parent_region_code) : null
  }

  return path
}

export function validateRegionCatalog(catalog) {
  const issues = []
  const regions = Array.isArray(catalog?.regions) ? catalog.regions : []
  const rootRegionCode = catalog?.root_region_code || ROOT_REGION_CODE
  const byCode = new Map()

  if (!regions.length) {
    return ['行政区目录为空']
  }

  regions.forEach((region, index) => {
    const code = String(region?.region_code || '').trim()
    if (!code) {
      issues.push(`第 ${index + 1} 个行政区缺少 region_code`)
      return
    }
    if (byCode.has(code)) {
      issues.push(`行政区标识重复：${code}`)
      return
    }
    if (!String(region?.name || '').trim()) {
      issues.push(`行政区缺少中文名称：${code}`)
    }
    if (!Array.isArray(region?.child_region_codes)) {
      issues.push(`行政区子级列表无效：${code}`)
    }
    byCode.set(code, region)
  })

  if (!byCode.has(rootRegionCode)) {
    issues.push(`缺少根行政区：${rootRegionCode}`)
  }

  regions.forEach((region) => {
    const code = region?.region_code
    if (!code || !byCode.has(code)) return

    if (code === rootRegionCode) {
      if (region.parent_region_code) {
        issues.push(`根行政区不应设置父级：${code}`)
      }
    } else if (!region.parent_region_code || !byCode.has(region.parent_region_code)) {
      issues.push(`行政区父级不存在：${code}`)
    }

    regionChildren(region).forEach((childCode) => {
      const child = byCode.get(childCode)
      if (!child) {
        issues.push(`行政区子级不存在：${code} -> ${childCode}`)
      } else if (child.parent_region_code !== code) {
        issues.push(`行政区父子关系不一致：${code} -> ${childCode}`)
      }
    })
  })

  const reachable = new Set()
  const active = new Set()
  function visit(code) {
    if (active.has(code)) {
      issues.push(`行政区层级存在循环：${code}`)
      return
    }
    if (reachable.has(code)) return
    const region = byCode.get(code)
    if (!region) return

    active.add(code)
    reachable.add(code)
    regionChildren(region).forEach(visit)
    active.delete(code)
  }
  visit(rootRegionCode)

  byCode.forEach((_, code) => {
    if (!reachable.has(code)) {
      issues.push(`行政区从根节点不可达：${code}`)
    }
  })

  return [...new Set(issues)]
}

export function buildRegionIndex(regions) {
  const byCode = new Map()
  ;(Array.isArray(regions) ? regions : []).forEach((region) => {
    if (region?.region_code && !byCode.has(region.region_code)) {
      byCode.set(region.region_code, { ...region })
    }
  })

  byCode.forEach((region, code) => {
    const path = buildPath(byCode, code)
    region.path = path.map((entry) => entry.region_code)
    region.full_path = path.map((entry) => entry.name).join(' / ')
    region.has_children = regionChildren(region).length > 0
  })

  return {
    byCode,
    regions: [...byCode.values()]
  }
}

export function findRegionView(index, requestedRegionCode) {
  const byCode = index?.byCode || new Map()
  const target = byCode.get(requestedRegionCode) || byCode.get(ROOT_REGION_CODE)
  if (!target) return null

  const hasChildren = regionChildren(target).length > 0
  const layerParent = hasChildren
    ? target
    : byCode.get(target.parent_region_code) || byCode.get(ROOT_REGION_CODE) || target

  return {
    region_code: target.region_code,
    region: target,
    layer_parent_code: layerParent.region_code,
    layer_parent: layerParent,
    selected_region_code: hasChildren ? null : target.region_code,
    breadcrumbs: buildPath(byCode, target.region_code)
  }
}

export function getParentRegionCode(index, regionCode) {
  const region = index?.byCode?.get(regionCode)
  return region?.parent_region_code || null
}

export function canGoBackFromView(view) {
  return Boolean(view?.region_code && view.region_code !== ROOT_REGION_CODE)
}

export function searchRegions(index, query, limit = 20) {
  const normalizedQuery = normalizeText(query)
  if (!normalizedQuery) return []

  const tokens = normalizedQuery.split(' ').filter(Boolean)
  return (index?.regions || [])
    .filter((region) => region.region_code !== ROOT_REGION_CODE)
    .map((region) => {
      const name = normalizeText(region.name)
      const code = normalizeText(region.region_code)
      const sourceCode = normalizeText(region.source_admin_code)
      const aliases = Array.isArray(region.aliases) ? region.aliases.map(normalizeText).join(' ') : ''
      const fullPath = normalizeText(region.full_path)
      const haystack = `${name} ${code} ${sourceCode} ${aliases} ${fullPath}`

      if (!tokens.every((token) => haystack.includes(token))) return null

      let score = 40
      if (name === normalizedQuery) score = 0
      else if (code === normalizedQuery || sourceCode === normalizedQuery) score = 5
      else if (name.startsWith(normalizedQuery)) score = 10
      else if (fullPath.includes(normalizedQuery)) score = 20

      return { ...region, score }
    })
    .filter(Boolean)
    .sort((left, right) => left.score - right.score || left.full_path.localeCompare(right.full_path, 'zh-CN'))
    .slice(0, Math.max(0, limit))
    .map(({ score: _score, ...region }) => region)
}

export function buildRegionQuery(currentQuery, regionCode) {
  const nextQuery = { ...(currentQuery || {}) }
  if (!regionCode || regionCode === ROOT_REGION_CODE) {
    delete nextQuery.region_code
  } else {
    nextQuery.region_code = regionCode
  }
  return nextQuery
}
