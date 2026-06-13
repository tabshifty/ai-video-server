package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvDolbyVisionDiagnosticsTest {
    @Test
    fun media3Route_usesShortDiagnosticMessage() {
        val message = buildTvDolbyVisionDiagnosticMessage(
            route = TvPlaybackRoute(
                kind = TvPlaybackRouteKind.MEDIA3_DOLBY_VISION,
                blockMessage = null,
            ),
            displayCapability = supportedDisplayCapability(),
            playbackUrl = "https://example.test/video.mp4",
            media3Available = true,
        )

        assertEquals(
            "当前路径：专用播放失败页\n当前原因：杜比视界专用播放链路\n兼容状态：显示支持杜比视界，播放源可用，专用链路可用",
            message,
        )
    }

    @Test
    fun blockedRoute_includesFailureSummary() {
        val message = buildTvDolbyVisionDiagnosticMessage(
            route = TvPlaybackRoute(
                kind = TvPlaybackRouteKind.BLOCKED,
                blockMessage = "当前电视或显示链路未声明支持杜比视界，无法安全播放该视频",
            ),
            displayCapability = unknownDisplayCapability(),
            playbackUrl = "",
            media3Available = false,
            failureMessage = "播放器初始化失败",
        )

        assertEquals(
            "当前路径：播放阻断页\n当前原因：当前电视或显示链路未声明支持杜比视界，无法安全播放该视频\n兼容状态：显示状态未知，播放源不可用，专用链路不可用\n错误摘要：播放器初始化失败",
            message,
        )
    }

    @Test
    fun diagnosticsAvailable_onlyForDvBlockedOrMedia3Routes() {
        assertTrue(
            isTvDolbyVisionDiagnosticsAvailable(
                route = TvPlaybackRoute(
                    kind = TvPlaybackRouteKind.BLOCKED,
                    blockMessage = "当前电视或显示链路未声明支持杜比视界，无法安全播放该视频",
                ),
            ),
        )
        assertTrue(
            isTvDolbyVisionDiagnosticsAvailable(
                route = TvPlaybackRoute(kind = TvPlaybackRouteKind.MEDIA3_DOLBY_VISION, blockMessage = null),
            ),
        )
        assertFalse(
            isTvDolbyVisionDiagnosticsAvailable(
                route = TvPlaybackRoute(kind = TvPlaybackRouteKind.BLOCKED, blockMessage = "播放兼容信息不完整，暂不能确认安全播放"),
            ),
        )
        assertFalse(
            isTvDolbyVisionDiagnosticsAvailable(
                route = TvPlaybackRoute(kind = TvPlaybackRouteKind.VLC, blockMessage = null),
            ),
        )
    }
}

private fun supportedDisplayCapability(): DolbyVisionDisplayCapability =
    DolbyVisionDisplayCapability(
        status = DolbyVisionDisplayCapabilityStatus.SUPPORTED,
        reason = DolbyVisionDisplayCapabilityReason.DOLBY_VISION_PRESENT,
        supportedHdrTypeNames = listOf("DOLBY_VISION"),
    )

private fun unknownDisplayCapability(): DolbyVisionDisplayCapability =
    DolbyVisionDisplayCapability(
        status = DolbyVisionDisplayCapabilityStatus.UNKNOWN,
        reason = DolbyVisionDisplayCapabilityReason.NO_DISPLAY,
        supportedHdrTypeNames = emptyList(),
    )
