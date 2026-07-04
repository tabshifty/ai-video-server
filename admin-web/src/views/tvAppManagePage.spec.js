import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const tvAppManage = readFileSync(new URL('./TvAppManage.vue', import.meta.url), 'utf8')
const qrHelper = readFileSync(new URL('./tvAppManage.qr.js', import.meta.url), 'utf8')
const queryBlock = tvAppManage.match(/const query = reactive\(\{[\s\S]*?\n\}\)/)?.[0] || ''
const resetQueryBlock = tvAppManage.match(/function resetQuery\(\) \{[\s\S]*?\n\}/)?.[0] || ''
const uploadAPKBlock = tvAppManage.match(/async function uploadAPK[\s\S]*?\n}\n\nasync function saveNotes/)?.[0] || ''
const qrBlock = tvAppManage.match(/async function refreshDownloadQRCode\(\) \{[\s\S]*?\n\}/)?.[0] || ''

describe('TV app package management page', () => {
  it('defaults to the full release list so uploaded draft releases are visible', () => {
    expect(queryBlock).toContain('current_published: false')
    expect(resetQueryBlock).toContain('query.current_published = false')
    expect(tvAppManage).toContain('active-text="只看家庭可见"')
    expect(tvAppManage).toContain('inactive-text="查看全部"')
    expect(tvAppManage).toContain('默认查看全部记录')
    expect(tvAppManage).not.toContain('默认只看家庭可见')
    expect(tvAppManage).not.toContain('默认先看当前家庭可见记录')
  })

  it('returns to the first page after uploading without changing the current visibility filter', () => {
    const uploadIndex = uploadAPKBlock.indexOf('await uploadAdminTVAppAPK(formData, clientType.value)')
    const firstPageIndex = uploadAPKBlock.indexOf('query.page = 1')
    const reloadIndex = uploadAPKBlock.indexOf('await load()')

    expect(uploadIndex).toBeGreaterThanOrEqual(0)
    expect(firstPageIndex).toBeGreaterThan(uploadIndex)
    expect(reloadIndex).toBeGreaterThan(firstPageIndex)
    expect(uploadAPKBlock).not.toContain('query.current_published = false')
    expect(uploadAPKBlock).not.toContain('query.current_published = true')
  })

  it('renders a single client-scoped qr card above the upload section', () => {
    const qrTitleIndex = tvAppManage.indexOf('<template #title>{{ downloadQRCodeTitle }}</template>')
    const uploadTitleIndex = tvAppManage.indexOf('<template #title>上传 APK</template>')

    expect(tvAppManage).toContain("getTVAppDownloadQRCodeTitle(clientType.value)")
    expect(qrHelper).toContain('TV 下载二维码')
    expect(qrHelper).toContain('手机端下载二维码')
    expect(tvAppManage).toContain('downloadQRCodeDataURL')
    expect(tvAppManage).toContain('download-qr-card')
    expect(tvAppManage).toContain(':alt="downloadQRCodeTitle"')
    expect(qrTitleIndex).toBeGreaterThanOrEqual(0)
    expect(uploadTitleIndex).toBeGreaterThan(qrTitleIndex)
  })

  it('builds qr urls from the current origin and falls back to the dev api target when needed', () => {
    expect(qrBlock).toContain('window.location.origin')
    expect(qrBlock).toContain('import.meta.env.DEV')
    expect(qrBlock).toContain('import.meta.env.VITE_API_PROXY_TARGET')
    expect(qrBlock).toContain('buildTVAppDownloadPageURL')
    expect(tvAppManage).not.toContain('download-link-address')
    expect(tvAppManage).not.toContain('打开下载页')
  })
})
