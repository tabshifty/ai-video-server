import { describe, expect, it, vi } from 'vitest'
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
  parentLayerQuery,
  selectRelationGeometryPayload,
  southChinaSeaQuery
} from './generation.mjs'

const snapshot = '2026-08-12T00:00:00Z'

describe('中国地图数据查询与附图生成', () => {
  it('将子关系元数据与几何合并为每个父级一次完整查询', async () => {
    const query = parentLayerQuery({ region_code: 'CN-44', osm_relation_id: 911844 }, snapshot)
    expect(query).toContain(`[date:"${snapshot}"]`)
    expect(query).toContain('rel(911844);rel(r:"subarea");')
    expect(query).toContain('out body;>;out skel qt;')

    const queryOverpass = vi.fn().mockResolvedValue({ elements: [] })
    const load = createParentPayloadLoader(queryOverpass, snapshot)
    const parent = { region_code: 'CN-44', osm_relation_id: 911844 }
    await load(parent)
    await load(parent)
    expect(queryOverpass).toHaveBeenCalledTimes(1)
  })

  it('为香港区级边界使用固定范围的同次完整查询', () => {
    const query = parentLayerQuery({ region_code: 'CN-81', osm_relation_id: 913110 }, snapshot)
    expect(query).toContain('22.14,113.82,22.58,114.52')
    expect(query).toContain('out body;>;out skel qt;')
  })

  it('从完整查询响应中排除节点、道路和父关系', () => {
    const candidates = childRelationCandidates(
      {
        elements: [
          { type: 'relation', id: 10, tags: { boundary: 'administrative', name: '父级' } },
          { type: 'relation', id: 11, tags: { boundary: 'administrative', name: '子级' } },
          { type: 'relation', id: 12, tags: { type: 'route', name: '公交线' } },
          { type: 'way', id: 13, tags: { highway: 'primary' } }
        ]
      },
      10
    )
    expect(candidates.map((candidate) => candidate.id)).toEqual([11])
  })

  it('只补载父级查询中缺少的固定关系几何并按元素身份去重', async () => {
    const payload = {
      elements: [
        { type: 'relation', id: 10, tags: { boundary: 'administrative' } },
        { type: 'node', id: 100, lat: 1, lon: 1 }
      ]
    }
    const loadMissing = vi.fn().mockResolvedValue({
      elements: [
        { type: 'relation', id: 20, tags: { boundary: 'administrative' } },
        { type: 'node', id: 100, lat: 2, lon: 2 },
        { type: 'node', id: 200, lat: 3, lon: 3 }
      ]
    })

    const merged = await ensureRelationGeometries(payload, [10, 20, 20], loadMissing)

    expect(loadMissing).toHaveBeenCalledWith([20])
    expect(merged.elements.filter((element) => element.type === 'node' && element.id === 100)).toHaveLength(1)
    expect(merged.elements.find((element) => element.type === 'node' && element.id === 100)?.lat).toBe(2)
    expect(merged.elements.some((element) => element.type === 'relation' && element.id === 20)).toBe(true)

    await ensureRelationGeometries(payload, [10], loadMissing)
    expect(loadMissing).toHaveBeenCalledTimes(1)
  })

  it('递归补载目标边界缺少的嵌套关系几何', async () => {
    const payload = {
      elements: [
        {
          type: 'relation',
          id: 10,
          members: [{ type: 'relation', ref: 20, role: 'outer' }]
        }
      ]
    }
    const loadMissing = vi.fn().mockResolvedValue({
      elements: [
        {
          type: 'relation',
          id: 20,
          members: [{ type: 'way', ref: 30, role: 'outer' }]
        },
        { type: 'way', id: 30, nodes: [100, 101] },
        { type: 'node', id: 100, lat: 1, lon: 1 },
        { type: 'node', id: 101, lat: 2, lon: 2 }
      ]
    })

    const merged = await ensureRelationGeometryDependencies(payload, [10], loadMissing)

    expect(loadMissing).toHaveBeenCalledWith([20])
    expect(merged.elements.some((element) => element.type === 'relation' && element.id === 20)).toBe(true)
    expect(selectRelationGeometryPayload(merged, [10]).elements.map((element) => `${element.type}/${element.id}`)).toEqual([
      'relation/10',
      'relation/20',
      'way/30',
      'node/100',
      'node/101'
    ])
  })

  it('按三个省一组并以三个批次并发收集全国层省界', async () => {
    const children = Array.from({ length: 7 }, (_, index) => ({
      name: `省级行政区${index + 1}`,
      osm_relation_id: 912940 + index
    }))
    let active = 0
    let maximumActive = 0
    const loadRelationGroup = vi.fn(async (group) => {
      active += 1
      maximumActive = Math.max(maximumActive, active)
      await Promise.resolve()
      active -= 1
      return new Map(group.map((child) => [child.osm_relation_id, { id: `relation/${child.osm_relation_id}` }]))
    })

    const features = await collectRootRelationFeatures(children, loadRelationGroup)

    expect(loadRelationGroup).toHaveBeenCalledTimes(3)
    expect(loadRelationGroup.mock.calls.map(([group]) => group.length)).toEqual([3, 3, 1])
    expect(maximumActive).toBe(3)
    expect([...features.keys()]).toEqual(children.map((child) => child.osm_relation_id))
  })

  it('裁剪 OSM 载荷时只保留目标边界的递归几何依赖', () => {
    const payload = {
      osm3s: { timestamp_osm_base: snapshot },
      elements: [
        {
          type: 'relation',
          id: 10,
          members: [
            { type: 'way', ref: 20, role: 'outer' },
            { type: 'relation', ref: 30, role: 'inner' },
            { type: 'relation', ref: 40, role: 'subarea' },
            { type: 'node', ref: 50, role: 'admin_centre' }
          ],
          tags: { boundary: 'administrative' }
        },
        { type: 'way', id: 20, nodes: [100, 101] },
        {
          type: 'relation',
          id: 30,
          members: [{ type: 'way', ref: 21, role: 'outer' }],
          tags: { type: 'multipolygon' }
        },
        { type: 'way', id: 21, nodes: [101, 102] },
        { type: 'relation', id: 40, members: [], tags: { boundary: 'administrative' } },
        { type: 'node', id: 50, lat: 1, lon: 1 },
        { type: 'node', id: 100, lat: 1, lon: 1 },
        { type: 'node', id: 101, lat: 2, lon: 2 },
        { type: 'node', id: 102, lat: 3, lon: 3 },
        { type: 'way', id: 99, nodes: [999] },
        { type: 'node', id: 999, lat: 9, lon: 9 }
      ]
    }

    const selected = selectRelationGeometryPayload(payload, [10])

    expect(selected.osm3s).toEqual(payload.osm3s)
    expect(selected.elements.map((element) => `${element.type}/${element.id}`)).toEqual([
      'relation/10',
      'way/20',
      'node/100',
      'node/101',
      'relation/30',
      'way/21',
      'node/102'
    ])
    expect(selected.elements[0].members).not.toContainEqual(expect.objectContaining({ role: 'subarea' }))
  })

  it('合并递归查询的重复骨架元素时保留完整关系标签', () => {
    const payload = {
      elements: [
        {
          type: 'relation',
          id: 10,
          members: [{ type: 'way', ref: 20, role: 'outer' }],
          tags: { type: 'boundary', boundary: 'administrative', name: '测试区' }
        },
        { type: 'relation', id: 10, members: [{ type: 'way', ref: 20, role: 'outer' }] },
        { type: 'way', id: 20, nodes: [100, 101, 102, 100] },
        { type: 'node', id: 100, lat: 1, lon: 1 },
        { type: 'node', id: 101, lat: 1, lon: 2 },
        { type: 'node', id: 102, lat: 2, lon: 1 }
      ]
    }

    const selected = selectRelationGeometryPayload(payload, [10])

    expect(selected.elements[0]).toMatchObject({
      type: 'relation',
      id: 10,
      tags: { type: 'boundary', boundary: 'administrative', name: '测试区' }
    })
  })

  it('按父子依赖将待查询父级分波', () => {
    const country = { region_code: 'CN', parent_region_code: null, child_region_codes: ['CN-44'] }
    const province = {
      region_code: 'CN-44',
      parent_region_code: 'CN',
      child_region_codes: ['CN-44-4401']
    }
    const prefecture = {
      region_code: 'CN-44-4401',
      parent_region_code: 'CN-44',
      child_region_codes: ['CN-44-4401-440106']
    }
    const county = {
      region_code: 'CN-44-4401-440106',
      parent_region_code: 'CN-44-4401',
      child_region_codes: []
    }

    expect(
      groupParentsByDepth({ root_region_code: 'CN', regions: [country, province, prefecture, county] }).map(
        (group) => group.map((region) => region.region_code)
      )
    ).toEqual([['CN-44'], ['CN-44-4401']])
  })

  it('仅使用固定 OSM 元素的中心点生成南海诸岛附图', () => {
    const payload = {
      elements: SOUTH_CHINA_SEA_ELEMENTS.map((entry, index) => ({
        type: entry.type,
        id: entry.id,
        center: { lon: 110 + index, lat: 18 - index * 3 },
        tags: { 'name:zh-Hans': entry.name }
      }))
    }
    const geoJson = buildSouthChinaSeaGeoJson(payload)

    expect(southChinaSeaQuery(snapshot)).toContain(`[date:"${snapshot}"]`)
    expect(geoJson.features).toHaveLength(4)
    expect(geoJson.features.every((feature) => feature.geometry.type === 'Point')).toBe(true)
    expect(geoJson.features.map((feature) => feature.properties.osm_element_id)).toEqual(
      SOUTH_CHINA_SEA_ELEMENTS.map((entry) => entry.id)
    )
  })

  it('当固定 OSM 元素缺失或被改名时停止发布', () => {
    expect(() => buildSouthChinaSeaGeoJson({ elements: [] })).toThrow('缺少 OSM')
    const first = SOUTH_CHINA_SEA_ELEMENTS[0]
    expect(() =>
      buildSouthChinaSeaGeoJson({
        elements: [
          {
            type: first.type,
            id: first.id,
            center: { lon: 112, lat: 16 },
            tags: { name: '名称已变更' }
          }
        ]
      })
    ).toThrow('不匹配')
  })

  it('拒绝早于固定快照或带超时备注的 Overpass 响应', () => {
    expect(() =>
      assertOverpassSnapshot(
        { elements: [], osm3s: { timestamp_osm_base: '2026-08-12T00:00:00Z' } },
        snapshot
      )
    ).not.toThrow()
    expect(() =>
      assertOverpassSnapshot(
        { elements: [], osm3s: { timestamp_osm_base: '2026-07-01T00:00:00Z' } },
        snapshot
      )
    ).toThrow('早于快照')
    expect(() =>
      assertOverpassSnapshot(
        {
          elements: [],
          osm3s: { timestamp_osm_base: '2026-08-13T00:00:00Z' },
          remark: 'runtime error: Query timed out'
        },
        snapshot
      )
    ).toThrow('不完整')
  })

  it('几何查询递归展开嵌套关系', () => {
    expect(geometryQuery([10, 20], snapshot)).toContain('rel(id:10,20);out body;>>;out skel qt;')
  })
})
