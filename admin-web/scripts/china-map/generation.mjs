export const SOUTH_CHINA_SEA_ELEMENTS = Object.freeze([
  { name: '西沙群岛', type: 'relation', id: 2417005 },
  { name: '中沙群岛', type: 'relation', id: 3959723 },
  { name: '南沙群岛', type: 'relation', id: 4856843 },
  { name: '曾母暗沙', type: 'way', id: 890873502 }
])

const HONG_KONG_DISTRICT_BBOX = '22.14,113.82,22.58,114.52'

export function parentLayerQuery(parent, snapshotTimestamp) {
  const prefix = `[out:json][date:"${snapshotTimestamp}"][timeout:300];`
  const selector =
    parent.region_code === 'CN-81'
      ? `relation["boundary"="administrative"]["admin_level"~"^(5|6|7|8)$"](${HONG_KONG_DISTRICT_BBOX});`
      : `rel(${parent.osm_relation_id});rel(r:"subarea");`
  return `${prefix}${selector}out body;>;out skel qt;`
}

export function geometryQuery(ids, snapshotTimestamp) {
  return `[out:json][date:"${snapshotTimestamp}"][timeout:300];rel(id:${ids.join(',')});out body;>>;out skel qt;`
}

export function southChinaSeaQuery(snapshotTimestamp) {
  const selectors = SOUTH_CHINA_SEA_ELEMENTS.map((entry) => `${entry.type}(${entry.id});`).join('')
  return `[out:json][date:"${snapshotTimestamp}"][timeout:120];(${selectors});out ids tags center;`
}

export function assertOverpassSnapshot(payload, snapshotTimestamp) {
  if (!Array.isArray(payload?.elements)) throw new Error('Overpass 响应缺少 elements')
  if (payload.remark) throw new Error(`Overpass 响应不完整：${payload.remark}`)

  const dataTimestamp = Date.parse(payload.osm3s?.timestamp_osm_base || '')
  const requiredTimestamp = Date.parse(snapshotTimestamp)
  if (!Number.isFinite(dataTimestamp)) throw new Error('Overpass 响应缺少有效数据基线时间')
  if (!Number.isFinite(requiredTimestamp)) throw new Error('地图快照时间无效')
  if (dataTimestamp < requiredTimestamp) {
    throw new Error(
      `Overpass 数据基线 ${payload.osm3s.timestamp_osm_base} 早于快照 ${snapshotTimestamp}`
    )
  }
}

export function createParentPayloadLoader(queryOverpass, snapshotTimestamp) {
  const cache = new Map()
  return (parent) => {
    const cacheKey = `${parent.region_code}:${parent.osm_relation_id || 'bbox'}`
    if (!cache.has(cacheKey)) {
      cache.set(
        cacheKey,
        queryOverpass(
          parentLayerQuery(parent, snapshotTimestamp),
          `layers/${parent.region_code}-${parent.osm_relation_id || 'bbox'}`
        )
      )
    }
    return cache.get(cacheKey)
  }
}

export function childRelationCandidates(payload, parentRelationId) {
  return (Array.isArray(payload?.elements) ? payload.elements : []).filter(
    (element) =>
      element.type === 'relation' &&
      element.tags &&
      element.id !== parentRelationId &&
      ['administrative', 'political'].includes(element.tags.boundary)
  )
}

export async function ensureRelationGeometries(payload, relationIds, loadMissing) {
  const expectedIds = [...new Set(relationIds)]
  const presentIds = new Set(
    (payload?.elements || [])
      .filter((element) => element.type === 'relation')
      .map((element) => element.id)
  )
  const missingIds = expectedIds.filter((id) => !presentIds.has(id))
  if (!missingIds.length) return payload

  const supplemental = await loadMissing(missingIds)
  return mergeElementPayloads(payload, supplemental)
}

function mergeElementPayloads(base, supplemental) {
  const elementsByIdentity = new Map()
  for (const element of [...(base?.elements || []), ...(supplemental?.elements || [])]) {
    const identity = `${element.type}/${element.id}`
    const existing = elementsByIdentity.get(identity)
    elementsByIdentity.set(identity, existing ? { ...existing, ...element, tags: element.tags || existing.tags } : element)
  }
  return { ...base, elements: [...elementsByIdentity.values()] }
}

function missingNestedRelationIds(payload, relationIds) {
  const elements = Array.isArray(payload?.elements) ? payload.elements : []
  const byIdentity = new Map(elements.map((element) => [`${element.type}/${element.id}`, element]))
  const missing = new Set()
  const visited = new Set()

  function visit(relationId) {
    if (visited.has(relationId)) return
    visited.add(relationId)
    const relation = byIdentity.get(`relation/${relationId}`)
    if (!relation) {
      missing.add(relationId)
      return
    }
    ;(relation.members || [])
      .filter((member) => member.type === 'relation' && isGeometryMember(member))
      .forEach((member) => visit(member.ref))
  }

  relationIds.forEach(visit)
  return [...missing]
}

