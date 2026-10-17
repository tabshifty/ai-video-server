import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import CloudDownloadList from './CloudDownloadList.vue'

const source = readFileSync(new URL('./CloudDownloadList.vue', import.meta.url), 'utf8')
const script = source.match(/<script setup>[\s\S]*?<\/script>/)?.[0] || ''
const template = source.match(/<template>[\s\S]*<\/template>/)?.[0] || ''
const style = source.match(/<style scoped>[\s\S]*?<\/style>/)?.[0] || ''

describe('115 云下载列表页', () => {
  it('通过真实 SFC 编译并复用管理端接口', () => {
    expect(CloudDownloadList).toBeTruthy()
    expect(script).toContain("import AdminTablePagination from '../components/AdminTablePagination.vue'")
    expect(script).toContain("import Layout from '../components/Layout.vue'")
    expect(script).toContain('listCloudDownloads')
    expect(script).toContain('retryCloudDownload')
    expect(script).toContain('getCloudBrowserState')
    expect(script).toContain('getCloudLoginQR')
  })

  it('失败和需处理任务都可以重试', () => {
    expect(template).toContain("row.status === 'failed' || row.status === 'blocked'")
    expect(template).toContain('@click="retryTask(row)"')
  })

  it('展开行展示失败原因、落地路径与触发信息', () => {
    expect(template).toContain('row.error_message')
    expect(template).toContain('row.local_path')
    expect(template).toContain('row.pick_code')
    expect(template).toContain('row.retry_count')
  })

  it('保留自动刷新、二维码倒计时和响应式布局', () => {
    expect(script).toContain('POLL_INTERVAL_MS')
    expect(script).toContain('syncPolling')
    expect(script).toContain('startQrCountdown')
    expect(style).toContain('@media (max-width: 63.9375rem)')
    expect(style).toContain('.cloud-detail__path')
  })

  it('按 source 显示工具箱、手动或原帖来源', () => {
    expect(template).toContain('label="来源"')
    expect(template).toContain("row.source === 'toolbox'")
    expect(template).toContain("row.source === 'manual'")
    expect(template).toContain("row.post_title || '原帖已删除'")
  })
})
