package com.chee.videos.feature.tv

import com.chee.videos.core.ui.TvLayoutSpec
import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvScrollableBottomPaddingTest {
    @Test
    fun `tv scrollable pages use shared bottom safe padding`() {
        assertTrue(TvLayoutSpec.scrollBottomSafePaddingDp >= 56f)

        assertSourceContains(
            path = "src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt",
            pattern = "bottom = TvLayoutSpec.scrollBottomSafePaddingDp.dp",
        )
        assertSourceContains(
            path = "src/main/java/com/chee/videos/feature/connection/ConnectionScreen.kt",
            pattern = "bottom = TvLayoutSpec.scrollBottomSafePaddingDp.dp",
        )
    }

    @Test
    fun `immersive tv player page does not reuse scroll bottom safe padding`() {
        assertSourceNotContains(
            path = "src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt",
            pattern = "TvLayoutSpec.scrollBottomSafePaddingDp",
        )
        assertSourceNotContains(
            path = "src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt",
            pattern = "TvLayoutSpec.scrollBottomSafePaddingDp",
        )
    }

    private fun assertSourceContains(path: String, pattern: String) {
        val sourcePath = Path.of(path)
        assertTrue("$path 必须存在", sourcePath.exists())
        assertTrue("$path 必须使用统一 TV 滚动内容底部安全留白", sourcePath.readText().contains(pattern))
    }

    private fun assertSourceNotContains(path: String, pattern: String) {
        val sourcePath = Path.of(path)
        assertTrue("$path 必须存在", sourcePath.exists())
        assertTrue("$path 不应使用滚动页底部安全留白", !sourcePath.readText().contains(pattern))
    }
}
