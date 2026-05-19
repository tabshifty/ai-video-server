package com.chee.videos.core.ui

import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

class AudioTrackSelectionTest {
    @Test
    fun buildAudioTrackPickerItems_includesAutoOptionAndMarksSelectedTrack() {
        val audioTracks = listOf(
            LongFormAudioTrack(
                id = "audio-0-0-eng-6",
                label = "英语 5.1",
                detail = "5.1",
                groupIndex = 0,
                trackIndex = 0,
                selected = false,
            ),
            LongFormAudioTrack(
                id = "audio-1-0-jpn-2",
                label = "日语 2.0",
                detail = "2.0",
                groupIndex = 1,
                trackIndex = 0,
                selected = true,
            ),
        )

        val items = buildAudioTrackPickerItems(audioTracks, selectedAudioTrackId = "audio-1-0-jpn-2")

        assertEquals(3, items.size)
        assertNull(items[0].trackId)
        assertEquals("自动选择", items[0].label)
        assertEquals("跟随视频默认音轨", items[0].detail)
        assertFalse(items[0].selected)
        assertEquals("audio-1-0-jpn-2", items[2].trackId)
        assertEquals("2.0", items[2].detail)
        assertTrue(items[2].selected)
    }

    @Test
    fun buildAudioTrackPickerItems_usesCurrentlyPlayingTrackAsSelectedAfterRuntimeSwitch() {
        val audioTracks = listOf(
            LongFormAudioTrack(
                id = "audio-0-0-eng-6",
                label = "英语",
                detail = "5.1",
                groupIndex = 0,
                trackIndex = 0,
                selected = false,
            ),
            LongFormAudioTrack(
                id = "audio-1-0-jpn-2",
                label = "日语",
                detail = "2.0",
                groupIndex = 1,
                trackIndex = 0,
                selected = true,
            ),
        )

        val items = buildAudioTrackPickerItems(audioTracks, selectedAudioTrackId = "audio-0-0-eng-6")

        assertFalse(items.first { it.trackId == "audio-0-0-eng-6" }.selected)
        assertTrue(items.first { it.trackId == "audio-1-0-jpn-2" }.selected)
    }

    @Test
    fun resolveAudioSelectionOnTrackLoad_keepsValidSavedPreference() {
        val selection = resolveAudioSelectionOnTrackLoad(
            storedSelection = "audio-1",
            tracks = listOf(
                LongFormAudioTrack(
                    id = "audio-1",
                    label = "英语 5.1",
                    detail = "5.1",
                    groupIndex = 0,
                    trackIndex = 0,
                    selected = false,
                ),
            ),
        )

        assertEquals("audio-1", selection)
    }

    @Test
    fun resolveAudioSelectionOnTrackLoad_fallsBackToAutoWhenSavedPreferenceIsMissing() {
        val selection = resolveAudioSelectionOnTrackLoad(
            storedSelection = "old-audio",
            tracks = listOf(
                LongFormAudioTrack(
                    id = "audio-1",
                    label = "英语 5.1",
                    detail = "5.1",
                    groupIndex = 0,
                    trackIndex = 0,
                    selected = false,
                ),
            ),
        )

        assertNull(selection)
    }

    @Test
    fun buildAudioTrackPickerItems_marksAutoSelectedWhenNoTrackSelected() {
        val items = buildAudioTrackPickerItems(
            tracks = listOf(
                LongFormAudioTrack(
                    id = "audio-0-0-eng-6",
                    label = "英语 5.1",
                    detail = "5.1",
                    groupIndex = 0,
                    trackIndex = 0,
                    selected = true,
                ),
            ),
            selectedAudioTrackId = null,
        )

        assertTrue(items[0].selected)
        assertFalse(items[1].selected)
    }
}
