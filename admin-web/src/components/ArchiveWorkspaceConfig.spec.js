import { createSSRApp, h } from 'vue'
import { renderToString } from '@vue/server-renderer'
import { describe, expect, it } from 'vitest'
import ArchiveWorkspaceConfig from './ArchiveWorkspaceConfig.vue'

async function renderConfig(root, modelValue = {}) {
  const app = createSSRApp({ render: () => h(ArchiveWorkspaceConfig, {
    root, modelValue, videoOptions: [{ id: 'v1', name: '视频目标' }],
    imageOptions: [{ id: 'i1', name: '图片目标' }], tagOptions: ['示例标签']
  }) })
  for (const name of ['el-select', 'el-option', 'el-button', 'el-checkbox', 'el-input']) {
    app.component(name, { setup: (_, { attrs, slots }) => () => h('div', attrs, slots.default?.()) })
  }
  return renderToString(app)
}

describe('分组独立配置编辑', () => {
  it('接口省略空字段时仍能选择标签及两类合集', async () => {
    const html = await renderConfig(true, { title_mode: 'filename', title: '组标题' })
    for (const label of ['标签', '视频合集', '图片合集']) expect(html).toContain(`aria-label="${label}"`)
    expect(html).toContain('新建视频合集')
    expect(html).toContain('新建图片合集')
    expect(html).not.toContain('采用所属分组配置')
  })
  it('文件未设置字段时仍保留继承，显式空数组显示编辑控件', async () => {
    const inherited = await renderConfig(false)
    expect(inherited).toContain('采用所属分组配置')
    expect(inherited).not.toContain('aria-label="标签"')
    const overridden = await renderConfig(false, { tags: [], video_collection_ids: [], image_collection_ids: [] })
    for (const label of ['标签', '视频合集', '图片合集']) expect(overridden).toContain(`aria-label="${label}"`)
  })
})
