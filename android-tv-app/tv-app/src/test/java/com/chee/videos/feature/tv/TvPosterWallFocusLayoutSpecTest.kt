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
    }

    @Test
    fun `poster wall cards use focus safe outer containers`() {
        assertTrue(TvPosterWallFocusLayoutSpec.posterCardsUseFocusSafeContainer)
    }

    @Test
    fun `poster wall cards use portrait artwork with flush dark title bar`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt")
        val source = sourcePath.toFile().readText()

        assertTrue("海报墙卡片应使用 9:16 竖向图片区域", source.contains(".aspectRatio(9f / 16f)"))
        assertTrue("海报墙图片应贴边显示，不应保留旧的 12dp 图片内边距", !source.contains(".padding(12.dp)"))
        assertTrue("海报墙标题条应使用深色背景并紧贴图片底部", source.contains("TvPosterWallTitleBackground"))
        assertTrue("海报墙卡片焦点态只能放大，不应使用带描边的 tvFocusableGlow", !source.contains(".tvFocusableGlow(shape = RoundedCornerShape(18.dp)"))
        assertTrue("海报墙卡片应使用无描边焦点修饰器", source.contains(".tvFocusableScaleOnly("))
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
