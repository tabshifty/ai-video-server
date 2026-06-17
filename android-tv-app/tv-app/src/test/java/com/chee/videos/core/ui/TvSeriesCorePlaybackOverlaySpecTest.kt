package com.chee.videos.core.ui

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSeriesCorePlaybackOverlaySpecTest {
    @Test
    fun overlaySupportsEpisodeSwitchCenterFeedbackAndPersistentFailureState() {
        val source = Path.of("src/main/java/com/chee/videos/core/ui/TvSeriesCorePlaybackOverlay.kt").readText()

        assertTrue(source.contains("episodeSwitchState: com.chee.videos.feature.tv.TvEpisodeSwitchUiState? = null"))
        assertTrue(source.contains("onDismissEpisodeSwitchFeedback: () -> Unit = {}"))
        assertTrue(source.contains("TvEpisodeSwitchUiState.Preparing"))
        assertTrue(source.contains("TvEpisodeSwitchUiState.Succeeded"))
        assertTrue(source.contains("TvEpisodeSwitchUiState.Canceled"))
        assertTrue(source.contains("TvEpisodeSwitchUiState.Failed"))
        assertTrue(source.contains("persistentCenterFailureMessage"))
        assertTrue(source.contains("text = \"留在当前分集\""))
    }
}
