<script setup>
import { computed, nextTick, onBeforeUnmount, onMounted, ref, shallowRef, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import * as echarts from 'echarts'
import {
  ArrowLeft,
  Back,
  Close,
  InfoFilled,
  Minus,
  Plus,
  RefreshRight,
  Search
} from '@element-plus/icons-vue'
import { createChinaMapDataLoader, loadChinaMapBootstrapData } from './chinaMap.data'
import {
  ROOT_REGION_CODE,
  buildRegionIndex,
  buildRegionQuery,
  canGoBackFromView,
  findRegionView,
  getParentRegionCode,
  searchRegions
} from './chinaMap.helpers'

const route = useRoute()
const router = useRouter()
const chartElement = ref(null)
const searchInput = ref(null)
const catalog = shallowRef(null)
const regionIndex = shallowRef(null)
const manifest = shallowRef(null)
const displayedView = shallowRef(null)
const pendingRegionCode = ref(null)
const selectedRegionCode = ref(null)
const loading = ref(true)
const errorMessage = ref('')
const informationOpen = ref(false)
const informationCloseButton = ref(null)
const searchText = ref('')
const searchOpen = ref(false)
const activeSearchIndex = ref(0)
const reducedMotion = ref(false)
const renderedMapName = ref('')
const insetLayer = shallowRef(null)

const dataBaseUrl = `${import.meta.env.BASE_URL || '/'}china-map/`
const dataLoader = createChinaMapDataLoader({ baseUrl: dataBaseUrl })
let chart = null
let resizeObserver = null
let motionQuery = null
let navigationSequence = 0
let informationReturnFocus = null

const requestedRegionCode = computed(() => String(route.query.region_code || ROOT_REGION_CODE))
const searchResults = computed(() => searchRegions(regionIndex.value, searchText.value, 12))
const selectedRegion = computed(() =>
  selectedRegionCode.value ? regionIndex.value?.byCode.get(selectedRegionCode.value) || null : null
)
const canGoBack = computed(() => canGoBackFromView(displayedView.value))
const currentAriaLabel = computed(() => {
  const label = displayedView.value?.region?.full_path || '中国'
  return `中国行政区划地图，当前位置：${label}`
})
const southChinaSeaPoints = computed(() => {
  const minimumLongitude = 110
  const maximumLongitude = 118
  const minimumLatitude = 3
  const maximumLatitude = 18
  return (insetLayer.value?.features || []).map((feature) => {
    const [longitude, latitude] = feature.geometry.coordinates
    const horizontalRatio = (longitude - minimumLongitude) / (maximumLongitude - minimumLongitude)
    const verticalRatio = (maximumLatitude - latitude) / (maximumLatitude - minimumLatitude)
    return {
      region_code: feature.properties.region_code,
      name: feature.properties.name,
      align_right: horizontalRatio > 0.58,
      style: {
        left: `${10 + Math.min(1, Math.max(0, horizontalRatio)) * 80}%`,
        top: `${18 + Math.min(1, Math.max(0, verticalRatio)) * 75}%`
      }
    }
  })
})

function readThemeColor(name) {
  const value = getComputedStyle(document.documentElement).getPropertyValue(name).trim()
  return value
}

function mapPalette() {
  return {
    water: readThemeColor('--map-water'),
    land: readThemeColor('--map-land'),
    border: readThemeColor('--line-strong'),
    text: readThemeColor('--text-secondary'),
    textStrong: readThemeColor('--text-primary'),
    hover: readThemeColor('--map-land-hover'),
    focus: readThemeColor('--map-selection'),
    focusSoft: readThemeColor('--map-selection-soft')
  }
}

function mapOption(mapName, view, layer) {
  const palette = mapPalette()
  const data = layer.features.map((feature) => ({
    name: feature.properties.name,
    region_code: feature.properties.region_code,
    selected: feature.properties.region_code === view.selected_region_code
  }))

  const option = {
    backgroundColor: palette.water,
    animation: !reducedMotion.value,
    animationDurationUpdate: reducedMotion.value ? 0 : 240,
    animationEasingUpdate: 'cubicOut',
    tooltip: {
      trigger: 'item',
      confine: true,
      backgroundColor: readThemeColor('--bg-surface'),
      borderColor: readThemeColor('--line-soft'),
      textStyle: { color: palette.textStrong, fontSize: 13 },
      formatter(params) {
        const feature = layer.features.find((entry) => entry.properties.region_code === params.data?.region_code)
        return feature?.properties?.full_path || params.name || ''
      }
    },
    series: [
      {
        id: 'administrative-map',
        type: 'map',
        map: mapName,
        nameProperty: 'name',
        roam: true,
        scaleLimit: { min: 0.8, max: 12 },
        selectedMode: 'single',
        data,
        label: {
          show: true,
          color: palette.text,
          fontSize: 11,
          formatter: '{b}'
        },
        labelLayout: { hideOverlap: true },
        itemStyle: {
          areaColor: palette.land,
          borderColor: palette.border,
          borderWidth: 0.8
        },
        emphasis: {
          itemStyle: { areaColor: palette.hover, borderColor: palette.text, borderWidth: 1.2 },
          label: { show: true, color: palette.textStrong, fontWeight: 600 }
        },
        select: {
          itemStyle: { areaColor: palette.focusSoft, borderColor: palette.focus, borderWidth: 1.8 },
          label: { show: true, color: palette.textStrong, fontWeight: 600 }
        }
      }
    ]
  }

  return option
}

function initializeChart() {
  if (!chartElement.value || chart) return
  chart = echarts.init(chartElement.value, null, { renderer: 'canvas' })
  chart.on('click', handleMapClick)
  resizeObserver = new ResizeObserver(() => chart?.resize())
  resizeObserver.observe(chartElement.value)
}

function renderLayer(layer, view) {
  initializeChart()
  const mapName = `china-administrative-${view.layer_parent_code}`
  echarts.registerMap(mapName, layer)
  renderedMapName.value = mapName
  selectedRegionCode.value = view.selected_region_code
  chart.setOption(mapOption(mapName, view, layer), true)
}

async function loadRegion(regionCode) {
  if (!regionIndex.value) return
  const sequence = ++navigationSequence
  const view = findRegionView(regionIndex.value, regionCode)
  if (!view) return

  loading.value = true
  errorMessage.value = ''
  pendingRegionCode.value = view.region_code
  try {
    const layer = await dataLoader.loadLayer(view.layer_parent_code)
    if (sequence !== navigationSequence) return
    renderLayer(layer, view)
    displayedView.value = view
  } catch (error) {
    if (sequence !== navigationSequence) return
    errorMessage.value = error?.message || '地图数据加载失败，请重试'
  } finally {
    if (sequence === navigationSequence) {
      loading.value = false
      pendingRegionCode.value = null
    }
  }
}

async function replaceRegionQuery(regionCode) {
  await router.replace({ query: buildRegionQuery(route.query, regionCode) })
}

async function navigateToRegion(regionCode) {
  const code = regionIndex.value?.byCode.has(regionCode) ? regionCode : ROOT_REGION_CODE
  searchOpen.value = false
  searchText.value = ''
  await router.push({ query: buildRegionQuery(route.query, code) })
}

function handleMapClick(params) {
  const code = params?.data?.region_code
  if (!code) return
  navigateToRegion(code)
}

function returnToToolbox() {
  router.push('/toolbox')
}

function goBackOneLevel() {
  if (!displayedView.value || !regionIndex.value) return
  const parentCode = getParentRegionCode(regionIndex.value, displayedView.value.region_code)
  if (parentCode) navigateToRegion(parentCode)
}

function retryPendingLayer() {
  if (!regionIndex.value) {
    bootstrap()
    return
  }
  loadRegion(pendingRegionCode.value || requestedRegionCode.value)
}

function zoomMap(factor) {
  if (!chart) return
  const width = chart.getWidth()
  const height = chart.getHeight()
  chart.dispatchAction({
    type: 'geoRoam',
    componentType: 'series',
    seriesId: 'administrative-map',
    zoom: factor,
    originX: width / 2,
    originY: height / 2
  })
}

function resetMap() {
  if (!chart || !displayedView.value || !renderedMapName.value) return
  chart.setOption({ series: [{ id: 'administrative-map', center: null, zoom: 1 }] })
}

function openSearch() {
  searchOpen.value = true
  activeSearchIndex.value = 0
}

function closeSearch() {
  searchOpen.value = false
  activeSearchIndex.value = 0
}

function handleSearchKeydown(event) {
  if (event.key === 'Escape') {
    closeSearch()
    searchInput.value?.blur()
    return
  }
  if (!searchResults.value.length) return

  if (event.key === 'ArrowDown') {
    event.preventDefault()
    searchOpen.value = true
    activeSearchIndex.value = (activeSearchIndex.value + 1) % searchResults.value.length
  } else if (event.key === 'ArrowUp') {
    event.preventDefault()
    searchOpen.value = true
    activeSearchIndex.value =
      (activeSearchIndex.value - 1 + searchResults.value.length) % searchResults.value.length
  } else if (event.key === 'Enter') {
    event.preventDefault()
    const target = searchResults.value[activeSearchIndex.value]
    if (target) navigateToRegion(target.region_code)
  }
}

function openInformation(event) {
  informationReturnFocus = event?.currentTarget || document.activeElement
  informationOpen.value = true
  nextTick(() => informationCloseButton.value?.focus())
}

function closeInformation() {
  informationOpen.value = false
  nextTick(() => {
    informationReturnFocus?.focus?.()
    informationReturnFocus = null
  })
}

function handleInformationKeydown(event) {
  if (event.key === 'Escape') {
    event.preventDefault()
    closeInformation()
    return
  }
  if (event.key === 'Tab') {
    event.preventDefault()
    informationCloseButton.value?.focus()
  }
}

function syncMotionPreference(event) {
  reducedMotion.value = Boolean(event?.matches ?? motionQuery?.matches)
}

async function bootstrap() {
  loading.value = true
  try {
    const { catalog: nextCatalog, manifest: nextManifest, inset } =
      await loadChinaMapBootstrapData(dataLoader)
    catalog.value = nextCatalog
    manifest.value = nextManifest
    regionIndex.value = buildRegionIndex(nextCatalog.regions)
    insetLayer.value = inset
    await nextTick()
    initializeChart()

    if (!regionIndex.value.byCode.has(requestedRegionCode.value)) {
      await replaceRegionQuery(ROOT_REGION_CODE)
    } else {
      await loadRegion(requestedRegionCode.value)
    }
  } catch (error) {
    errorMessage.value = error?.message || '地图目录加载失败，请重试'
    loading.value = false
  }
}

watch(requestedRegionCode, (regionCode) => {
  if (regionIndex.value) loadRegion(regionCode)
})

watch(searchText, () => {
  activeSearchIndex.value = 0
  searchOpen.value = true
})

onMounted(() => {
  motionQuery = window.matchMedia('(prefers-reduced-motion: reduce)')
  syncMotionPreference(motionQuery)
  motionQuery.addEventListener?.('change', syncMotionPreference)
  bootstrap()
})

onBeforeUnmount(() => {
  navigationSequence += 1
  motionQuery?.removeEventListener?.('change', syncMotionPreference)
  resizeObserver?.disconnect()
  chart?.off('click', handleMapClick)
  chart?.dispose()
  chart = null
})
</script>

<template>
  <main class="china-map-workspace" :aria-busy="loading ? 'true' : 'false'">
    <div ref="chartElement" class="china-map-canvas" role="img" :aria-label="currentAriaLabel"></div>

    <nav class="map-floating map-navigation" aria-label="地图导航">
      <button
        class="map-command map-command--text"
        type="button"
        title="返回工具箱"
        aria-label="返回工具箱"
        @click="returnToToolbox"
      >
        <el-icon aria-hidden="true"><Back /></el-icon>
        <span>返回工具箱</span>
      </button>
      <span class="map-navigation__divider" aria-hidden="true"></span>
      <button
        class="map-icon-button"
        type="button"
        title="返回上一级"
        aria-label="返回上一级"
        :disabled="!canGoBack"
        @click="goBackOneLevel"
      >
        <el-icon><ArrowLeft /></el-icon>
      </button>
      <ol class="map-breadcrumbs" aria-label="当前位置">
        <li v-for="(entry, index) in displayedView?.breadcrumbs || []" :key="entry.region_code">
          <span v-if="index" class="map-breadcrumbs__separator" aria-hidden="true">/</span>
          <button
            type="button"
            :aria-current="index === displayedView.breadcrumbs.length - 1 ? 'location' : undefined"
            @click="navigateToRegion(entry.region_code)"
          >
            {{ entry.name }}
          </button>
        </li>
        <li v-if="!displayedView"><span class="map-breadcrumbs__placeholder">中国行政区划地图</span></li>
      </ol>
    </nav>

    <div class="map-floating map-search" @focusin="openSearch" @focusout="closeSearch">
      <el-icon class="map-search__icon" aria-hidden="true"><Search /></el-icon>
      <label class="sr-only" for="china-map-search">搜索行政区</label>
      <input
        id="china-map-search"
        ref="searchInput"
        v-model="searchText"
        type="search"
        autocomplete="off"
        placeholder="搜索省、市、县级行政区"
        aria-autocomplete="list"
        aria-controls="china-map-search-results"
        :aria-expanded="searchOpen && Boolean(searchText)"
        :aria-activedescendant="
          searchOpen && searchResults[activeSearchIndex]
            ? `china-map-result-${searchResults[activeSearchIndex].region_code}`
            : undefined
        "
        @keydown="handleSearchKeydown"
      />
      <button
        v-if="searchText"
        class="map-search__clear"
        type="button"
        title="清空搜索"
        aria-label="清空搜索"
        @mousedown.prevent
        @click="searchText = ''"
      >
        <el-icon><Close /></el-icon>
      </button>
      <div
        v-if="searchOpen && searchText"
        id="china-map-search-results"
        class="map-search-results"
        role="listbox"
      >
        <button
          v-for="(result, index) in searchResults"
          :id="`china-map-result-${result.region_code}`"
          :key="result.region_code"
          class="map-search-result"
          :class="{ 'is-active': index === activeSearchIndex }"
          type="button"
          role="option"
          :aria-selected="index === activeSearchIndex"
          @mousedown.prevent="navigateToRegion(result.region_code)"
          @mouseenter="activeSearchIndex = index"
        >
          <strong>{{ result.name }}</strong>
          <span>{{ result.full_path }}</span>
        </button>
        <p v-if="!searchResults.length" class="map-search-empty">未找到匹配的行政区</p>
      </div>
    </div>

    <div class="map-floating map-zoom-controls" aria-label="地图缩放">
      <button class="map-icon-button" type="button" title="放大" aria-label="放大" @click="zoomMap(1.25)">
        <el-icon><Plus /></el-icon>
      </button>
      <button class="map-icon-button" type="button" title="缩小" aria-label="缩小" @click="zoomMap(0.8)">
        <el-icon><Minus /></el-icon>
      </button>
      <span class="map-zoom-controls__divider" aria-hidden="true"></span>
      <button class="map-icon-button" type="button" title="复位地图" aria-label="复位地图" @click="resetMap">
        <el-icon><RefreshRight /></el-icon>
      </button>
    </div>

    <section v-if="selectedRegion" class="map-floating map-selection" aria-live="polite">
      <div>
        <span>县级行政区</span>
        <strong>{{ selectedRegion.name }}</strong>
      </div>
      <p>{{ selectedRegion.full_path }}</p>
      <dl>
        <dt>区划标识</dt>
        <dd>{{ selectedRegion.region_code }}</dd>
      </dl>
    </section>

    <aside
      v-if="displayedView?.layer_parent_code === ROOT_REGION_CODE && southChinaSeaPoints.length"
      class="south-china-sea-inset"
      aria-label="南海诸岛地理附图"
    >
      <strong>南海诸岛</strong>
      <span
        v-for="point in southChinaSeaPoints"
        :key="point.region_code"
        class="south-china-sea-point"
        :class="{ 'is-right-aligned': point.align_right }"
        :style="point.style"
      >
        <i aria-hidden="true"></i>
        <span>{{ point.name }}</span>
      </span>
    </aside>

    <div class="map-bottom-actions">
      <button
        class="map-floating map-icon-button map-info-button"
        type="button"
        title="地图信息"
        aria-label="地图信息"
        @click="openInformation"
      >
        <el-icon><InfoFilled /></el-icon>
      </button>
    </div>

    <div v-if="loading" class="map-floating map-status" role="status">
      <span class="map-status__spinner" aria-hidden="true"></span>
      <span>{{ displayedView ? '正在加载下一级边界' : '正在加载中国行政区划地图' }}</span>
    </div>

    <section v-if="errorMessage" class="map-floating map-error" role="alert">
      <div>
        <strong>地图数据加载失败</strong>
        <p>{{ errorMessage }}</p>
      </div>
      <div class="map-error__actions">
        <button type="button" @click="retryPendingLayer">重试</button>
        <button v-if="canGoBack" type="button" @click="goBackOneLevel">返回上一级</button>
        <button v-else type="button" @click="returnToToolbox">返回工具箱</button>
      </div>
    </section>

    <div v-if="informationOpen" class="map-information-backdrop" @click.self="closeInformation">
      <section
        class="map-information"
        role="dialog"
        aria-modal="true"
        aria-labelledby="map-information-title"
        @keydown="handleInformationKeydown"
      >
        <header>
          <div>
            <span>地图信息</span>
            <h1 id="map-information-title">{{ manifest?.dataset_name || '中国行政区划地图' }}</h1>
          </div>
          <button
            ref="informationCloseButton"
            class="map-icon-button"
            type="button"
            title="关闭"
            aria-label="关闭"
            @click="closeInformation"
          >
            <el-icon><Close /></el-icon>
          </button>
        </header>
        <dl>
          <div>
            <dt>数据版本</dt>
            <dd>{{ manifest?.version || '未知' }}</dd>
          </div>
          <div>
            <dt>边界来源</dt>
            <dd>{{ manifest?.geometry_source || 'OpenStreetMap' }}</dd>
          </div>
          <div>
            <dt>行政目录</dt>
            <dd>{{ manifest?.directory_source || '公开行政区划目录' }}</dd>
          </div>
          <div>
            <dt>数据许可</dt>
            <dd>{{ manifest?.license || 'ODbL 1.0' }}</dd>
          </div>
        </dl>
        <p>{{ manifest?.attribution || '© OpenStreetMap contributors' }}</p>
      </section>
    </div>
  </main>
</template>

<style scoped>
.china-map-workspace {
  position: relative;
  width: 100vw;
  height: 100vh;
  height: 100dvh;
  min-width: 0;
  overflow: hidden;
  color: var(--text-primary);
  background: var(--map-water);
}

.china-map-canvas {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  outline: none;
}

.map-floating {
  border: 1px solid color-mix(in srgb, var(--line-strong) 82%, transparent);
  border-radius: var(--radius-md);
  box-shadow: var(--shadow-md);
  background: color-mix(in srgb, var(--bg-surface) 94%, transparent);
  backdrop-filter: blur(10px);
}

.map-navigation {
  position: absolute;
  top: var(--space-4);
  left: var(--space-4);
  z-index: 4;
  display: flex;
  align-items: center;
  min-width: 0;
  max-width: min(48rem, calc(100vw - 31rem));
  height: 44px;
  padding: 4px 6px;
}

.map-navigation__divider,
.map-zoom-controls__divider {
  width: 1px;
  height: 22px;
  margin: 0 4px;
  background: var(--line-soft);
}

.map-command,
.map-icon-button,
.map-breadcrumbs button,
.map-search__clear,
.map-search-result,
.map-error button {
  border: 0;
  color: inherit;
  background: transparent;
  cursor: pointer;
}

.map-command:focus-visible,
.map-icon-button:focus-visible,
.map-breadcrumbs button:focus-visible,
.map-search__clear:focus-visible,
.map-search-result:focus-visible,
.map-error button:focus-visible,
.map-search input:focus-visible {
  outline: 3px solid var(--line-focus);
  outline-offset: 1px;
}

.map-command--text {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  flex: 0 0 auto;
  height: 34px;
  padding: 0 8px;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: var(--text-small);
  transition: color var(--motion-duration-base) var(--motion-easing-standard),
    background-color var(--motion-duration-base) var(--motion-easing-standard);
}

.map-command--text:hover {
  color: var(--text-primary);
  background: var(--bg-surface-muted);
}

.map-icon-button {
  display: inline-grid;
  width: 34px;
  height: 34px;
  flex: 0 0 34px;
  place-items: center;
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  font-size: 17px;
  transition: color var(--motion-duration-base) var(--motion-easing-standard),
    background-color var(--motion-duration-base) var(--motion-easing-standard);
}

.map-icon-button:hover:not(:disabled) {
  color: var(--text-primary);
  background: var(--bg-surface-muted);
}

.map-icon-button:disabled {
  cursor: not-allowed;
  opacity: 0.42;
}

.map-breadcrumbs {
  display: flex;
  align-items: center;
  min-width: 0;
  margin: 0;
  padding: 0 4px;
  overflow: hidden;
  list-style: none;
}

.map-breadcrumbs li {
  display: inline-flex;
  align-items: center;
  min-width: 0;
}

.map-breadcrumbs button,
.map-breadcrumbs__placeholder {
  max-width: 11rem;
  padding: 6px;
  overflow: hidden;
  border-radius: var(--radius-sm);
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--text-small);
}

