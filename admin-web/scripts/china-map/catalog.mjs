import { createHash } from 'node:crypto'
import OpenCC from 'opencc-js/t2cn'

export const ROOT_REGION_CODE = 'CN'
const traditionalToSimplified = OpenCC.Converter({ from: 'tw', to: 'cn' })

export const PROVINCE_OSM_BY_CODE = Object.freeze({
  '11': 912940,
  '12': 912999,
  '13': 912998,
  '14': 913105,
  '15': 161349,
  '21': 912942,
  '22': 198590,
  '23': 199073,
  '31': 913067,
  '32': 913012,
  '33': 553302,
  '34': 913011,
  '35': 553303,
  '36': 913109,
  '37': 913006,
  '41': 407492,
  '42': 913106,
  '43': 913073,
  '44': 911844,
  '45': 286342,
  '46': 2128285,
  '50': 913069,
  '51': 913068,
  '52': 286937,
  '53': 913094,
  '54': 153292,
  '61': 913100,
  '62': 153314,
  '63': 153269,
  '64': 913101,
  '65': 153310,
  '71': 449220,
  '81': 913110,
  '82': 1867188
})

export const COUNTY_OSM_BY_SOURCE_CODE = Object.freeze({
  '210702': 5684760,
  '210703': 5684762,
  '410108': 4566399,
  '460302': 6753150,
  '460303': 6753263,
  '500157': 19798221,
  '522628': 2777837,
  '654201': 2751616,
  '654225': 2751619
})

const CURRENT_COUNTY_OVERRIDES = Object.freeze({
  '460321': { code: '460302', name: '西沙区', aliases: ['西沙群岛'] },
  '460322': { code: '460303', name: '南沙区', aliases: ['南沙群岛'] },
  '460323': null,
  '500105': { code: '500157', name: '两江新区', aliases: ['江北区', '渝北区'] },
  '500112': null
})

const REGION_PREFIX = Object.freeze({
  '71': 'CN-71',
  '81': 'CN-81',
  '82': 'CN-82'
})

export function normalizeChineseName(value) {
  return traditionalToSimplified(String(value || ''))
    .normalize('NFKC')
    .replace(/澳門/g, '澳门')
    .replace(/香港\s*Hong Kong/gi, '香港')
    .replace(/澳门\s*Macau/gi, '澳门')
    .replace(/坵/gu, '丘')
    .replace(/\s+/g, '')
    .trim()
}

function isStatisticalFunctionalZone(sourceCode) {
  const code = String(sourceCode || '')
  if (!/^\d{6}$/u.test(code)) return false
  const suffix = Number(code.slice(-2))
  return suffix >= 71 && suffix <= 80
}

export function shortChineseName(value) {
  return normalizeChineseName(value)
    .replace(/特别行政区$/u, '')
    .replace(/壮族自治区$/u, '')
    .replace(/回族自治区$/u, '')
    .replace(/维吾尔自治区.*$/u, '')
    .replace(/自治区.*$/u, '')
    .replace(/[省市县区旗]$/u, '')
}

function slugHash(parts) {
  return createHash('sha256').update(parts.join('/')).digest('hex').slice(0, 10).toUpperCase()
}

function specialRegionCode(provinceCode, path, sourceCode) {
  if (sourceCode) return `${REGION_PREFIX[provinceCode]}-${sourceCode}`
  return `${REGION_PREFIX[provinceCode]}-${slugHash(path)}`
}

function addRegion(regions, region) {
  regions.push({
    source_admin_code: null,
    source_admin_code_origin: null,
    osm_relation_id: null,
    aliases: [],
    ...region,
    child_region_codes: []
  })
  return regions.at(-1)
}

function addChild(parent, child) {
  parent.child_region_codes.push(child.region_code)
}

function normalizeCountySource(countySource) {
  const sourceCode = String(countySource.code)
  if (!Object.hasOwn(CURRENT_COUNTY_OVERRIDES, sourceCode)) {
    return {
      code: sourceCode,
      name: normalizeChineseName(countySource.name),
      aliases: [],
      source_admin_code_origin: '中华人民共和国民政部/国家统计局公开行政区划目录'
    }
  }
  const override = CURRENT_COUNTY_OVERRIDES[sourceCode]
  if (!override) return null
  return {
    ...override,
    aliases: override.aliases.map(normalizeChineseName),
    source_admin_code_origin: '国务院行政区划调整批复/中华人民共和国民政部公开信息'
  }
}

