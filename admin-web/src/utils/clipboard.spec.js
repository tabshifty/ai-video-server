import { describe, expect, it } from 'vitest'
import { copyTextToClipboard } from './clipboard'

describe('copyTextToClipboard', () => {
  it('空文本不触发任何复制动作', async () => {
    expect(await copyTextToClipboard('')).toBe(false)
    expect(await copyTextToClipboard(null)).toBe(false)
    expect(await copyTextToClipboard(undefined)).toBe(false)
  })

  it('没有 DOM 环境时返回失败而不是抛错', async () => {
    expect(typeof document).toBe('undefined')
    expect(await copyTextToClipboard('错误详情')).toBe(false)
  })
})