.map-breadcrumbs button:hover {
  background: var(--bg-surface-muted);
}

.map-breadcrumbs button[aria-current='location'] {
  color: var(--text-primary);
  font-weight: 600;
}

.map-breadcrumbs__separator {
  color: var(--text-muted);
}

.map-search {
  position: absolute;
  top: var(--space-4);
  right: var(--space-4);
  z-index: 6;
  display: grid;
  grid-template-columns: 36px minmax(0, 1fr) 34px;
  align-items: center;
  width: min(28rem, calc(100vw - 2rem));
  min-height: 44px;
}

.map-search__icon {
  justify-self: center;
  color: var(--text-muted);
  font-size: 17px;
}

.map-search input {
  width: 100%;
  min-width: 0;
  height: 42px;
  padding: 0;
  border: 0;
  outline: 0;
  color: var(--text-primary);
  background: transparent;
  font-size: var(--text-body);
}

.map-search input::placeholder {
  color: var(--text-muted);
}

.map-search__clear {
  display: grid;
  width: 30px;
  height: 30px;
  place-items: center;
  border-radius: var(--radius-sm);
  color: var(--text-muted);
}

.map-search-results {
  position: absolute;
  top: calc(100% + var(--space-2));
  right: 0;
  left: 0;
  max-height: min(25rem, calc(100dvh - 6rem));
  padding: var(--space-2);
  overflow-y: auto;
  border: 1px solid var(--line-strong);
  border-radius: var(--radius-md);
  background: var(--bg-surface);
}

