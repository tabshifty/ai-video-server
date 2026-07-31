<script setup>
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { FolderOpened, Key, Link, Picture, Search, TopRight } from '@element-plus/icons-vue'
import Layout from '../components/Layout.vue'

const router = useRouter()

const ed2kToolHref = computed(() => router.resolve('/toolbox/ed2k').href)
const archiveImportHref = computed(() => router.resolve('/toolbox/archive-import').href)
const imageWorkbenchHref = computed(() => router.resolve('/toolbox/image-workbench').href)
const orphanFilesHref = computed(() => router.resolve('/toolbox/orphan-files').href)
const passwordVaultHref = computed(() => router.resolve('/toolbox/password-vault').href)
</script>

<template>
  <Layout>
    <div class="page-shell toolbox-page" data-density="form">
      <p class="page-context-note">工具会在独立标签页打开，并保持当前管理端上下文。</p>

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
          <a class="tool-menu-item" :href="archiveImportHref" target="_blank" rel="noopener noreferrer">
            <span class="tool-menu-item__icon">
              <el-icon><FolderOpened /></el-icon>
            </span>
            <span class="tool-menu-item__content">
              <strong>压缩包导入</strong>
              <span>管理员上传 zip、rar、7z 压缩包后解包审核并导入媒体。</span>
            </span>
            <span class="tool-menu-item__meta">
              <el-icon><TopRight /></el-icon>
              <span>新标签页打开</span>
            </span>
          </a>
          <a class="tool-menu-item" :href="imageWorkbenchHref" target="_blank" rel="noopener noreferrer">
            <span class="tool-menu-item__icon">
              <el-icon><Picture /></el-icon>
            </span>
            <span class="tool-menu-item__content">
              <strong>图像生成工作台</strong>
              <span>基于提示词和参考图生成、编辑、保存创作结果。</span>
            </span>
            <span class="tool-menu-item__meta">
              <el-icon><TopRight /></el-icon>
              <span>新标签页打开</span>
            </span>
          </a>
          <a class="tool-menu-item" :href="orphanFilesHref" target="_blank" rel="noopener noreferrer">
            <span class="tool-menu-item__icon">
              <el-icon><Search /></el-icon>
            </span>
            <span class="tool-menu-item__content">
              <strong>孤儿文件扫描</strong>
              <span>盘点存储目录中的未引用持久文件，并按扫描结果统一处理。</span>
            </span>
            <span class="tool-menu-item__meta">
              <el-icon><TopRight /></el-icon>
              <span>新标签页打开</span>
            </span>
          </a>
          <a class="tool-menu-item" :href="passwordVaultHref" target="_blank" rel="noopener noreferrer">
            <span class="tool-menu-item__icon">
              <el-icon><Key /></el-icon>
            </span>
            <span class="tool-menu-item__content">
              <strong>密码管理</strong>
              <span>记录外部服务、网站和设备的账号密码。</span>
            </span>
            <span class="tool-menu-item__meta">
              <el-icon><TopRight /></el-icon>
              <span>新标签页打开</span>
            </span>
          </a>
      </div>
    </div>
  </Layout>
</template>

<style scoped>
.toolbox-page {
  display: grid;
  min-width: 0;
  gap: var(--space-4);
  overflow-x: clip;
}

.tool-menu-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: var(--space-3);
}

.tool-menu-item {
  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  align-items: start;
  gap: var(--space-3);
  min-width: 0;
  min-height: 44px;
  padding: var(--space-4);
  border: 1px solid var(--line-soft);
  border-radius: 8px;
  box-shadow: none;
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
  overflow-wrap: anywhere;
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

@media (max-width: 80rem) {
  .tool-menu-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 64rem) {
  .tool-menu-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 48rem) {
  .tool-menu-grid {
    grid-template-columns: 1fr;
  }
}
</style>
