function trimText(value) {
  return String(value ?? '').trim()
}

function toPositiveInt(value) {
  const number = Number.parseInt(String(value ?? '').trim(), 10)
  return Number.isFinite(number) && number > 0 ? number : 1
}

export function buildTvSeriesPayload(form) {
  return {
    title: trimText(form.title),
    overview: trimText(form.overview),
    poster_url: trimText(form.poster_url),
    backdrop_url: trimText(form.backdrop_url),
    first_air_date: trimText(form.first_air_date),
    active: !!form.active
  }
}

export function buildTvSeasonPayload(form) {
  return {
    season_number: toPositiveInt(form.season_number),
    title: trimText(form.title),
    overview: trimText(form.overview),
    poster_url: trimText(form.poster_url),
    air_date: trimText(form.air_date)
  }
}

export function buildTvEpisodePayload(form) {
  return {
    episode_number: toPositiveInt(form.episode_number),
    title: trimText(form.title),
    overview: trimText(form.overview),
    runtime: Number.parseInt(String(form.runtime ?? '').trim(), 10) || 0,
    air_date: trimText(form.air_date),
    still_url: trimText(form.still_url),
    video_id: trimText(form.video_id)
  }
}

function videoStatusLabel(status) {
  const map = {
    ready: '可播放',
    processing: '处理中',
    uploaded: '已上传',
    failed: '失败'
  }
  return map[status] || status || '未知状态'
}

export function mapEpisodeVideoOption(item) {
  return {
    value: item.id,
    label: `${item.title || item.id} · ${videoStatusLabel(item.status)}`
  }
}
