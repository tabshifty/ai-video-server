import { describe, expect, it } from 'vitest'
import { getRouteTransitionName } from './transition'

describe('getRouteTransitionName', () => {
  it('为公开页面保留淡入淡出过渡', () => {
    expect(getRouteTransitionName({ meta: { public: true } })).toBe('fade-slide')
  })

  it('为后台页面禁用路由过渡', () => {
    expect(getRouteTransitionName({ meta: {} })).toBeUndefined()
    expect(getRouteTransitionName()).toBeUndefined()
  })
})
