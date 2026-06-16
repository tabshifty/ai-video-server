package com.chee.videos.feature.tv

import java.nio.file.Path
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSeriesMixedPlaybackControlsSpecTest {
    @Test
    fun seriesMedia3RouteUsesSharedCoreControls() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").toFile().readText()
        val media3Branch = source
            .substringAfter("} else if (isMedia3Route)")
            .substringBefore("} else if (uiState.playbackPreparing)")

        assertTrue(
            "DV Media3 分支必须复用电视剧核心播放控制层，不能只渲染裸 PlayerView",
            media3Branch.contains("TvSeriesCorePlaybackOverlay("),
        )
        assertTrue(
            "DV Media3 第二阶段必须展示字幕和音轨入口",
            media3Branch.contains("showTrackActions = true") &&
                media3Branch.contains("TvMedia3TrackPickerLayer("),
        )
    }

    @Test
    fun media3FailureKeepsEpisodeSelectionButNoForcePlay() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").toFile().readText()
        val media3Branch = source
            .substringAfter("} else if (isMedia3Route)")
            .substringBefore("} else if (uiState.playbackPreparing)")

        assertTrue(
            "DV Media3 失败态应允许打开选集轨切到其它分集",
            media3Branch.contains("tertiaryActionLabel = if (episodeRailItems.isNotEmpty()) \"选集\" else null") &&
                media3Branch.contains("openEpisodeRailRequestKey += 1"),
        )
        assertFalse(
            "DV Media3 失败态不能提供强行播放或回退 LibVLC 入口",
            media3Branch.contains("强行播放") || media3Branch.contains("继续播放") || media3Branch.contains("LibVLC"),
        )
    }

    @Test
    fun sharedSeriesControlsExposeEpisodeRailButDoNotOwnTracks() {
        val source = Path.of("src/main/java/com/chee/videos/core/ui/TvSeriesCorePlaybackOverlay.kt").toFile().readText()

        assertTrue(source.contains("episodeRailItems: List<TvEpisodeRailItem>"))
        assertTrue(source.contains("onSelectEpisodeRailItem: ((TvEpisodeRailItem) -> Unit)?"))
        assertTrue(source.contains("showTrackActions: Boolean = true"))
        assertFalse(
            "电视剧核心控制层不应持有字幕轨道模型，字幕能力留在 LibVLC 外层/后续阶段",
            source.contains("SubtitleTrackDto"),
        )
        assertFalse(
            "电视剧核心控制层不应持有音轨偏好模型，音轨能力留在 LibVLC 外层/后续阶段",
            source.contains("TvTrackPreference"),
        )
    }

    @Test
    fun historySnapshotRouteUsesTargetEpisodeRoute() {
        assertFalse(
            shouldUseMedia3SnapshotForTvSeriesHistory(
                currentVideoId = "dv-next",
                targetVideoId = "vlc-prev",
                currentRouteIsMedia3 = true,
                lastHistoryVideoId = "vlc-prev",
                lastHistoryUsesMedia3Snapshot = false,
            ),
        )
        assertTrue(
            shouldUseMedia3SnapshotForTvSeriesHistory(
                currentVideoId = "vlc-next",
                targetVideoId = "dv-prev",
                currentRouteIsMedia3 = false,
                lastHistoryVideoId = "dv-prev",
                lastHistoryUsesMedia3Snapshot = true,
            ),
        )
        assertTrue(
            shouldUseMedia3SnapshotForTvSeriesHistory(
                currentVideoId = "dv-current",
                targetVideoId = "dv-current",
                currentRouteIsMedia3 = true,
                lastHistoryVideoId = "old",
                lastHistoryUsesMedia3Snapshot = false,
            ),
        )
    }
}
