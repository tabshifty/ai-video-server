export const AV_SITE_OPTIONS = [
  'theporndb',
  'dmm',
  'javdb',
  'jav321',
  'mgstage',
  'javbus',
  'javlibrary',
  'airav_cc',
  'avsex',
  'mywife',
  'fc2ppvdb',
  'fc2club',
  'fc2',
  'fc2hub'
]

export const AV_POSTER_CROP_MODE_OPTIONS = [
  'portrait_center',
  'portrait_left',
  'portrait_right'
]

export function buildAVManualScrapePreviewPayload(form) {
  return {
    video_id: normalizeText(form?.video_id),
    title: normalizeText(form?.title),
    site_category: normalizeText(form?.site_category),
    site_source: normalizeText(form?.site_source),
    bypass_cache: form?.bypass_cache !== false
  }
}

export function applyAVManualScrapeRouteQuery(form, edit, query) {
  const videoID = normalizeText(query?.video_id)
  const title = normalizeText(query?.title)
  const externalID = normalizeText(query?.external_id)
  if (videoID) {
    form.video_id = videoID
  }
  if (title) {
    form.title = title
  } else if (externalID) {
    form.title = externalID
  }
  if (externalID) {
    edit.external_id = externalID
  }
}

export function buildAVManualScrapeConfirmPayload(form, edit) {
  return {
    video_id: normalizeText(form?.video_id || edit?.video_id),
    external_id: normalizeText(edit?.external_id),
    title: normalizeText(edit?.title),
    overview: normalizeText(edit?.overview),
    poster_url: normalizeText(edit?.poster_url),
    release_date: normalizeText(edit?.release_date),
    site_category: normalizeText(form?.site_category),
    site_source: normalizeText(form?.site_source),
    metadata: edit?.metadata && typeof edit.metadata === 'object' ? edit.metadata : {}
  }
}

export function parseSiteOrderInput(value) {
  if (Array.isArray(value)) {
    return value.map((item) => normalizeText(item)).filter(Boolean)
  }
  return String(value || '')
    .split(/[\n,]/)
    .map((item) => normalizeText(item))
    .filter(Boolean)
}

export function buildAVScrapeConfigPayload(form) {
  return {
    enabled_sites: Array.isArray(form?.enabled_sites) ? form.enabled_sites.map((item) => normalizeText(item)).filter(Boolean) : [],
    category_site_order: {
      fc2: parseSiteOrderInput(form?.fc2_order),
      western: parseSiteOrderInput(form?.western_order),
      japanese: parseSiteOrderInput(form?.japanese_order)
    },
    poster_crop_enabled: form?.poster_crop_enabled !== false,
    poster_crop_mode: normalizePosterCropMode(form?.poster_crop_mode)
  }
}

export function applyAVScrapeConfig(form, payload) {
  const categoryOrder = payload?.category_site_order && typeof payload.category_site_order === 'object'
    ? payload.category_site_order
    : {}
  form.enabled_sites = Array.isArray(payload?.enabled_sites) ? payload.enabled_sites.map((item) => normalizeText(item)).filter(Boolean) : []
  form.fc2_order = toSiteOrderText(categoryOrder.fc2)
  form.western_order = toSiteOrderText(categoryOrder.western)
  form.japanese_order = toSiteOrderText(categoryOrder.japanese)
  form.poster_crop_enabled = payload?.poster_crop_enabled !== false
  form.poster_crop_mode = normalizePosterCropMode(payload?.poster_crop_mode)
}

export function toSiteOrderText(value) {
  return parseSiteOrderInput(value).join(', ')
}

function normalizeText(value) {
  return typeof value === 'string' ? value.trim() : ''
}

function normalizePosterCropMode(value) {
  const mode = normalizeText(value)
  return AV_POSTER_CROP_MODE_OPTIONS.includes(mode) ? mode : 'portrait_center'
}
