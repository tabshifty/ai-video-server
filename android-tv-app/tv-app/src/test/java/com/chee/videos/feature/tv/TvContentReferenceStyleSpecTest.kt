package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvContentReferenceStyleSpecTest {
    private val contentPageSources = listOf(
        "TV 首页" to Path.of("src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt"),
        "海报墙" to Path.of("src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt"),
        "长视频详情页" to Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt"),
    )

    @Test
    fun `content pages do not keep legacy cold placeholder palette`() {
        val legacyColdColors = listOf(
            "0xFF211527",
            "0xFF0B101A",
            "0xE80B0F17",
            "0xC4111520",
            "0xAA111827",
            "0xFF1A2031",
            "0xFF111827",
            "0xE61A2031",
            "0xF0111827",
            "0xFF23314A",
            "0xFF0B0F18",
            "0xFF2D1A48",
            "0xFF3C1D2E",
            "0xFF1F3A56",
            "0xFF1C3D36",
            "0xE80A0E15",
            "0xFF20142D",
            "0xFF0C1018",
            "0x6610151F",
            "0x3310151F",
            "0xDD070A10",
            "0x66080B11",
            "0xFF23131F",
            "0xFF090C12",
            "0xFF253044",
            "0x33090C13",
        )

        contentPageSources.forEach { (label, path) ->
            assertTrue("$label 源文件必须存在", path.exists())
            val source = path.readText()
            legacyColdColors.forEach { color ->
                assertFalse("$label 不应保留旧冷色/紫蓝参考色 $color", source.contains(color))
            }
        }
    }

    @Test
    fun `content page gold controls use dark foreground instead of white`() {
        contentPageSources.forEach { (label, path) ->
            assertTrue("$label 源文件必须存在", path.exists())
            val source = path.readText()
            assertFalse("$label 不应再直接使用 Color.White 作为参考图内容页前景", source.contains("Color.White"))
        }

        val catalog = Path.of("src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt").readText()
        assertTrue("TV 首页金色主按钮文字/图标应使用深色画布前景", catalog.contains("tint = AppChrome.Canvas"))
        assertTrue("TV 首页金色标签文字应使用深色画布前景", catalog.contains("color = AppChrome.Canvas"))

        val detail = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt").readText()
        assertTrue("长视频详情页金色主按钮文字/图标应使用深色画布前景", detail.contains("tint = primaryContentColor"))
        assertTrue("长视频详情页金色主按钮文字应跟随 primaryContentColor", detail.contains("color = primaryContentColor"))
    }
}
