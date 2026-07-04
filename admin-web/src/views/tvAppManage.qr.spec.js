import { describe, expect, it } from 'vitest'

import {
  buildTVAppDownloadPageURL,
  getTVAppDownloadPagePath,
  getTVAppDownloadQRCodeTitle,
  resolveTVAppDownloadQRCodeOrigin
} from './tvAppManage.qr'

describe('tv app manage qr helpers', () => {
  it('maps client type to the correct download page path and title', () => {
    expect(getTVAppDownloadPagePath('android_tv')).toBe('/downloads/android-tv')
    expect(getTVAppDownloadPagePath('android_phone')).toBe('/downloads/android-phone')
    expect(getTVAppDownloadQRCodeTitle('android_tv')).toBe('TV 下载二维码')
    expect(getTVAppDownloadQRCodeTitle('android_phone')).toBe('手机端下载二维码')
  })

  it('keeps the current origin when not in dev mode', () => {
    const origin = resolveTVAppDownloadQRCodeOrigin({
      currentOrigin: 'http://192.168.1.24:8080',
      isDev: false,
      apiProxyTarget: 'http://10.0.0.2:8080'
    })

    expect(origin).toBe('http://192.168.1.24:8080')
    expect(buildTVAppDownloadPageURL('android_tv', {
      currentOrigin: 'http://192.168.1.24:8080',
      isDev: false,
      apiProxyTarget: 'http://10.0.0.2:8080'
    })).toBe('http://192.168.1.24:8080/downloads/android-tv')
  })

  it('falls back to the API proxy target origin in dev mode when available', () => {
    expect(buildTVAppDownloadPageURL('android_phone', {
      currentOrigin: 'http://192.168.1.8:5173',
      isDev: true,
      apiProxyTarget: 'http://192.168.1.24:8080/api/v1'
    })).toBe('http://192.168.1.24:8080/downloads/android-phone')
  })

  it('falls back to the current origin when the dev proxy target is missing or invalid', () => {
    expect(resolveTVAppDownloadQRCodeOrigin({
      currentOrigin: 'http://127.0.0.1:5173',
      isDev: true,
      apiProxyTarget: ''
    })).toBe('http://127.0.0.1:5173')

    expect(buildTVAppDownloadPageURL('android_tv', {
      currentOrigin: 'http://127.0.0.1:5173',
      isDev: true,
      apiProxyTarget: 'not-a-url'
    })).toBe('http://127.0.0.1:5173/downloads/android-tv')
  })
})
