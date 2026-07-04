export function normalizeTVAppClientType(clientType) {
  return clientType === 'android_phone' ? 'android_phone' : 'android_tv'
}

export function getTVAppDownloadPagePath(clientType) {
  return normalizeTVAppClientType(clientType) === 'android_phone'
    ? '/downloads/android-phone'
    : '/downloads/android-tv'
}

export function getTVAppDownloadQRCodeTitle(clientType) {
  return normalizeTVAppClientType(clientType) === 'android_phone'
    ? '手机端下载二维码'
    : 'TV 下载二维码'
}

export function resolveTVAppDownloadQRCodeOrigin({ currentOrigin, isDev, apiProxyTarget }) {
  const fallbackOrigin = String(currentOrigin || '').trim()
  if (!isDev) return fallbackOrigin
  const rawTarget = String(apiProxyTarget || '').trim()
  if (!rawTarget) return fallbackOrigin
  try {
    return new URL(rawTarget).origin
  } catch {
    return fallbackOrigin
  }
}

export function buildTVAppDownloadPageURL(clientType, options) {
  const origin = resolveTVAppDownloadQRCodeOrigin(options)
  return new URL(getTVAppDownloadPagePath(clientType), origin).toString()
}
