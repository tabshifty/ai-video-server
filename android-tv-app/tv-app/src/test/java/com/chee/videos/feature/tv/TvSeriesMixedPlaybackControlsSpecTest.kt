package com.chee.videos.feature.tv

import java.nio.file.Path
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSeriesMixedPlaybackControlsSpecTest {
    @Test
    fun seriesExoPlayerRouteUsesSharedCoreControls() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").toFile().readText()
        val media3Branch = source
            .substringAfter("if (isMedia3Route)")
            .substringBefore("} else if (uiState.playbackPreparing)")

        assertTrue(
            "剧集 ExoPlayer 分支必须复用电视剧核心播放控制层，不能只渲染裸 PlayerView",
            media3Branch.contains("TvSeriesCorePlaybackOverlay("),
        )
        assertTrue(
            "剧集 ExoPlayer 分支必须展示字幕和音轨入口",
            media3Branch.contains("showTrackActions = true") &&
                media3Branch.contains("TvMedia3TrackPickerLayer("),
        )
        assertTrue(
            "剧集 ExoPlayer 分支必须把切集中心状态挂给核心控制层",
            media3Branch.contains("episodeSwitchState = uiState.episodeSwitchState"),
        )
    }

    @Test
    fun media3FailureKeepsEpisodeSelectionButNoForcePlay() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").toFile().readText()
        val media3Branch = source
            .substringAfter("if (isMedia3Route)")
            .substringBefore("} else if (uiState.playbackPreparing)")

        assertTrue(
            "Media3 失败态应允许打开选集轨切到其它分集",
            media3Branch.contains("tertiaryActionLabel = if (episodeRailItems.isNotEmpty()) \"选集\" else null") &&
                media3Branch.contains("openEpisodeRailRequestKey += 1"),
        )
        assertFalse(
            "Media3 失败态不能提供强行播放或回退 LibVLC 入口",
            media3Branch.contains("强行播放") || media3Branch.contains("继续播放") || media3Branch.contains("LibVLC"),
        )
        assertTrue(
            "准备中按 BACK 应优先取消切集，失败态按 BACK 应先关闭失败中心态",
            source.contains("viewModel.cancelEpisodeSwitch()") &&
                source.contains("viewModel.clearEpisodeSwitchFeedback()"),
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
    fun seriesScreenNoLongerKeepsMixedVlcAndMedia3SnapshotRouting() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").toFile().readText()

        assertTrue(source.contains("reportTvSeriesMedia3History("))
        assertFalse(source.contains("shouldUseMedia3SnapshotForTvSeriesHistory"))
        assertFalse(source.contains("lastHistoryUsesMedia3Snapshot"))
        assertFalse(source.contains("isVlcRoute"))
        assertFalse(source.contains("LongFormVideoPlayer("))
        assertFalse(source.contains("org.videolan.libvlc"))
    }

    @Test
    fun seriesScreenUsesSingleLongFormMedia3Component() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").toFile().readText()

        assertTrue(source.contains("TvLongFormMedia3Player("))
        assertFalse(
            "剧集页不应再保留 DV 专用 Media3 组件名",
            source.contains("TvDolbyVisionMedia3Player("),
        )
    }
}
