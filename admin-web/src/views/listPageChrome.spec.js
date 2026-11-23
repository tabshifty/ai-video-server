import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const listPages = [
  { file: 'VideoList.vue', redundant: [
    '<template #title>视频列表</template>',
    '<span class="result-total">共 {{ total }} 条</span>',
  ] },
  { file: 'ImageManage.vue', redundant: [
    '<template #title>图片资产</template>',
  ] },
  { file: 'ActorManage.vue', redundant: [
    '<template #title>演员列表</template>',
    '<template #description>支持演员资料的新增、编辑、停用与刮削回填</template>',
  ] },
  { file: 'CollectionManage.vue', redundant: [
    '<template #title>合集列表</template>',
    '<template #description>支持短视频合集的新增、编辑、停用与删除</template>',
  ] },
  { file: 'UserManage.vue', redundant: [
    '<template #title>账号列表</template>',
    '<template #description>可直接调整用户角色，变更即时生效。</template>',
  ] },
  { file: 'TaskMonitor.vue', redundant: [
    '<template #title>任务列表</template>',
  ] },
  { file: 'ForumPostList.vue', redundant: [
    '<template #title>资源列表</template>',
    '<el-tag effect="plain">共 {{ total }} 条</el-tag>',
  ] },
  { file: 'CloudDownloadList.vue', redundant: [
    '<template #title>云下载任务</template>',
    '<el-tag effect="plain">共 {{ total }} 条</el-tag>',
  ] },
  { file: 'ImageCollectionManage.vue', redundant: [
    '<template #title>合集列表</template>',
    '<template #description>维护图片合集并统一管理图片归档关系。</template>',
  ] },
  { file: 'ToolboxPasswordVault.vue', redundant: [
    '<template #title>密码库</template>',
    '<template #description>支持新增、修改、删除、显示密码和复制账号密码。</template>',
  ] },
]

describe('单表列表页信息层级', () => {
  it('不重复页头标题或泛说明', () => {
    listPages.forEach(({ file, redundant }) => {
      const source = readFileSync(new URL(`./${file}`, import.meta.url), 'utf8')
      redundant.forEach((snippet) => {
        expect(source, `${file} should not repeat: ${snippet}`).not.toContain(snippet)
      })
    })
  })

  it('分页栏承接总数与每页条数', () => {
    listPages.forEach(({ file }) => {
      const source = readFileSync(new URL(`./${file}`, import.meta.url), 'utf8')
      expect(source, file).toMatch(/<AdminTablePagination[\s\S]*?layout="total[^"]*"/)
    })
  })
})
