package com.chee.videos.core.ui

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * 共享操作按钮审计：TV 端主/次操作按钮统一由 core/ui/TvActionButton 承载，
 * 各屏禁止再各自手写胶囊按钮（TvSeriesDetailScreen 按 CONTEXT.md「TV 参考图」授权豁免）。
 */
class TvActionButtonSpecTest {

    private val source: String by lazy {
        Path.of("src/main/java/com/chee/videos/core/ui/TvActionButton.kt").readText()
    }

    @Test
    fun sharedActionButtonCarriesCanonicalStyle() {
        assertTrue("TvActionButton.kt 必须存在", Path.of("src/main/java/com/chee/videos/core/ui/TvActionButton.kt").exists())
        assertTrue("必须暴露 fun TvActionButton(", source.contains("fun TvActionButton("))
        assertTrue("焦点反馈必须走共享 scale-only 修饰器", source.contains(".tvFocusableScaleOnly("))
        assertTrue("形状必须是胶囊 PillShape", source.contains("AppChrome.PillShape"))
        assertTrue("主操作底色必须是暖金 Accent", source.contains("tone == TvActionButtonTone.Primary -> AppChrome.Accent"))
        assertTrue("主操作前景必须是深色画布", source.contains("tone == TvActionButtonTone.Primary -> AppChrome.Canvas"))
        assertTrue(
            "内边距必须是 canonical 的 18/12（保证 10-foot 焦点目标高度）",
            source.contains("horizontal = 18.dp, vertical = 12.dp"),
        )
        assertTrue("按钮文字必须单行省略", source.contains("maxLines = 1"))
        assertFalse("不得使用默认 Material Button", source.contains("import androidx.compose.material3.Button"))
        assertFalse("不得使用默认 Material IconButton", source.contains("import androidx.compose.material3.IconButton"))
    }

    @Test
    fun screensNoLongerDefineLocalActionButtons() {
        val offenders = listOf(
            "TvCatalogScreen" to "src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt",
            "TvLongFormDetailScreen" to "src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt",
            "TvPosterWallScreen" to "src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt",
        )
        val localButtonNames = listOf(
            "private fun TvHeroActionButton(",
            "private fun TvHeroSecondaryActionButton(",
            "private fun TvDetailPrimaryActionButton(",
            "private fun TvDetailSecondaryActionButton(",
            "private fun TvPosterWallSortButton(",
        )
        offenders.forEach { (label, path) ->
            val text = Path.of(path).readText()
            localButtonNames.forEach { name ->
                assertFalse("$label 不得再保留本地按钮实现 $name（应委托共享 TvActionButton）", text.contains(name))
            }
            assertTrue("$label 必须调用共享 TvActionButton", text.contains("TvActionButton("))
        }
    }

    @Test
    fun connectionAndPairingDelegateThroughThinWrappers() {
        val connection = Path.of("src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt").readText()
        assertTrue(
            "ConnectionActionButton 薄封装必须委托 TvActionButton",
            connection.contains("ConnectionActionButton(") && connection.contains("TvActionButton("),
        )
        val pairing = Path.of("src/main/java/com/chee/videos/tv/TvPairingScreen.kt").readText()
        assertTrue(
            "TvPairingActionButton 薄封装必须委托 TvActionButton",
            pairing.contains("TvPairingActionButton(") && pairing.contains("TvActionButton("),
        )
    }
}
