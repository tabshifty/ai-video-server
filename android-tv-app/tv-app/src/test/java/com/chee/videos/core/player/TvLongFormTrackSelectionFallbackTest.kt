package com.chee.videos.core.player

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

/**
 * Type-only fallback tests for [resolveLongFormTrackByLanguage]：当 preference 的 language
 * 字段为空但 type 不空时，应当退化为按 type（"default" / "forced" / "commentary"）匹配第一条
 * 同 type 的 track，而不是直接返回 null。
 */
class TvLongFormTrackSelectionFallbackTest {

    @Test
    fun typeOnlyDefault_matchesFirstTrackOfSameType() {
        val tracks = listOf(
            TvLongFormVlcTrack(id = 10, name = "Stereo", language = "", type = "commentary"),
            TvLongFormVlcTrack(id = 11, name = "Stereo", language = "", type = "default"),
            TvLongFormVlcTrack(id = 12, name = "Stereo", language = "", type = "default"),
        )
        val match = resolveLongFormTrackByLanguage(
            tracks = tracks,
            preference = TvLongFormLanguagePreference(language = "", type = "default"),
        )
        assertEquals(11, match?.id)
    }

    @Test
    fun typeOnlyForced_matchesByNameWhenTypeFieldMissing() {
        val tracks = listOf(
            TvLongFormVlcTrack(id = 20, name = "Track 1", language = "", type = ""),
            TvLongFormVlcTrack(id = 21, name = "Forced subtitle", language = "", type = ""),
        )
        val match = resolveLongFormTrackByLanguage(
            tracks = tracks,
            preference = TvLongFormLanguagePreference(language = "", type = "forced"),
        )
        assertEquals(21, match?.id)
    }

    @Test
    fun typeOnlyButNoMatch_returnsNull() {
        val tracks = listOf(
            TvLongFormVlcTrack(id = 30, name = "Track 1", language = "", type = ""),
            TvLongFormVlcTrack(id = 31, name = "Track 2", language = "", type = ""),
        )
        val match = resolveLongFormTrackByLanguage(
            tracks = tracks,
            preference = TvLongFormLanguagePreference(language = "", type = "default"),
        )
        assertNull(match)
    }

    @Test
    fun bothBlank_stillReturnsNull() {
        val match = resolveLongFormTrackByLanguage(
            tracks = listOf(TvLongFormVlcTrack(id = 40, name = "Stereo", language = "en", type = "default")),
            preference = TvLongFormLanguagePreference(language = "", type = ""),
        )
        assertNull(match)
    }

    @Test
    fun nullLanguageWithType_alsoFallsBack() {
        val tracks = listOf(
            TvLongFormVlcTrack(id = 50, name = "Stereo", language = null, type = "default"),
        )
        val match = resolveLongFormTrackByLanguage(
            tracks = tracks,
            preference = TvLongFormLanguagePreference(language = null, type = "default"),
        )
        assertEquals(50, match?.id)
    }

    @Test
    fun typeOnlyDoesNotInterfereWithLanguageMatching() {
        // 同时给 language 和 type，preference 优先按 language 走
        val tracks = listOf(
            TvLongFormVlcTrack(id = 60, name = "Stereo", language = "ja", type = "default"),
            TvLongFormVlcTrack(id = 61, name = "Stereo", language = "en", type = "default"),
        )
        val match = resolveLongFormTrackByLanguage(
            tracks = tracks,
            preference = TvLongFormLanguagePreference(language = "en", type = "default"),
        )
        assertEquals(61, match?.id)
    }

    @Test
    fun preexistingLanguageMatchPathUnchanged() {
        // 已有 happy path 仍然工作
        val tracks = listOf(
            TvLongFormVlcTrack(id = 70, name = "English", language = "en", type = ""),
        )
        val match = resolveLongFormTrackByLanguage(
            tracks = tracks,
            preference = TvLongFormLanguagePreference(language = "en", type = ""),
        )
        assertEquals(70, match?.id)
    }
}
