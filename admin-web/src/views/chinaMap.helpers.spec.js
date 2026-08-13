import { describe, expect, it } from 'vitest'
import {
  ROOT_REGION_CODE,
  buildRegionIndex,
  buildRegionQuery,
  canGoBackFromView,
  findRegionView,
  getParentRegionCode,
  searchRegions,
  validateRegionCatalog
} from './chinaMap.helpers'

const regions = [
  {
    region_code: 'CN',
    name: '中国',
    parent_region_code: null,
    level: 'country',
    child_region_codes: ['CN-11', 'CN-44', 'CN-71']
  },
  {
    region_code: 'CN-11',
    name: '北京市',
    parent_region_code: 'CN',
    level: 'province',
    child_region_codes: ['CN-11-110105']
  },
  {
    region_code: 'CN-11-110105',
    name: '朝阳区',
    parent_region_code: 'CN-11',
    level: 'county',
    child_region_codes: []
  },
  {
    region_code: 'CN-44',
    name: '广东省',
    parent_region_code: 'CN',
    level: 'province',
    child_region_codes: ['CN-44-440100']
  },
  {
    region_code: 'CN-44-440100',
    name: '广州市',
    parent_region_code: 'CN-44',
    level: 'prefecture',
    child_region_codes: ['CN-44-440100-440106']
  },
  {
    region_code: 'CN-44-440100-440106',
    name: '天河区',
    parent_region_code: 'CN-44-440100',
    level: 'county',
    child_region_codes: []
  },
  {
    region_code: 'CN-71',
    name: '台湾省',
    parent_region_code: 'CN',
    level: 'province',
    child_region_codes: ['CN-71-TPE']
  },
  {
    region_code: 'CN-71-TPE',
    name: '台北市',
    parent_region_code: 'CN-71',
    level: 'county',
    child_region_codes: []
  }
]

describe('China map hierarchy helpers', () => {
  it('validates a closed variable-depth hierarchy and builds full paths', () => {
    expect(validateRegionCatalog({ root_region_code: ROOT_REGION_CODE, regions })).toEqual([])

    const index = buildRegionIndex(regions)
    expect(index.byCode.get('CN-44-440100-440106').full_path).toBe('中国 / 广东省 / 广州市 / 天河区')
    expect(index.byCode.get('CN-71-TPE').full_path).toBe('中国 / 台湾省 / 台北市')
  })

  it('rejects duplicate identifiers, broken parents and unreachable regions', () => {
    const invalid = [
      ...regions,
      { ...regions[1] },
      {
        region_code: 'CN-X',
        name: '孤立区域',
        parent_region_code: 'CN-MISSING',
        level: 'county',
        child_region_codes: []
      }
    ]

    const issues = validateRegionCatalog({ root_region_code: ROOT_REGION_CODE, regions: invalid })
    expect(issues.some((issue) => issue.includes('重复'))).toBe(true)
    expect(issues.some((issue) => issue.includes('父级'))).toBe(true)
    expect(issues.some((issue) => issue.includes('不可达'))).toBe(true)
  })

  it('shows child layers for branches and the parent layer for terminal focus', () => {
    const index = buildRegionIndex(regions)

    expect(findRegionView(index, 'CN-44-440100')).toMatchObject({
      region_code: 'CN-44-440100',
      layer_parent_code: 'CN-44-440100',
      selected_region_code: null
    })
    expect(findRegionView(index, 'CN-44-440100-440106')).toMatchObject({
      region_code: 'CN-44-440100-440106',
      layer_parent_code: 'CN-44-440100',
      selected_region_code: 'CN-44-440100-440106'
    })
    expect(findRegionView(index, 'not-found')).toMatchObject({
      region_code: ROOT_REGION_CODE,
      layer_parent_code: ROOT_REGION_CODE
    })
  })

  it('navigates upward without assuming every region has three levels', () => {
    const index = buildRegionIndex(regions)
    expect(getParentRegionCode(index, 'CN-71-TPE')).toBe('CN-71')
    expect(getParentRegionCode(index, 'CN-71')).toBe(ROOT_REGION_CODE)
    expect(getParentRegionCode(index, ROOT_REGION_CODE)).toBeNull()
  })

  it('only enables upward navigation after a non-root view is displayed', () => {
    expect(canGoBackFromView(null)).toBe(false)
    expect(canGoBackFromView({ region_code: ROOT_REGION_CODE })).toBe(false)
    expect(canGoBackFromView({ region_code: 'CN-44' })).toBe(true)
  })

  it('searches every level, disambiguates by full path and ranks exact names first', () => {
    const index = buildRegionIndex(regions)
    expect(searchRegions(index, '朝阳区')).toEqual([
      expect.objectContaining({ region_code: 'CN-11-110105', full_path: '中国 / 北京市 / 朝阳区' })
    ])
    expect(searchRegions(index, '广东 广州 天河')[0]).toMatchObject({
      region_code: 'CN-44-440100-440106'
    })
    expect(searchRegions(index, 'CN-71-TPE')[0]).toMatchObject({ name: '台北市' })
    expect(searchRegions(index, '不存在')).toEqual([])
  })

  it('writes stable region_code URL state while preserving unrelated query values', () => {
    expect(buildRegionQuery({ tab: 'map' }, 'CN-44')).toEqual({ tab: 'map', region_code: 'CN-44' })
    expect(buildRegionQuery({ tab: 'map', region_code: 'CN-44' }, ROOT_REGION_CODE)).toEqual({ tab: 'map' })
  })
})
