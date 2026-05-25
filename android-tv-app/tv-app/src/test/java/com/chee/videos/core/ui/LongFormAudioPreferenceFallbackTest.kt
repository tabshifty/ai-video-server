package com.chee.videos.core.ui

import com.chee.videos.core.model.TvTrackPreference
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class LongFormAudioPreferenceFallbackTest {

    private fun track(
        id: String,
        vlcTrackId: Int,
        label: String = id,
        languageCode: String = "",
        preferenceType: String = "",
    ) = LongFormAudioTrack(
        id = id,
        vlcTrackId = vlcTrackId,
        label = label,
        detail = "",
        groupIndex = 0,
        trackIndex = vlcTrackId,
        selected = false,
        languageCode = languageCode,
        preferenceType = preferenceType,
    )

    @Test
    fun currentSelectionWins_whenMatchingTrackExists() {
        val tracks = listOf(
            track("audio-1-stereo-english", 1, languageCode = "en"),
            track("audio-2-stereo", 2, languageCode = ""),
        )
        val resolved = resolveAudioSelectionOnTrackLoad(
            currentSelection = "audio-1-stereo-english",
            storedPreference = TvTrackPreference(language = "ja", type = ""),
            tracks = tracks,
        )
        // currentSelection 命中已存在 track 时直接返回，不再走 preference
        assertEquals("audio-1-stereo-english", resolved)
    }

    @Test
    fun fallsBackToLanguagePreference_whenCurrentSelectionMissing() {
        val tracks = listOf(
            track("a1", 1, languageCode = "en", label = "English"),
            track("a2", 2, languageCode = "ja", label = "Japanese"),
        )
        val resolved = resolveAudioSelectionOnTrackLoad(
            currentSelection = null,
            storedPreference = TvTrackPreference(language = "en", type = ""),
            tracks = tracks,
        )
        assertEquals("a1", resolved)
    }

    @Test
    fun fallsBackToTypeOnly_whenLanguagePreferenceBlank() {
        // 用户上次选了 isDefault=true 但无 language 的音轨
        val tracks = listOf(
            track("a1", 1, preferenceType = ""),
            track("a2", 2, preferenceType = "default"),
        )
        val resolved = resolveAudioSelectionOnTrackLoad(
            currentSelection = null,
            storedPreference = TvTrackPreference(language = "", type = "default"),
            tracks = tracks,
        )
        assertEquals("a2", resolved)
    }

    @Test
    fun blankPreferenceReturnsNull() {
        val tracks = listOf(track("a1", 1))
        val resolved = resolveAudioSelectionOnTrackLoad(
            currentSelection = null,
            storedPreference = TvTrackPreference(language = "", type = ""),
            tracks = tracks,
        )
        assertNull(resolved)
    }

    @Test
    fun nullPreferenceReturnsNull() {
        val tracks = listOf(track("a1", 1))
        val resolved = resolveAudioSelectionOnTrackLoad(
            currentSelection = null,
            storedPreference = null,
            tracks = tracks,
        )
        assertNull(resolved)
    }

    @Test
    fun staleSelectionFallsThroughToPreference() {
        // currentSelection 是上次的 track id，但现在 track 列表已经变了（VLC track id 重排）
        val tracks = listOf(
            track("a1", 1, languageCode = "en"),
        )
        val resolved = resolveAudioSelectionOnTrackLoad(
            currentSelection = "stale-id-that-doesnt-exist",
            storedPreference = TvTrackPreference(language = "en", type = ""),
            tracks = tracks,
        )
        // currentSelection 不匹配 → 退到 preference 解析 → 命中 a1
        assertEquals("a1", resolved)
    }
}
