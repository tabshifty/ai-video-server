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
        val preparingMessage = source.indexOf("正在切到第")
        val overlayState = source.indexOf("episodeSwitchState = uiState.episodeSwitchState")

        assertTrue(preparingBranch >= 0)
        assertTrue(preparingMessage > preparingBranch)
        assertTrue(overlayState > 0)
    }
}
