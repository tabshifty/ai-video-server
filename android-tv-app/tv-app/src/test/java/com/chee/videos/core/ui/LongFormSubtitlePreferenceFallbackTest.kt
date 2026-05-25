package com.chee.videos.core.ui

import com.chee.videos.core.model.SubtitleTrackDto
import com.chee.videos.core.model.TvTrackPreference
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

/**
 * Type-only fallback tests for [resolveSelectedSubtitleTrackByPreference]：当 preference 的
 * language 字段为空但 type 不空时（典型场景：上次选的字幕在服务端 isDefault=true 但没有
 * languageCode），按 type 匹配第一条 [subtitlePreferenceType] 同 type 的可用 track。
 */
class LongFormSubtitlePreferenceFallbackTest {

    private fun track(
        id: String,
        languageCode: String = "",
        label: String = "",
        isDefault: Boolean = false,
        available: Boolean = true,
    ) = SubtitleTrackDto(
        id = id,
        languageCode = languageCode,
        label = label,
        isDefault = isDefault,
        available = available,
        url = "https://example.com/$id.ass",
    )

    @Test
    fun typeOnlyDefault_matchesIsDefaultTrack() {
        val tracks = listOf(
            track(id = "t1", label = "中文字幕"),
            track(id = "t2", isDefault = true),
        )
        val match = resolveSelectedSubtitleTrackByPreference(
            tracks = tracks,
            preference = TvTrackPreference(language = "", type = "default"),
        )
        assertEquals("t2", match?.id)
    }

    @Test
    fun typeOnlyForced_matchesByLabel() {
        val tracks = listOf(
            track(id = "s1", label = "中文 forced"),
            track(id = "s2", label = "中文"),
        )
        val match = resolveSelectedSubtitleTrackByPreference(
            tracks = tracks,
            preference = TvTrackPreference(language = "", type = "forced"),
        )
        assertEquals("s1", match?.id)
    }

    @Test
    fun typeOnlyButNoTrackMatches_returnsNull() {
        val tracks = listOf(
            track(id = "x1", label = "中文字幕"),
            track(id = "x2", label = "English"),
        )
        val match = resolveSelectedSubtitleTrackByPreference(
            tracks = tracks,
            preference = TvTrackPreference(language = "", type = "default"),
        )
        assertNull(match)
    }

    @Test
    fun typeOnlyExcludesUnavailableTracks() {
        val tracks = listOf(
            track(id = "u1", isDefault = true, available = false),
            track(id = "u2", isDefault = true, available = true),
        )
        val match = resolveSelectedSubtitleTrackByPreference(
            tracks = tracks,
            preference = TvTrackPreference(language = "", type = "default"),
        )
        assertEquals("u2", match?.id)
    }

    @Test
    fun blankPreferenceStillReturnsNull() {
        val tracks = listOf(track(id = "any", isDefault = true))
        val match = resolveSelectedSubtitleTrackByPreference(
            tracks = tracks,
            preference = TvTrackPreference(language = "", type = ""),
        )
        assertNull(match)
    }

    @Test
    fun nullPreferenceReturnsNull() {
        val tracks = listOf(track(id = "any", isDefault = true))
        val match = resolveSelectedSubtitleTrackByPreference(tracks = tracks, preference = null)
        assertNull(match)
    }

    @Test
    fun preexistingLanguagePathUnchanged() {
        // 既有的 language match 仍然工作
        val tracks = listOf(track(id = "zh", languageCode = "zh"))
        val match = resolveSelectedSubtitleTrackByPreference(
            tracks = tracks,
            preference = TvTrackPreference(language = "zh", type = ""),
        )
        assertEquals("zh", match?.id)
    }

    @Test
    fun languagePathPrefersTypeAmongLanguageMatches() {
        val tracks = listOf(
            track(id = "zh-a", languageCode = "zh", label = "中文 forced"),
            track(id = "zh-b", languageCode = "zh", isDefault = true),
        )
        val match = resolveSelectedSubtitleTrackByPreference(
            tracks = tracks,
            preference = TvTrackPreference(language = "zh", type = "default"),
        )
        assertEquals("zh-b", match?.id)
    }
}
