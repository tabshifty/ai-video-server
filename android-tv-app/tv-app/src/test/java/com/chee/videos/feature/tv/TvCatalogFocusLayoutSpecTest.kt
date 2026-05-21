package com.chee.videos.feature.tv

import com.chee.videos.core.ui.TvFocusSafeSpec
import com.chee.videos.core.ui.TvLayoutSpec
import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvCatalogFocusLayoutSpecTest {
    @Test
    fun `horizontal shelves keep edge and item focus safe space`() {
        assertTrue(TvCatalogFocusLayoutSpec.shelfHorizontalPaddingDp >= TvFocusSafeSpec.posterFocusSafeSpaceDp)
        assertTrue(TvCatalogFocusLayoutSpec.shelfVerticalPaddingDp >= TvFocusSafeSpec.posterFocusSafeSpaceDp)
        assertTrue(TvCatalogFocusLayoutSpec.shelfItemSpacingDp >= TvFocusSafeSpec.posterFocusSafeSpaceDp * 2)
    }

    @Test
    fun `poster shelf cards use focus safe outer containers`() {
        assertTrue(TvCatalogFocusLayoutSpec.posterCardsUseFocusSafeContainer)
    }

    @Test
    fun `poster shelf cards use scale only focus language`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt")
        assertTrue("TV 首页必须存在", sourcePath.exists())

        val source = sourcePath.readText()
        assertTrue("TV 首页海报卡应导入无描边焦点修饰器", source.contains("tvFocusableScaleOnly"))
        assertTrue("TV 首页海报卡不应再用 tvFocusableGlow 包裹海报安全容器", !source.contains(".tvFocusableGlow(shape = RoundedCornerShape(16.dp), focusedScale = TvFocusSafeSpec.posterFocusedScale)"))
    }

    @Test
    fun `catalog scrollable content keeps bottom safe padding`() {
        assertTrue(TvLayoutSpec.scrollBottomSafePaddingDp >= 56f)
        assertTrue(TvCatalogFocusLayoutSpec.contentBottomPaddingDp >= TvLayoutSpec.scrollBottomSafePaddingDp)

        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt")
        assertTrue("TV 首页必须存在", sourcePath.exists())

        val source = sourcePath.readText()
        assertTrue(source.contains("bottom = TvCatalogFocusLayoutSpec.contentBottomPaddingDp.dp"))
    }
}