.map-search-result {
  display: grid;
  width: 100%;
  gap: 2px;
  padding: 9px 10px;
  border-radius: var(--radius-sm);
  text-align: left;
}

.map-search-result:hover,
.map-search-result.is-active {
  background: var(--bg-surface-muted);
}

.map-search-result strong {
  font-size: var(--text-body);
  font-weight: 600;
}

.map-search-result span {
  overflow: hidden;
  color: var(--text-muted);
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: var(--text-caption);
}

.map-search-empty {
  margin: 0;
  padding: var(--space-5) var(--space-3);
  color: var(--text-muted);
  text-align: center;
  font-size: var(--text-small);
}

.map-zoom-controls {
  position: absolute;
  top: 50%;
  right: var(--space-4);
  z-index: 4;
  display: grid;
  justify-items: center;
  padding: 5px;
  transform: translateY(-50%);
}

.map-zoom-controls__divider {
  width: 22px;
  height: 1px;
  margin: 4px 0;
}

.map-selection {
  position: absolute;
  bottom: var(--space-4);
  left: var(--space-4);
  z-index: 4;
  width: min(23rem, calc(100vw - 2rem));
  padding: var(--space-4);
}

.map-selection > div:first-child {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: var(--space-3);
}

.map-selection span,
.map-information header span {
  color: var(--map-selection-text);
  font-size: var(--text-caption);
  font-weight: 600;
}

