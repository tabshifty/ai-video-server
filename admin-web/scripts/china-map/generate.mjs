import { createHash } from 'node:crypto'
import { execFile } from 'node:child_process'
import { mkdir, readFile, readdir, rename, rm, stat, writeFile } from 'node:fs/promises'
import { tmpdir } from 'node:os'
import { dirname, join, resolve } from 'node:path'
import { promisify } from 'node:util'
import { fileURLToPath } from 'node:url'
import osmtogeojson from 'osmtogeojson'
import { buildCatalog, catalogSummary, matchRegionToOsm } from './catalog.mjs'
import {
  SOUTH_CHINA_SEA_ELEMENTS,
  assertOverpassSnapshot,
  buildSouthChinaSeaGeoJson,
  childRelationCandidates,
  collectRootRelationFeatures,
  createParentPayloadLoader,
  ensureRelationGeometryDependencies,
  ensureRelationGeometries,
  geometryQuery,
  groupParentsByDepth,
  selectRelationGeometryPayload,
  southChinaSeaQuery
} from './generation.mjs'
import { validateSnapshot } from './validate.mjs'

const execFileAsync = promisify(execFile)
const SNAPSHOT_TIMESTAMP = '2026-08-12T00:00:00Z'
const SNAPSHOT_VERSION = '2026.08.12-osm'
const DIRECTORY_COMMIT = 'c49d495b40ac73eb1a66f6eeae5f8fd10696f035'
const DIRECTORY_BASE_URLS = [
  `https://cdn.jsdelivr.net/gh/modood/Administrative-divisions-of-China@${DIRECTORY_COMMIT}/dist`,
  `https://raw.githubusercontent.com/modood/Administrative-divisions-of-China/${DIRECTORY_COMMIT}/dist`
]
const DIRECTORY_SHA256 = Object.freeze({
  'pca-code': '83b7536f853ad16beb4d37b92890a3fd7bb9d33d4f37e7c8885fb948749a9bc4',
  'HK-MO-TW': '546616aa6ec7239d9bcf02cb063539652604162b1ab74ef65293537c54ea3534'
})
const OVERPASS_ENDPOINTS = [
  'https://overpass-api.de/api/interpreter',
  'https://overpass.openstreetmap.fr/api/interpreter',
  'https://maps.mail.ru/osm/tools/overpass/api/interpreter'
]
const REQUEST_USER_AGENT = 'ai-video-server-china-map-generator/1.0 (+https://github.com/tabshifty/ai-video-server)'
const SCRIPT_DIRECTORY = dirname(fileURLToPath(import.meta.url))
const ADMIN_WEB_DIRECTORY = resolve(SCRIPT_DIRECTORY, '../..')
const PUBLIC_DIRECTORY = join(ADMIN_WEB_DIRECTORY, 'public', 'china-map')
const CACHE_DIRECTORY = join(tmpdir(), `ai-video-server-china-map-${SNAPSHOT_VERSION}`)
const STAGING_DIRECTORY = join(dirname(PUBLIC_DIRECTORY), `.china-map-staging-${process.pid}`)
const MAPSHAPER = join(ADMIN_WEB_DIRECTORY, 'node_modules', '.bin', 'mapshaper')
const CATALOG_ONLY = process.argv.includes('--catalog-only')
let overpassEndpointCursor = 0
const overpassEndpointLocks = new Map()
const nextOverpassRequestAt = new Map()

function sha(value) {
  return createHash('sha256').update(value).digest('hex')
}

function sleep(milliseconds) {
  return new Promise((resolvePromise) => setTimeout(resolvePromise, milliseconds))
}

async function cachedJson(cacheName, loader, validate = () => {}) {
  const path = join(CACHE_DIRECTORY, cacheName)
  try {
    const payload = JSON.parse(await readFile(path, 'utf8'))
    validate(payload)
    return payload
  } catch {
    const payload = await loader()
    validate(payload)
    await mkdir(dirname(path), { recursive: true })
    await writeFile(path, `${JSON.stringify(payload)}\n`, 'utf8')
    return payload
  }
}

