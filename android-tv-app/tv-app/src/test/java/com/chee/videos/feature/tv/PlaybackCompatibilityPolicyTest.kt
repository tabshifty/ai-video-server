package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class PlaybackCompatibilityPolicyTest {
    @Test
    fun missingPlaybackCompatibilityMetadata_allowsHistoricalVideos() {
        val decision = resolveTvPlaybackCompatibilityDecision(metadata = emptyMap())

        assertTrue(decision.allowed)
        assertEquals(null, decision.blockMessage)
    }

    @Test
    fun probeFailed_blocksPlayback() {
        val decision = resolveTvPlaybackCompatibilityDecision(
            mapOf(
                "playback_compat" to mapOf(
                    "version" to 1,
                    "status" to "probe_failed",
                ),
            ),
        )

        assertFalse(decision.allowed)
        assertEquals("该视频播放兼容性未确认，当前 TV 端暂不自动播放", decision.blockMessage)
    }

    @Test
    fun incompleteOkPayload_blocksPlayback() {
        val decision = resolveTvPlaybackCompatibilityDecision(
            mapOf(
                "playback_compat" to mapOf(
                    "version" to 1,
                    "status" to "ok",
                ),
            ),
        )

        assertFalse(decision.allowed)
        assertEquals("该视频播放兼容性未确认，当前 TV 端暂不自动播放", decision.blockMessage)
    }

    @Test
    fun sourceDolbyVisionButOutputIsNotDolbyVision_blocksTranscodedOutput() {
        val decision = resolveTvPlaybackCompatibilityDecision(
            mapOf(
                "playback_compat" to mapOf(
                    "version" to 1.0,
                    "status" to "ok",
                    "source" to mapOf("dolby_vision" to true),
                    "output" to mapOf("dolby_vision" to false),
                ),
            ),
        )

        assertFalse(decision.allowed)
        assertEquals("该视频来源为杜比视界，当前压缩结果可能无法安全播放", decision.blockMessage)
    }

    @Test
    fun outputDolbyVision_blocksUntilDedicatedSystemPlaybackExists() {
        val decision = resolveTvPlaybackCompatibilityDecision(
            mapOf(
                "playback_compat" to mapOf(
                    "version" to "1",
                    "status" to "ok",
                    "source" to mapOf("dolby_vision" to "true"),
                    "output" to mapOf("dolby_vision" to true),
                ),
            ),
        )

        assertFalse(decision.allowed)
        assertEquals("该视频可能为杜比视界，当前设备或播放链路不支持安全播放", decision.blockMessage)
    }

    @Test
    fun nonDolbyVisionOkPayload_allowsPlayback() {
        val decision = resolveTvPlaybackCompatibilityDecision(
            mapOf(
                "playback_compat" to mapOf(
                    "version" to 1,
                    "status" to "ok",
                    "source" to mapOf("dolby_vision" to false),
                    "output" to mapOf("dolby_vision" to false),
                ),
            ),
        )

        assertTrue(decision.allowed)
        assertEquals(null, decision.blockMessage)
    }

    @Test
    fun episodePlayablePolicyRequiresPlaybackCompatibilityAllowed() {
        val episode = TvEpisodeUiModel(
            id = "e1",
            number = 1,
            title = "第1集",
            durationLabel = "45 分钟",
            summary = "剧情",
            videoId = "video-1",
            videoStatus = "ready",
            playable = true,
            metadata = mapOf(
                "playback_compat" to mapOf(
                    "version" to 1,
                    "status" to "probe_failed",
                ),
            ),
        )

        assertFalse(isTvEpisodePlayableForPlayback(episode))
    }
}
