package com.chee.videos.feature.tv

import org.junit.Assert.assertTrue
import org.junit.Test

class TvLongFormDetailGlassPanelSpecTest {

    private val detailSource: String by lazy {
        java.nio.file.Path
            .of("src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt")
            .toFile()
            .readText()
    }

    private val panelHelperSource: String by lazy {
        java.nio.file.Path
            .of("src/main/java/com/chee/videos/core/ui/TvDetailPanelBackground.kt")
            .toFile()
            .readText()
    }

    @Test
    fun `detail screen routes bottom panel through TvDetailGlassPanel`() {
        assertTrue(
            "TvLongFormDetailScreen 必须 import 玻璃面板 helper",
            detailSource.contains("import com.chee.videos.core.ui.TvDetailGlassPanel"),
        )
        assertTrue(
            "TvLongFormDetailScreen 必须 import 玻璃面板 token",
            detailSource.contains("import com.chee.videos.core.ui.TvDetailPanelTokens"),
        )
        assertTrue(
            "TvLongFormDetailScreen 底部面板必须用 TvDetailGlassPanel 容器",
            detailSource.contains("TvDetailGlassPanel("),
        )
        assertTrue(
            "面板内 Column 必须用 TvDetailPanelTokens.ContentPaddingHorizontalDp",
            detailSource.contains("TvDetailPanelTokens.ContentPaddingHorizontalDp"),
        )
        assertTrue(
            "面板内 Column 必须用 TvDetailPanelTokens.ContentPaddingVerticalDp",
            detailSource.contains("TvDetailPanelTokens.ContentPaddingVerticalDp"),
        )
    }

    @Test
    fun `detail screen no longer contains legacy panel color literal`() {
        assertTrue(
            "TvLongFormDetailScreen 禁止保留旧版面板硬编码颜色 0xD20B1018",
            !detailSource.contains("0xD20B1018"),
        )
    }

    @Test
    fun `panel helper performs api gating and reuses tokens`() {
        assertTrue(
            "TvDetailPanelBackground.kt 必须用 Build.VERSION.SDK_INT >= Build.VERSION_CODES.S 做 API 31+ 判定",
            panelHelperSource.contains("Build.VERSION.SDK_INT >= Build.VERSION_CODES.S"),
        )
        assertTrue(
            "TvDetailPanelBackground.kt 必须使用 Modifier.blur 实现 API 31+ 玻璃感",
            panelHelperSource.contains("Modifier.blur(") ||
                panelHelperSource.contains(".blur("),
        )
        assertTrue(
            "TvDetailPanelBackground.kt 必须 import androidx.compose.ui.draw.blur",
            panelHelperSource.contains("import androidx.compose.ui.draw.blur"),
        )
        assertTrue(
            "TvDetailPanelBackground.kt 必须用 Brush.verticalGradient 实现上沿渐变 scrim",
            panelHelperSource.contains("Brush.verticalGradient"),
        )
        assertTrue(
            "TvDetailPanelBackground.kt 必须读 TvDetailPanelTokens.BlurRadiusDp 而非裸字面量",
            panelHelperSource.contains("TvDetailPanelTokens.BlurRadiusDp"),
        )
        assertTrue(
            "TvDetailPanelBackground.kt 必须读 TvDetailPanelTokens.ScrimColorBlurred",
            panelHelperSource.contains("TvDetailPanelTokens.ScrimColorBlurred"),
        )
        assertTrue(
            "TvDetailPanelBackground.kt 必须读 TvDetailPanelTokens.ScrimColorFallback",
            panelHelperSource.contains("TvDetailPanelTokens.ScrimColorFallback"),
        )
        assertTrue(
            "TvDetailPanelBackground.kt 必须读 TvDetailPanelTokens.UpperGradientHeightDp",
            panelHelperSource.contains("TvDetailPanelTokens.UpperGradientHeightDp"),
        )
        assertTrue(
            "TvDetailPanelBackground.kt 必须读 TvDetailPanelTokens.TopCornerRadiusDp",
            panelHelperSource.contains("TvDetailPanelTokens.TopCornerRadiusDp"),
        )
    }

    @Test
    fun `panel helper exposes composable TvDetailGlassPanel`() {
        assertTrue(
            "TvDetailGlassPanel 必须是 @Composable",
            panelHelperSource.contains("@Composable") &&
                panelHelperSource.contains("fun TvDetailGlassPanel("),
        )
    }
}
