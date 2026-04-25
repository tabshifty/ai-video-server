package com.chee.videos.core.ui

import com.chee.videos.core.model.SubtitleTrackDto
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class SubtitleSelectionTest {
    @Test
    fun resolveInitialSubtitleTrackId_prefersUploadedDefault() {
        val trackId = resolveInitialSubtitleTrackId(
            listOf(
                SubtitleTrackDto(
                    id = "embedded-1",
                    sourceType = "embedded",
                    languageCode = "zh",
                    label = "内嵌中文字幕",
                    format = "vtt",
                    url = "/api/v1/videos/v1/subtitles/embedded-1/file",
                    mimeType = "text/vtt",
                    isDefault = true,
                ),
                SubtitleTrackDto(
                    id = "uploaded-1",
                    sourceType = "uploaded",
                    languageCode = "zh-CN",
                    label = "外挂简中",
                    format = "srt",
                    url = "/api/v1/videos/v1/subtitles/uploaded-1/file",
                    mimeType = "application/x-subrip",
                    isDefault = true,
                ),
            ),
        )

        assertEquals("uploaded-1", trackId)
    }

    @Test
    fun resolveInitialSubtitleTrackId_fallsBackToEmbeddedDefault() {
        val trackId = resolveInitialSubtitleTrackId(
            listOf(
                SubtitleTrackDto(
                    id = "embedded-1",
                    sourceType = "embedded",
                    languageCode = "en",
                    label = "English",
                    format = "vtt",
                    url = "/api/v1/videos/v1/subtitles/embedded-1/file",
                    mimeType = "text/vtt",
                    isDefault = true,
                ),
            ),
        )

        assertEquals("embedded-1", trackId)
    }

    @Test
    fun resolveInitialSubtitleTrackId_returnsNullWhenNoDefaultTrack() {
        val trackId = resolveInitialSubtitleTrackId(
            listOf(
                SubtitleTrackDto(
                    id = "uploaded-1",
                    sourceType = "uploaded",
                    languageCode = "ja",
                    label = "日语",
                    format = "vtt",
                    url = "/api/v1/videos/v1/subtitles/uploaded-1/file",
                    mimeType = "text/vtt",
                ),
            ),
        )

        assertNull(trackId)
    }

    @Test
    fun resolveSubtitleSelectionOnTrackLoad_keepsDefaultNullAfterPlaybackStarted() {
        val trackId = resolveSubtitleSelectionOnTrackLoad(
            currentSelection = null,
            tracks = listOf(
                SubtitleTrackDto(
                    id = "uploaded-1",
                    sourceType = "uploaded",
                    languageCode = "zh-CN",
                    label = "外挂简中",
                    format = "srt",
                    url = "/api/v1/videos/v1/subtitles/uploaded-1/file",
                    mimeType = "application/x-subrip",
                    isDefault = true,
                    available = true,
                ),
            ),
            hasStartedPlayback = true,
        )

        assertNull(trackId)
    }

    @Test
    fun resolveSubtitleSelectionOnTrackLoad_appliesDefaultBeforePlaybackStarts() {
        val trackId = resolveSubtitleSelectionOnTrackLoad(
            currentSelection = null,
            tracks = listOf(
                SubtitleTrackDto(
                    id = "uploaded-1",
                    sourceType = "uploaded",
                    languageCode = "zh-CN",
                    label = "外挂简中",
                    format = "srt",
                    url = "/api/v1/videos/v1/subtitles/uploaded-1/file",
                    mimeType = "application/x-subrip",
                    isDefault = true,
                    available = true,
                ),
            ),
            hasStartedPlayback = false,
        )

        assertEquals("uploaded-1", trackId)
    }
}
