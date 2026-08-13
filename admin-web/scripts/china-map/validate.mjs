import { createHash } from 'node:crypto'
import { readFile, readdir, stat } from 'node:fs/promises'
import { basename, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { validateGeoJsonLayer } from '../../src/views/chinaMap.data.js'
import { validateRegionCatalog } from '../../src/views/chinaMap.helpers.js'
import { SOUTH_CHINA_SEA_ELEMENTS } from './generation.mjs'

const EXPECTED_PROVINCE_CODES = new Set([
  'CN-11',
  'CN-12',
  'CN-13',
  'CN-14',
  'CN-15',
  'CN-21',
  'CN-22',
  'CN-23',
  'CN-31',
  'CN-32',
  'CN-33',
  'CN-34',
  'CN-35',
  'CN-36',
  'CN-37',
  'CN-41',
  'CN-42',
  'CN-43',
  'CN-44',
  'CN-45',
  'CN-46',
  'CN-50',
  'CN-51',
  'CN-52',
  'CN-53',
  'CN-54',
  'CN-61',
  'CN-62',
  'CN-63',
  'CN-64',
  'CN-65',
  'CN-71',
  'CN-81',
  'CN-82'
])

async function readJson(path) {
  return JSON.parse(await readFile(path, 'utf8'))
}

async function sha256(path) {
  const bytes = await readFile(path)
  return createHash('sha256').update(bytes).digest('hex')
}

function sorted(values) {
  return [...values].sort((left, right) => left.localeCompare(right))
}

export async function validateSnapshot(snapshotDirectory) {
  const issues = []
  let manifest
  let catalog

  try {
    manifest = await readJson(join(snapshotDirectory, 'manifest.json'))
  } catch (error) {
    return { issues: [`无法读取 manifest.json：${error.message}`], summary: null }
  }
  try {
    catalog = await readJson(join(snapshotDirectory, 'catalog.json'))
  } catch (error) {
    return { issues: [`无法读取 catalog.json：${error.message}`], summary: null }
  }

  if (manifest.status !== 'complete') issues.push('地图清单状态不是 complete')
  if (!manifest.snapshot_timestamp) issues.push('地图清单缺少固定快照时间')
  if (!String(manifest.geometry_source || '').includes('OpenStreetMap')) {
    issues.push('地图清单未声明 OpenStreetMap 几何来源')
  }
  if (!String(manifest.license || '').includes('ODbL')) issues.push('地图清单未声明 ODbL 许可')
  issues.push(...validateRegionCatalog(catalog))

  const matchReportPath = join(snapshotDirectory, 'match-report.json')
  let matchReport
  try {
    matchReport = await readJson(matchReportPath)
    if (!Array.isArray(matchReport.missing) || matchReport.missing.length) {
      issues.push('OSM 行政区关系匹配仍有缺失')
    }
  } catch (error) {
    issues.push(`缺少或无法读取行政区关系匹配报告：${error.message}`)
  }

  const byCode = new Map((catalog.regions || []).map((region) => [region.region_code, region]))
  const root = byCode.get(catalog.root_region_code)
  const actualProvinceCodes = new Set(root?.child_region_codes || [])
  const missingProvinces = sorted([...EXPECTED_PROVINCE_CODES].filter((code) => !actualProvinceCodes.has(code)))
  const unexpectedProvinces = sorted([...actualProvinceCodes].filter((code) => !EXPECTED_PROVINCE_CODES.has(code)))
  if (missingProvinces.length) issues.push(`省级范围缺失：${missingProvinces.join('、')}`)
  if (unexpectedProvinces.length) issues.push(`存在未审核的省级范围：${unexpectedProvinces.join('、')}`)

  const nonRootRegions = (catalog.regions || []).filter((region) => region.region_code !== catalog.root_region_code)
  const missingOsmRelations = nonRootRegions.filter((region) => !Number.isInteger(region.osm_relation_id))
  if (missingOsmRelations.length) {
    issues.push(`行政区缺少 OSM 关系：${missingOsmRelations.slice(0, 10).map((region) => region.region_code).join('、')}`)
  }
  const matchedCodes = new Set((matchReport?.matched || []).map((entry) => entry.region_code))
  const unreportedCodes = nonRootRegions.filter((region) => !matchedCodes.has(region.region_code))
  if (unreportedCodes.length) {
    issues.push(`行政区未进入 OSM 匹配报告：${unreportedCodes.slice(0, 10).map((region) => region.region_code).join('、')}`)
  }

  const nonTerminalRegions = (catalog.regions || []).filter(
    (region) => Array.isArray(region.child_region_codes) && region.child_region_codes.length > 0
  )
  let featureCount = 0
  for (const parent of nonTerminalRegions) {
    const layerPath = join(snapshotDirectory, 'layers', `${parent.region_code}.geojson`)
    let layer
    try {
      layer = await readJson(layerPath)
    } catch (error) {
      issues.push(`缺少或无法读取图层 ${parent.region_code}：${error.message}`)
      continue
    }

    const layerIssues = validateGeoJsonLayer(layer, parent.region_code)
    issues.push(...layerIssues.map((issue) => `${parent.region_code}：${issue}`))
    const actualCodes = new Set((layer.features || []).map((feature) => feature.properties?.region_code).filter(Boolean))
    const expectedCodes = new Set(parent.child_region_codes)
    const missing = sorted([...expectedCodes].filter((code) => !actualCodes.has(code)))
    const extra = sorted([...actualCodes].filter((code) => !expectedCodes.has(code)))
    if (missing.length) issues.push(`${parent.region_code} 图层缺少子级：${missing.join('、')}`)
    if (extra.length) issues.push(`${parent.region_code} 图层包含目录外子级：${extra.join('、')}`)
    featureCount += layer.features?.length || 0
  }

  try {
    const inset = await readJson(join(snapshotDirectory, 'south-china-sea.geojson'))
    issues.push(
      ...validateGeoJsonLayer(inset, 'CN-SOUTH-CHINA-SEA', ['Point']).map(
        (issue) => `南海诸岛附图：${issue}`
      )
    )
    const insetSources = new Set(
      (inset.features || []).map(
        (feature) => `${feature.properties?.osm_element_type}/${feature.properties?.osm_element_id}`
      )
    )
    SOUTH_CHINA_SEA_ELEMENTS.forEach((entry) => {
      if (!insetSources.has(`${entry.type}/${entry.id}`)) {
        issues.push(`南海诸岛附图缺少固定 OSM 来源：${entry.type}/${entry.id}`)
      }
    })
  } catch (error) {
    issues.push(`缺少或无法读取南海诸岛附图：${error.message}`)
  }

  const files = Array.isArray(manifest.files) ? manifest.files : []
  const manifestPaths = new Set(files.map((entry) => entry.path))
  const requiredPaths = [
    'catalog.json',
    'match-report.json',
    'south-china-sea.geojson',
    ...nonTerminalRegions.map((region) => `layers/${region.region_code}.geojson`)
  ]
  requiredPaths.forEach((path) => {
    if (!manifestPaths.has(path)) issues.push(`清单未覆盖必需文件：${path}`)
  })
  for (const entry of files) {
    const path = join(snapshotDirectory, entry.path)
    try {
      const details = await stat(path)
      if (details.size !== entry.bytes) issues.push(`文件大小与清单不一致：${entry.path}`)
      if ((await sha256(path)) !== entry.sha256) issues.push(`文件哈希与清单不一致：${entry.path}`)
    } catch (error) {
      issues.push(`清单文件不存在：${entry.path}`)
    }
  }

  const layerFileNames = await readdir(join(snapshotDirectory, 'layers')).catch(() => [])
  if (layerFileNames.length !== nonTerminalRegions.length) {
    issues.push(`图层文件数不一致：期望 ${nonTerminalRegions.length}，实际 ${layerFileNames.length}`)
  }

  const terminalCount = (catalog.regions || []).filter((region) => !region.child_region_codes?.length).length
  const summary = {
    regions: catalog.regions?.length || 0,
    provinces: actualProvinceCodes.size,
    terminal_regions: terminalCount,
    layers: layerFileNames.length,
    features: featureCount,
    snapshot: manifest.snapshot_timestamp || null
  }

  if (manifest.counts) {
    for (const [key, actual] of Object.entries({
      regions: summary.regions,
      provinces: summary.provinces,
      terminal_regions: summary.terminal_regions,
      layers: summary.layers,
      features: summary.features
    })) {
      if (manifest.counts[key] !== actual) issues.push(`清单计数不一致：${key}`)
    }
  }

  return { issues: [...new Set(issues)], summary }
}

async function main() {
  const defaultDirectory = fileURLToPath(new URL('../../public/china-map', import.meta.url))
  const target = process.argv[2] || defaultDirectory
  const result = await validateSnapshot(target)
  if (result.issues.length) {
    console.error(`中国行政区划地图快照校验失败（${result.issues.length} 项）：`)
    result.issues.forEach((issue) => console.error(`- ${issue}`))
    process.exitCode = 1
    return
  }
  console.log(`中国行政区划地图快照校验通过：${basename(target)}`)
  console.log(JSON.stringify(result.summary, null, 2))
}

if (process.argv[1] && fileURLToPath(import.meta.url) === process.argv[1]) {
  await main()
}
