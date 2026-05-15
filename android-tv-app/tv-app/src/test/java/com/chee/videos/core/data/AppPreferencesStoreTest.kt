package com.chee.videos.core.data

import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import com.chee.videos.core.model.ShortPlaybackMode
import com.chee.videos.core.testing.MainDispatcherRule
import com.google.gson.Gson
import java.io.File
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Rule
import org.junit.Test

class AppPreferencesStoreTest {
    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    @Test
    fun shortPlaybackMode_defaultsToLoopOne_andCanPersist() = runTest {
        val dataStore = PreferenceDataStoreFactory.create(
            scope = backgroundScope,
            produceFile = {
                File.createTempFile("app-preferences-store", ".preferences_pb").apply {
                    deleteOnExit()
                }
            },
        )
        val store = AppPreferencesStore(
            dataStore = dataStore,
            gson = Gson(),
        )

        assertEquals(ShortPlaybackMode.LOOP_ONE, store.shortPlaybackModeFlow.first())
        store.saveShortPlaybackMode(ShortPlaybackMode.AUTO_NEXT)
        assertEquals(ShortPlaybackMode.AUTO_NEXT, store.shortPlaybackModeFlow.first())
    }

    @Test
    fun tvSubtitlePreference_persistsPerVideoId() = runTest {
        val dataStore = PreferenceDataStoreFactory.create(
            scope = backgroundScope,
            produceFile = {
                File.createTempFile("app-preferences-store", ".preferences_pb").apply {
                    deleteOnExit()
                }
            },
        )
        val store = AppPreferencesStore(
            dataStore = dataStore,
            gson = Gson(),
        )

        assertNull(store.readTvSubtitlePreference("video-1"))

        store.saveTvSubtitlePreference("video-1", "subtitle-zh")
        store.saveTvSubtitlePreference("video-2", "")

        assertEquals("subtitle-zh", store.readTvSubtitlePreference("video-1"))
        assertEquals("", store.readTvSubtitlePreference("video-2"))
        assertNull(store.readTvSubtitlePreference("video-3"))
    }
}
