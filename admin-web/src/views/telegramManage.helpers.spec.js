import { describe, expect, it } from 'vitest'
import {
  TELEGRAM_AUTHORIZATION_POLL_INTERVAL_MS,
  TELEGRAM_STATUS_POLL_INTERVAL_MS,
  getTelegramPollingInterval,
  isPrivateTelegramInviteReference,
  isTelegramAuthorizationActive,
  shouldPollTelegramPage
} from './telegramManage.helpers'

describe('Telegram 管理页轮询规则', () => {
  it('uses a shorter interval only while an authorization remains active', () => {
    expect(isTelegramAuthorizationActive({ status: 'awaiting_code' })).toBe(true)
    expect(isTelegramAuthorizationActive({ status: 'awaiting_password' })).toBe(true)
    expect(isTelegramAuthorizationActive({ status: 'scanning' })).toBe(true)
    expect(isTelegramAuthorizationActive({ status: 'succeeded' })).toBe(false)
    expect(isTelegramAuthorizationActive({ status: 'failed' })).toBe(false)
    expect(getTelegramPollingInterval({ status: 'awaiting_code' })).toBe(TELEGRAM_AUTHORIZATION_POLL_INTERVAL_MS)
    expect(getTelegramPollingInterval({ status: 'succeeded' })).toBe(TELEGRAM_STATUS_POLL_INTERVAL_MS)
  })

  it('stops automatic refreshes while the page is hidden or inactive', () => {
    expect(shouldPollTelegramPage({ pageActive: true, documentVisible: true })).toBe(true)
    expect(shouldPollTelegramPage({ pageActive: true, documentVisible: false })).toBe(false)
    expect(shouldPollTelegramPage({ pageActive: false, documentVisible: true })).toBe(false)
  })
})

describe('Telegram 管理页私密来源判断', () => {
  it('recognizes the invitation formats that require a second confirmation', () => {
    expect(isPrivateTelegramInviteReference('https://t.me/+privateInviteToken')).toBe(true)
    expect(isPrivateTelegramInviteReference('https://telegram.me/joinchat/privateInviteToken')).toBe(true)
    expect(isPrivateTelegramInviteReference('tg://join?invite=privateInviteToken')).toBe(true)
    expect(isPrivateTelegramInviteReference('@public_channel')).toBe(false)
    expect(isPrivateTelegramInviteReference('https://t.me/public_channel')).toBe(false)
  })
})
