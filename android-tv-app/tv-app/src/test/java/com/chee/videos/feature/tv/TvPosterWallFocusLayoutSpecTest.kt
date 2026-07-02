package com.chee.videos.feature.tv

import com.chee.videos.core.ui.TvFocusSafeSpec
import com.chee.videos.core.ui.TvLayoutSpec
import org.junit.Assert.assertTrue
import org.junit.Test
import java.nio.file.Path

class TvPosterWallFocusLayoutSpecTest {
    @Test
    fun `poster wall grid keeps focus safe padding and spacing`() {
        assertTrue(TvPosterWallFocusLayoutSpec.gridHorizontalPaddingDp >= TvFocusSafeSpec.posterFocusSafeSpaceDp)
        assertTrue(TvPosterWallFocusLayoutSpec.gridTopPaddingDp >= TvFocusSafeSpec.posterFocusSafeSpaceDp)
        assertTrue(TvPosterWallFocusLayoutSpec.gridBottomPaddingDp >= TvFocusSafeSpec.posterFocusSafeSpaceDp)
        assertTrue(TvPosterWallFocusLayoutSpec.gridBottomPaddingDp >= TvLayoutSpec.scrollBottomSafePaddingDp)
        assertTrue(TvPosterWallFocusLayoutSpec.gridItemSpacingDp >= TvFocusSafeSpec.posterFocusSafeSpaceDp * 2)
        // 纯 9:16 卡 + 1.08 放大 + 中性落影抬升，顶行聚焦时垂直方向既要吸收缩放溢出又要吸收落影外溢：
        // 顶部内边距必须 ≥ 焦点垂直溢出 + 落影 elevation，否则聚焦卡的落影/圆角会被顶部栏裁切。
        // 最坏列宽取 Adaptive(minSize=90) 单列上限（略低于 2×minSize=180dp）——窄容器单列场景下卡最宽、
        // 垂直溢出最大；TV 宽屏实际是多列近 90dp 地板，但护栏按最坏算。
        val worstColumnWidthDp = 179f
        val worstCardHeightDp = worstColumnWidthDp * 16f / 9f
        val scaleOverflowDp = TvFocusSafeSpec.requiredSafeSpaceDp(
            baseSizeDp = worstCardHeightDp,
            focusedScale = TvPosterWallFocusLayoutSpec.posterWallFocusedScale,
            focusedHaloPaddingDp = TvFocusSafeSpec.focusedHaloPaddingDp,
        )
        val requiredTopPaddingDp = scaleOverflowDp + TvPosterWallFocusLayoutSpec.posterWallFocusedShadowElevationDp
        assertTrue(
            "海报墙顶部内边距 ${TvPosterWallFocusLayoutSpec.gridTopPaddingDp}dp 应 ≥ 顶行聚焦垂直溢出 $scaleOverflowDp dp + 落影 ${TvPosterWallFocusLayoutSpec.posterWallFocusedShadowElevationDp}dp（最坏列宽 ${worstColumnWidthDp}dp × 16/9）",
            TvPosterWallFocusLayoutSpec.gridTopPaddingDp >= requiredTopPaddingDp,
        )
    }

    @Test
    fun `poster wall cards use focus safe outer containers`() {
        assertTrue(TvPosterWallFocusLayoutSpec.posterCardsUseFocusSafeContainer)
    }

    @Test
    fun `poster wall cards use portrait artwork with gradient scrim title`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt")
        val source = sourcePath.toFile().readText()

