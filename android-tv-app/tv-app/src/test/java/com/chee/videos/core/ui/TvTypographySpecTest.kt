package com.chee.videos.core.ui

import androidx.compose.ui.unit.sp
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvTypographySpecTest {
    @Test
    fun `typography follows compact reference image hierarchy`() {
        assertTrue(
            "参考图主标题 token 应保持 30sp 紧凑基线",
            TvTypographyTokens.HeroTitleSp == 30,
        )
        assertTrue(
            "headlineLarge 应等于参考图主标题 token",
            TvTypography.headlineLarge.fontSize.value == TvTypographyTokens.HeroTitleSp.toFloat(),
        )
        assertTrue(
            "titleSmall 应落在参考图卡片标题尺寸，不再保留 22sp 字号地板",
            TvTypography.titleSmall.fontSize.value == 15f,
        )
        assertTrue(
            "bodyMedium 应落在参考图正文尺寸，不再保留 18sp 字号地板",
            TvTypography.bodyMedium.fontSize.value == 13f,
        )
    }

    @Test
    fun `typography token names describe reference roles instead of 10-foot floors`() {
        val source = java.nio.file.Path.of("src/main/java/com/chee/videos/core/ui/TvTypography.kt").toFile().readText()

        assertTrue(
            "TvTypographyTokens 应暴露 HeroTitleSp",
            source.contains("HeroTitleSp"),
        )
        assertTrue(
            "TvTypographyTokens 应暴露 SectionTitleSp",
            source.contains("SectionTitleSp"),
        )
        assertFalse(
            "参考图换代后不再暴露 MainTitleSp 这类 10-foot 地板 token",
            source.contains("MainTitleSp"),
        )
        assertFalse(
            "参考图换代后不再暴露 SubtitleFloorSp 这类 10-foot 地板 token",
            source.contains("SubtitleFloorSp"),
        )
        assertFalse(
            "参考图换代后不再暴露 HelperFloorSp 这类 10-foot 地板 token",
            source.contains("HelperFloorSp"),
        )
    }

    @Test
    fun `body and label roles stay compact but readable`() {
        assertTrue(
            "bodyLarge 应等于参考图正文 token",
            TvTypography.bodyLarge.fontSize.value == TvTypographyTokens.BodySp.toFloat(),
        )
        assertTrue(
            "bodySmall 应等于参考图辅助文字 token",
            TvTypography.bodySmall.fontSize.value == TvTypographyTokens.HelperSp.toFloat(),
        )
        assertTrue(
            "labelSmall 应等于参考图角标 token",
            TvTypography.labelSmall.fontSize.value == TvTypographyTokens.CaptionSp.toFloat(),
        )
        assertTrue(
            "紧凑基线下 labelSmall 仍不应小于 11sp",
            TvTypography.labelSmall.fontSize.value >= 11f,
        )
        assertTrue(
            "正文基线不应大于 14sp，避免回退到旧 TV 大字号风格",
            TvTypographyTokens.BodySp <= 14,
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
                "$name lineHeight 必须 ≥ fontSize，避免紧凑排版行距塌陷",
                style.lineHeight.value >= style.fontSize.value,
            )
        }
    }

    @Test
    fun `letter spacing stays neutral for reference typography`() {
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
                "$name letterSpacing 必须为 0sp，避免旧 Material 默认 tracking 破坏参考图字面",
                style.letterSpacing.value == 0f,
            )
        }
    }

    @Test
    fun `tv shell app wires typography into materialtheme`() {
        val source = java.nio.file.Path.of(
            "src/main/java/com/chee/videos/tv/TvShellApp.kt",
        ).toFile().readText()

        assertTrue(
            "TvShellApp 必须显式向 MaterialTheme 注入 TvTypography，否则全量调用点无法继承参考图紧凑字号",
            source.contains("typography = TvTypography"),
        )
        assertTrue(
            "TvShellApp 必须 import TvTypography",
            source.contains("import com.chee.videos.core.ui.TvTypography"),
        )
    }

    @Test
    fun `tv typography tokens stay in sync with compact role constants`() {
        assertTrue(
            "HeroTitleSp 必须等于 headlineLarge.fontSize",
            TvTypographyTokens.HeroTitleSp.toFloat() == TvTypography.headlineLarge.fontSize.value,
        )
        assertTrue(
            "SectionTitleSp 必须等于 titleLarge.fontSize",
            TvTypographyTokens.SectionTitleSp.toFloat() == TvTypography.titleLarge.fontSize.value,
        )
        assertTrue(
            "HelperSp 必须等于 bodySmall.fontSize",
            TvTypographyTokens.HelperSp.toFloat() == TvTypography.bodySmall.fontSize.value,
        )
    }

    @Test
    fun `tv typography tokens expose sp helpers for downstream consumers`() {
        assertTrue(
            "HeroTitleSp 应转 sp",
            TvTypographyTokens.HeroTitleSp.sp.value == 30f,
        )
    }
}
