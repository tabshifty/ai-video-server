package com.chee.videos.core.repository

import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.ApiEnvelope
import com.chee.videos.core.model.RefreshPayload
import com.chee.videos.core.model.SessionTokens
import com.chee.videos.core.model.TvScreensaverImage
import com.chee.videos.core.model.TvScreensaverImagesPayload
import com.chee.videos.core.network.ApiService
import com.chee.videos.core.player.PlaybackProfileResolver
import com.chee.videos.core.testing.MainDispatcherRule
import com.google.gson.Gson
import java.io.File
import java.lang.reflect.Proxy
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.test.TestScope
import kotlinx.coroutines.test.runTest
import org.junit.Assert.*
import org.junit.Rule
import org.junit.Test

class TvScreensaverRepositoryTest {
    @get:Rule val dispatcher = MainDispatcherRule()

    private suspend fun TestScope.store(): AppPreferencesStore {
        val store = AppPreferencesStore(
            PreferenceDataStoreFactory.create(scope = backgroundScope,
                produceFile = { File.createTempFile("tv-screensaver-repo", ".preferences_pb").apply { deleteOnExit() } }),
            Gson(),
        )
        store.setActiveBaseUrl("https://gallery.example")
        store.saveTokens(SessionTokens("access", "refresh"))
        return store
    }

    private fun api(handler: (String, Array<out Any?>) -> Any?): ApiService =
        Proxy.newProxyInstance(ApiService::class.java.classLoader, arrayOf(ApiService::class.java)) { _, method, args ->
            handler(method.name, args ?: emptyArray())
        } as ApiService

    private fun repository(api: ApiService, store: AppPreferencesStore) =
        VideoRepository(api, store, AuthRepository(api, store), PlaybackProfileResolver())

    @Test
    fun followsCursorAndDeduplicatesFullGalleryWithoutAdminCalls() = runTest {
        val cursors = mutableListOf<String?>()
        val first = TvScreensaverImage("a", "/a")
        val second = TvScreensaverImage("b", "/b")
        val api = api { name, args ->
            assertEquals("tvScreensaverImages", name)
            assertEquals("https://gallery.example/api/v1/tv/screensaver/images", args[0])
            assertEquals("Bearer access", args[1])
            assertEquals(200, args[3])
            cursors += args[2] as String?
            ApiEnvelope(code = 0, data = if (args[2] == null)
                TvScreensaverImagesPayload(listOf(first), "cursor-a")
            else TvScreensaverImagesPayload(listOf(first, second)))
        }
        assertEquals(listOf(first, second), repository(api, store()).fetchTvScreensaverImages())
        assertEquals(listOf(null, "cursor-a"), cursors)
    }

    @Test
    fun repeatedCursorFailsInsteadOfLoopingForever() = runTest {
        var requests = 0
        val api = api { _, _ ->
            requests++
            ApiEnvelope(code = 0, data = TvScreensaverImagesPayload(emptyList(), "same"))
        }
        try {
            repository(api, store()).fetchTvScreensaverImages()
            fail("重复游标应失败")
        } catch (_: com.chee.videos.core.model.AppException) {
            assertEquals(2, requests)
        }
    }

    @Test
    fun cancellationDuringRefreshDoesNotLogOutUser() = runTest {
        val store = store()
        val api = api { name, _ ->
            when (name) {
                "tvScreensaverImages" -> ApiEnvelope<TvScreensaverImagesPayload>(code = 401)
                "refresh" -> throw CancellationException("用户已唤醒")
                else -> error("unexpected API: $name")
            }
        }
        try {
            repository(api, store).fetchTvScreensaverImages()
            fail("应传播取消")
        } catch (_: CancellationException) {
            assertEquals("access", store.readAccessToken())
            assertEquals("refresh", store.readRefreshToken())
        }
    }

    @Test
    fun refreshRetriesWithNewBearerAndReturnsCandidate() = runTest {
        var requests = 0
        val expected = TvScreensaverImage("a", "/a")
        val api = api { name, args ->
            when (name) {
                "tvScreensaverImages" -> {
                    requests++
                    if (args[1] == "Bearer access") ApiEnvelope<TvScreensaverImagesPayload>(code = 401)
                    else {
                        assertEquals("Bearer renewed-access", args[1])
                        ApiEnvelope(code = 0, data = TvScreensaverImagesPayload(listOf(expected)))
                    }
                }
                "refresh" -> ApiEnvelope(code = 0, data = RefreshPayload("renewed-access", "renewed-refresh"))
                else -> error("unexpected API: $name")
            }
        }
        assertEquals(listOf(expected), repository(api, store()).fetchTvScreensaverImages())
        assertEquals(2, requests)
    }
}
