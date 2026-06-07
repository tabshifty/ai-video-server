<script setup>
import { computed, ref } from 'vue'
import Layout from '../components/Layout.vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'
import EmptyState from '../components/base/EmptyState.vue'
import { parseEd2kLinks } from './toolbox.helpers'

const ed2kInput = ref('')
const ed2kClickedLinks = ref(new Set())

const ed2kParseResult = computed(() => parseEd2kLinks(ed2kInput.value))
const ed2kLinks = computed(() => ed2kParseResult.value.links)
const ed2kInvalidCount = computed(() => ed2kParseResult.value.invalidCount)
const ed2kHasInput = computed(() => ed2kInput.value.trim() !== '')

function isEd2kLinkClicked(link) {
  return ed2kClickedLinks.value.has(link.href)
}

function markEd2kLinkClicked(link) {
  ed2kClickedLinks.value = new Set([...ed2kClickedLinks.value, link.href])
}

function clearEd2kInput() {
  ed2kInput.value = ''
  ed2kClickedLinks.value = new Set()
}
</script>

<template>
  <Layout>
    <div class="page-shell toolbox-page">
      <PageHeader title="工具箱" subtitle="放置后台常用小工具" />

      <SectionCard>
        <template #title>ED2K 链接生成器</template>
        <template #description>把多行 ED2K 文本转换为可点击链接，点击后在本页标记状态。</template>
        <template #actions>
          <el-button :disabled="!ed2kHasInput" @click="clearEd2kInput">清空</el-button>
        </template>

        <div class="ed2k-tool">
          <el-input
            v-model="ed2kInput"
            type="textarea"
            :rows="6"
            resize="vertical"
            placeholder="每行一个 ed2k:// 链接"
            class="ed2k-tool__input"
          />
          <div class="ed2k-tool__summary">
            <span>有效链接：{{ ed2kLinks.length }}</span>
            <span v-if="ed2kInvalidCount > 0">已忽略 {{ ed2kInvalidCount }} 行非 ED2K 文本</span>
          </div>

          <EmptyState
            v-if="ed2kHasInput && ed2kLinks.length === 0"
            title="未识别到 ED2K 链接"
            description="请输入以 ed2k:// 开头的链接"
          />

          <div v-else-if="ed2kLinks.length > 0" class="ed2k-link-list" aria-label="生成的 ED2K 链接">
            <a
              v-for="link in ed2kLinks"
              :key="link.id"
              class="ed2k-link"
              :class="{ 'is-clicked': isEd2kLinkClicked(link) }"
              :href="link.href"
              @click="markEd2kLinkClicked(link)"
            >
              <span class="ed2k-link__line tabular-num">{{ link.lineNumber }}</span>
              <span class="ed2k-link__text">{{ link.label }}</span>
              <el-tag v-if="isEd2kLinkClicked(link)" size="small" type="info" effect="plain">已点击</el-tag>
              <el-tag v-else size="small" type="success" effect="plain">未点击</el-tag>
            </a>
          </div>
        </div>
      </SectionCard>
    </div>
  </Layout>
</template>

<style scoped>
.toolbox-page {
  display: grid;
  gap: var(--space-6);
}

.ed2k-tool {
  display: grid;
  gap: var(--space-3);
}

.ed2k-tool__input {
  max-width: 56rem;
}

.ed2k-tool__summary {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-3);
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.ed2k-link-list {
  display: grid;
  gap: var(--space-2);
}

.ed2k-link {
  display: grid;
  grid-template-columns: 3rem minmax(0, 1fr) auto;
  align-items: center;
  gap: var(--space-3);
  padding: var(--space-3);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  background: var(--bg-surface-muted);
  text-decoration: none;
}

.ed2k-link:hover {
  border-color: var(--primary);
  color: var(--primary);
  background: var(--bg-surface);
}

.ed2k-link.is-clicked {
  color: var(--text-muted);
  background: var(--bg-surface);
}

.ed2k-link__line {
  color: var(--text-muted);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.ed2k-link__text {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: var(--font-mono);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

@media (max-width: 48rem) {
  .ed2k-link {
    grid-template-columns: 2.5rem minmax(0, 1fr);
  }

  .ed2k-link .el-tag {
    grid-column: 2;
    justify-self: start;
  }
}
</style>
