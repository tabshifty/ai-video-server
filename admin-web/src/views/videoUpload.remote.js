export function mergeRemoteStringOptions(existing, incoming) {
  const out = []
  const seen = new Set()

  for (const item of [...(existing || []), ...(incoming || [])]) {
    const value = String(item || '').trim()
    if (!value) continue
    const key = value.toLowerCase()
    if (seen.has(key)) continue
    seen.add(key)
    out.push(value)
  }

  return out
}

export function mergeRemoteValueOptions(existing, incoming) {
  const optionMap = new Map()

  for (const item of existing || []) {
    const value = String(item?.value || '').trim()
    if (!value) continue
    optionMap.set(value.toLowerCase(), { ...item, value })
  }
  for (const item of incoming || []) {
    const value = String(item?.value || '').trim()
    if (!value) continue
    optionMap.set(value.toLowerCase(), { ...item, value })
  }

  return Array.from(optionMap.values())
}

export function createRemoteSuggestionLoader({
  delay = 180,
  fetcher,
  getOptions,
  setOptions,
  setLoading,
  mergeOptions
}) {
  let timer = null
  let requestId = 0

  return (keyword = '') => {
    const query = String(keyword || '').trim()
    requestId += 1
    const currentRequestId = requestId
    if (timer) {
      clearTimeout(timer)
    }

    timer = setTimeout(async () => {
      setLoading?.(true)
      try {
        const incoming = await fetcher(query)
        if (currentRequestId !== requestId) return
        const existing = typeof getOptions === 'function' ? getOptions() : []
        setOptions(mergeOptions(existing, incoming))
      } catch (_) {
        if (currentRequestId !== requestId) return
      } finally {
        if (currentRequestId === requestId) {
          setLoading?.(false)
        }
      }
    }, delay)
  }
}
