package com.chee.videos.core.data

import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import com.chee.videos.core.model.ShortPlaybackMode
import com.chee.videos.core.model.TvSubtitlePreferenceMode
import com.chee.videos.core.model.TvTrackPreference
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
    fun tvSeekStepSeconds_defaultsToTen_persistsAllowedValuesAndRejectsInvalidValues() = runTest {
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

        assertEquals(10, store.tvSeekStepSecondsFlow.first())

        store.saveTvSeekStepSeconds(20)
        assertEquals(20, store.tvSeekStepSecondsFlow.first())

        store.saveTvSeekStepSeconds(7)
        assertEquals(10, store.tvSeekStepSecondsFlow.first())
    }

    @Test
    fun tvSubtitlePreference_persistsNormalizedAccountSemanticAndCanClear() = runTest {
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

        assertNull(store.readTvSubtitlePreference())

        store.saveTvSubtitlePreference(
            TvTrackPreference(
                language = " zh-CN ",
                type = "DEFAULT",
                subtitleMode = TvSubtitlePreferenceMode.SPECIFIC.storageValue,
            ),
        )
        assertEquals(
            TvTrackPreference(
                language = "zh-CN",
                type = "default",
                subtitleMode = TvSubtitlePreferenceMode.SPECIFIC.storageValue,
            ),
            store.readTvSubtitlePreference(),
        )

        store.saveTvSubtitlePreference(
            TvTrackPreference(subtitleMode = TvSubtitlePreferenceMode.OFF.storageValue),
        )
        assertEquals(
            TvTrackPreference(subtitleMode = TvSubtitlePreferenceMode.OFF.storageValue),
            store.readTvSubtitlePreference(),
        )

        store.saveTvSubtitlePreference(
            TvTrackPreference(subtitleMode = TvSubtitlePreferenceMode.AUTO.storageValue),
        )
        assertEquals(
            TvTrackPreference(subtitleMode = TvSubtitlePreferenceMode.AUTO.storageValue),
            store.readTvSubtitlePreference(),
        )

        store.saveTvSubtitlePreference(TvTrackPreference())
        assertNull(store.readTvSubtitlePreference())
    }

    @Test
    fun tvAudioPreference_persistsNormalizedAccountSemanticAndCanClear() = runTest {
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

        assertNull(store.readTvAudioPreference())

        store.saveTvAudioPreference(TvTrackPreference(language = "ja", type = "commentary"))

        assertEquals(TvTrackPreference(language = "ja", type = "commentary"), store.readTvAudioPreference())

        store.saveTvAudioPreference(null)
        assertNull(store.readTvAudioPreference())
    }

    @Test
    fun tvTrackPreference_readClearsLegacyTrackIdKeys() = runTest {
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
        val legacySubtitleKey = stringPreferencesKey("tv_subtitle_preferences")
        val legacyAudioKey = stringPreferencesKey("tv_audio_preferences")

        dataStore.edit { prefs ->
            prefs[legacySubtitleKey] = """{"video-1":"subtitle-zh"}"""
            prefs[legacyAudioKey] = """{"video-1":"audio-zh"}"""
        }

        assertNull(store.readTvSubtitlePreference())

        val prefs = dataStore.data.first()
        assertNull(prefs[legacySubtitleKey])
        assertNull(prefs[legacyAudioKey])
    }

    @Test
    fun clearTokens_clearsAccountTrackPreferences() = runTest {
        val dataStore = PreferenceDataStoreFactory.create(
            scope = backgroundScope,
            produceFile = {
                File.createTempFile("app-preferences-store", ".preferences_pb").apply {
                    deleteOnExit()
                }
            },
        )
        val store = AppPreferencesStore(dataStore = dataStore, gson = Gson())
        store.saveTvSubtitlePreference(TvTrackPreference(language = "zh-CN", type = "default"))
        store.saveTvAudioPreference(TvTrackPreference(language = "ja", type = "commentary"))

        store.clearTokens()

        assertNull(store.readTvSubtitlePreference())
        assertNull(store.readTvAudioPreference())
    }
}
