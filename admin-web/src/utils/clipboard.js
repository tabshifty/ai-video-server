/**
 * copyTextToClipboard：优先用异步剪贴板，失败或非安全上下文时回退到
 * 隐藏 textarea + execCommand。管理端常以 http://内网 IP 访问，此时
 * navigator.clipboard 不存在，不能只依赖它。
 */
export async function copyTextToClipboard(text) {
  const value = String(text || '')
  if (value === '') return false

  if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(value)
      return true
    } catch (_) {
      // 权限被拒或页面失焦时继续尝试同步回退
    }
  }

  if (typeof document === 'undefined' || !document.body) return false

  const textarea = document.createElement('textarea')
  textarea.value = value
  textarea.setAttribute('readonly', '')
  textarea.setAttribute('aria-hidden', 'true')
  textarea.style.position = 'fixed'
  textarea.style.top = '-1000px'
  textarea.style.opacity = '0'
  document.body.appendChild(textarea)
  try {
    textarea.select()
    return document.execCommand('copy') === true
  } catch (_) {
    return false
  } finally {
    document.body.removeChild(textarea)
  }
}
