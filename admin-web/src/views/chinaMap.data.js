import { validateRegionCatalog } from './chinaMap.helpers.js'

function normalizeBaseUrl(baseUrl) {
  const value = String(baseUrl || '/')
  return value.endsWith('/') ? value : `${value}/`
}

function hasFiniteCoordinates(value) {
  if (!Array.isArray(value) || value.length === 0) return false
  if (typeof value[0] === 'number') {
    return value.length >= 2 && value.every(Number.isFinite)
  }
  return value.every(hasFiniteCoordinates)
}

export function validateGeoJsonLayer(
  layer,
  expectedParentCode,
  allowedGeometryTypes = ['Polygon', 'MultiPolygon']
) {
  const issues = []
  if (layer?.type !== 'FeatureCollection') {
    return ['地图数据不是 FeatureCollection']
  }
  if (!Array.isArray(layer.features) || layer.features.length === 0) {
    return ['地图数据不包含行政区要素']
  }

  const seen = new Set()
  layer.features.forEach((feature, index) => {
    const properties = feature?.properties || {}
    const code = String(properties.region_code || '').trim()
    const label = code || `第 ${index + 1} 个要素`

    if (feature?.type !== 'Feature') {
      issues.push(`${label} 不是 GeoJSON Feature`)
    }
    if (!code) {
      issues.push(`${label} 缺少 region_code`)
    } else if (seen.has(code)) {
      issues.push(`地图要素标识重复：${code}`)
    } else {
      seen.add(code)
    }
    if (!String(properties.name || '').trim()) {
      issues.push(`${label} 缺少中文名称`)
    }
    if (properties.parent_region_code !== expectedParentCode) {
      issues.push(`${label} 的父级与图层不一致`)
    }
    if (!allowedGeometryTypes.includes(feature?.geometry?.type)) {
      issues.push(`${label} 的几何类型无效`)
    } else if (!hasFiniteCoordinates(feature.geometry.coordinates)) {
      issues.push(`${label} 的几何坐标无效`)
    }
  })

  return issues
}

function createResourceError(message, cause) {
  const error = new Error(message)
  if (cause) error.cause = cause
  return error
}

export async function loadChinaMapBootstrapData(loader) {
  const [catalog, manifest, inset] = await Promise.all([
    loader.loadCatalog(),
    loader.loadManifest(),
    loader.loadInset().catch(() => null)
  ])
  return { catalog, manifest, inset }
}

export function createChinaMapDataLoader({ fetchImpl = globalThis.fetch, baseUrl = '/' } = {}) {
  if (typeof fetchImpl !== 'function') {
    throw new TypeError('缺少可用的 fetch 实现')
  }

  const resourceBaseUrl = normalizeBaseUrl(baseUrl)
  const cache = new Map()

  async function loadJson(cacheKey, relativePath, accept, validate) {
    if (cache.has(cacheKey)) return cache.get(cacheKey)

    const pending = (async () => {
      let response
      try {
        response = await fetchImpl(`${resourceBaseUrl}${relativePath}`, {
          headers: { Accept: accept }
        })
      } catch (error) {
        throw createResourceError('地图数据加载失败，请检查网络后重试', error)
      }

      if (!response?.ok) {
        throw createResourceError(`地图数据加载失败（HTTP ${response?.status || '未知'}）`)
      }

      let payload
      try {
        payload = await response.json()
      } catch (error) {
        throw createResourceError('地图数据无法解析，请重试或返回上一级', error)
      }

      const issues = validate(payload)
      if (issues.length) {
        throw createResourceError(`地图数据损坏：${issues[0]}`)
      }
      return payload
    })()

    cache.set(cacheKey, pending)
    try {
      const payload = await pending
      cache.set(cacheKey, payload)
      return payload
    } catch (error) {
      cache.delete(cacheKey)
      throw error
    }
  }

  return {
    loadManifest() {
      return loadJson('manifest', 'manifest.json', 'application/json', (manifest) => {
        const issues = []
        if (!String(manifest?.dataset_name || '').trim()) issues.push('清单缺少数据集名称')
        if (!String(manifest?.version || '').trim()) issues.push('清单缺少版本')
        if (!String(manifest?.license || '').trim()) issues.push('清单缺少许可说明')
        return issues
      })
    },
    loadCatalog() {
      return loadJson('catalog', 'catalog.json', 'application/json', validateRegionCatalog)
    },
    loadLayer(parentRegionCode) {
      const code = String(parentRegionCode || '').trim()
      if (!code) return Promise.reject(new TypeError('缺少图层父级 region_code'))
      return loadJson(
        `layer:${code}`,
        `layers/${encodeURIComponent(code)}.geojson`,
        'application/geo+json, application/json',
        (layer) => validateGeoJsonLayer(layer, code)
      )
    },
    loadInset() {
      return loadJson(
        'inset',
        'south-china-sea.geojson',
        'application/geo+json, application/json',
        (layer) => validateGeoJsonLayer(layer, ROOT_INSET_PARENT_CODE, ['Point'])
      )
    },
    clear() {
      cache.clear()
    }
  }
}

const ROOT_INSET_PARENT_CODE = 'CN-SOUTH-CHINA-SEA'