.map-selection strong {
  font-size: var(--text-h2);
}

.map-selection p {
  margin: var(--space-2) 0 var(--space-3);
  color: var(--text-secondary);
  font-size: var(--text-small);
  line-height: var(--leading-small);
}

.map-selection dl {
  display: flex;
  justify-content: space-between;
  gap: var(--space-3);
  margin: 0;
  padding-top: var(--space-2);
  border-top: 1px solid var(--line-soft);
  font-size: var(--text-caption);
}

.map-selection dt {
  color: var(--text-muted);
}

.map-selection dd {
  margin: 0;
  overflow-wrap: anywhere;
  font-family: var(--font-mono);
}

.map-bottom-actions {
  position: absolute;
  right: var(--space-4);
  bottom: var(--space-4);
  z-index: 4;
  display: flex;
  align-items: end;
  gap: var(--space-2);
}

.south-china-sea-inset {
  position: absolute;
  right: var(--space-4);
  bottom: 4.5rem;
  z-index: 3;
  width: 148px;
  height: 190px;
  overflow: hidden;
  border: 1px solid color-mix(in srgb, var(--line-strong) 82%, transparent);
  border-radius: var(--radius-sm);
  color: var(--text-secondary);
  background: color-mix(in srgb, var(--map-water) 92%, var(--bg-surface));
  pointer-events: none;
}

