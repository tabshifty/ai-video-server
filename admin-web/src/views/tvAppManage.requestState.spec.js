import { describe, expect, it } from 'vitest'
import {
  buildTVAppReleaseRequest,
  isCurrentTVAppReleaseRequest,
  selectTVAppReleaseCache
} from './tvAppManage.requestState'

const baseQuery = Object.freeze({
  page: 1,
  page_size: 20,
  q: '',
  status: '',
  abi_completeness: '',
  current_published: false
})

function buildRequest(query = {}, options = {}) {
  return buildTVAppReleaseRequest({
    query: { ...baseQuery, ...query },
    clientType: options.clientType || 'android_tv',
    supportsAbi: options.supportsAbi ?? true
  })
}

describe('安装包查询请求状态', () => {
  it('以实际发送口径生成参数与稳定查询身份', () => {
    const request = buildRequest({
      page: 2,
      page_size: 50,
      q: '  2.4.0  ',
      status: 'draft',
      abi_completeness: 'missing',
      current_published: true
    })

    expect(request.params).toEqual({
      page: 2,
      page_size: 50,
      current_published: 1,
      client_type: 'android_tv',
      q: '2.4.0',
      status: 'draft',
      abi_completeness: 'missing'
    })
    expect(request.key).toBe(JSON.stringify(request.params))
    expect(buildRequest({ q: '  2.4.0  ' }).key).toBe(buildRequest({ q: '2.4.0' }).key)

    const phoneWithIgnoredAbi = buildRequest(
      { abi_completeness: 'missing' },
      { clientType: 'android_phone', supportsAbi: false }
    )
    const phoneWithoutAbi = buildRequest(
      { abi_completeness: '' },
      { clientType: 'android_phone', supportsAbi: false }
    )

    expect(phoneWithIgnoredAbi.params).not.toHaveProperty('abi_completeness')
    expect(phoneWithIgnoredAbi.key).toBe(phoneWithoutAbi.key)
  })

  it('任一实际请求参数变化都会产生不同身份', () => {
    const baseKey = buildRequest().key
    const variants = [
      buildRequest({ page: 2 }).key,
      buildRequest({ page_size: 50 }).key,
      buildRequest({ q: '2.4.0' }).key,
      buildRequest({ status: 'offline' }).key,
      buildRequest({ abi_completeness: 'complete' }).key,
      buildRequest({ current_published: true }).key,
      buildRequest({}, { clientType: 'android_phone', supportsAbi: false }).key
    ]

    expect(new Set(variants).size).toBe(variants.length)
    variants.forEach((key) => expect(key).not.toBe(baseKey))
  })

  it('同身份复用原缓存，异身份隐藏旧行但不修改原数据', () => {
    const tvRequest = buildRequest()
    const phoneRequest = buildRequest({}, { clientType: 'android_phone', supportsAbi: false })
    const items = [{ id: 7, version_name: '2.4.0' }]

    expect(selectTVAppReleaseCache({
      activeRequestKey: tvRequest.key,
      cachedRequestKey: tvRequest.key,
      items,
      totalCount: 3
    })).toEqual({ hasCurrentResult: true, items, totalCount: 3 })

    expect(selectTVAppReleaseCache({
      activeRequestKey: phoneRequest.key,
      cachedRequestKey: tvRequest.key,
      items,
      totalCount: 3
    })).toEqual({ hasCurrentResult: false, items: [], totalCount: 0 })
    expect(items).toEqual([{ id: 7, version_name: '2.4.0' }])
  })

  it('同身份当前页无行但总数非零时保留结果总数', () => {
    const request = buildRequest({ page: 3 })

    expect(selectTVAppReleaseCache({
      activeRequestKey: request.key,
      cachedRequestKey: request.key,
      items: [],
      totalCount: 41
    })).toEqual({ hasCurrentResult: true, items: [], totalCount: 41 })
  })

  it('只接受最新序号且身份匹配的请求结果', () => {
    const tvRequest = buildRequest()
    const phoneRequest = buildRequest({}, { clientType: 'android_phone', supportsAbi: false })
    const latestRequest = { sequence: 2, key: phoneRequest.key }

    expect(isCurrentTVAppReleaseRequest({ sequence: 1, key: tvRequest.key }, latestRequest)).toBe(false)
    expect(isCurrentTVAppReleaseRequest({ sequence: 1, key: phoneRequest.key }, latestRequest)).toBe(false)
    expect(isCurrentTVAppReleaseRequest({ sequence: 2, key: tvRequest.key }, latestRequest)).toBe(false)
    expect(isCurrentTVAppReleaseRequest({ sequence: 2, key: phoneRequest.key }, latestRequest)).toBe(true)
  })
})