async function fetchJson(fileName, label, expectedSha256) {
  const cached = await cachedJson(
    `directory/${label}-${expectedSha256}.json`,
    async () => {
      let lastError
      for (let attempt = 0; attempt < 6; attempt += 1) {
        const url = `${DIRECTORY_BASE_URLS[attempt % DIRECTORY_BASE_URLS.length]}/${fileName}`
        const controller = new AbortController()
        const timeout = setTimeout(() => controller.abort(), 60_000)
        try {
          const response = await fetch(url, {
            headers: { Accept: 'application/json', 'User-Agent': REQUEST_USER_AGENT },
            signal: controller.signal
          })
          if (!response.ok) throw new Error(`HTTP ${response.status}`)
          const body = await response.text()
          const actualSha256 = sha(body)
          if (actualSha256 !== expectedSha256) {
            throw new Error(`来源哈希不匹配：期望 ${expectedSha256}，实际 ${actualSha256}`)
          }
          return { source_sha256: actualSha256, data: JSON.parse(body) }
        } catch (error) {
          lastError = error
          await sleep(Math.min(8_000, 1_000 * 2 ** attempt))
        } finally {
          clearTimeout(timeout)
        }
      }
      throw new Error(`${label} 下载失败：${lastError?.message || '未知错误'}`)
    },
    (payload) => {
      if (payload?.source_sha256 !== expectedSha256 || !payload.data) {
        throw new Error(`${label} 缓存未经来源哈希校验`)
      }
    }
  )
  return cached.data
}

async function queryOverpass(query, cacheName) {
  return cachedJson(`overpass/${cacheName}.json`, async () => {
    let lastError
    for (let attempt = 0; attempt < 9; attempt += 1) {
      const endpoint = OVERPASS_ENDPOINTS[overpassEndpointCursor % OVERPASS_ENDPOINTS.length]
      overpassEndpointCursor += 1
      const previousRequest = overpassEndpointLocks.get(endpoint) || Promise.resolve()
      let releaseEndpoint
      const currentRequest = new Promise((resolvePromise) => {
        releaseEndpoint = resolvePromise
      })
      overpassEndpointLocks.set(endpoint, currentRequest)
      try {
        await previousRequest.catch(() => {})
        const waitMilliseconds = Math.max(0, (nextOverpassRequestAt.get(endpoint) || 0) - Date.now())
        if (waitMilliseconds) await sleep(waitMilliseconds)
        nextOverpassRequestAt.set(endpoint, Date.now() + 750)
        const controller = new AbortController()
        const timeout = setTimeout(() => controller.abort(), 240_000)
        const response = await fetch(endpoint, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded;charset=UTF-8',
            'User-Agent': REQUEST_USER_AGENT
          },
          body: new URLSearchParams({ data: query }),
          signal: controller.signal
        })
        try {
          const body = await response.text()
          if (!response.ok || body.trimStart().startsWith('<')) {
            throw new Error(`HTTP ${response.status}：${body.replace(/<[^>]+>/g, ' ').replace(/\s+/g, ' ').slice(0, 180)}`)
          }
          const payload = JSON.parse(body)
          assertOverpassSnapshot(payload, SNAPSHOT_TIMESTAMP)
          return payload
        } finally {
          clearTimeout(timeout)
        }
      } catch (error) {
        lastError = error
        await sleep(Math.min(12_000, 1_000 * 2 ** Math.min(attempt, 3)))
      } finally {
        releaseEndpoint()
        if (overpassEndpointLocks.get(endpoint) === currentRequest) {
          overpassEndpointLocks.delete(endpoint)
        }
      }
    }
    throw new Error(`${cacheName} 查询失败：${lastError?.message || '未知错误'}`)
  }, (payload) => assertOverpassSnapshot(payload, SNAPSHOT_TIMESTAMP))
}

function fullPath(byCode, region) {
  const names = []
  const visited = new Set()
  let current = region
  while (current && !visited.has(current.region_code)) {
    names.unshift(current.name)
    visited.add(current.region_code)
    current = current.parent_region_code ? byCode.get(current.parent_region_code) : null
  }
  return names.join(' / ')
}

