import { describe, expect, it } from 'vitest'
import {
  COUNTY_OSM_BY_SOURCE_CODE,
  buildCatalog,
  catalogSummary,
  matchRegionToOsm,
  normalizeChineseName
} from './catalog.mjs'

const mainland = [
  {
    code: '11',
    name: '北京市',
    children: [
      {
        code: '1101',
        name: '市辖区',
        children: [
          { code: '110105', name: '朝阳区' },
          { code: '110106', name: '丰台区' },
          { code: '110171', name: '北京经济技术开发区' }
        ]
      }
    ]
  },
  {
    code: '44',
    name: '广东省',
    children: [
      {
        code: '4401',
        name: '广州市',
        children: [{ code: '440106', name: '天河区' }]
      },
      {
        code: '4419',
        name: '东莞市',
        children: [{ code: '441900003', name: '东城街道' }]
      }
    ]
  }
]

const special = {
  香港特别行政区: { 香港岛: ['中西区'] },
  澳门特别行政区: { 澳门半岛: ['大堂区'] },
  台湾省: { 台北市: ['大安区', '内湖区'] }
}

describe('China map catalog generator', () => {
  it('builds direct county children for municipalities and variable special-region branches', () => {
    const catalog = buildCatalog(mainland, special)
    const beijing = catalog.regions.find((region) => region.region_code === 'CN-11')
    const beijingCounty = catalog.regions.find((region) => region.source_admin_code === '110105')
    const guangzhou = catalog.regions.find((region) => region.source_admin_code === '4401')
    const dongguan = catalog.regions.find((region) => region.source_admin_code === '4419')
    const taipei = catalog.regions.find((region) => region.name === '台北市')
    const hongKong = catalog.regions.find((region) => region.region_code === 'CN-81')
    const macau = catalog.regions.find((region) => region.region_code === 'CN-82')

    expect(beijing.child_region_codes).toContain(beijingCounty.region_code)
    expect(catalog.regions.some((region) => region.source_admin_code === '110171')).toBe(false)
    expect(beijingCounty.parent_region_code).toBe(beijing.region_code)
    expect(guangzhou.child_region_codes).toHaveLength(1)
    expect(dongguan.child_region_codes).toHaveLength(0)
    expect(taipei.level).toBe('prefecture')
    expect(taipei.child_region_codes).toHaveLength(2)
    expect(hongKong.child_region_codes).toHaveLength(1)
    expect(macau.child_region_codes).toHaveLength(2)
    expect(catalogSummary(catalog)).toMatchObject({ country: 1, province: 5, terminal: 19 })
  })

  it('normalizes traditional variants and matches OSM by source code before name', () => {
    expect(normalizeChineseName('臺北市')).toBe('台北市')
    expect(normalizeChineseName('萬華區')).toBe('万华区')
    expect(normalizeChineseName('烏坵鄉')).toBe('乌丘乡')
    const region = { name: '广州市', source_admin_code: '4401', aliases: [] }
    const candidates = [
      { id: 1, tags: { name: '同名错误', 'ref:admin:CN': '4401' } },
      { id: 2, tags: { name: '广州市' } }
    ]

    expect(matchRegionToOsm(region, candidates)).toEqual(candidates[0])
  })

  it('requires an unambiguous normalized name when no source code is available', () => {
    const region = { name: '台北市', source_admin_code: null, aliases: ['臺北市'] }
    const match = { id: 1293250, tags: { name: '臺北市' } }
    expect(matchRegionToOsm(region, [match])).toEqual(match)
    expect(matchRegionToOsm(region, [match, { id: 2, tags: { name: '台北市' } }])).toBeNull()
  })

  it('匹配台湾繁体标签和带蒙古文后缀的双语名称', () => {
    const taiwan = { name: '万华区', source_admin_code: null, aliases: [] }
    const banner = { name: '乌拉特中旗', source_admin_code: '150824', aliases: [] }
    const candidates = [
      { id: 1, tags: { name: '萬華區' } },
      { id: 2, tags: { name: '乌拉特中旗 ᠤᠷᠠᠳ ᠤᠨ ᠳᠤᠮᠳᠠᠳᠤ ᠬᠣᠰ᠃ᠢᠭᠤ' } }
    ]

    expect(matchRegionToOsm(taiwan, candidates)).toEqual(candidates[0])
    expect(matchRegionToOsm(banner, candidates)).toEqual(candidates[1])
  })

  it('将澳门目录括号内的常用地名作为匹配别名', () => {
    const catalog = buildCatalog(mainland, {
      ...special,
      澳门特别行政区: { 离岛: ['嘉模堂区(氹仔)'] }
    })
    const cotai = catalog.regions.find((region) => region.name === '嘉模堂区')
    const candidate = { id: 5758868, tags: { name: '氹仔' } }

    expect(cotai.aliases).toContain('氹仔')
    expect(matchRegionToOsm(cotai, [candidate])).toEqual(candidate)
  })

  it('将滞后的重庆与三沙目录校正为快照时点的现行区划', () => {
    const catalog = buildCatalog(
      [
        {
          code: '46',
          name: '海南省',
          children: [
            {
              code: '4603',
              name: '三沙市',
              children: [
                { code: '460321', name: '西沙群岛' },
                { code: '460322', name: '南沙群岛' },
                { code: '460323', name: '中沙群岛的岛礁及其海域' }
              ]
            }
          ]
        },
        {
          code: '50',
          name: '重庆市',
          children: [
            {
              code: '5001',
              name: '市辖区',
              children: [
                { code: '500105', name: '江北区' },
                { code: '500112', name: '渝北区' }
              ]
            }
          ]
        }
      ],
      special
    )

    const liangjiang = catalog.regions.find((region) => region.source_admin_code === '500157')
    const xisha = catalog.regions.find((region) => region.source_admin_code === '460302')
    const nansha = catalog.regions.find((region) => region.source_admin_code === '460303')

    expect(liangjiang).toMatchObject({
      name: '两江新区',
      osm_relation_id: COUNTY_OSM_BY_SOURCE_CODE['500157'],
      aliases: expect.arrayContaining(['江北区', '渝北区'])
    })
    expect(xisha).toMatchObject({ name: '西沙区', osm_relation_id: COUNTY_OSM_BY_SOURCE_CODE['460302'] })
    expect(nansha).toMatchObject({ name: '南沙区', osm_relation_id: COUNTY_OSM_BY_SOURCE_CODE['460303'] })
    expect(catalog.regions.some((region) => ['500105', '500112', '460321', '460322', '460323'].includes(region.source_admin_code))).toBe(false)
  })

  it('为父级 subarea 缺漏的县级边界使用经过核对的固定 OSM 关系', () => {
    const catalog = buildCatalog(
      [
        {
          code: '21',
          name: '辽宁省',
          children: [
            {
              code: '2107',
              name: '锦州市',
              children: [
                { code: '210702', name: '古塔区' },
                { code: '210703', name: '凌河区' }
              ]
            }
          ]
        },
        {
          code: '52',
          name: '贵州省',
          children: [
            { code: '5226', name: '黔东南苗族侗族自治州', children: [{ code: '522628', name: '锦屏县' }] }
          ]
        }
      ],
      special
    )

    for (const sourceCode of ['210702', '210703', '522628']) {
      expect(catalog.regions.find((region) => region.source_admin_code === sourceCode)?.osm_relation_id).toBe(
        COUNTY_OSM_BY_SOURCE_CODE[sourceCode]
      )
    }
    const wuqiu = catalog.regions.find((region) => region.name === '乌丘乡')
    expect(matchRegionToOsm(wuqiu, [{ id: 3339702, tags: { name: '烏坵鄉' } }])?.id).toBe(3339702)
  })
})
