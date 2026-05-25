package com.chee.videos.core.player

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class TvLongFormTrackSelectionTest {
    @Test
    fun normalizeLongFormLanguageCode_handlesCommonAliases() {
        assertEquals("zh-cn", normalizeLongFormLanguageCode("zh_CN"))
        assertEquals("ja", normalizeLongFormLanguageCode("jpn"))
        assertEquals("en", normalizeLongFormLanguageCode("eng"))
    }

    @Test
    fun resolveLongFormTrackByLanguage_prefersLanguageThenType() {
        val tracks = listOf(
            TvLongFormVlcTrack(id = 1, name = "English", language = "eng", type = "default"),
            TvLongFormVlcTrack(id = 2, name = "Japanese Commentary", language = "jpn", type = "commentary"),
            TvLongFormVlcTrack(id = 3, name = "Japanese Default", language = "ja", type = "default"),
        )

        val selected = resolveLongFormTrackByLanguage(
            tracks = tracks,
            preference = TvLongFormLanguagePreference(language = "ja", type = "default"),
        )

        assertEquals(3, selected?.id)
    }

    @Test
    fun resolveLongFormTrackByLanguage_returnsNullWhenPreferenceMissing() {
        val selected = resolveLongFormTrackByLanguage(
            tracks = listOf(TvLongFormVlcTrack(id = 1, name = "English", language = "en", type = "default")),
            preference = TvLongFormLanguagePreference(language = null, type = null),
        )

        assertNull(selected)
    }
}