.south-china-sea-inset > strong {
  position: absolute;
  top: 7px;
  left: 9px;
  font-size: var(--text-caption);
  font-weight: 600;
}

.south-china-sea-point {
  position: absolute;
  white-space: nowrap;
  font-size: 10px;
  line-height: 1;
}

.south-china-sea-point i {
  position: absolute;
  top: -3px;
  left: -3px;
  width: 7px;
  height: 7px;
  border: 1px solid var(--map-selection-strong);
  border-radius: var(--radius-md);
  background: var(--map-selection-soft);
}

.south-china-sea-point span {
  position: absolute;
  top: -5px;
  left: 7px;
}

.south-china-sea-point.is-right-aligned span {
  right: 7px;
  left: auto;
}

.map-info-button {
  position: static;
  width: 44px;
  height: 44px;
  flex-basis: 44px;
}

.map-status {
  position: absolute;
  bottom: var(--space-4);
  left: 50%;
  z-index: 7;
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  min-height: 38px;
  padding: 0 var(--space-3);
  color: var(--text-secondary);
  transform: translateX(-50%);
  font-size: var(--text-small);
}

.map-status__spinner {
  width: 14px;
  height: 14px;
  border: 2px solid var(--line-strong);
  border-top-color: var(--map-selection);
  border-radius: var(--radius-md);
  animation: map-spin 800ms linear infinite;
}

