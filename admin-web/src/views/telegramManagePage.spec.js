import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const page = readFileSync(new URL('./TelegramManage.vue', import.meta.url), 'utf8')

describe('Telegram 管理页', () => {
  it('通过同一页面工作流完成授权、预解析确认与来源控制', () => {
    expect(page).toContain('startAdminTelegramPhoneAuthorization')
    expect(page).toContain('startAdminTelegramQRAuthorization')
    expect(page).toContain('previewAdminTelegramSource')
    expect(page).toContain('confirmAdminTelegramSource')
    expect(page).toContain("requestSourceAction(row, 'pause')")
    expect(page).toContain("requestSourceAction(row, 'resume')")
    expect(page).toContain("requestSourceAction(row, 'recover')")
    expect(page).toContain("requestSourceAction(row, 'backfill')")
  })

  it('只在页面内存保留授权秘密和邀请引用，并在不可见或离开时停止轮询', () => {
    expect(page).toContain("const sourceInput = ref('')")
    expect(page).toContain("const authorizationForm = reactive({ code: '', password: '' })")
    expect(page).toContain("confirmation.password = ''")
    expect(page).toContain('document.addEventListener(\'visibilitychange\', handleVisibilityChange)')
    expect(page).toContain('clearPollTimer()')
    expect(page).toContain('onUnmounted(() => {')
    expect(page).not.toContain('localStorage.setItem')
    expect(page).not.toContain('sessionStorage.setItem')
  })
})
