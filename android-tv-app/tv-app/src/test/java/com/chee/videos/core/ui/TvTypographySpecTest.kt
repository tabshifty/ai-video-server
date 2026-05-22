package com.chee.videos.core.ui

import androidx.compose.ui.unit.sp
import org.junit.Assert.assertTrue
import org.junit.Test

class TvTypographySpecTest {
    @Test
    fun `headline large meets 10-foot main title floor of 34sp`() {
        assertTrue(
            "10-foot 主标题地板 34sp",
            TvTypography.headlineLarge.fontSize.value >= 34f,
        )
        assertTrue(
            "TvTypographyTokens.MainTitleSp 必须 ≥ 34",
            TvTypographyTokens.MainTitleSp >= 34,
        )
    }

    @Test
    fun `title small meets 10-foot subtitle floor of 22sp`() {
        assertTrue(
            "海报卡标题（titleSmall）必须 ≥ 22sp",
            TvTypography.titleSmall.fontSize.value >= TvTypographyTokens.SubtitleFloorSp.toFloat(),
        )
        assertTrue(
            "TvTypographyTokens.SubtitleFloorSp 必须 ≥ 22",
            TvTypographyTokens.SubtitleFloorSp >= 22,
        )
    }

    @Test
    fun `title medium does not drop below subtitle floor`() {
        assertTrue(
            "section / 卡片标题（titleMedium）必须 ≥ 22sp",
            TvTypography.titleMedium.fontSize.value >= 22f,
        )
    }

    @Test
    fun `body and label roles meet 10-foot helper floor of 18sp`() {
        assertTrue(
            "bodyMedium 必须 ≥ HelperFloorSp",
            TvTypography.bodyMedium.fontSize.value >= TvTypographyTokens.HelperFloorSp.toFloat(),
        )
        assertTrue(
            "bodySmall 必须 ≥ 18sp",
            TvTypography.bodySmall.fontSize.value >= 18f,
        )
        assertTrue(
            "labelLarge 必须 ≥ 18sp",
            TvTypography.labelLarge.fontSize.value >= 18f,
        )
        assertTrue(
            "labelMedium 必须 ≥ 18sp",
            TvTypography.labelMedium.fontSize.value >= 18f,
        )
        assertTrue(
            "labelSmall 必须 ≥ TightHelperSp",
            TvTypography.labelSmall.fontSize.value >= TvTypographyTokens.TightHelperSp.toFloat(),
        )
        assertTrue(
            "TvTypographyTokens.HelperFloorSp 必须 ≥ 18",
            TvTypographyTokens.HelperFloorSp >= 18,
        )
        assertTrue(
            "TvTypographyTokens.TightHelperSp 必须 ≥ 18",
            TvTypographyTokens.TightHelperSp >= 18,
        )
    }

    @Test
    fun `material3 monotonicity holds across heading and title roles`() {
        val hl = TvTypography.headlineLarge.fontSize.value
        val hm = TvTypography.headlineMedium.fontSize.value
        val hs = TvTypography.headlineSmall.fontSize.value
        val tl = TvTypography.titleLarge.fontSize.value
        val tm = TvTypography.titleMedium.fontSize.value
        val ts = TvTypography.titleSmall.fontSize.value
        assertTrue("headlineLarge ≥ headlineMedium", hl >= hm)
        assertTrue("headlineMedium ≥ headlineSmall", hm >= hs)
        assertTrue("headlineSmall ≥ titleLarge", hs >= tl)
        assertTrue("titleLarge ≥ titleMedium", tl >= tm)
        assertTrue("titleMedium ≥ titleSmall", tm >= ts)
    }

    @Test
    fun `body large stays above body medium`() {
        assertTrue(
            "bodyLarge 必须 ≥ bodyMedium",
            TvTypography.bodyLarge.fontSize.value >= TvTypography.bodyMedium.fontSize.value,
        )
    }

    @Test
    fun `line heights stay above font sizes for each role`() {
        val roles = listOf(
            "headlineLarge" to TvTypography.headlineLarge,
            "headlineMedium" to TvTypography.headlineMedium,
            "headlineSmall" to TvTypography.headlineSmall,
            "titleLarge" to TvTypography.titleLarge,
            "titleMedium" to TvTypography.titleMedium,
            "titleSmall" to TvTypography.titleSmall,
            "bodyLarge" to TvTypography.bodyLarge,
            "bodyMedium" to TvTypography.bodyMedium,
            "bodySmall" to TvTypography.bodySmall,
            "labelLarge" to TvTypography.labelLarge,
            "labelMedium" to TvTypography.labelMedium,
            "labelSmall" to TvTypography.labelSmall,
        )
        for ((name, style) in roles) {
            assertTrue(
                "$name lineHeight 必须 ≥ fontSize，避免 10-foot 行距塌陷",
                style.lineHeight.value >= style.fontSize.value,
            )
        }
    }

    @Test
    fun `tv shell app wires typography into materialtheme`() {
        val source = java.nio.file.Path.of(
            "src/main/java/com/chee/videos/tv/TvShellApp.kt",
        ).toFile().readText()

        assertTrue(
            "TvShellApp 必须显式向 MaterialTheme 注入 TvTypography，否则全量调用点继承 Material3 默认 sp，无法满足 10-foot",
            source.contains("typography = TvTypography"),
        )
        assertTrue(
            "TvShellApp 必须 import TvTypography",
            source.contains("import com.chee.videos.core.ui.TvTypography"),
        )
    }

    @Test
    fun `tv typography tokens stay in sync with floor constants`() {
        // 主标题地板必须等于 headlineLarge 实际配置，避免 token 与实例脱节
        assertTrue(
            "MainTitleSp 必须等于 headlineLarge.fontSize",
            TvTypographyTokens.MainTitleSp.toFloat() == TvTypography.headlineLarge.fontSize.value,
        )
        assertTrue(
            "SubtitleFloorSp 必须等于 titleSmall.fontSize",
            TvTypographyTokens.SubtitleFloorSp.toFloat() == TvTypography.titleSmall.fontSize.value,
        )
    }

    @Test
    fun `tv typography tokens expose sp helpers for downstream consumers`() {
        assertTrue(
            "MainTitleSp 应转 sp",
            TvTypographyTokens.MainTitleSp.sp.value >= 34f,
        )
    }
}
