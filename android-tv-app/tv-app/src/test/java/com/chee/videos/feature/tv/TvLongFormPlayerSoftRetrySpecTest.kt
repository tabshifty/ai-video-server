package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
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
        val diagnosticsBlock = source
            .substringAfter("fun closeDiagnosticsPanel() {")
            .substringBefore("fun requestPlaybackRetry()")

        assertTrue(source.contains("fun closeDiagnosticsPanel()"))
        assertTrue(source.contains("actionLabel = \"返回\""))
        assertTrue(source.contains("closeDiagnosticsPanel()"))
        assertTrue(source.contains("if (showDolbyVisionDiagnostics && !playerErrorMessage.isNullOrBlank())"))
        assertFalse(diagnosticsBlock.contains("requestPlaybackRetry"))
        assertFalse(diagnosticsBlock.contains("routeRetryNonce += 1"))
    }

    @Test
    fun `long form retry cancel keeps the current playback session active`() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt").readText()
        val cancelBlock = source
            .substringAfter("fun cancelCurrentPlaybackRetry() {")
            .substringBefore("fun dismissSoftRetryFailure()")

        assertTrue(source.contains("fun cancelCurrentPlaybackRetry()"))
        assertTrue(source.contains("cancelPrepareRequestKey += 1"))
        assertTrue(source.contains("softRetryUiState = TvLongFormSoftRetryUiState.Canceled(preparingState.retryKey, \"已取消重试\")"))
        assertTrue(source.contains("ignoredRetryAttemptKey = preparingState.retryKey"))
        assertTrue(source.contains("playerErrorMessage = null"))
        assertFalse(cancelBlock.contains("hasStartedPlayback = false"))

        val media3Source = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt").readText()
        val cancelEffectBlock = media3Source
            .substringAfter("LaunchedEffect(player, cancelPrepareRequestKey) {")
            .substringBefore("LaunchedEffect(player, preparedSourceKey, initialPositionMs) {")

        assertTrue(media3Source.contains("LaunchedEffect(player, cancelPrepareRequestKey)"))
        assertTrue(media3Source.contains("latestOnSnapshotChanged(player.readTvMedia3PlaybackSnapshot())"))
        assertFalse(cancelEffectBlock.contains("player.pause()"))
        assertFalse(cancelEffectBlock.contains("player.stop()"))
        assertFalse(cancelEffectBlock.contains("latestOnPlayingChanged(false)"))
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