const loadParentPayload = createParentPayloadLoader(queryOverpass, SNAPSHOT_TIMESTAMP)

async function resolveOsmRelations(catalog) {
  const byCode = new Map(catalog.regions.map((region) => [region.region_code, region]))
  const groups = groupParentsByDepth(catalog)
  const queue = groups.flat()
  const root = byCode.get('CN')
  const report = {
    matched: root.child_region_codes.map((code) => {
      const region = byCode.get(code)
      return { region_code: code, osm_relation_id: region.osm_relation_id, mode: 'fixed' }
    }),
    missing: [],
    extras: []
  }

  let completed = 0
  async function resolveParent(parent) {
    const result = { matched: [], missing: [], extras: [] }
    if (!parent.osm_relation_id) {
      result.missing.push({ region_code: parent.region_code, name: parent.name, reason: '父级缺少 OSM 关系' })
      return result
    }
    const payload = await loadParentPayload(parent)
    const candidates = childRelationCandidates(payload, parent.osm_relation_id)
    const unused = new Map(candidates.map((candidate) => [candidate.id, candidate]))

    for (const childCode of parent.child_region_codes) {
      const child = byCode.get(childCode)
      if (child.osm_relation_id) {
        unused.delete(child.osm_relation_id)
        result.matched.push({ region_code: child.region_code, osm_relation_id: child.osm_relation_id, mode: 'fixed' })
        continue
      }
      const match = matchRegionToOsm(child, [...unused.values()])
      if (!match) {
        result.missing.push({
          region_code: child.region_code,
          name: child.name,
          parent_region_code: parent.region_code,
          candidates: [...unused.values()].map((entry) => ({ id: entry.id, name: entry.tags?.['name:zh-Hans'] || entry.tags?.['name:zh'] || entry.tags?.name }))
        })
        continue
      }
      child.osm_relation_id = match.id
      unused.delete(match.id)
      result.matched.push({ region_code: child.region_code, osm_relation_id: match.id, mode: 'catalog' })
    }

    if (unused.size) {
      result.extras.push({
        parent_region_code: parent.region_code,
        regions: [...unused.values()].map((entry) => ({
          osm_relation_id: entry.id,
          name: entry.tags?.['name:zh-Hans'] || entry.tags?.['name:zh'] || entry.tags?.name,
          admin_level: entry.tags?.admin_level
        }))
      })
    }
    completed += 1
    process.stdout.write(`\r匹配行政关系 ${completed}/${queue.length}：${parent.name}                    `)
    return result
  }

  for (const group of groups) {
    for (let index = 0; index < group.length; index += 3) {
      const results = await Promise.all(group.slice(index, index + 3).map(resolveParent))
      results.forEach((result) => {
        report.matched.push(...result.matched)
        report.missing.push(...result.missing)
        report.extras.push(...result.extras)
      })
    }
  }
  process.stdout.write('\n')
  return report
}

async function simplifyLayer(inputPath, outputPath, percentage) {
  await execFileAsync(MAPSHAPER, [
    inputPath,
    '-simplify',
    `${percentage}%`,
    'keep-shapes',
    '-clean',
    '-o',
    'format=geojson',
    'precision=0.0001',
    outputPath
  ], { maxBuffer: 4 * 1024 * 1024 })
}

function convertedRelations(payload, relationIds) {
  const selectedPayload = selectRelationGeometryPayload(payload, relationIds)
  const converted = osmtogeojson(selectedPayload, { flatProperties: true })
  return new Map(
    converted.features
      .filter((feature) => feature.id?.startsWith('relation/'))
      .map((feature) => [Number(feature.id.slice('relation/'.length)), feature])
  )
}

async function loadRootRelationFeatures(children, layerIndex, layerCount) {
  let completed = 0
  return collectRootRelationFeatures(children, async (group) => {
    completed += group.length
    process.stdout.write(
      `\r生成图层 ${layerIndex}/${layerCount}：中国（省界 ${completed}/${children.length}）                    `
    )
    const ids = group.map((child) => child.osm_relation_id)
    const payload = await queryOverpass(
      geometryQuery(ids, SNAPSHOT_TIMESTAMP),
      `root-geometry/${sha(ids.join(','))}`
    )
    return convertedRelations(payload, ids)
  })
}