.map-error {
  position: absolute;
  bottom: var(--space-4);
  left: 50%;
  z-index: 8;
  display: flex;
  align-items: center;
  width: min(38rem, calc(100vw - 2rem));
  gap: var(--space-4);
  padding: var(--space-3) var(--space-4);
  transform: translateX(-50%);
}

.map-error > div:first-child {
  min-width: 0;
  flex: 1;
}

.map-error strong {
  font-size: var(--text-small);
}

.map-error p {
  margin: 2px 0 0;
  overflow-wrap: anywhere;
  color: var(--text-muted);
  font-size: var(--text-caption);
}

.map-error__actions {
  display: flex;
  flex: 0 0 auto;
  gap: var(--space-2);
}

.map-error button {
  height: 32px;
  padding: 0 var(--space-3);
  border: 1px solid var(--line-strong);
  border-radius: var(--radius-sm);
  background: var(--bg-surface);
  font-size: var(--text-small);
}

.map-error button:first-child {
  border-color: var(--map-selection);
  color: var(--map-selection-strong);
  background: var(--map-selection-surface);
}

.map-information-backdrop {
  position: fixed;
  inset: 0;
  z-index: 20;
  display: grid;
  place-items: center;
  padding: var(--space-4);
  background: var(--bg-overlay);
}

