import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import CloudDownloadImportDialog from './CloudDownloadImportDialog.vue'

const source = readFileSync(new URL('./CloudDownloadImportDialog.vue', import.meta.url), 'utf8')
const script = source.match(/<script setup>[\s\S]*?<\/script>/)?.[0] || ''
const template = source.match(/<template>[\s\S]*<\/template>/)?.[0] || ''
const style = source.match(/<style scoped>[\s\S]*?<\/style>/)?.[0] || ''

describe('115 云下载导入弹窗', () => {
  it('通过真实 SFC 编译并复用元数据候选接口', () => {
    expect(CloudDownloadImportDialog).toBeTruthy()
    expect(script).toContain('getAdminActors')
    expect(script).toContain('getAdminCollections')
    expect(script).toContain('getAdminImageCollections')
    expect(script).toContain('getAdminPopularVideoTags')
    expect(script).toContain('getAdminVideoTags')
    expect(script).toContain('createRemoteSuggestionLoader')
    expect(script).toContain('filterRemoteOptionsByValues')
  })

  it('只复用元数据，不复制文件上传流程', () => {
    expect(script).toContain('importCloudDownload')
    expect(script).not.toContain('uploadInit')
    expect(script).not.toContain('uploadChunk')
    expect(script).not.toContain('uploadComplete')
    expect(script).not.toContain('sha256File')
    expect(template).not.toContain('<el-upload')
  })

  it('视频弹窗字段与上传页一致（去掉文件）', () => {
    expect(template).toContain('label="类型"')
    expect(template).toContain('label="AV 地区分类"')
    expect(template).toContain('label="标题"')
    expect(template).toContain('label="描述"')
    expect(template).toContain('label="视频标签"')
    expect(template).toContain('label="图片图集"')
    expect(template).toContain('label="所属合集"')
    expect(template).toContain('label="关联演员"')
    expect(template).toContain('v-if="isShortType"')
    expect(template).toContain('v-if="isAVType"')
    expect(template).toContain('allow-create')
    expect(script).toContain("type: 'short'")
    expect(script).toContain("site_category: 'japanese'")
    expect(script).toContain('image_collection_id')
    expect(script).toContain('videoForm.collections')
  })

  it('压缩包配置保留主分组默认值和密码字段', () => {
    expect(template).toContain('label="视频默认标题"')
    expect(template).toContain('label="默认说明"')
    expect(template).toContain('label="默认标签"')
    expect(template).toContain('label="默认视频合集"')
    expect(template).toContain('label="默认图片合集"')
    expect(template).toContain('label="有密码"')
    expect(template).toContain('label="压缩包密码"')
    expect(script).toContain('default_video_collection_ids')
    expect(script).toContain('default_image_collection_ids')
    expect(script).toContain('has_password')
    expect(script).toContain('validateArchiveImportForm')
  })

  it('压缩包配置可内联新建视频/图片合集并自动选中', () => {
    expect(script).toContain('createAdminCollection')
    expect(script).toContain('createAdminImageCollection')
    expect(script).toContain('quickCollectionDialogVisible')
    expect(script).toContain('async function saveQuickCollection()')
    expect(script).toContain('archiveForm.default_video_collection_ids')
    expect(script).toContain('archiveForm.default_image_collection_ids')
    expect(template).toContain('新建视频合集')
    expect(template).toContain('新建图片合集')
    expect(template).toContain('v-model="quickCollectionDialogVisible"')
    expect(template).toContain(':inline="inline"')
    expect(script).toContain("import CloudImportFrame from './CloudImportFrame.vue'")
    expect(template).toContain('创建并选中')
  })

  it('默认标题取 local_path basename，提交走任务 import 接口', () => {
    expect(script).toContain('resolveCloudImportDefaultTitle')
    expect(script).toContain('buildVideoImportPayload')
    expect(script).toContain('buildArchiveImportPayload')
    expect(script).toContain("importCloudDownload(taskId, payload)")
    expect(script).toContain("ElMessage.success('已提交导入任务')")
    expect(script).toContain("emit('submitted'")
    expect(script).toContain("emit('update:modelValue', false)")
    expect(script).toContain('extractApiErrorMessage')
  })

  it('提交期间防重复点击并保持按钮 loading', () => {
    expect(script).toContain('if (submitting.value || quickCollectionSaving.value) return')
    expect(template).toContain(':loading="submitting"')
    expect(template).toContain(':disabled="submitting || quickCollectionSaving"')
    expect(script).toContain('handleBeforeClose')
  })

  it('窄屏表单标签改为纵向排列', () => {
    expect(style).toContain('@media (max-width: 48rem)')
    expect(style).toContain('.cloud-import-form :deep(.el-form-item)')
    expect(style).toContain('.cloud-import-form :deep(.el-form-item__label)')
  })
})