async function buildLayer(parent, byCode, layerIndex, layerCount) {
  const children = parent.child_region_codes.map((code) => byCode.get(code))
  const ids = children.map((child) => child.osm_relation_id)
  process.stdout.write(`\r生成图层 ${layerIndex}/${layerCount}：${parent.name}                    `)
  let convertedById
  if (parent.region_code === 'CN') {
    convertedById = await loadRootRelationFeatures(children, layerIndex, layerCount)
  } else {
    let payload = await loadParentPayload(parent)
    payload = await ensureRelationGeometries(payload, ids, (missingIds) =>
      queryOverpass(
        geometryQuery(missingIds, SNAPSHOT_TIMESTAMP),
        `geometry/${sha(missingIds.join(','))}`
      )
    )
    payload = await ensureRelationGeometryDependencies(payload, ids, (missingIds) =>
      queryOverpass(
        geometryQuery(missingIds, SNAPSHOT_TIMESTAMP),
        `geometry-dependencies/${sha(missingIds.join(','))}`
      )
    )
    convertedById = convertedRelations(payload, ids)
  }

  const features = children.map((child) => {
    const feature = convertedById.get(child.osm_relation_id)
    if (!feature || !['Polygon', 'MultiPolygon'].includes(feature.geometry?.type)) {
      throw new Error(`${child.region_code} 缺少有效的 OSM 面几何`)
    }
    return {
      type: 'Feature',
      properties: {
        region_code: child.region_code,
        parent_region_code: parent.region_code,
        name: child.name,
        level: child.level,
        has_children: child.child_region_codes.length > 0,
        full_path: fullPath(byCode, child),
        source_admin_code: child.source_admin_code,
        osm_relation_id: child.osm_relation_id
      },
      geometry: feature.geometry
    }
  })

  const rawPath = join(CACHE_DIRECTORY, 'layers-raw', `${parent.region_code}.geojson`)
  const outputPath = join(STAGING_DIRECTORY, 'layers', `${parent.region_code}.geojson`)
  await mkdir(dirname(rawPath), { recursive: true })
  await mkdir(dirname(outputPath), { recursive: true })
  await writeFile(rawPath, `${JSON.stringify({ type: 'FeatureCollection', features })}\n`, 'utf8')
  const percentage = parent.region_code === 'CN' ? 3 : 8
  await simplifyLayer(rawPath, outputPath, percentage)
}

async function buildSouthChinaSeaInset() {
  const payload = await queryOverpass(southChinaSeaQuery(SNAPSHOT_TIMESTAMP), 'south-china-sea/elements')
  const geoJson = buildSouthChinaSeaGeoJson(payload)
  await writeFile(
    join(STAGING_DIRECTORY, 'south-china-sea.geojson'),
    `${JSON.stringify(geoJson)}\n`,
    'utf8'
  )
}

async function fileEntry(path, relativePath) {
  const bytes = await readFile(path)
  return {
    path: relativePath,
    bytes: bytes.length,
    sha256: createHash('sha256').update(bytes).digest('hex')
  }
}

async function collectFiles(directory, prefix = '') {
  const entries = []
  for (const name of await readdir(directory)) {
    const path = join(directory, name)
    const relativePath = prefix ? `${prefix}/${name}` : name
    const details = await stat(path)
    if (details.isDirectory()) entries.push(...(await collectFiles(path, relativePath)))
    else if (relativePath !== 'manifest.json') entries.push(await fileEntry(path, relativePath))
  }
  return entries.sort((left, right) => left.path.localeCompare(right.path))
}

async function publishSnapshot() {
  const backup = `${PUBLIC_DIRECTORY}.previous-${process.pid}`
  await mkdir(dirname(PUBLIC_DIRECTORY), { recursive: true })
  let hadPrevious = false
  try {
    await rename(PUBLIC_DIRECTORY, backup)
    hadPrevious = true
  } catch (error) {
    if (error.code !== 'ENOENT') throw error
  }
  try {
    await rename(STAGING_DIRECTORY, PUBLIC_DIRECTORY)
    if (hadPrevious) await rm(backup, { recursive: true, force: true })
  } catch (error) {
    if (hadPrevious) await rename(backup, PUBLIC_DIRECTORY)
    throw error
  }
}

