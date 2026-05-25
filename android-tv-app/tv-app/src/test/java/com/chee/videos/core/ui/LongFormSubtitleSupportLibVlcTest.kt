package com.chee.videos.core.ui

import com.chee.videos.core.model.SubtitleTrackDto
import com.chee.videos.core.model.TvTrackPreference
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class LongFormSubtitleSupportLibVlcTest {
    @Test
    fun resolveSelectedSubtitleTrackByLanguage_prefersMatchingAvailableTrack() {
        val selected = resolveSelectedSubtitleTrackByLanguage(
            tracks = listOf(
                subtitleTrack(id = "sub-en", languageCode = "eng", label = "English"),
                subtitleTrack(id = "sub-zh", languageCode = "zh-CN", label = "简中"),
            ),
            preferredLanguage = "zh_cn",
        )

        assertEquals("sub-zh", selected?.id)
    }

    @Test
    fun resolveSelectedSubtitleTrackByLanguage_ignoresUnavailableTracks() {
        val selected = resolveSelectedSubtitleTrackByLanguage(
            tracks = listOf(subtitleTrack(id = "sub-zh", languageCode = "zh", available = false)),
            preferredLanguage = "zh",
        )

        assertNull(selected)
    }

    @Test
    fun resolveSelectedSubtitleTrackByPreference_prefersMatchingTypeWithinLanguage() {
        val selected = resolveSelectedSubtitleTrackByPreference(
            tracks = listOf(
                subtitleTrack(id = "sub-ja-main", languageCode = "ja", label = "日语"),
                subtitleTrack(id = "sub-ja-commentary", languageCode = "jpn", label = "Japanese Commentary"),
            ),
            preference = TvTrackPreference(language = "ja", type = "commentary"),
        )

        assertEquals("sub-ja-commentary", selected?.id)
    }

    @Test
    fun buildSubtitleTrackPreference_usesLanguageAndTypeNotTrackId() {
        val preference = buildSubtitleTrackPreference(
            subtitleTrack(
                id = "runtime-sub-id",
                languageCode = "zh_CN",
                label = "简中强制字幕",
            ),
        )

        assertEquals(TvTrackPreference(language = "zh-cn", type = "forced"), preference)
    }

    private fun subtitleTrack(
        id: String,
        languageCode: String,
        label: String = "",
        available: Boolean = true,
    ) = SubtitleTrackDto(
        id = id,
        sourceType = "uploaded",
        languageCode = languageCode,
        label = label,
        format = "ass",
        url = "/api/v1/videos/v1/subtitles/$id/file",
        mimeType = "text/x-ssa",
        available = available,
    )
}
