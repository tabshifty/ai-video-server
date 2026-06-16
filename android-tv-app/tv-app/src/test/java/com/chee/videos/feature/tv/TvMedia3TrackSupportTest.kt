package com.chee.videos.feature.tv

import androidx.media3.common.C
import androidx.media3.common.MimeTypes
import java.nio.file.Path
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvMedia3TrackSupportTest {
    @Test
    fun subtitleMimeTypeFallsBackFromFormat() {
        assertEquals(MimeTypes.APPLICATION_SUBRIP, resolveTvMedia3SubtitleMimeType(mimeType = "", format = "srt"))
        assertEquals(MimeTypes.TEXT_VTT, resolveTvMedia3SubtitleMimeType(mimeType = "", format = "webvtt"))
        assertEquals(MimeTypes.TEXT_SSA, resolveTvMedia3SubtitleMimeType(mimeType = "", format = "ass"))
    }

    @Test
    fun subtitleMimeTypeNormalizesCommonServerMimeTypes() {
        assertEquals(MimeTypes.APPLICATION_SUBRIP, resolveTvMedia3SubtitleMimeType(mimeType = "application/x-subrip", format = ""))
        assertEquals(MimeTypes.TEXT_VTT, resolveTvMedia3SubtitleMimeType(mimeType = "text/webvtt", format = ""))
        assertEquals(MimeTypes.TEXT_SSA, resolveTvMedia3SubtitleMimeType(mimeType = "text/x-ass", format = ""))
    }

    @Test
    fun media3TrackSupportSourcePreloadsExternalSubtitlesAndUsesOverridesForSwitching() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvMedia3TrackSupport.kt").toFile().readText()
        val player = Path.of("src/main/java/com/chee/videos/feature/tv/TvDolbyVisionMedia3Player.kt").toFile().readText()

        assertTrue(source.contains("MediaItem.SubtitleConfiguration.Builder"))
        assertTrue(player.contains(".setSubtitleConfigurations(subtitleConfigurations)"))
        assertTrue(source.contains("TrackSelectionOverride"))
        assertTrue(source.contains("setOverrideForType(TrackSelectionOverride"))
        assertTrue(source.contains("setTrackTypeDisabled(C.TRACK_TYPE_TEXT, true)"))
        assertTrue(source.contains("vlcTrackId = buildTvMedia3AudioPreferenceTrackId(groupIndex, trackIndex)"))
        assertFalse(
            "字幕/音轨切换不能通过 setMediaSource 重建播放源实现",
            source.contains("setMediaSource(") || source.contains("prepare()"),
        )
    }

    @Test
    fun media3TrackLoadDoesNotOverwriteStoredAudioPreferenceWithRuntimeDefault() {
        val player = Path.of("src/main/java/com/chee/videos/feature/tv/TvDolbyVisionMedia3Player.kt").toFile().readText()

        assertTrue(player.contains("onAudioTracksChanged: (List<LongFormAudioTrack>) -> Unit"))
        assertFalse(
            "轨道加载时不能把 Media3 默认选中音轨回写为父层选择，否则会覆盖按语言/类型恢复出的用户偏好",
            player.contains("onSelectedAudioTrackChanged") ||
                player.contains("latestOnSelectedAudioTrackChanged") ||
                player.contains("firstOrNull { it.selected }"),
        )
    }

    @Test
    fun textAndAudioTrackTypesStayDistinct() {
        assertEquals(1, C.TRACK_TYPE_AUDIO)
        assertEquals(3, C.TRACK_TYPE_TEXT)
    }

    @Test
    fun media3TrackPickerIsSharedBySingleAndSeriesRoutes() {
        val picker = Path.of("src/main/java/com/chee/videos/feature/tv/TvMedia3TrackPickerLayer.kt").toFile().readText()
        val single = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt").toFile().readText()
        val series = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").toFile().readText()

        assertTrue(picker.contains("internal enum class TvMedia3TrackPickerKind"))
        assertTrue(picker.contains("TvSubtitlePickerDialog("))
        assertTrue(picker.contains("TvAudioTrackPickerDialog("))
        assertTrue(single.contains("TvMedia3TrackPickerLayer("))
        assertTrue(series.contains("TvMedia3TrackPickerLayer("))
        assertFalse("共享弹层不应继续私有在剧集页面里", series.contains("private fun TvMedia3TrackPickerLayer"))
    }

    @Test
    fun dolbyVisionExitToDetailUsesBlackCoverOnSingleAndSeriesMedia3Routes() {
        val cover = Path.of("src/main/java/com/chee/videos/feature/tv/TvDolbyVisionExitToDetailCover.kt").toFile().readText()
        val single = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt").toFile().readText()
        val series = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").toFile().readText()

        assertTrue(cover.contains("background(Color.Black)"))
        assertTrue(cover.contains("TvDolbyVisionExitToDetailCoverDelayMillis"))
        listOf(single, series).forEach { source ->
            assertTrue(source.contains("pendingDvExitToDetail"))
            assertTrue(source.contains("TvDolbyVisionExitToDetailCover()"))
            assertTrue(source.contains("delay(TvDolbyVisionExitToDetailCoverDelayMillis)"))
            assertTrue(source.contains("handlePlaybackBack(isMedia3DolbyVisionRoute = isMedia3Route)"))
        }
    }
}
