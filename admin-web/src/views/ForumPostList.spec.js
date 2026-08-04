import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'
import ForumPostList from './ForumPostList.vue'

const source = readFileSync(new URL('./ForumPostList.vue', import.meta.url), 'utf8')
const script = source.match(/<script setup>[\s\S]*?<\/script>/)?.[0] || ''
const template = source.match(/<template>[\s\S]*<\/template>/)?.[0] || ''
const style = source.match(/<style scoped>[\s\S]*?<\/style>/)?.[0] || ''

describe('论坛资源列表页', () => {
  it('通过真实 SFC 编译并复用管理端读取能力', () => {
    expect(ForumPostList).toBeTruthy()
    expect(script).toContain("import Layout from '../components/Layout.vue'")
    expect(script).toContain("import AdminTablePagination from '../components/AdminTablePagination.vue'")
    expect(script).toContain("import { getAdminForumPosts } from '../api/admin'")
    expect(script).toContain("import { getEd2kLinkLabel } from './toolbox.helpers'")
  })

  it('固定每页 20 条并只提供进入加载、手动刷新与分页', () => {
    expect(script).toContain("const query = reactive({ page: 1, page_size: 20 })")
    expect(script).toContain('onMounted(load)')
    expect(script).not.toContain('setInterval')
    expect(template).toContain('aria-label="刷新论坛资源"')
    expect(template).toContain('title="刷新论坛资源"')
    expect(template).toContain('<AdminTablePagination')
    expect(template).not.toContain('<el-input')
    expect(template).not.toContain('<el-select')
  })

  it('一帖一行展示四个已确认字段', () => {
    expect([...template.matchAll(/<el-table-column[^>]*label="([^"]+)"/g)].map((match) => match[1])).toEqual([
      '标题',
      '文件链接',
      'ED2K',
      '发现时间'
    ])
    expect(template).toContain(':data="list"')
    expect(template).toContain('formatAdminDateTime(row.observed_at)')
  })

  it('使用原生锚点并保留重复资源的索引身份', () => {
    expect(template).toContain(':href="row.url"')
    expect(template).toContain('target="_blank"')
    expect(template).toContain('rel="noopener noreferrer"')
    expect(template).toContain('v-for="(attachment, index) in row.attachments"')
    expect(template).toContain(':href="attachment"')
    expect(template).toContain('文件 {{ index + 1 }}')
    expect(template).toContain('v-for="(ed2kLink, index) in row.ed2k_links"')
    expect(template).toContain(':href="ed2kLink"')
    expect(template).toContain('getEd2kLinkLabel(ed2kLink)')
    expect(template).toContain("`${row.id}:attachment:${index}`")
    expect(template).toContain("`${row.id}:ed2k:${index}`")
  })

  it('明确空资源、加载失败和首次空列表状态', () => {
    expect(template).toContain('暂无论坛资源')
    expect(template).toContain('最近 30 天内暂无可展示的论坛资源')
    expect(template).toContain("row.attachments?.length")
    expect(template).toContain("row.ed2k_links?.length")
    expect(template.match(/>无<\/span>/g)).toHaveLength(2)
    expect(template).toContain('<el-alert v-if="loadError"')
    expect(template).toContain('@click="load">重试</el-button>')
  })

  it('让长链接在单元格内收敛并为移动端保留可操作高度', () => {
    expect(style).toContain('overflow-wrap: anywhere;')
    expect(style).toContain('@media (max-width: 63.9375rem)')
    expect(style).toContain('min-height: 44px;')
  })
})
