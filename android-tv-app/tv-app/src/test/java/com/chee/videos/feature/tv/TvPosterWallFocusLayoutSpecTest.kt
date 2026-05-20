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
}
