package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvRemotePlaybackControlsSpecTest {
    @Test
    fun remotePlaybackScreenAddsSeekChromeAndTopInfoOverlay() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvRemotePlaybackScreen.kt").readText()

        listOf(
            ".focusRequester(rootFocusRequester)",
            ".focusable()",
            "rootFocusRequester.tryRequestFocus()",
            "AndroidKeyEvent.KEYCODE_DPAD_LEFT",
            "AndroidKeyEvent.KEYCODE_DPAD_RIGHT",
            "showSeekOverlay = true",
            "delay(TvRemoteChromeAutoHideDurationMillis)",
            "delay(TvRemoteSeekOverlayDurationMillis)",
            "Modifier.align(Alignment.TopStart)",
            ".align(Alignment.BottomCenter)",
            "TvRemotePlaybackBottomProgressBar(",
            "repository.readTvSeekStepSeconds()",
            "contentScale = ContentScale.Fit",
            "resizeMode = AspectRatioFrameLayout.RESIZE_MODE_FIT",
        ).forEach { line ->
            assertTrue("TvRemotePlaybackScreen 必须包含 $line", source.contains(line))
        }
    }

    @Test
    fun remotePlaybackScreenAutoAdvancesSharedSessionWhenCurrentItemEnds() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvRemotePlaybackScreen.kt").readText()

        listOf(
            "Player.STATE_ENDED",
            "shouldTvRemoteAutoplayNext(",
            "viewModel.autoNext()",
        ).forEach { line ->
            assertTrue("TV 远程投放页自动连播必须包含 $line", source.contains(line))
        }
    }

    @Test
    fun remotePlaybackUsesSharedPrimaryShortSourceAndDiagnostics() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvRemotePlaybackScreen.kt").readText()

        listOf(
            "buildTvShortPlaybackSourceUrl(baseUrl, currentVideoId)",
            "TvShortPlaybackDiagnostics()",
            "diagnostics.attach(sharedPlayer)",
            "diagnostics.sourcePrepared(attempt)",
            "diagnostics.detach(sharedPlayer)",
            "createTvShortPlaybackAttempt(",
            "entry = TvShortPlaybackEntry.Remote",
        ).forEach { line ->
            assertTrue("TV 远程投屏页必须包含 $line", source.contains(line))
        }
        assertFalse(
            "TV 远程投屏页不得再复用默认走 compat 的长视频源构造",
            source.contains("repository.buildSourceUrl(currentVideoId)"),
        )
    }
}
