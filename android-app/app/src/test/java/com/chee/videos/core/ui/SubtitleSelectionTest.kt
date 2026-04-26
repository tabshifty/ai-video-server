package com.chee.videos.core.ui

import com.chee.videos.core.model.SubtitleTrackDto
import androidx.media3.common.C
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
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

    @Test
    fun resolveLongFormPlayerUpdate_clearsWhenNextUrlMissing() {
        val decision = resolveLongFormPlayerUpdate(
            preparedUrl = "https://example.com/video.mp4",
            nextUrl = "",
            preparedSubtitleTrackId = null,
            nextSubtitleTrackId = null,
        )

        assertTrue(decision.shouldClear)
        assertFalse(decision.shouldReplaceSource)
        assertFalse(decision.preservePosition)
    }

    @Test
    fun resolveLongFormPlayerUpdate_replacesAndPreservesPositionForSubtitleOnlyChange() {
        val decision = resolveLongFormPlayerUpdate(
            preparedUrl = "https://example.com/video.mp4",
            nextUrl = "https://example.com/video.mp4",
            preparedSubtitleTrackId = null,
            nextSubtitleTrackId = "sub-1",
        )

        assertFalse(decision.shouldClear)
        assertTrue(decision.shouldReplaceSource)
        assertTrue(decision.preservePosition)
    }

    @Test
    fun resolveLongFormPlayerUpdate_noopWhenTargetUnchanged() {
        val decision = resolveLongFormPlayerUpdate(
            preparedUrl = "https://example.com/video.mp4",
            nextUrl = "https://example.com/video.mp4",
            preparedSubtitleTrackId = "sub-1",
            nextSubtitleTrackId = "sub-1",
        )

        assertFalse(decision.shouldClear)
        assertFalse(decision.shouldReplaceSource)
        assertFalse(decision.preservePosition)
    }

    @Test
    fun resolvePlaybackAssetUrl_buildsAbsoluteSubtitleUrlForRelativePathWhenBaseUrlAvailable() {
        val method = Class
            .forName("com.chee.videos.core.ui.LongFormSubtitleSupportKt")
            .getDeclaredMethod("resolvePlaybackAssetUrl", String::class.java, String::class.java)
            .apply { isAccessible = true }

        val resolvedUrl = method.invoke(
            null,
            "https://example.com",
            "/api/v1/videos/video-1/subtitles/sub-1/file",
        ) as String?

        assertEquals(
            "https://example.com/api/v1/videos/video-1/subtitles/sub-1/file",
            resolvedUrl,
        )
    }

    @Test
    fun resolveSubtitleSelectionFlags_marksExplicitlySelectedEmbeddedTrackAsDefault() {
        val method = Class
            .forName("com.chee.videos.core.ui.LongFormSubtitleSupportKt")
            .getDeclaredMethod("resolveSubtitleSelectionFlags", SubtitleTrackDto::class.java)
            .apply { isAccessible = true }

        val selectionFlags = method.invoke(
            null,
            SubtitleTrackDto(
                id = "embedded-2",
                sourceType = "embedded",
                languageCode = "zh-CN",
                label = "简中（内嵌）",
                format = "vtt",
                url = "/api/v1/videos/video-1/subtitles/embedded-2/file",
                mimeType = "text/vtt",
                isDefault = false,
                available = true,
            ),
        ) as Int

        assertEquals(C.SELECTION_FLAG_DEFAULT, selectionFlags)
    }
}
