import { createSSRApp, h } from 'vue'
import { renderToString } from '@vue/server-renderer'
import { describe, expect, it } from 'vitest'
import CloudImportFrame from './CloudImportFrame.vue'

// 实际渲染验证内联模式不创建遮罩/表单弹窗，同时保留操作按钮。
describe('云导入全屏表单容器', () => {
  it('内联配置及合集创建渲染为页面内容和操作区', async () => {
    const html = await renderToString(createSSRApp({ render: () => h(CloudImportFrame, { inline: true, modelValue: true, title: '压缩包入媒体库' }, {
      default: () => h('input', { 'aria-label': '默认标题' }),
      footer: () => h('button', '提交导入')
    }) }))
    expect(html).toContain('<section')
    expect(html).toContain('默认标题')
    expect(html).toContain('提交导入')
    expect(html).not.toContain('el-dialog')
    expect(html).not.toContain('el-overlay')
  })
  it('内联新建合集关闭后不残留可操作表单', async () => {
    const html = await renderToString(createSSRApp({ render: () => h(CloudImportFrame, { inline: true, modelValue: false }, { default: () => h('input', { 'aria-label': '合集名称' }) }) }))
    expect(html).not.toContain('<input')
    expect(html).not.toContain('el-dialog')
  })
})