export async function ensureRelationGeometryDependencies(payload, relationIds, loadMissing) {
  let merged = payload
  for (let depth = 0; depth < 8; depth += 1) {
    const missingIds = missingNestedRelationIds(merged, relationIds)
    if (!missingIds.length) return merged
    const supplemental = await loadMissing(missingIds)
    merged = mergeElementPayloads(merged, supplemental)
  }
  throw new Error('OSM 嵌套关系几何依赖超过最大深度')
}

function isGeometryMember(member) {
  const role = String(member?.role || '')
  if (['subarea', 'admin_centre', 'admin_center', 'label'].includes(role)) return false
  if (member?.type === 'node') return ['', 'outer', 'inner'].includes(role)
  return ['way', 'relation'].includes(member?.type)
}

export function selectRelationGeometryPayload(payload, relationIds) {
  const elements = Array.isArray(payload?.elements) ? payload.elements : []
  const byIdentity = new Map()
  for (const element of elements) {
    const identity = `${element.type}/${element.id}`
    const existing = byIdentity.get(identity)
    byIdentity.set(identity, existing ? { ...existing, ...element, tags: element.tags || existing.tags } : element)
  }
  const selected = []
  const visited = new Set()

  function visit(type, id) {
    const identity = `${type}/${id}`
    if (visited.has(identity)) return
    const element = byIdentity.get(identity)
    if (!element) throw new Error(`OSM 几何依赖缺失：${identity}`)
    visited.add(identity)

    if (type === 'relation') {
      const members = (element.members || []).filter(isGeometryMember)
      selected.push({ ...element, members })
      members.forEach((member) => visit(member.type, member.ref))
      return
    }
    selected.push(element)
    if (type === 'way') {
      ;(element.nodes || []).forEach((nodeId) => visit('node', nodeId))
    }
  }

  ;[...new Set(relationIds)].forEach((id) => visit('relation', id))
  return { ...payload, elements: selected }
}

export async function collectRootRelationFeatures(children, loadRelationGroup) {
  const features = new Map()
  const groups = []
  for (let index = 0; index < children.length; index += 3) {
    groups.push(children.slice(index, index + 3))
  }
  for (let index = 0; index < groups.length; index += 3) {
    const loadedGroups = await Promise.all(groups.slice(index, index + 3).map(loadRelationGroup))
    loadedGroups.forEach((loaded) => {
      loaded.forEach((feature, relationId) => features.set(relationId, feature))
    })
  }
  return features
}

export function groupParentsByDepth(catalog) {
  const regions = Array.isArray(catalog?.regions) ? catalog.regions : []
  const byCode = new Map(regions.map((region) => [region.region_code, region]))
  const rootCode = catalog?.root_region_code
  const groups = []

  regions.forEach((region) => {
    if (region.region_code === rootCode || !region.child_region_codes?.length) return
    let depth = 0
    let current = region
    const visited = new Set()
    while (current?.parent_region_code && !visited.has(current.region_code)) {
      visited.add(current.region_code)
      depth += 1
      current = byCode.get(current.parent_region_code)
    }
    if (!current || current.region_code !== rootCode) {
      throw new Error(`无法计算行政区目录深度：${region.region_code}`)
    }
    ;(groups[depth] ||= []).push(region)
  })

  return groups.filter(Boolean)
}

function elementCoordinates(element) {
  const latitude = element?.center?.lat ?? element?.lat
  const longitude = element?.center?.lon ?? element?.lon
  if (!Number.isFinite(latitude) || !Number.isFinite(longitude)) return null
  return [longitude, latitude]
}

export function buildSouthChinaSeaGeoJson(payload) {
  const elements = Array.isArray(payload?.elements) ? payload.elements : []
  const byIdentity = new Map(elements.map((element) => [`${element.type}/${element.id}`, element]))

  const features = SOUTH_CHINA_SEA_ELEMENTS.map((source, index) => {
    const element = byIdentity.get(`${source.type}/${source.id}`)
    const coordinates = elementCoordinates(element)
    if (!coordinates) {
      throw new Error(`南海诸岛附图缺少 OSM ${source.type}/${source.id} 中心点`)
    }
    const names = [element.tags?.['name:zh-Hans'], element.tags?.['name:zh'], element.tags?.name]
      .filter(Boolean)
      .join(' ')
    if (!names.includes(source.name)) {
      throw new Error(`OSM ${source.type}/${source.id} 与预期名称“${source.name}”不匹配`)
    }

    return {
      type: 'Feature',
      properties: {
        region_code: `CN-SOUTH-CHINA-SEA-${index + 1}`,
        parent_region_code: 'CN-SOUTH-CHINA-SEA',
        name: source.name,
        level: 'inset',
        has_children: false,
        full_path: `中国 / 南海诸岛 / ${source.name}`,
        source_admin_code: null,
        osm_element_type: source.type,
        osm_element_id: source.id,
        osm_relation_id: source.type === 'relation' ? source.id : null
      },
      geometry: { type: 'Point', coordinates }
    }
  })

  return { type: 'FeatureCollection', features }
}
