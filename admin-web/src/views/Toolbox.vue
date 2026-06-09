<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Link, TopRight } from '@element-plus/icons-vue'
import Layout from '../components/Layout.vue'
import PageHeader from '../components/base/PageHeader.vue'
import SectionCard from '../components/base/SectionCard.vue'

const router = useRouter()

const ed2kToolHref = computed(() => router.resolve('/toolbox/ed2k').href)
</script>

<template>
  <Layout>
    <div class="page-shell toolbox-page">
      <PageHeader title="工具箱" subtitle="打开独立标签页使用后台小工具" />

      <SectionCard>
        <template #title>工具菜单</template>
        <template #description>点击工具入口后会在新的浏览器标签页打开具体功能。</template>

        <div class="tool-menu-grid" aria-label="工具箱菜单">
          <a class="tool-menu-item" :href="ed2kToolHref" target="_blank" rel="noopener noreferrer">
            <span class="tool-menu-item__icon">
              <el-icon><Link /></el-icon>
            </span>
            <span class="tool-menu-item__content">
              <strong>ED2K 链接生成器</strong>
              <span>把多行 ED2K 文本转换为可点击链接。</span>
            </span>
            <span class="tool-menu-item__meta">
              <el-icon><TopRight /></el-icon>
              <span>新标签页打开</span>
            </span>
          </a>
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

.tool-menu-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(18rem, 1fr));
  gap: var(--space-3);
}

.tool-menu-item {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  align-items: start;
  gap: var(--space-3);
  min-width: 0;
  padding: var(--space-4);
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-lg);
  color: var(--text-primary);
  background: var(--bg-surface-muted);
  text-decoration: none;
  transition: border-color var(--motion-duration-base) var(--motion-easing-standard),
    box-shadow var(--motion-duration-base) var(--motion-easing-standard),
    background-color var(--motion-duration-base) var(--motion-easing-standard);
}

.tool-menu-item:hover,
.tool-menu-item:focus-visible {
  border-color: var(--primary);
  background: var(--bg-surface);
  box-shadow: var(--shadow-sm);
}

.tool-menu-item:focus-visible {
  outline: 3px solid var(--line-focus);
  outline-offset: 2px;
}

.tool-menu-item__icon {
  display: grid;
  width: calc(var(--space-8) + var(--space-2));
  height: calc(var(--space-8) + var(--space-2));
  place-items: center;
  border: 1px solid var(--line-soft);
  border-radius: var(--radius-md);
  color: var(--primary);
  background: var(--primary-soft);
  font-size: var(--text-h1);
}

.tool-menu-item__content {
  display: grid;
  gap: var(--space-1);
  min-width: 0;
}

.tool-menu-item__content strong {
  overflow: hidden;
  font-size: var(--text-body);
  line-height: var(--leading-body);
  font-weight: 600;
}

.tool-menu-item__content span {
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.tool-menu-item__meta {
  grid-column: 2;
  display: inline-flex;
  align-items: center;
  gap: var(--space-1);
  color: var(--text-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

@media (max-width: 48rem) {
  .tool-menu-grid {
    grid-template-columns: 1fr;
  }
}
</style>
