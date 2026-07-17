import { readFileSync } from 'node:fs'
import { describe, expect, it } from 'vitest'

const tvAppManage = readFileSync(new URL('./TvAppManage.vue', import.meta.url), 'utf8')
const qrHelper = readFileSync(new URL('./tvAppManage.qr.js', import.meta.url), 'utf8')
const queryBlock = tvAppManage.match(/const query = reactive\(\{[\s\S]*?\n\}\)/)?.[0] || ''
const resetQueryBlock = tvAppManage.match(/function resetQuery\(\) \{[\s\S]*?\n\}/)?.[0] || ''
const uploadAPKBlock = tvAppManage.match(/async function uploadAPK[\s\S]*?\n}\n\nasync function saveNotes/)?.[0] || ''
const qrBlock = tvAppManage.match(/async function refreshDownloadQRCode\(\) \{[\s\S]*?\n\}/)?.[0] || ''
const loadBlock = tvAppManage.match(/async function load\(\) \{[\s\S]*?\n\}(?=\n\nfunction resetQuery)/)?.[0] || ''
const loadCatchBlock = loadBlock.match(/} catch \(error\) \{[\s\S]*?(?=\n  } finally \{)/)?.[0] || ''
const loadFinallyBlock = loadBlock.match(/} finally \{[\s\S]*?(?=\n  \}\n\})/)?.[0] || ''
const confirmActionBlock = tvAppManage.match(/async function confirmAction[\s\S]*?\n\}(?=\n\nfunction downloadHref)/)?.[0] || ''
const template = tvAppManage.match(/<template>([\s\S]*?)<\/template>\s*\n\s*<style scoped>/)?.[1] || ''
const style = tvAppManage.match(/<style scoped>([\s\S]*?)<\/style>/)?.[1] || ''

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

  it('使用紧凑工作区且不丢失安装包命令', () => {
    expect(tvAppManage).toContain('<template #header-actions>')
    expect(tvAppManage).toContain('data-density="compact"')
    expect(tvAppManage).not.toContain('<PageHeader')
    expect(tvAppManage).not.toContain('pageTitle:')
    expect(tvAppManage).not.toContain('pageSubtitle:')
    expect(tvAppManage).toContain('@click="uploadAPK(false)"')
    expect(tvAppManage).toContain("@click=\"confirmAction(row, 'publish')\"")
    expect(tvAppManage).toContain("@click=\"confirmAction(row, 'offline')\"")
    expect(tvAppManage).toContain("@click=\"confirmAction(row, 'delete')\"")
    expect(tvAppManage).toContain('下载 APK')
  })

  it('读取失败只显示行内错误并在刷新时保留已有安装包', () => {
    const alertIndex = template.indexOf('<el-alert v-if="loadError"')
    const qrIndex = template.indexOf('<template #title>{{ downloadQRCodeTitle }}</template>')
    const uploadIndex = template.indexOf('<template #title>上传 APK</template>')
    const skeletonIndex = template.indexOf('<el-skeleton v-if="initialLoading"')
    const contentIndex = template.indexOf('<template v-else-if="!loadError || hasCurrentResult">')
    const dataContent = template.slice(contentIndex)

    expect(tvAppManage).toContain("import { shouldShowCrudCollectionSkeleton } from './crudCollectionState'")
    expect(tvAppManage).toContain('const loading = ref(true)')
    expect(tvAppManage).toContain("const loadError = ref('')")
    expect(tvAppManage).toMatch(
      /const initialLoading = computed\(\(\) => shouldShowCrudCollectionSkeleton\(\{\s*loading: loading\.value,\s*rowCount: currentItems\.value\.length\s*\}\)\)/
    )
    expect(loadBlock.indexOf("loadError.value = ''")).toBeGreaterThanOrEqual(0)
    expect(loadBlock.indexOf("loadError.value = ''")).toBeLessThan(loadBlock.indexOf('try {'))
    expect(loadCatchBlock).toContain("loadError.value = extractErrorMessage(error, '加载安装包列表失败')")
    expect(loadCatchBlock).not.toContain('applyResult(')
    expect(loadCatchBlock).not.toContain('data.items')
    expect(loadCatchBlock).not.toContain('data.total_count')
    expect(loadCatchBlock).not.toContain('ElMessage.error')
    expect(alertIndex).toBeGreaterThanOrEqual(0)
    expect(qrIndex).toBeGreaterThan(alertIndex)
    expect(uploadIndex).toBeGreaterThan(qrIndex)
    expect(uploadIndex).toBeLessThan(skeletonIndex)
    expect(skeletonIndex).toBeGreaterThan(alertIndex)
    expect(contentIndex).toBeGreaterThan(skeletonIndex)
    expect(dataContent).toContain('<MetricStrip :items="summaryMetrics" aria-label="安装包摘要" />')
    expect(dataContent).toContain('<template #title>发布记录</template>')
    expect(dataContent).not.toContain('<template #title>{{ downloadQRCodeTitle }}</template>')
    expect(dataContent).not.toContain('<template #title>上传 APK</template>')
    expect(dataContent).toContain('<EmptyState v-if="currentTotalCount === 0"')
  })

  it('只渲染当前请求身份的缓存并拒绝迟到响应改写状态', () => {
    const requestIndex = loadBlock.indexOf('const request = buildTVAppReleaseRequest({')
    const activeIndex = loadBlock.indexOf('activeRequestKey.value = request.key')
    const fetchIndex = loadBlock.indexOf('const result = await getAdminTVAppReleases(request.params)')
    const successGuardIndex = loadBlock.indexOf('if (!isCurrentTVAppReleaseRequest(token, latestRequest)) return', fetchIndex)
    const applyIndex = loadBlock.indexOf('applyResult(result)')
    const cacheIndex = loadBlock.indexOf('cachedRequestKey.value = request.key')

    expect(tvAppManage).toContain("from './tvAppManage.requestState'")
    expect(tvAppManage).toContain('let loadSequence = 0')
    expect(tvAppManage).toContain("const activeRequestKey = ref('')")
    expect(tvAppManage).toContain('const cachedRequestKey = ref(null)')
    expect(tvAppManage).toContain('const currentCollection = computed(() => selectTVAppReleaseCache({')
    expect(tvAppManage).toContain('activeRequestKey: activeRequestKey.value')
    expect(tvAppManage).toContain('cachedRequestKey: cachedRequestKey.value')
    expect(tvAppManage).toContain('const hasCurrentResult = computed(() => currentCollection.value.hasCurrentResult)')
    expect(tvAppManage).toContain('const currentItems = computed(() => currentCollection.value.items)')
    expect(tvAppManage).toContain('const currentTotalCount = computed(() => currentCollection.value.totalCount)')
    expect(template).toContain(':data="currentItems"')
    expect(template).not.toContain(':data="data.items"')

    expect(requestIndex).toBeGreaterThanOrEqual(0)
    expect(activeIndex).toBeGreaterThan(requestIndex)
    expect(fetchIndex).toBeGreaterThan(activeIndex)
    expect(successGuardIndex).toBeGreaterThan(fetchIndex)
    expect(applyIndex).toBeGreaterThan(successGuardIndex)
    expect(cacheIndex).toBeGreaterThan(applyIndex)
    expect(loadCatchBlock).toContain('if (!isCurrentTVAppReleaseRequest(token, latestRequest)) return')
    expect(loadCatchBlock.indexOf('if (!isCurrentTVAppReleaseRequest(token, latestRequest)) return')).toBeLessThan(
      loadCatchBlock.indexOf("loadError.value = extractErrorMessage(error, '加载安装包列表失败')")
    )
    expect(loadFinallyBlock).toContain('if (isCurrentTVAppReleaseRequest(token, latestRequest))')
    expect(loadFinallyBlock.indexOf('if (isCurrentTVAppReleaseRequest(token, latestRequest))')).toBeLessThan(
      loadFinallyBlock.indexOf('loading.value = false')
    )
  })

  it('按当前查询总数判断真正空态并保留非零总数的分页入口', () => {
    const releaseSection = template.slice(template.indexOf('<template #title>发布记录</template>'))

    expect(releaseSection).toContain('<EmptyState v-if="currentTotalCount === 0"')
    expect(releaseSection).toContain('<template v-else>')
    expect(releaseSection).toContain(':data="currentItems"')
    expect(releaseSection).toContain(':total="currentTotalCount"')
    expect(releaseSection).not.toContain('v-if="data.items.length === 0"')
    expect(releaseSection).not.toContain('v-if="!data.total_count"')
  })

  it('摘要只展示既有事实并逐项标明当前筛选或当前页口径', () => {
    expect(tvAppManage).toContain("import MetricStrip from '../components/base/MetricStrip.vue'")
    expect(tvAppManage).toContain('<MetricStrip :items="summaryMetrics" aria-label="安装包摘要" />')
    expect(tvAppManage).toContain("{ key: 'total', label: '记录总数', value: currentTotalCount.value, scope: '当前筛选·全部页' }")
    expect(tvAppManage).toContain("{ key: 'visible', label: '家庭可见', value: visibleCount.value, scope: '当前页' }")
    expect(tvAppManage).toContain("scope: '当前页'")
    expect(tvAppManage).toContain("{ key: 'recommended', label: '推荐版本', value: latestItem.value ? `${latestItem.value.version_name} (${latestItem.value.version_code})` : '暂无', scope: '当前页' }")
    expect(tvAppManage).not.toContain('健康度')
    expect(tvAppManage).not.toContain('告警')
  })

  it('状态有文字且四项发布命令进入更多操作菜单', () => {
    expect(tvAppManage).toContain("import StatusIndicator from '../components/base/StatusIndicator.vue'")
    expect(template).toContain('<StatusIndicator :label="statusText(row.publish_status)" :tone="statusTone(row.publish_status)" />')
    expect(template).toContain('<el-dropdown trigger="click"')
    expect(template).toContain('更多操作')
    expect(template).toContain("@click=\"confirmAction(row, 'publish')\"")
    expect(template).toContain("@click=\"confirmAction(row, 'offline')\"")
    expect(template).toContain("@click=\"confirmAction(row, 'restore')\"")
    expect(template).toContain("@click=\"confirmAction(row, 'delete')\"")
    expect(template).toContain(':href="downloadHref(row, abi.abi)"')
  })

  it('长版本与 ABI 文本独立收敛且保留完整值提示', () => {
    const compactTextRule = style.match(/\.compact-text\s*\{[^}]*\}/s)?.[0] || ''
    const tableRowRule = style.match(/\.package-table\s+:deep\(\.el-table__row\)\s*\{[^}]*\}/s)?.[0] || ''

    expect(template).toContain('<el-tooltip :content="row.version_name || \'--\'" placement="top">')
    expect(template).toMatch(
      /class="compact-text version-name"\s+tabindex="0"\s+:aria-label="`版本名称：\$\{row\.version_name \|\| '--'\}`"/
    )
    expect(template).toContain('<el-tooltip :content="`版本号：${row.version_code ?? \'--\'}`" placement="top">')
    expect(template).toMatch(
      /class="compact-text version-code"\s+tabindex="0"\s+:aria-label="`版本号：\$\{row\.version_code \?\? '--'\}`"/
    )
    expect(template).toContain('<el-tag v-if="row.latest_recommended" size="small" type="success">推荐</el-tag>')
    expect(template).toContain('<el-tooltip :content="abiLine(row)" placement="top">')
    expect(template).toMatch(
      /class="abi-line compact-text"\s+tabindex="0"\s+:aria-label="`ABI 概览：\$\{abiLine\(row\)\}`"/
    )
    expect(template).toContain(':content="`${abi.abi} ${formatBytes(abi.file_size)}`"')
    expect(template).toMatch(
      /class="abi-entry compact-text"\s+tabindex="0"\s+:aria-label="`ABI 文件：\$\{abi\.abi\}，大小 \$\{formatBytes\(abi\.file_size\)\}`"/
    )
    expect(compactTextRule).toContain('min-width: 0;')
    expect(compactTextRule).toContain('overflow: hidden;')
    expect(compactTextRule).toContain('text-overflow: ellipsis;')
    expect(compactTextRule).toContain('white-space: nowrap;')
    expect(tableRowRule).not.toContain('overflow: hidden;')
  })

  it('危险确认取消或关闭时终止动作且不产生未处理拒绝', () => {
    const confirmIndex = confirmActionBlock.indexOf('await ElMessageBox.confirm(')
    const catchIndex = confirmActionBlock.indexOf('} catch (error) {')
    const cancelIndex = confirmActionBlock.indexOf("if (error === 'cancel' || error === 'close')")
    const actionIndex = confirmActionBlock.lastIndexOf('await runDangerAction(item, action)')

    expect(confirmIndex).toBeGreaterThanOrEqual(0)
    expect(catchIndex).toBeGreaterThan(confirmIndex)
    expect(cancelIndex).toBeGreaterThan(catchIndex)
    expect(confirmActionBlock).toContain("ElMessage.error('操作确认失败，请重试')")
    expect(actionIndex).toBeGreaterThan(cancelIndex)
  })

  it('替换上传仍是页面文件选择器的全局命令', () => {
    expect(template).toContain('@click="uploadAPK(false)">上传并建档</el-button>')
    expect(template).toContain('@click="uploadAPK(true)">替换上传</el-button>')
    expect(tvAppManage).not.toContain('uploadAPK(row')
    expect(uploadAPKBlock).toContain("if (replaceExisting) formData.append('replace_existing', 'true')")
  })
})
