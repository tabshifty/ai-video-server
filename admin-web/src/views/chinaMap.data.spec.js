import { describe, expect, it, vi } from 'vitest'
import {
  createChinaMapDataLoader,
  loadChinaMapBootstrapData,
  validateGeoJsonLayer
} from './chinaMap.data'

const layer = {
  type: 'FeatureCollection',
  features: [
    {
      type: 'Feature',
      properties: {
        region_code: 'CN-44',
        parent_region_code: 'CN',
        name: '广东省',
        level: 'province',
        has_children: true,
        full_path: '中国 / 广东省'
      },
      geometry: {
        type: 'Polygon',
        coordinates: [
          [
            [110, 20],
            [116, 20],
            [116, 25],
            [110, 20]
          ]
        ]
      }
    }
  ]
}

function jsonResponse(payload, ok = true) {
  return {
    ok,
    status: ok ? 200 : 500,
    json: vi.fn().mockResolvedValue(payload)
  }
}

describe('China map data loader', () => {
  it('keeps required startup data when the optional South China Sea inset fails', async () => {
    const catalog = { root_region_code: 'CN', regions: [] }
    const manifest = { version: 'test' }
    const loader = {
      loadCatalog: vi.fn().mockResolvedValue(catalog),
      loadManifest: vi.fn().mockResolvedValue(manifest),
      loadInset: vi.fn().mockRejectedValue(new Error('附图不可用'))
    }

    await expect(loadChinaMapBootstrapData(loader)).resolves.toEqual({
      catalog,
      manifest,
      inset: null
    })
  })

  it('loads a parent layer once and reuses the in-memory session cache', async () => {
    const fetchImpl = vi.fn().mockResolvedValue(jsonResponse(layer))
    const loader = createChinaMapDataLoader({ fetchImpl, baseUrl: '/admin/china-map/' })

    await expect(loader.loadLayer('CN')).resolves.toEqual(layer)
    await expect(loader.loadLayer('CN')).resolves.toEqual(layer)
    expect(fetchImpl).toHaveBeenCalledTimes(1)
    expect(fetchImpl).toHaveBeenCalledWith('/admin/china-map/layers/CN.geojson', {
      headers: { Accept: 'application/geo+json, application/json' }
    })
  })

  it('does not cache a failed or corrupt layer so the user can retry', async () => {
    const fetchImpl = vi
      .fn()
      .mockResolvedValueOnce(jsonResponse({ type: 'FeatureCollection', features: [] }))
      .mockResolvedValueOnce(jsonResponse(layer))
    const loader = createChinaMapDataLoader({ fetchImpl, baseUrl: '/china-map/' })

    await expect(loader.loadLayer('CN')).rejects.toThrow('地图数据损坏')
    await expect(loader.loadLayer('CN')).resolves.toEqual(layer)
    expect(fetchImpl).toHaveBeenCalledTimes(2)
  })

  it('rejects mismatched parents, invalid geometry and duplicate feature identifiers', () => {
    const duplicate = { ...layer.features[0], geometry: { type: 'Point', coordinates: [0, 0] } }
    const issues = validateGeoJsonLayer(
      { ...layer, features: [layer.features[0], duplicate] },
      'CN-OTHER'
    )

    expect(issues.some((issue) => issue.includes('父级'))).toBe(true)
    expect(issues.some((issue) => issue.includes('重复'))).toBe(true)
    expect(issues.some((issue) => issue.includes('几何'))).toBe(true)
  })

  it('encodes region codes before composing a static resource URL', async () => {
    const encodedLayer = {
      ...layer,
      features: layer.features.map((feature) => ({
        ...feature,
        properties: { ...feature.properties, parent_region_code: 'CN/HK' }
      }))
    }
    const fetchImpl = vi.fn().mockResolvedValue(jsonResponse(encodedLayer))
    const loader = createChinaMapDataLoader({ fetchImpl, baseUrl: '/china-map' })

    await loader.loadLayer('CN/HK')
    expect(fetchImpl.mock.calls[0][0]).toBe('/china-map/layers/CN%2FHK.geojson')
  })

  it('接受南海诸岛附图中来自 OSM 的点位要素', async () => {
    const inset = {
      type: 'FeatureCollection',
      features: [
        {
          type: 'Feature',
          properties: {
            region_code: 'CN-SOUTH-CHINA-SEA-1',
            parent_region_code: 'CN-SOUTH-CHINA-SEA',
            name: '西沙群岛'
          },
          geometry: { type: 'Point', coordinates: [112.0369, 16.411] }
        }
      ]
    }
    const fetchImpl = vi.fn().mockResolvedValue(jsonResponse(inset))
    const loader = createChinaMapDataLoader({ fetchImpl, baseUrl: '/china-map/' })

    await expect(loader.loadInset()).resolves.toEqual(inset)
  })
})
