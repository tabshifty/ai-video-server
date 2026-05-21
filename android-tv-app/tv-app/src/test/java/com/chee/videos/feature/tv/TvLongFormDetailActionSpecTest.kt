package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvLongFormDetailActionSpecTest {
    @Test
    fun `long form detail uses shared tv icon action for back action`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt")
        assertTrue("长视频详情页必须存在", sourcePath.exists())

        val source = sourcePath.readText()
        assertTrue("长视频详情页返回操作应复用共享 TV 图标操作", source.contains("TvIconActionButton("))
        assertFalse("长视频详情页不应导入默认 Material IconButton", source.contains("import androidx.compose.material3.IconButton"))
        assertFalse("长视频详情页不应导入默认 Material Button", source.contains("import androidx.compose.material3.Button"))
        assertFalse("长视频详情页不应导入默认 Material TextButton", source.contains("import androidx.compose.material3.TextButton"))
        assertFalse("长视频详情页不应使用默认 Material IconButton", source.contains("IconButton("))
        assertFalse("长视频详情页不应使用默认 Material TextButton", source.contains("TextButton("))
        assertTrue("长视频详情页播放和收藏仍应使用共享焦点视觉", source.contains(".tvFocusableGlow("))
    }

    @Test
    fun `immersive long form detail does not use scroll bottom safe padding`() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt").readText()

        assertTrue("长视频详情页应保持沉浸式首屏背景", source.contains("TvLongFormDetailBackground("))
        assertFalse(
            "沉浸式长视频详情首屏不应套用滚动页底部安全留白",
            source.contains("TvLayoutSpec.scrollBottomSafePaddingDp"),
        )
    }
}