async function main() {
  await mkdir(CACHE_DIRECTORY, { recursive: true })
  await rm(STAGING_DIRECTORY, { recursive: true, force: true })
  await mkdir(STAGING_DIRECTORY, { recursive: true })

  const [mainland, special] = await Promise.all([
    fetchJson('pca-code.json', 'pca-code', DIRECTORY_SHA256['pca-code']),
    fetchJson('HK-MO-TW.json', 'HK-MO-TW', DIRECTORY_SHA256['HK-MO-TW'])
  ])
  const catalog = buildCatalog(mainland, special)
  console.log('行政目录：', catalogSummary(catalog))
  const matchReport = await resolveOsmRelations(catalog)
  await writeFile(join(STAGING_DIRECTORY, 'match-report.json'), `${JSON.stringify(matchReport, null, 2)}\n`, 'utf8')
  await writeFile(join(STAGING_DIRECTORY, 'catalog.json'), `${JSON.stringify(catalog)}\n`, 'utf8')

  if (matchReport.missing.length) {
    console.error(`存在 ${matchReport.missing.length} 个未匹配行政区，报告：${join(STAGING_DIRECTORY, 'match-report.json')}`)
    process.exitCode = 1
    return
  }
  if (CATALOG_ONLY) {
    console.log(`行政目录与 OSM 关系匹配通过：${matchReport.matched.length} 个区域`)
    console.log(`报告：${join(STAGING_DIRECTORY, 'match-report.json')}`)
    return
  }

  const byCode = new Map(catalog.regions.map((region) => [region.region_code, region]))
  const parents = catalog.regions.filter((region) => region.child_region_codes.length > 0)
  for (let index = 0; index < parents.length; index += 1) {
    await buildLayer(parents[index], byCode, index + 1, parents.length)
  }
  process.stdout.write('\n')
  await buildSouthChinaSeaInset()

  const summary = catalogSummary(catalog)
  const featureCount = catalog.regions.length - 1
  const manifest = {
    status: 'complete',
    dataset_name: '中国行政区划地图 OSM 离线快照',
    version: SNAPSHOT_VERSION,
    snapshot_timestamp: SNAPSHOT_TIMESTAMP,
    geometry_source: 'OpenStreetMap 固定时间点行政边界',
    directory_source: `modood/Administrative-divisions-of-China @ ${DIRECTORY_COMMIT}`,
    directory_sha256: DIRECTORY_SHA256,
    south_china_sea_osm_elements: SOUTH_CHINA_SEA_ELEMENTS,
    license: 'Open Database License (ODbL) 1.0',
    attribution: '© OpenStreetMap contributors；行政目录来自 modood/Administrative-divisions-of-China。衍生边界数据库按 ODbL 1.0 提供。',
    generated_at: new Date().toISOString(),
    counts: {
      regions: catalog.regions.length,
      provinces: byCode.get('CN').child_region_codes.length,
      terminal_regions: summary.terminal,
      layers: parents.length,
      features: featureCount
    },
    files: await collectFiles(STAGING_DIRECTORY)
  }
  await writeFile(join(STAGING_DIRECTORY, 'manifest.json'), `${JSON.stringify(manifest, null, 2)}\n`, 'utf8')

  const validation = await validateSnapshot(STAGING_DIRECTORY)
  if (validation.issues.length) {
    console.error('生成结果未通过完整性门禁：')
    validation.issues.forEach((issue) => console.error(`- ${issue}`))
    process.exitCode = 1
    return
  }

  await publishSnapshot()
  console.log('中国行政区划地图快照已发布到 admin-web/public/china-map')
  console.log(JSON.stringify(validation.summary, null, 2))
}

await main().catch((error) => {
  console.error(error.stack || error.message)
  process.exitCode = 1
})
