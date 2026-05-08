export function resolvePageJump(rawPage, options = {}) {
  const currentPage = toPositiveInt(options.currentPage, 1)
  const pageSize = toPositiveInt(options.pageSize, 20)
  const total = toNonNegativeInt(options.total)
  const maxPage = total > 0 ? Math.max(1, Math.ceil(total / pageSize)) : 0

  if (maxPage === 0) {
    return {
      disabled: true,
      shouldJump: false,
      page: currentPage,
      displayValue: String(currentPage)
    }
  }

  const normalized = typeof rawPage === 'string' ? rawPage.trim() : String(rawPage ?? '').trim()
  if (!/^\d+$/.test(normalized)) {
    return {
      disabled: false,
      shouldJump: false,
      page: currentPage,
      displayValue: String(currentPage)
    }
  }

  const page = clamp(Number(normalized), 1, maxPage)
  return {
    disabled: false,
    shouldJump: page !== currentPage,
    page,
    displayValue: String(page)
  }
}

function toPositiveInt(value, fallback) {
  const parsed = Number(value)
  if (!Number.isInteger(parsed) || parsed <= 0) {
    return fallback
  }
  return parsed
}

function toNonNegativeInt(value) {
  const parsed = Number(value)
  if (!Number.isFinite(parsed) || parsed <= 0) {
    return 0
  }
  return Math.trunc(parsed)
}

function clamp(value, min, max) {
  return Math.min(Math.max(value, min), max)
}
