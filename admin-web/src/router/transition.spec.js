import { describe, expect, it } from 'vitest'
import { getRouteTransitionName, getRouteViewKey } from './transition'

describe('getRouteTransitionName', () => {
  it('为公开页面保留淡入淡出过渡', () => {
    expect(getRouteTransitionName({ meta: { public: true } })).toBe('fade-slide')
  })

  it('为后台页面禁用路由过渡', () => {
    expect(getRouteTransitionName({ meta: {} })).toBeUndefined()
    expect(getRouteTransitionName()).toBeUndefined()
  })

  it('中国地图只按路径复用页面实例，其它页面仍按完整地址重建', () => {
    expect(getRouteViewKey({ path: '/toolbox/china-map', fullPath: '/toolbox/china-map?region_code=CN-44' })).toBe(
      '/toolbox/china-map'
    )
    expect(getRouteViewKey({ path: '/videos', fullPath: '/videos?page=2' })).toBe('/videos?page=2')
  })
})
