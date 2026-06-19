package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvLongFormPlayerSoftRetrySpecTest {
    @Test
    fun `long form player keeps fullscreen error only before first frame and otherwise uses center soft retry feedback`() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt").readText()

        assertTrue(source.contains("if (!playerErrorMessage.isNullOrBlank() && !hasRenderedFirstFrame)"))
        assertTrue(source.contains("shouldUseTvLongFormSoftRetryFeedback(hasRenderedFirstFrame, currentSoftRetryUiState, showDolbyVisionDiagnostics)"))
        assertTrue(source.contains("TvLongFormSoftRetryFeedback("))
        assertTrue(source.contains("var hasRenderedFirstFrame by rememberSaveable(detail.id)"))
    }

    @Test
    fun `long form diagnostics panel only returns to failure state instead of triggering another retry`() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt").readText()

        assertTrue(source.contains("fun closeDiagnosticsPanel()"))
        assertTrue(source.contains("actionLabel = \"返回\""))
        assertTrue(source.contains("closeDiagnosticsPanel()"))
    }

    @Test
    fun `long form media3 player reports rendered first frame and supports prepare cancellation without recreating player`() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt").readText()

        assertTrue(source.contains("cancelPrepareRequestKey: Int = 0"))
        assertTrue(source.contains("onRenderedFirstFrame: () -> Unit = {}"))
        assertTrue(source.contains("override fun onRenderedFirstFrame()"))
        assertTrue(source.contains("val player = remember(accessToken) { ExoPlayer.Builder(context).build() }"))
        assertTrue(source.contains("LaunchedEffect(player, cancelPrepareRequestKey)"))
    }
}
