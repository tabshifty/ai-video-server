<script setup lang="ts">
import { computed } from 'vue'
const model = defineModel<Record<string, any>>({ required: true })
const props = defineProps<{ root?: boolean; image?: boolean; disabled?: boolean; videoOptions: any[]; imageOptions: any[]; tagOptions: string[] }>()
const emit = defineEmits(['search-video', 'search-image', 'search-tags', 'create'])
const fields = computed(() => [
  { key: 'title_mode', label: '标题模式', empty: 'uniform', kind: 'mode' },
  { key: 'title', label: '统一标题', empty: '', kind: 'text' },
  { key: 'description', label: '说明', empty: '', kind: 'description' },
  ...(!props.image ? [{ key: 'tags', label: '标签', empty: [], kind: 'tags' }, { key: 'video_type', label: '视频类型', empty: 'short', kind: 'type' }, { key: 'video_collection_ids', label: '视频合集', empty: [], kind: 'video' }] : []),
  { key: 'image_collection_ids', label: props.image ? '所属图片合集' : '关联图片合集', empty: [], kind: 'image' }
])
function toggle(field: any, override: boolean) {
  const next = { ...model.value }
  if (override) next[field.key] = Array.isArray(field.empty) ? [] : field.empty
  else delete next[field.key]
  model.value = next
}
function options(kind: string, key: string) {
  const list = kind === 'video' ? props.videoOptions : props.imageOptions
  const selected = model.value[key] || []
  return [...list, ...selected.filter((id: string) => !list.some((o: any) => o.id === id)).map((id: string) => ({ id, name: id }))]
}
</script>

<template>
  <div class="workspace-config">
    <div v-for="field in fields" :key="field.key" class="config-field">
      <div class="config-field__label">
        <span>{{ field.label }}</span>
        <el-checkbox v-if="!root" :model-value="model[field.key] !== undefined" :disabled="disabled" @change="toggle(field, Boolean($event))">单独设置</el-checkbox>
      </div>
      <small v-if="model[field.key] === undefined">继承上级配置</small>
      <template v-else>
        <el-select v-if="field.kind === 'mode'" v-model="model[field.key]" :disabled="disabled" aria-label="标题模式">
          <el-option label="统一标题" value="uniform" /><el-option label="按文件名" value="filename" />
        </el-select>
        <el-select v-else-if="field.kind === 'type'" v-model="model[field.key]" :disabled="disabled" aria-label="视频类型">
          <el-option label="短视频" value="short" /><el-option label="电影" value="movie" /><el-option label="剧集" value="episode" /><el-option label="AV" value="av" />
        </el-select>
        <el-select v-else-if="field.kind === 'tags'" v-model="model[field.key]" multiple filterable remote allow-create default-first-option :remote-method="(q: string) => emit('search-tags', q)" :disabled="disabled" aria-label="标签">
          <el-option v-for="tag in [...new Set([...tagOptions, ...(model[field.key] || [])])]" :key="tag" :label="tag" :value="tag" />
        </el-select>
        <template v-else-if="field.kind === 'video' || field.kind === 'image'">
          <el-select v-model="model[field.key]" multiple filterable remote clearable :remote-method="(q: string) => emit(field.kind === 'video' ? 'search-video' : 'search-image', q)" :disabled="disabled" :aria-label="field.label" placeholder="可不选，可多选">
            <el-option v-for="o in options(field.kind, field.key)" :key="o.id" :label="o.name" :value="o.id" />
          </el-select>
          <el-button text type="primary" :disabled="disabled" @click="emit('create', field.kind)">新建{{ field.kind === 'video' ? '视频' : '图片' }}合集</el-button>
        </template>
        <el-input v-else v-model="model[field.key]" :type="field.kind === 'description' ? 'textarea' : 'text'" :rows="3" :maxlength="field.kind === 'text' ? 200 : undefined" :disabled="disabled" :aria-label="field.label" :placeholder="field.kind === 'text' ? '统一模式必填，文件名模式不使用此字段' : ''" />
      </template>
    </div>
  </div>
</template>

<style scoped>
.workspace-config{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:14px}.config-field{min-width:0}.config-field__label{display:flex;align-items:center;justify-content:space-between;gap:6px;min-height:32px;font-size:13px;font-weight:600}.config-field small{display:block;color:var(--el-text-color-secondary);padding:8px 0}.config-field :deep(.el-select){width:100%}.config-field :deep(.el-checkbox__label){font-size:11px;font-weight:400}
@media(max-width:1100px){.workspace-config{grid-template-columns:1fr}}
</style>