.map-information {
  width: min(31rem, 100%);
  max-height: calc(100dvh - 2rem);
  padding: var(--space-5);
  overflow-y: auto;
  border: 1px solid var(--line-strong);
  border-radius: var(--radius-md);
  background: var(--bg-surface);
}

.map-information header {
  display: flex;
  align-items: start;
  justify-content: space-between;
  gap: var(--space-4);
}

.map-information h1 {
  margin: 3px 0 0;
  font-size: var(--text-h1);
  line-height: var(--leading-h1);
}

.map-information dl {
  display: grid;
  gap: var(--space-3);
  margin: var(--space-5) 0;
}

.map-information dl div {
  display: grid;
  grid-template-columns: 6rem minmax(0, 1fr);
  gap: var(--space-3);
  padding-bottom: var(--space-3);
  border-bottom: 1px solid var(--line-soft);
}

.map-information dt {
  color: var(--text-muted);
  font-size: var(--text-small);
}

.map-information dd {
  margin: 0;
  overflow-wrap: anywhere;
  color: var(--text-secondary);
  font-size: var(--text-small);
}

.map-information > p {
  margin: 0;
  color: var(--text-muted);
  font-size: var(--text-caption);
  line-height: var(--leading-caption);
}

.sr-only {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

@keyframes map-spin {
  to {
    transform: rotate(360deg);
  }
}

@media (max-width: 72rem) {
  .map-navigation {
    max-width: calc(100vw - 26rem);
  }

  .map-search {
    width: 23rem;
  }
}

@media (max-width: 56rem) {
  .map-navigation {
    right: var(--space-4);
    max-width: none;
  }

  .map-search {
    top: calc(var(--space-4) + 52px);
    left: var(--space-4);
    width: min(28rem, calc(100vw - 2rem));
  }

  .map-command--text span,
  .map-breadcrumbs li:not(:last-child) {
    display: none;
  }
}

@media (max-width: 40rem) {
  .map-navigation,
  .map-search {
    left: var(--space-2);
    right: var(--space-2);
  }

  .map-navigation {
    top: var(--space-2);
  }

  .map-search {
    top: calc(var(--space-2) + 52px);
    width: auto;
  }

  .map-zoom-controls {
    right: var(--space-2);
  }

  .map-selection {
    right: var(--space-2);
    bottom: var(--space-2);
    left: var(--space-2);
    width: auto;
  }

  .map-bottom-actions {
    right: var(--space-2);
    bottom: var(--space-2);
  }

  .south-china-sea-inset {
    right: var(--space-2);
    bottom: 3.75rem;
    width: 116px;
    height: 154px;
  }

  .map-error {
    bottom: var(--space-2);
    align-items: stretch;
    flex-direction: column;
  }
}

@media (prefers-reduced-motion: reduce) {
  .map-command--text,
  .map-icon-button {
    transition-duration: 0.01ms;
  }

  .map-status__spinner {
    animation-duration: 1.8s;
  }
}
</style>
