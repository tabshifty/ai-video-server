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
    fun oldPlaybackCompatibilityVersion_allowsHistoricalVideosOnNormalPath() {
        val decision = resolveTvPlaybackCompatibilityDecision(
            mapOf(
                "playback_compat" to mapOf(
                    "version" to 0,
                    "status" to "ok",
                ),
            ),
        )

        assertTrue(decision.allowed)
        assertEquals(null, decision.blockMessage)
    }

    @Test
    fun futurePlaybackCompatibilityVersion_blocksAsIncomplete() {
        val decision = resolveTvPlaybackCompatibilityDecision(
            mapOf(
                "playback_compat" to mapOf(
                    "version" to 2,
                    "status" to "ok",
                    "source" to mapOf("dolby_vision" to false),
                    "output" to mapOf("dolby_vision" to false),
                ),
            ),
        )

        assertFalse(decision.allowed)
        assertEquals("播放兼容信息不完整，暂不能确认安全播放", decision.blockMessage)
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
        assertEquals("播放兼容信息不完整，暂不能确认安全播放", decision.blockMessage)
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
        assertEquals("播放兼容信息不完整，暂不能确认安全播放", decision.blockMessage)
    }

    @Test
    fun sourceDolbyVisionButOutputIsNotDolbyVision_allowsPlayback() {
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

        assertTrue(decision.allowed)
        assertEquals(null, decision.blockMessage)
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

    @Test
    fun episodePlayablePolicyKeepsOutputDolbyVisionAsPlaybackCandidate() {
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
                    "status" to "ok",
                    "source" to mapOf("dolby_vision" to true),
                    "output" to mapOf("dolby_vision" to true),
                ),
            ),
        )

        assertTrue(isTvEpisodePlayableForPlayback(episode))
    }

    @Test
    fun episodePlayablePolicyKeepsDolbyVisionSdrOutputAsCandidate() {
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
                    "status" to "ok",
                    "source" to mapOf("dolby_vision" to true),
                    "output" to mapOf("dolby_vision" to false),
                ),
            ),
        )

        assertTrue(isTvEpisodePlayableForPlayback(episode))
    }
}
