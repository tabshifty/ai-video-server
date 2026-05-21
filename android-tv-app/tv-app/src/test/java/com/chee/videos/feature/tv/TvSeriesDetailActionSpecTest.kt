package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSeriesDetailActionSpecTest {
    @Test
    fun `series detail uses shared tv icon action for back action`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt")
        assertTrue("电视剧详情页必须存在", sourcePath.exists())

        val source = sourcePath.readText()
        assertTrue("电视剧详情页返回操作应复用共享 TV 图标操作", source.contains("TvIconActionButton("))
        assertFalse("电视剧详情页不应导入默认 Material IconButton", source.contains("import androidx.compose.material3.IconButton"))
        assertFalse("电视剧详情页不应导入默认 Material Button", source.contains("import androidx.compose.material3.Button"))
        assertFalse("电视剧详情页不应导入默认 Material TextButton", source.contains("import androidx.compose.material3.TextButton"))
        assertFalse("电视剧详情页不应使用默认 Material IconButton", source.contains("IconButton("))
        assertFalse("电视剧详情页不应使用默认 Material TextButton", source.contains("TextButton("))
        assertTrue("电视剧详情页播放、季选择和集选择仍应使用共享焦点视觉", source.contains(".tvFocusableGlow("))
    }
}
