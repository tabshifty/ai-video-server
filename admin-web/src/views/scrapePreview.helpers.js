export function toPositiveInt(value) {
  const parsed = Number(value)
  if (!Number.isInteger(parsed) || parsed <= 0) {
    return 0
  }
  return parsed
}

export function buildScrapePreviewPayload(form) {
  const payload = {
    video_id: normalizeText(form?.video_id),
    title: normalizeText(form?.title),
    type: normalizeText(form?.type),
  }

  const year = toPositiveInt(form?.year)
  if (year > 0) {
    payload.year = year
  }

  if (payload.type === 'movie') {
    payload.bypass_cache = true
  }

  if (payload.type === 'tv') {
    const seasonNumber = toPositiveInt(form?.season_number)
    const episodeNumber = toPositiveInt(form?.episode_number)
    if (seasonNumber > 0) {
      payload.season_number = seasonNumber
    }
    if (episodeNumber > 0) {
      payload.episode_number = episodeNumber
    }
  }

  return payload
}

export function buildScrapeConfirmPayload(form, edit) {
  const payload = {
    video_id: normalizeText(edit?.video_id),
    tmdb_id: toPositiveInt(edit?.tmdb_id),
    external_id: normalizeText(edit?.external_id),
    title: normalizeText(edit?.title),
    overview: normalizeText(edit?.overview),
    poster_url: normalizeText(edit?.poster_url),
    backdrop_url: normalizeText(edit?.backdrop_url),
    release_date: normalizeText(edit?.release_date),
    metadata: edit?.metadata && typeof edit.metadata === 'object' ? edit.metadata : {}
  }

  if (normalizeText(form?.type) === 'tv') {
    const seasonNumber = toPositiveInt(form?.season_number)
    const episodeNumber = toPositiveInt(form?.episode_number)
    if (seasonNumber > 0) {
      payload.season_number = seasonNumber
    }
    if (episodeNumber > 0) {
      payload.episode_number = episodeNumber
    }
  }

  return payload
}

export function resolveTVPreviewState(candidate, fallback = {}) {
  const title = normalizeText(candidate?.parsed_title) || normalizeText(fallback.title)
  const seasonNumber = toPositiveInt(candidate?.parsed_season_number) || toPositiveInt(fallback.season_number)
  const episodeNumber = toPositiveInt(candidate?.parsed_episode_number) || toPositiveInt(fallback.episode_number)
  return {
    title,
    season_number: seasonNumber > 0 ? seasonNumber : null,
    episode_number: episodeNumber > 0 ? episodeNumber : null
  }
}

function normalizeText(value) {
  return typeof value === 'string' ? value.trim() : ''
}
