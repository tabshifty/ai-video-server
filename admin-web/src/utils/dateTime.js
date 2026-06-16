function pad2(value) {
  return String(value).padStart(2, '0')
}

function normalizeDateInput(value) {
  if (value === null || value === undefined || value === '') {
    return null
  }
  if (value instanceof Date) {
    return value
  }
  if (typeof value === 'number' || (typeof value === 'string' && /^\d+$/.test(value.trim()))) {
    const numericValue = Number(value)
    const timestamp = numericValue > 0 && numericValue < 100000000000 ? numericValue * 1000 : numericValue
    return new Date(timestamp)
  }
  return new Date(value)
}

export function formatAdminDateTime(value, fallback = '--') {
  const date = normalizeDateInput(value)
  if (!date || Number.isNaN(date.getTime())) {
    return fallback
  }
  return [
    `${date.getFullYear()}/${pad2(date.getMonth() + 1)}/${pad2(date.getDate())}`,
    `${pad2(date.getHours())}:${pad2(date.getMinutes())}:${pad2(date.getSeconds())}`
  ].join(' ')
}
