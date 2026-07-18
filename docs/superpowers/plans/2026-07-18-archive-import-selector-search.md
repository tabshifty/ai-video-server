# 压缩包导入选择框搜索候选收敛实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 修复 `ToolboxArchiveImport` 远程选择器搜索后持续显示历史未选候选的问题，同时保留当前已选标签和合集的可读名称。

**Architecture:** 保持 `createRemoteSuggestionLoader` 的防抖、请求序号和 `mergeOptions(existing, incoming)` 接口不变。新增一个通用的纯过滤函数，只从压缩包页当前四个表单模型的已选值中提取旧选项；三个压缩包 loader 将“已选旧选项 + 当前接口结果”交给现有合并器，普通 `VideoUpload` 不改接线。

**Tech Stack:** Vue 3 `script setup`、Element Plus `el-select`、Vitest、Vite。

## Global Constraints

- 只修改管理端远程候选显示，不修改后端 API、查询参数、请求 payload、上传中心行为或 Android 模块。
- 保留远程搜索 180ms 防抖和 latest-wins 旧请求丢弃规则。
- 远程候选按大小写不敏感的规范化值匹配；返回选项对象时保留原对象和显示标签。
- 所有新增测试和界面文案使用中文，文件保持有效 UTF-8 且不得出现替换字符。
- 完成后必须更新 `plan.md`，精确提交本次文件，不纳入现有 `android-tv-app` 删除和 `.superpowers/` 未跟踪目录。

---

## 文件职责

- `admin-web/src/views/videoUpload.remote.js`：新增无副作用的远程选项按已选值过滤 helper；不改变现有合并器和加载器的时序行为。
- `admin-web/src/views/videoUpload.remote.spec.js`：验证字符串选项、值对象选项、空输入和大小写/空白规范化。
- `admin-web/src/views/ToolboxArchiveImport.vue`：收集上传、单文件、批量编辑、分组默认值四类表单的当前已选值，并将过滤结果接入标签、视频合集、图片合集三个 loader。
- `admin-web/src/views/ToolboxArchiveImport.spec.js`：验证压缩包页三个 loader 使用已选值过滤结果，而不是完整候选数组；保留既有上传、保存和 UI 契约。
- `docs/superpowers/specs/2026-07-18-archive-import-selector-search-design.md`：已确认的行为边界和非目标。
- `CONTEXT.md`：已记录压缩包远程选择候选边界，无需重复追加相同术语。
- `plan.md`：追加红灯、实现、验证和提交记录。

## Task 1: 添加候选过滤红灯测试

**Files:**
- Modify: `admin-web/src/views/videoUpload.remote.spec.js`
- Modify: `admin-web/src/views/ToolboxArchiveImport.spec.js`

**Interfaces:**
- Expected new export: `filterRemoteOptionsByValues(options, selectedValues, getValue)` returns a filtered array and defaults `getValue` to the item itself.
- The archive page contract must use `filterRemoteOptionsByValues` for `tagOptions`, `collectionOptions` and `imageCollectionOptions` before calling `mergeOptions`.

- [ ] **Step 1: Write the failing pure-helper tests**

Add a namespace import so the test fails as a missing behavior rather than changing an existing named import:

```js
import * as remoteOptions from './videoUpload.remote'
```

Add this suite before the existing `createRemoteSuggestionLoader` suite:

```js
describe('filterRemoteOptionsByValues', () => {
  it('只保留当前已选的字符串候选并规范化空白和大小写', () => {
    expect(typeof remoteOptions.filterRemoteOptionsByValues).toBe('function')
    expect(remoteOptions.filterRemoteOptionsByValues(
      [' 剧情 ', '动作', '爱情'],
      [' 动作 ']
    )).toEqual(['动作'])
  })

  it('按值字段保留当前已选对象并忽略空值', () => {
    expect(remoteOptions.filterRemoteOptionsByValues(
      [
        { value: 'collection-a', label: '甲' },
        { value: 'collection-b', label: '乙' },
        { value: '', label: '无效' }
      ],
      [' COLLECTION-B '],
      (item) => item.value
    )).toEqual([{ value: 'collection-b', label: '乙' }])
  })
})
```

- [ ] **Step 2: Write the failing archive-page wiring test**

Add this test to `admin-web/src/views/ToolboxArchiveImport.spec.js`:

```js
it('远程搜索只把当前已选值作为历史候选传入合并器', () => {
  expect(source).toContain('function selectedArchiveTagValues()')
  expect(source).toContain('function selectedArchiveCollectionValues()')
  expect(source).toContain('function selectedArchiveImageCollectionValues()')
  expect(source).toContain(
    'filterRemoteOptionsByValues(tagOptions.value, selectedArchiveTagValues())'
  )
  expect(source).toMatch(
    /filterRemoteOptionsByValues\(\s*collectionOptions\.value,\s*selectedArchiveCollectionValues\(\),\s*\(item\) => item\?\.value\s*\)/
  )
  expect(source).toMatch(
    /filterRemoteOptionsByValues\(\s*imageCollectionOptions\.value,\s*selectedArchiveImageCollectionValues\(\),\s*\(item\) => item\?\.value\s*\)/
  )
  expect(source).not.toContain('getOptions: () => tagOptions.value')
  expect(source).not.toContain('getOptions: () => collectionOptions.value')
  expect(source).not.toContain('getOptions: () => imageCollectionOptions.value')
})
```

- [ ] **Step 3: Run the focused tests and verify the failure is correct**

Run:

```bash
cd admin-web && npm test -- src/views/videoUpload.remote.spec.js src/views/ToolboxArchiveImport.spec.js
```

Expected: Vitest exits with failure. The pure-helper test reports that `filterRemoteOptionsByValues` is not a function, and the archive-page test reports the missing selected-value loader wiring. Existing remote latest-wins and archive behavior tests must remain collected without syntax or environment errors.

## Task 2: Implement selected-value filtering and archive loader wiring

**Files:**
- Modify: `admin-web/src/views/videoUpload.remote.js`
- Modify: `admin-web/src/views/ToolboxArchiveImport.vue:45,1951-2026`
- Test: `admin-web/src/views/videoUpload.remote.spec.js`
- Test: `admin-web/src/views/ToolboxArchiveImport.spec.js`

**Interfaces:**
- `filterRemoteOptionsByValues(options, selectedValues, getValue = (item) => item)` returns the original option objects whose normalized selected key is present in `selectedValues`.
- `selectedArchiveTagValues()` returns normalized selected tag strings from `uploadForm.default_tags`, `selectedFile.value?.tags`, `batchEditForm.tags` and `archiveGroupForm.tags`.
- `selectedArchiveCollectionValues()` returns normalized video collection IDs from `uploadForm.default_video_collection_ids`, `selectedFile.value?.video_collection_ids`, `batchEditForm.video_collection_ids` and `archiveGroupForm.video_collection_ids`.
- `selectedArchiveImageCollectionValues()` returns normalized image collection IDs from upload defaults, the selected file, `batchEditForm.video_image_collection_id`, `batchEditForm.image_collection_ids` and `archiveGroupForm.image_collection_ids`.

- [ ] **Step 1: Add the minimal pure helper**

Add this function to `admin-web/src/views/videoUpload.remote.js` without changing the existing merge functions:

```js
export function filterRemoteOptionsByValues(options, selectedValues, getValue = (item) => item) {
  const selected = new Set(
    (selectedValues || [])
      .map((value) => String(value || '').trim().toLowerCase())
      .filter(Boolean)
  )

  return (options || []).filter((item) => {
    const value = String(getValue(item) || '').trim().toLowerCase()
    return value !== '' && selected.has(value)
  })
}
```

- [ ] **Step 2: Import the helper in the archive page**

Extend the existing import in `ToolboxArchiveImport.vue`:

```js
import {
  createRemoteSuggestionLoader,
  filterRemoteOptionsByValues,
  mergeRemoteStringOptions,
  mergeRemoteValueOptions
} from './videoUpload.remote'
```

- [ ] **Step 3: Add selected-value collectors before the three loader declarations**

Insert the following functions after `searchImageCollections` and before `const loadTagSuggestions`:

```js
function selectedArchiveTagValues() {
  return [
    ...normalizeTagSelection(uploadForm.default_tags),
    ...normalizeTagSelection(selectedFile.value?.tags),
    ...normalizeTagSelection(batchEditForm.tags),
    ...normalizeTagSelection(archiveGroupForm.tags)
  ]
}

function selectedArchiveCollectionValues() {
  return [
    ...normalizeUUIDSelection(uploadForm.default_video_collection_ids),
    ...normalizeUUIDSelection(selectedFile.value?.video_collection_ids),
    ...normalizeUUIDSelection(batchEditForm.video_collection_ids),
    ...normalizeUUIDSelection(archiveGroupForm.video_collection_ids)
  ]
}

function selectedArchiveImageCollectionValues() {
  return [
    ...normalizeUUIDSelection(uploadForm.default_image_collection_ids),
    ...normalizeUUIDSelection(selectedFile.value?.image_collection_ids),
    ...normalizeUUIDSelection([batchEditForm.video_image_collection_id]),
    ...normalizeUUIDSelection(batchEditForm.image_collection_ids),
    ...normalizeUUIDSelection(archiveGroupForm.image_collection_ids)
  ]
}
```

- [ ] **Step 4: Change only the three archive loader `getOptions` callbacks**

Use these exact callbacks and leave `fetcher`, `setOptions`, `setLoading` and `mergeOptions` unchanged:

```js
getOptions: () => filterRemoteOptionsByValues(tagOptions.value, selectedArchiveTagValues())
```

```js
getOptions: () => filterRemoteOptionsByValues(
  collectionOptions.value,
  selectedArchiveCollectionValues(),
  (item) => item?.value
)
```

```js
getOptions: () => filterRemoteOptionsByValues(
  imageCollectionOptions.value,
  selectedArchiveImageCollectionValues(),
  (item) => item?.value
)
```

Do not change the corresponding `VideoUpload.vue` callbacks. The existing archive form normalizers already provide UUID and tag canonicalization; no new payload normalization is needed.

- [ ] **Step 5: Run the focused tests and verify GREEN**

Run:

```bash
cd admin-web && npm test -- src/views/videoUpload.remote.spec.js src/views/ToolboxArchiveImport.spec.js
```

Expected: both files pass, including the new helper behavior, archive loader wiring, remote debounce/latest-wins tests, and existing archive import contracts.

- [ ] **Step 6: Commit the implementation**

Run:

```bash
git add admin-web/src/views/videoUpload.remote.js admin-web/src/views/videoUpload.remote.spec.js admin-web/src/views/ToolboxArchiveImport.vue admin-web/src/views/ToolboxArchiveImport.spec.js plan.md
git commit -m "修复压缩包选择框搜索候选"
```

Expected: one Chinese commit containing only the implementation files, tests and the new `plan.md` progress entry. Do not stage the existing TV splash deletion or `.superpowers/` directory.

## Task 3: Verify the finished change

**Files:**
- Verify: `admin-web/src/views/videoUpload.remote.js`
- Verify: `admin-web/src/views/videoUpload.remote.spec.js`
- Verify: `admin-web/src/views/ToolboxArchiveImport.vue`
- Verify: `admin-web/src/views/ToolboxArchiveImport.spec.js`
- Modify: `plan.md`

- [ ] **Step 1: Run the complete admin-web test suite**

Run:

```bash
cd admin-web && npm test
```

Expected: Vitest exits 0 with all existing and new test files passing.

- [ ] **Step 2: Build the admin web**

Run:

```bash
cd admin-web && npm run build
```

Expected: Vite build exits 0; the existing chunk-size warning may remain, but there are no compile errors.

- [ ] **Step 3: Run repository hygiene checks**

Run:

```bash
git diff --check
rg -n $'\\uFFFD' admin-web/src/views/videoUpload.remote.js admin-web/src/views/videoUpload.remote.spec.js admin-web/src/views/ToolboxArchiveImport.vue admin-web/src/views/ToolboxArchiveImport.spec.js plan.md
git status --short
```

Expected: `git diff --check` passes, the replacement-character scan returns no matches, and status shows only the pre-existing TV splash deletion and `.superpowers/` directory after the implementation commit.

- [ ] **Step 4: Append the final verification record**

Add a new reverse-chronological entry to `plan.md` recording the focused RED/GREEN result, complete `npm test`, `npm run build`, hygiene checks, implementation commit, and the fact that unrelated worktree changes were not staged.

- [ ] **Step 5: Commit the verification record**

Run:

```bash
git add plan.md
git commit -m "记录压缩包选择框搜索验证"
```

Expected: the progress record is committed separately, and `git status --short` still reports only unrelated pre-existing worktree items.
