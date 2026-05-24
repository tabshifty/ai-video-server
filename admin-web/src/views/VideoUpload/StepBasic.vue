<script setup>
import SectionCard from '../../components/base/SectionCard.vue'

defineProps({
  form: {
    type: Object,
    required: true
  },
  isAvType: {
    type: Boolean,
    default: false
  },
  avSiteCategoryOptions: {
    type: Array,
    default: () => []
  }
})
</script>

<template>
  <SectionCard>
    <template #title>基础信息</template>
    <el-form label-width="100px">
      <el-form-item label="类型">
        <el-select v-model="form.type" style="width: 220px">
          <el-option label="短视频" value="short" />
          <el-option label="电影" value="movie" />
          <el-option label="剧集分集" value="episode" />
          <el-option label="AV" value="av" />
        </el-select>
      </el-form-item>
      <el-form-item v-if="isAvType" label="AV 地区分类">
        <el-radio-group v-model="form.siteCategory">
          <el-radio-button v-for="item in avSiteCategoryOptions" :key="item.value" :label="item.value">
            {{ item.label }}
          </el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="标题"><el-input v-model="form.title" /></el-form-item>
      <el-form-item label="描述"><el-input v-model="form.description" type="textarea" rows="3" /></el-form-item>
    </el-form>
  </SectionCard>
</template>
