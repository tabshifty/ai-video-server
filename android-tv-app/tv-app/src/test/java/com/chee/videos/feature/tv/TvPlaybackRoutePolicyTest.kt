package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Test

class TvPlaybackRoutePolicyTest {
    @Test
    fun missingPlaybackCompatibility_usesVlcForHistoricalVideos() {
        val route = resolveTvPlaybackRoute(
            metadata = emptyMap(),
            displayCapability = supportedDisplayCapability(),
            playbackUrl = "https://example.test/video.mp4",
            media3Available = true,
        )

        assertEquals(TvPlaybackRouteKind.VLC, route.kind)
        assertEquals(null, route.blockMessage)
    }

    @Test
    fun oldPlaybackCompatibilityVersion_usesVlcForHistoricalVideos() {
        val route = resolveTvPlaybackRoute(
            metadata = mapOf(
                "playback_compat" to mapOf(
                    "version" to 0,
                    "status" to "ok",
                ),
            ),
            displayCapability = supportedDisplayCapability(),
            playbackUrl = "https://example.test/video.mp4",
            media3Available = true,
        )

        assertEquals(TvPlaybackRouteKind.VLC, route.kind)
        assertEquals(null, route.blockMessage)
    }

    @Test
    fun sourceDolbyVisionButOutputIsNotDolbyVision_blocksTranscodedOutput() {
        val route = resolveTvPlaybackRoute(
            metadata = playbackCompat(sourceDolbyVision = true, outputDolbyVision = false),
            displayCapability = supportedDisplayCapability(),
            playbackUrl = "https://example.test/video.mp4",
            media3Available = true,
        )

        assertEquals(TvPlaybackRouteKind.BLOCKED, route.kind)
        assertEquals("该视频来源为杜比视界，当前压缩结果可能无法安全播放", route.blockMessage)
    }

    @Test
    fun sourceDolbyVisionTrustedToneMappedSdrOutput_usesVlcRoute() {
        val route = resolveTvPlaybackRoute(
            metadata = playbackCompat(
                sourceDolbyVision = true,
                outputDolbyVision = false,
                trustedToneMappedSdr = true,
            ),
            displayCapability = unsupportedDisplayCapability(),
            playbackUrl = "https://example.test/video.mp4",
            media3Available = false,
        )

        assertEquals(TvPlaybackRouteKind.VLC, route.kind)
        assertEquals(null, route.blockMessage)
    }

    @Test
    fun outputDolbyVisionWithSupportedDisplayAndMedia3_usesMedia3DedicatedRoute() {
        val route = resolveTvPlaybackRoute(
            metadata = playbackCompat(sourceDolbyVision = true, outputDolbyVision = true),
            displayCapability = supportedDisplayCapability(),
            playbackUrl = "https://example.test/video.mp4",
            media3Available = true,
        )

        assertEquals(TvPlaybackRouteKind.MEDIA3_DOLBY_VISION, route.kind)
        assertEquals(null, route.blockMessage)
    }

    @Test
    fun outputDolbyVisionWithUnsupportedDisplay_blocksWithDisplayUnsupportedMessage() {
        val route = resolveTvPlaybackRoute(
            metadata = playbackCompat(sourceDolbyVision = true, outputDolbyVision = true),
            displayCapability = unsupportedDisplayCapability(),
            playbackUrl = "https://example.test/video.mp4",
            media3Available = true,
        )

        assertEquals(TvPlaybackRouteKind.BLOCKED, route.kind)
        assertEquals("当前电视或显示链路未声明支持杜比视界，无法安全播放该视频", route.blockMessage)
    }

    @Test
    fun outputDolbyVisionWithUnknownDisplay_blocksWithDisplayUnknownMessage() {
        val route = resolveTvPlaybackRoute(
            metadata = playbackCompat(sourceDolbyVision = true, outputDolbyVision = true),
            displayCapability = unknownDisplayCapability(),
            playbackUrl = "https://example.test/video.mp4",
            media3Available = true,
        )

        assertEquals(TvPlaybackRouteKind.BLOCKED, route.kind)
        assertEquals("无法确认当前电视或显示链路是否支持杜比视界，暂不能安全播放", route.blockMessage)
    }

