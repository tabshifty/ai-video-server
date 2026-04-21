export const IMAGE_COLLECTION_PREVIEW_PARAMS = Object.freeze({
  w: 240,
  h: 240,
  fit: 'cover',
  q: 82
})

function normalizeSortOrder(raw) {
  const value = Number(raw)
  if (!Number.isFinite(value)) return 0
  return Math.trunc(value)
}

export function buildImageCollectionPayload(form) {
  const rawCoverImageID = form?.cover_image_id
  const coverImageID = typeof rawCoverImageID === 'string' ? rawCoverImageID.trim() : rawCoverImageID || ''
  return {
    name: form?.name?.trim() || '',
    description: form?.description?.trim() || '',
    cover_url: form?.cover_url?.trim() || '',
    cover_image_id: coverImageID || null,
    sort_order: normalizeSortOrder(form?.sort_order),
    active: !!form?.active
  }
}

export function revokePreviewURLs(urlMap, revoke = (value) => URL.revokeObjectURL(value)) {
  for (const value of Object.values(urlMap || {})) {
    if (!value) continue
    revoke(value)
  }
  return {}
}
