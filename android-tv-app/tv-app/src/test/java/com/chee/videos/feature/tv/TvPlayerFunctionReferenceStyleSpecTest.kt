package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvPlayerFunctionReferenceStyleSpecTest {
    private val playerAndFunctionSources = listOf(
        "播放器" to Path.of("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt"),
        "轨道选择器" to Path.of("src/main/java/com/chee/videos/core/ui/SubtitlePicker.kt"),
        "返回确认提示" to Path.of("src/main/java/com/chee/videos/feature/tv/TvPlayerBackConfirm.kt"),
        "连接服务器页" to Path.of("src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt"),
        "配对页" to Path.of("src/main/java/com/chee/videos/tv/TvPairingScreen.kt"),
    )

    @Test
    fun `third batch player and function pages do not use direct white foregrounds`() {
        playerAndFunctionSources.forEach { (label, path) ->
            assertTrue("$label 源文件必须存在", path.exists())
            val source = path.readText()
            assertFalse("$label 不应直接使用 Color.White 作为参考图第三批前景", source.contains("Color.White"))
        }
    }

    @Test
    fun `third batch removes legacy cold player surface literals`() {
        val legacyColdColors = listOf(
            "0xD610131A",
            "0xCC0D1016",
            "0x8C0D1016",
            "0x910D1016",
            "0xE8141820",
            "0x24000000",
            "0xCC121212",
            "0x99000000",
            "0x66FFFFFF",
            "0x1AFFFFFF",
            "0xE610161F",
            "0xCC241C11",
            "0xE6040508",
            "0x26FFFFFF",
            "0x1AFFFFFF",
            "0x12000000",
        )

        playerAndFunctionSources.forEach { (label, path) ->
            assertTrue("$label 源文件必须存在", path.exists())
            val source = path.readText()
            legacyColdColors.forEach { color ->
                assertFalse("$label 不应保留旧播放器/功能页冷色字面量 $color", source.contains(color))
            }
        }
    }

    @Test
    fun `gold controls in connection and pairing pages use canvas foreground`() {
        val connection = Path.of("src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt").readText()
        assertTrue("连接页金色主操作文字应使用深色画布前景", connection.contains("primary -> AppChrome.Canvas"))

        val pairing = Path.of("src/main/java/com/chee/videos/tv/TvPairingScreen.kt").readText()
        assertTrue("配对页金色主操作文字应使用深色画布前景", pairing.contains("if (primary) AppChrome.Canvas"))
    }
}