    @Test
    fun outputDolbyVisionWithoutMedia3_blocksDedicatedRouteUnavailable() {
        val route = resolveTvPlaybackRoute(
            metadata = playbackCompat(sourceDolbyVision = true, outputDolbyVision = true),
            displayCapability = supportedDisplayCapability(),
            playbackUrl = "https://example.test/video.mp4",
            media3Available = false,
        )

        assertEquals(TvPlaybackRouteKind.BLOCKED, route.kind)
        assertEquals("杜比视界专用播放链路暂不可用，无法安全播放该视频", route.blockMessage)
    }

    @Test
    fun outputDolbyVisionWithoutPlaybackUrl_blocksDedicatedRouteUnavailable() {
        val route = resolveTvPlaybackRoute(
            metadata = playbackCompat(sourceDolbyVision = true, outputDolbyVision = true),
            displayCapability = supportedDisplayCapability(),
            playbackUrl = "",
            media3Available = true,
        )

        assertEquals(TvPlaybackRouteKind.BLOCKED, route.kind)
        assertEquals("杜比视界专用播放链路暂不可用，无法安全播放该视频", route.blockMessage)
    }

    @Test
    fun incompleteCompatibility_blocksAsIncomplete() {
        val route = resolveTvPlaybackRoute(
            metadata = mapOf(
                "playback_compat" to mapOf(
                    "version" to 1,
                    "status" to "probe_failed",
                ),
            ),
            displayCapability = supportedDisplayCapability(),
            playbackUrl = "https://example.test/video.mp4",
            media3Available = true,
        )

        assertEquals(TvPlaybackRouteKind.BLOCKED, route.kind)
        assertEquals("播放兼容信息不完整，暂不能确认安全播放", route.blockMessage)
    }

    @Test
    fun futureCompatibilityVersion_blocksAsIncomplete() {
        val route = resolveTvPlaybackRoute(
            metadata = mapOf(
                "playback_compat" to mapOf(
                    "version" to 2,
                    "status" to "ok",
                    "source" to mapOf("dolby_vision" to false),
                    "output" to mapOf("dolby_vision" to false),
                ),
            ),
            displayCapability = supportedDisplayCapability(),
            playbackUrl = "https://example.test/video.mp4",
            media3Available = true,
        )

        assertEquals(TvPlaybackRouteKind.BLOCKED, route.kind)
        assertEquals("播放兼容信息不完整，暂不能确认安全播放", route.blockMessage)
    }
}

private fun playbackCompat(
    sourceDolbyVision: Boolean,
    outputDolbyVision: Boolean,
    trustedToneMappedSdr: Boolean = false,
): Map<String, Any?> =
    mapOf(
        "playback_compat" to buildMap<String, Any?> {
            put("version", 1)
            put("status", "ok")
            put("source", mapOf("dolby_vision" to sourceDolbyVision))
            put("output", mapOf("dolby_vision" to outputDolbyVision))
            if (trustedToneMappedSdr) {
                put("trusted_compat_output", "dv_sdr_bt709")
                put("tone_mapped_sdr", true)
            }
        },
    )

private fun supportedDisplayCapability(): DolbyVisionDisplayCapability =
    DolbyVisionDisplayCapability(
        status = DolbyVisionDisplayCapabilityStatus.SUPPORTED,
        reason = DolbyVisionDisplayCapabilityReason.DOLBY_VISION_PRESENT,
        supportedHdrTypeNames = listOf("DOLBY_VISION"),
    )

private fun unsupportedDisplayCapability(): DolbyVisionDisplayCapability =
    DolbyVisionDisplayCapability(
        status = DolbyVisionDisplayCapabilityStatus.UNSUPPORTED,
        reason = DolbyVisionDisplayCapabilityReason.DOLBY_VISION_MISSING,
        supportedHdrTypeNames = listOf("HDR10"),
    )

private fun unknownDisplayCapability(): DolbyVisionDisplayCapability =
    DolbyVisionDisplayCapability(
        status = DolbyVisionDisplayCapabilityStatus.UNKNOWN,
        reason = DolbyVisionDisplayCapabilityReason.NO_DISPLAY,
        supportedHdrTypeNames = emptyList(),
    )
