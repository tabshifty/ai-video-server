<script setup>
import { useEmptyCollectionCleanup } from '../views/emptyCollectionCleanup'

const props = defineProps({
  kind: { type: String, required: true, validator: (value) => ['video', 'image'].includes(value) }
})
const emit = defineEmits(['completed'])
const { busy, run } = useEmptyCollectionCleanup(props.kind, () => emit('completed'))
</script>

<template>
  <el-tooltip content="清理全部无成员且无业务引用的合集，包含停用合集，不受筛选和分页限制" placement="bottom">
    <el-button type="danger" plain :loading="busy" :disabled="busy" @click="run">清理空合集</el-button>
  </el-tooltip>
</template>