        assertTrue("海报墙卡片应使用 9:16 竖向图片区域", source.contains(".aspectRatio(9f / 16f)"))
        assertTrue("海报墙图片应贴边显示，不应保留旧的 12dp 图片内边距", !source.contains(".padding(12.dp)"))
        assertTrue("海报墙标题应融进海报底部渐变遮罩，不应保留独立深色标题色块", source.contains("TvPosterWallTitleScrimBrush"))
        assertTrue("海报墙不应再使用旧的独立标题色块常量", !source.contains("TvPosterWallTitleBackground"))
        assertTrue("海报墙卡片焦点态不应使用已删除的 tvFocusableGlow", !source.contains(".tvFocusableGlow("))
        assertTrue("海报墙卡片应使用只缩放焦点修饰器", source.contains(".tvFocusableScaleOnly("))
        assertTrue(
            "海报墙卡片焦点放大倍率应收口到海报墙本地 1.08f，不应改共享 token",
            source.contains("posterWallFocusedScale: Float = 1.08f") &&
                source.contains("focusedScale = TvPosterWallFocusLayoutSpec.posterWallFocusedScale"),
        )
        assertTrue(
            "海报墙不应再直接用共享 posterFocusedScale 作为卡片放大倍率，避免连带改首页/目录页",
            !source.contains("focusedScale = TvFocusSafeSpec.posterFocusedScale"),
        )
        assertTrue("共享焦点放大倍率应保持 1.04f，不应被海报墙连带改动", TvFocusSafeSpec.posterFocusedScale == 1.04f)
        assertTrue(
            "海报墙卡片应叠加中性落影抬升作为只缩放口径的例外",
            source.contains("posterWallFocusedShadowElevationDp") && source.contains(".shadow("),
        )
        assertTrue("海报墙中性落影应为黑色而非暖金", source.contains("ambientColor = Color.Black") && source.contains("spotColor = Color.Black"))
    }

    @Test
    fun `poster wall shared element binds to poster image only`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt")
        val source = sourcePath.toFile().readText()
        val cardSource = source.substringAfter("private fun TvPosterWallCard(")

        // tvSharedSeriesPoster 必须只绑在海报图片/占位节点上，不能挂在含 scrim+标题的外层 Box 上，
        // 否则 shared-element 过渡会把标题和遮罩一起「飞」进详情页（详情页端是纯背景图节点）。
        assertTrue(
            "海报墙 shared-element 应绑在海报图片节点上",
            cardSource.contains(".tvSharedSeriesPoster(item.id)"),
        )
        // 外层 9:16 Box 的修饰符链是 aspectRatio → clip → background，这段链上不应出现 tvSharedSeriesPoster。
        val outerBoxChain = cardSource.substringAfter(".aspectRatio(9f / 16f)")
            .substringBefore(".background(TvPosterWallPlaceholderBrush)")
        assertTrue(
            "海报墙 9:16 外层 Box 修饰符链不应携带 tvSharedSeriesPoster（否则过渡会捕获 scrim+标题）",
            !outerBoxChain.contains("tvSharedSeriesPoster"),
        )
        // shared-element 应是 fillMaxSize 的图片/占位节点上调用，而非外层 Box。
        assertTrue(
            "海报墙 shared-element 应跟在 fillMaxSize 之后（绑在图片/占位节点上，非外层 Box）",
            cardSource.contains("Modifier.fillMaxSize().then(sharedModifier)"),
        )
    }

    @Test
    fun `poster wall grid targets six columns on dense tv displays`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt")
        val source = sourcePath.toFile().readText()

        assertTrue(
            "海报墙网格 minSize 应降到 90dp 量级以在 4K TV 上塞进 6 列，不应保留旧的 170dp",
            source.contains("GridCells.Adaptive(minSize = 90.dp)"),
        )
    }

    @Test
    fun `poster wall top bar exposes sort controls`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt")
        val source = sourcePath.toFile().readText()

        assertTrue("海报墙顶部应提供排序字段切换", source.contains("tvPosterWallSortByLabel"))
        assertTrue("海报墙顶部应提供排序方向切换", source.contains("tvPosterWallSortOrderLabel"))
        assertTrue("海报墙排序应支持添加时间", source.contains("添加时间"))
        assertTrue("海报墙排序应支持发售时间", source.contains("发售时间"))
        assertTrue("海报墙排序应支持正序", source.contains("正序"))
        assertTrue("海报墙排序应支持倒序", source.contains("倒序"))
    }

    @Test
    fun `poster wall keeps grid visible during soft refresh`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt")
        val source = sourcePath.toFile().readText()
        val gridSource = source.substringAfter("LazyVerticalGrid(")

        assertTrue(
            "海报墙已有内容刷新/排序时必须在网格内显示行内更新状态，不能替换为整页 loading",
            gridSource.contains("if (uiState.refreshing)") &&
                gridSource.contains("TvInlineLoadingState") &&
                gridSource.contains("正在更新"),
        )
        assertTrue(
            "海报墙行内更新状态必须先于 itemsIndexed，保证旧内容仍保留在同一个网格里",
            gridSource.indexOf("if (uiState.refreshing)") < gridSource.indexOf("itemsIndexed("),
        )
        assertTrue(
            "海报墙已有内容刷新/排序失败时必须在网格内显示紧凑错误条，不能替换为整页错误态",
            gridSource.contains("if (!uiState.errorMessage.isNullOrBlank())") &&
                gridSource.contains("TvPosterWallInlineError"),
        )
        assertTrue(
            "海报墙行内错误必须先于 itemsIndexed，保证旧内容仍保留在同一个网格里",
            gridSource.indexOf("if (!uiState.errorMessage.isNullOrBlank())") < gridSource.indexOf("itemsIndexed("),
        )
    }

    @Test
    fun `poster wall soft update does not steal focus back to first poster`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt")
        val source = sourcePath.toFile().readText()
        val focusSource = source.substringAfter("var initialFocusRequested")
            .substringBefore("LaunchedEffect(gridState")
        val focusEffectKeys = source.substringAfter("LaunchedTvInitialFocus(")
            .substringBefore(") {")

        assertTrue("海报墙首屏焦点必须用一次性标记，软刷新/排序完成后不得再次抢回第一张海报", focusSource.contains("initialFocusRequested"))
        assertTrue("海报墙首焦点 effect 不得把 refreshing 作为 key，否则软更新完成会重新请求首项焦点", !focusEffectKeys.contains("uiState.refreshing"))
        assertTrue("海报墙首焦点请求成功后必须置位，避免后续软更新再次抢焦点", focusSource.contains("initialFocusRequested = true"))
    }
}
