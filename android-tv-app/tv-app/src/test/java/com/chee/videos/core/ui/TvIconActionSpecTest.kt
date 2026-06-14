package com.chee.videos.core.ui

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvIconActionSpecTest {
    @Test
    fun `shared tv icon action component exists`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/core/ui/TvIconAction.kt")
        assertTrue("TV 端应提供共享图标操作组件", sourcePath.exists())

        val source = sourcePath.readText()
        assertTrue("共享图标操作组件应命名为 TvIconActionButton", source.contains("fun TvIconActionButton("))
        assertTrue("TV 图标操作必须使用共享焦点视觉", source.contains(".tvFocusableGlow("))
        assertTrue("TV 图标操作默认容器应使用暗玻璃 surface", source.contains("containerColor: Color = AppChrome.Surface.copy(alpha = 0.62f)"))
        assertTrue("TV 图标操作默认图标色应使用参考图主文字色", source.contains("contentColor: Color = AppChrome.TextPrimary"))
        assertFalse("TV 图标操作组件内部不应使用默认 Material IconButton 作为主要焦点控件", source.contains("IconButton("))
    }

    @Test
    fun `tv screens do not use material icon button for primary icon actions`() {
        assertSourceDoesNotImportIconButton("src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt")
        assertSourceDoesNotImportIconButton("src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt")
        assertSourceDoesNotImportIconButton("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt")
    }

    @Test
    fun `iptv root focus remains a key event target not an icon action`() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt").readText()

        assertTrue("IPTV 播放页根节点仍需要焦点以接收遥控按键", source.contains(".focusable()"))
        assertTrue("IPTV 播放页根节点仍需要处理遥控按键", source.contains(".onPreviewKeyEvent"))
    }

    private fun assertSourceDoesNotImportIconButton(path: String) {
        val sourcePath = Path.of(path)
        assertTrue("$path 必须存在", sourcePath.exists())

        val source = sourcePath.readText()
        assertFalse("$path 不应导入默认 Material IconButton", source.contains("import androidx.compose.material3.IconButton"))
    }
}
