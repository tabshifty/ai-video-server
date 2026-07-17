export function buildTVAppReleaseRequest({ query, clientType, supportsAbi }) {
  const params = {
    page: query.page,
    page_size: query.page_size,
    current_published: query.current_published ? 1 : 0,
    client_type: clientType
  }
  const search = query.q.trim()
  if (search) params.q = search
  if (query.status) params.status = query.status
  if (query.abi_completeness && supportsAbi) params.abi_completeness = query.abi_completeness

  return { params, key: JSON.stringify(params) }
}

export function isCurrentTVAppReleaseRequest(request, latestRequest) {
  return Boolean(request?.key)
    && request.sequence === latestRequest?.sequence
    && request.key === latestRequest?.key
}

export function selectTVAppReleaseCache({ activeRequestKey, cachedRequestKey, items, totalCount }) {
  const hasCurrentResult = Boolean(activeRequestKey) && activeRequestKey === cachedRequestKey

  return {
    hasCurrentResult,
    items: hasCurrentResult && Array.isArray(items) ? items : [],
    totalCount: hasCurrentResult ? Number(totalCount || 0) : 0
  }
}
