package com.chee.videos.core.data

import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import com.chee.videos.core.model.ShortPlaybackMode
import com.google.gson.Gson
import java.io.File
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.TestScope
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.Assert.assertEquals
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class AppPreferencesStoreTest {
    @Test
    fun shortPlaybackMode_defaultsToLoopOne_andCanPersist() = runTest {
        withMainDispatcher {
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
    }
}

@OptIn(ExperimentalCoroutinesApi::class)
private suspend fun TestScope.withMainDispatcher(block: suspend TestScope.() -> Unit) {
    Dispatchers.setMain(StandardTestDispatcher(testScheduler))
    try {
        block()
    } finally {
        Dispatchers.resetMain()
    }
}
