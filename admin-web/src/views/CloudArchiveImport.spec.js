import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import CloudArchiveImport from './CloudArchiveImport.vue'

const source = readFileSync(new URL('./CloudArchiveImport.vue', import.meta.url), 'utf8')
describe('115 压缩包全屏导入页面', () => {
  it('真实 SFC 编译，并以任务 ID 查询导入状态和压缩包预览', () => {
    expect(CloudArchiveImport).toBeTruthy()
    expect(source).toContain('getCloudDownloadImportTask(taskID)')
    expect(source).toContain('previewCloudDownloadImport(taskID)')
    expect(source).toContain('router.replace(importDestination(result))')
    expect(source).not.toContain('listCloudDownloads(')
  })
  it('复用内联主分组表单，未提交编辑有离开保护', () => {
    expect(source).toContain('inline :model-value="true"')
    expect(source).toContain('onBeforeRouteLeave')
    expect(source).toContain('beforeunload')
    expect(source).toContain('@submit-error="handleSubmitError"')
    expect(source).not.toContain('<el-dialog')
    expect(source).not.toContain('<el-drawer')
  })
})
