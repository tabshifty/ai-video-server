package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvRemotePlaybackControlsSpecTest {
    @Test
    fun remotePlaybackScreenAddsSeekChromeAndTopInfoOverlay() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvRemotePlaybackScreen.kt").readText()

        listOf(
            "AndroidKeyEvent.KEYCODE_DPAD_LEFT",
            "AndroidKeyEvent.KEYCODE_DPAD_RIGHT",
            "showSeekOverlay = true",
            "delay(TvRemoteChromeAutoHideDurationMillis)",
            "delay(TvRemoteSeekOverlayDurationMillis)",
            "Modifier.align(Alignment.TopStart)",
            ".align(Alignment.BottomCenter)",
            "TvRemotePlaybackBottomProgressBar(",
            "repository.readTvSeekStepSeconds()",
        ).forEach { line ->
            assertTrue("TvRemotePlaybackScreen 必须包含 $line", source.contains(line))
        }
    }
}