function addMainlandCounty(regions, parent, countySource) {
  if (isStatisticalFunctionalZone(countySource.code)) return
  if (String(countySource.code) === '350527') return
  const normalized = normalizeCountySource(countySource)
  if (!normalized) return
  const county = addRegion(regions, {
    region_code: `${parent.region_code}-${normalized.code}`,
    name: normalized.name,
    aliases: normalized.aliases,
    level: 'county',
    parent_region_code: parent.region_code,
    source_admin_code: normalized.code,
    source_admin_code_origin: normalized.source_admin_code_origin,
    osm_relation_id: COUNTY_OSM_BY_SOURCE_CODE[normalized.code] || null
  })
  addChild(parent, county)
}

function buildMainlandCatalog(directory) {
  const regions = []
  const root = addRegion(regions, {
    region_code: ROOT_REGION_CODE,
    name: '中国',
    level: 'country',
    parent_region_code: null,
    osm_relation_id: null
  })

  directory.forEach((provinceSource) => {
    const provinceCode = String(provinceSource.code)
    const province = addRegion(regions, {
      region_code: `${ROOT_REGION_CODE}-${provinceCode}`,
      name: normalizeChineseName(provinceSource.name),
      level: 'province',
      parent_region_code: ROOT_REGION_CODE,
      source_admin_code: provinceCode,
      source_admin_code_origin: '中华人民共和国民政部/国家统计局公开行政区划目录',
      osm_relation_id: PROVINCE_OSM_BY_CODE[provinceCode] || null
    })
    addChild(root, province)

    const directCounty = ['11', '12', '31', '50'].includes(provinceCode)
    for (const prefectureSource of provinceSource.children || []) {
      const sixDigitCounties = (prefectureSource.children || []).filter((entry) => /^\d{6}$/.test(String(entry.code)))
      const isDirectCountyGroup = /直辖县级行政区划/u.test(prefectureSource.name)
      if (directCounty || isDirectCountyGroup) {
        sixDigitCounties.forEach((countySource) => addMainlandCounty(regions, province, countySource))
        continue
      }

      const prefectureCode = String(prefectureSource.code)
      const prefecture = addRegion(regions, {
        region_code: `${province.region_code}-${prefectureCode}`,
        name: normalizeChineseName(prefectureSource.name),
        level: 'prefecture',
        parent_region_code: province.region_code,
        source_admin_code: prefectureCode,
        source_admin_code_origin: '中华人民共和国民政部/国家统计局公开行政区划目录'
      })
      addChild(province, prefecture)

      sixDigitCounties.forEach((countySource) => addMainlandCounty(regions, prefecture, countySource))
    }
  })

  return { regions, root }
}

function findRegion(regions, code) {
  const region = regions.find((entry) => entry.region_code === code)
  if (!region) throw new Error(`目录内部缺少行政区：${code}`)
  return region
}

function addNamedSpecialBranch(regions, province, branchName, children, level) {
  const branch = addRegion(regions, {
    region_code: specialRegionCode(province.source_admin_code, [province.name, branchName]),
    name: normalizeChineseName(branchName),
    level,
    parent_region_code: province.region_code
  })
  addChild(province, branch)

  children.forEach((childName) => {
    const child = addRegion(regions, {
      region_code: specialRegionCode(province.source_admin_code, [province.name, branchName, childName]),
      name: normalizeChineseName(childName).replace(/（.*?）/gu, ''),
      aliases: [normalizeChineseName(childName)],
      level: 'county',
      parent_region_code: branch.region_code
    })
    addChild(branch, child)
  })
}

const MACAU_ALIASES = Object.freeze({
  圣安多尼堂区: ['花王堂区'],
  嘉模堂区: ['氹仔'],
  圣方济各堂区: ['路环']
})

function splitDirectoryName(value) {
  const normalized = normalizeChineseName(value)
  const match = normalized.match(/^(.+?)[（(](.+?)[）)]$/u)
  if (!match) return { name: normalized, aliases: [] }
  return { name: match[1], aliases: [normalized, match[2]] }
}

function addDirectSpecialChildren(regions, province, branches) {
  Object.values(branches).flat().forEach((childName) => {
    const parsed = splitDirectoryName(childName)
    const child = addRegion(regions, {
      region_code: specialRegionCode(province.source_admin_code, [province.name, childName]),
      name: parsed.name,
      aliases: [...parsed.aliases, ...(MACAU_ALIASES[parsed.name] || [])],
      level: 'county',
      parent_region_code: province.region_code
    })
    addChild(province, child)
  })
}

