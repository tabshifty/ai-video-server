export const TELEGRAM_STATUS_POLL_INTERVAL_MS = 5000
export const TELEGRAM_AUTHORIZATION_POLL_INTERVAL_MS = 2000

const ACTIVE_AUTHORIZATION_STATUSES = new Set([
  'pending',
  'awaiting_code',
  'awaiting_password',
  'scanning'
])

const TELEGRAM_WEB_HOSTS = new Set([
  't.me',
  'www.t.me',
  'telegram.me',
  'www.telegram.me'
])

export function isTelegramAuthorizationActive(authorization) {
  const status = typeof authorization === 'string'
    ? authorization
    : authorization?.status
  return ACTIVE_AUTHORIZATION_STATUSES.has(String(status || '').trim().toLowerCase())
}

export function getTelegramPollingInterval(authorization) {
  return isTelegramAuthorizationActive(authorization)
    ? TELEGRAM_AUTHORIZATION_POLL_INTERVAL_MS
    : TELEGRAM_STATUS_POLL_INTERVAL_MS
}

export function shouldPollTelegramPage({ pageActive = true, documentVisible = true } = {}) {
  return Boolean(pageActive && documentVisible)
}

export function isPrivateTelegramInviteReference(value) {
  const reference = String(value || '').trim()
  if (!reference) return false

  try {
    const parsed = new URL(reference)
    const host = parsed.hostname.toLowerCase()
    if (parsed.protocol === 'tg:') {
      return host === 'join' && Boolean(parsed.searchParams.get('invite')?.trim())
    }
    if (!TELEGRAM_WEB_HOSTS.has(host)) return false
    const path = parsed.pathname.replace(/^\/+/, '')
    return path.startsWith('+') || /^joinchat\//i.test(path)
  } catch (_) {
    return false
  }
}
