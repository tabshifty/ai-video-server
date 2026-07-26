package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvShortPlaybackDiagnosticsTest {
    @Test
    fun sharedSourceUrlUsesPrimaryEndpointWithoutProfileOverride() {
        val url = buildTvShortPlaybackSourceUrl(
            baseUrl = "https://media.example.test/",
            videoId = "video-1",
        )

        assertEquals("https://media.example.test/api/v1/videos/video-1/source", url)
    }

    @Test
    fun playbackAttemptKeepsOnlyComparableSourceIdentity() {
        val attempt = createTvShortPlaybackAttempt(
            entry = TvShortPlaybackEntry.Remote,
            attemptId = 7,
            videoId = "video-1",
            sourceUrl = "https://192.168.1.24/api/v1/videos/video-1/source?token=secret&profile=compat",
        )

        assertEquals("/api/v1/videos/video-1/source", attempt.sourcePath)
        assertEquals("compat", attempt.profile)
        assertFalse(attempt.diagnosticPrefix("source_prepared").contains("192.168.1.24"))
        assertFalse(attempt.diagnosticPrefix("source_prepared").contains("secret"))
        assertFalse(attempt.diagnosticPrefix("source_prepared").contains("token"))
    }

    @Test
    fun sourceWithoutExplicitProfileIsReportedAsPrimary() {
        val attempt = createTvShortPlaybackAttempt(
            entry = TvShortPlaybackEntry.Local,
            attemptId = 1,
            videoId = "video-2",
            sourceUrl = "http://server.test/api/v1/videos/video-2/source",
        )

        assertEquals("primary", attempt.profile)
        assertEquals(
            "event=source_prepared entry=local attempt=1 videoId=video-2 " +
                "source=/api/v1/videos/video-2/source profile=primary",
            attempt.diagnosticPrefix("source_prepared"),
        )
    }

    @Test
    fun unknownProfileIsNotCopiedIntoDiagnostics() {
        val attempt = createTvShortPlaybackAttempt(
            entry = TvShortPlaybackEntry.Remote,
            attemptId = 2,
            videoId = "video-3",
            sourceUrl = "http://server.test/api/v1/videos/video-3/source?profile=private-value",
        )

        assertEquals("unknown", attempt.profile)
        assertFalse(attempt.diagnosticPrefix("source_prepared").contains("private-value"))
    }

    @Test
    fun diagnosticsAreLimitedToApprovedPlaybackEvents() {
        val source = Path.of(
            "src/main/java/com/chee/videos/feature/tv/TvShortPlaybackDiagnostics.kt",
        ).readText()

        listOf(
            "source_prepared",
            "decoder_initialized",
            "video_input_format",
            "first_frame",
            "player_error",
        ).forEach { event ->
            assertTrue("短视频诊断必须记录 $event", source.contains("\"$event\""))
        }
        assertFalse("短视频诊断不得记录逐帧事件", source.contains("onVideoFrameProcessingOffset"))
        assertFalse("短视频诊断不得记录掉帧轮询", source.contains("onDroppedVideoFrames"))
    }
}