const TAIWAN_SUPPLEMENTS = Object.freeze({
  金门县: ['金城镇', '金沙镇', '金湖镇', '金宁乡', '烈屿乡', '乌丘乡'],
  连江县: ['南竿乡', '北竿乡', '莒光乡', '东引乡']
})

function appendSpecialCatalog(catalog, specialDirectory) {
  const { regions, root } = catalog
  const specs = [
    ['81', '香港特别行政区'],
    ['82', '澳门特别行政区'],
    ['71', '台湾省']
  ]

  specs.forEach(([code, name]) => {
    const province = addRegion(regions, {
      region_code: REGION_PREFIX[code],
      name,
      level: 'province',
      parent_region_code: ROOT_REGION_CODE,
      source_admin_code: code,
      source_admin_code_origin: 'modood/Administrative-divisions-of-China 港澳台公开目录',
      osm_relation_id: PROVINCE_OSM_BY_CODE[code]
    })
    addChild(root, province)

    const branches = specialDirectory[name] || {}
    if (code === '71') {
      const taiwanBranches = { ...branches, ...TAIWAN_SUPPLEMENTS }
      Object.entries(taiwanBranches).forEach(([branchName, childNames]) => {
        addNamedSpecialBranch(regions, province, branchName, childNames, 'prefecture')
      })
    } else {
      addDirectSpecialChildren(regions, province, branches)
      if (code === '82') {
        const cotai = addRegion(regions, {
          region_code: specialRegionCode(code, [province.name, '路氹填海区']),
          name: '路氹填海区',
          level: 'county',
          parent_region_code: province.region_code,
          osm_relation_id: 5758867,
          source_admin_code_origin: 'OpenStreetMap 行政边界补充'
        })
        addChild(province, cotai)
      }
    }
  })
}

export function buildCatalog(mainlandDirectory, specialDirectory) {
  const catalog = buildMainlandCatalog(mainlandDirectory)
  appendSpecialCatalog(catalog, specialDirectory)
  return {
    root_region_code: ROOT_REGION_CODE,
    regions: catalog.regions
  }
}

export function matchRegionToOsm(region, candidates) {
  const sourceCode = String(region.source_admin_code || '')
  const exactCodeMatches = candidates.filter((candidate) => {
    const ref = String(candidate.tags?.['ref:admin:CN'] || '').replace(/00$/u, '')
    return sourceCode && (ref === sourceCode || ref === sourceCode.replace(/00$/u, ''))
  })
  if (exactCodeMatches.length === 1) return exactCodeMatches[0]

  const names = new Set([region.name, ...(region.aliases || [])].map(normalizeChineseName).filter(Boolean))
  const exactNameMatches = candidates.filter((candidate) => {
    const candidateNames = [
      candidate.tags?.['name:zh-Hans'],
      candidate.tags?.['name:zh'],
      candidate.tags?.name,
      candidate.tags?.alt_name
    ]
      .flatMap((value) => String(value || '').split(';'))
      .flatMap((value) => {
        const normalized = normalizeChineseName(value)
        const leadingChinese = normalized.match(/^[\p{Script=Han}·•・]+/u)?.[0]
        return leadingChinese && leadingChinese !== normalized ? [normalized, leadingChinese] : [normalized]
      })
      .filter(Boolean)
    return candidateNames.some((name) => names.has(name))
  })
  if (exactNameMatches.length === 1) return exactNameMatches[0]

  const shortNames = new Set([...names].map(shortChineseName).filter((name) => name.length >= 2))
  const shortMatches = candidates.filter((candidate) => {
    const normalized = normalizeChineseName(
      candidate.tags?.['name:zh-Hans'] || candidate.tags?.['name:zh'] || candidate.tags?.name
    )
    const name = normalized.match(/^[\p{Script=Han}·•・]+/u)?.[0] || normalized
    return shortNames.has(shortChineseName(name))
  })
  return shortMatches.length === 1 ? shortMatches[0] : null
}

export function catalogSummary(catalog) {
  return catalog.regions.reduce(
    (summary, region) => {
      summary.total += 1
      summary[region.level] = (summary[region.level] || 0) + 1
      if (region.child_region_codes.length === 0) summary.terminal += 1
      return summary
    },
    { total: 0, terminal: 0 }
  )
}
