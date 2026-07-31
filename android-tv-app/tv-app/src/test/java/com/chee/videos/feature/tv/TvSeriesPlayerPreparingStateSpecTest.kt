package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSeriesPlayerPreparingStateSpecTest {
    @Test
    fun `series player keeps fullscreen preparing only for initial no-source state`() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").readText()
        val preparingBranch = source.indexOf("uiState.playbackPreparing && uiState.currentSourceUrl.isBlank()")
        val media3Player = source.indexOf("TvLongFormMedia3Player(")
        val newTargetSource = source.indexOf("sourceUrl = uiState.currentSourceUrl.takeIf { isMedia3Route }.orEmpty()")

        assertTrue(preparingBranch >= 0)
        assertTrue(media3Player > 0)
        assertTrue(newTargetSource > media3Player)
        assertTrue(source.contains("delay(TvLongFormStartupFeedbackDelayMillis)"))
        assertTrue(source.contains("if (sourcePreparingFeedbackVisible)"))
    }
}
