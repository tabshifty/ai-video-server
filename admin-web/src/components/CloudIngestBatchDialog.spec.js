import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

// 结构化断言：把模板真解析一遍，而不是只用字符串包含。
// 触发原因：一次属性改写多留了一个引号（`="false""`），Vue 会把它当成一个名为
// `"` 的属性透传到根 div，挂载时抛 InvalidCharacterError——build 与字符串断言
// 都发现不了，只有真正解析模板才能拦住。
const source = readFileSync(new URL('./CloudIngestBatchDialog.vue', import.meta.url), 'utf8')
const template = source.match(/<template>([\s\S]*)<\/template>/)[1]

// 极简模板扫描：收集开始标签里的属性名，足以发现引号错位这类问题。
function collectAttributeNames(markup) {
  const names = []
  const tagPattern = /<([a-zA-Z][\w-]*)((?:\s+[^\s=>/]+(?:\s*=\s*"[^"]*")?)*)\s*\/?>/g
  let match
  while ((match = tagPattern.exec(markup)) !== null) {
    const attrPattern = /([^\s=]+)(?:\s*=\s*"([^"]*)")?/g
    let attr
    while ((attr = attrPattern.exec(match[2] ?? '')) !== null) {
      names.push(attr[1])
    }
  }
  return names
}

describe('CloudIngestBatchDialog 结构', () => {
  it('模板属性名合法，没有多余引号造成的伪属性', () => {
    const names = collectAttributeNames(template)
    expect(names.length).toBeGreaterThan(0)
    for (const name of names) {
      expect(name).toMatch(/^[a-zA-Z@:#][\w@:.#-]*$/)
    }
  })

  it('遮罩点击不关闭，关闭与冻结交给 submitting 守卫', () => {
    expect(template).toContain(':close-on-click-modal="false"')
    expect(template).not.toContain(':close-on-click-modal="false""')
    expect(template).toContain(':before-close="handleBeforeClose"')
    expect(template).toContain(':disabled="submitting"')
  })

  it('入库前置要求字段与可访问名称齐备', () => {
    for (const label of ['视频类型', '标题模式', '统一标题', '描述', '视频标签', '所属合集', '关联演员', '图片图集']) {
      expect(template).toContain(label)
    }
    expect(template).toContain('aria-label="视频类型"')
    expect(template).toContain('aria-label="标题模式"')
    expect(template).toContain('aria-label="AV 地区"')
    expect(template).toContain('aria-label="本批统计"')
  })

  it('短视须专属联动的合集只在 short 时渲染', () => {
    expect(template).toContain('v-if="isShortType"')
    expect(template).toContain('v-if="isAVType"')
  })

  it('无法识别分区默认不强制，并提供一键强制与逐条勾选', () => {
    expect(template).toContain('强制按视频：')
    expect(template).toContain('全部强制按视频')
    expect(template).toContain('全部取消')
  })

  it('超过 20 条时提示下载槽位占用', () => {
    expect(template).toContain('本批将占用')
  })

  it('四类候选都走远程检索，不会因为只拉一页而选不到', () => {
    for (const method of ['remoteTags', 'remoteActors', 'remoteCollections', 'remoteImageCollections']) {
      expect(template).toContain(`:remote-method="${method}"`)
      expect(source).toContain(`function ${method}(`)
    }
    expect(template).not.toContain('page_size: 100')
  })

  it('脚本侧使用新的文件名清洗与 payload 构造函数，不误用上传页旧语义', () => {
    expect(source).toContain('buildEd2kIngestPayload')
    expect(source).toContain('validateEd2kIngestForm')
    expect(source).not.toContain('buildTitleFromFilename')
  })
})
