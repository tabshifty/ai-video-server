package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvAutoplayPromptCardSpecTest {
    @Test
    fun `series player imports shared autoplay prompt and end overlay`() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").readText()

        assertTrue(source.contains("TvAutoplayPromptCard("))
        assertTrue(source.contains("TvSeriesEndOverlay("))
    }

    @Test
    fun `autoplay prompt and overlay use shared motion and shape tokens`() {
        val prompt = Path.of("src/main/java/com/chee/videos/feature/tv/TvAutoplayPromptCard.kt").readText()
        val overlay = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesEndOverlay.kt").readText()

        assertTrue(prompt.contains("TvMotionTokens.DurationStandardMs"))
        assertTrue(prompt.contains("TvMotionTokens.EasingStandard"))
        assertTrue(prompt.contains("AppChrome.SurfaceShape"))
        assertTrue(prompt.contains("LaunchedTvInitialFocus("))
        assertTrue(prompt.contains("tryRequestFocus()"))
        assertTrue(prompt.contains("观看片尾"))
        assertFalse(prompt.contains("RoundedCornerShape("))
        assertFalse(prompt.contains("tween("))

        assertTrue(overlay.contains("TvLongFormCompletionOverlay("))
        assertTrue(overlay.contains("onBackToDetail = onBackToDetail"))
        assertTrue(overlay.contains("\"重播本集\""))
        assertFalse(overlay.contains("RoundedCornerShape("))
        assertFalse(overlay.contains("tween("))
    }

    @Test
    fun `next episode control uses skip next icon`() {
        val source = Path.of("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt").readText()
        val nextEpisodeBlock = source.substringAfter("onNextEpisode?.let").substringBefore("contentDescription = \"字幕\"")

        assertTrue(source.contains("import androidx.compose.material.icons.filled.SkipNext"))
        assertTrue(nextEpisodeBlock.contains("Icons.Filled.SkipNext"))
        assertFalse(nextEpisodeBlock.contains("Icons.Filled.FastForward"))
    }
}
