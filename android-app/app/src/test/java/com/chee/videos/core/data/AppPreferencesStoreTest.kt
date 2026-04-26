package com.chee.videos.core.data

import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import com.google.gson.Gson
import java.io.File
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class AppPreferencesStoreTest {
    @Test
    fun tvSubtitlePreference_persistsPerVideoId() = runTest {
        val dataStore = PreferenceDataStoreFactory.create(
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
