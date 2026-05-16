package com.chee.videos.feature.shortsearch

import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.ActionTogglePayload
import com.chee.videos.core.model.ApiEnvelope
import com.chee.videos.core.model.ActorWorksPayload
import com.chee.videos.core.model.ContinueHistoryPayload
import com.chee.videos.core.model.FeedPayload
import com.chee.videos.core.model.ImageCollectionDetailDto
import com.chee.videos.core.model.ImageCollectionsPayload
import com.chee.videos.core.model.LoginPayload
import com.chee.videos.core.model.LoginRequest
import com.chee.videos.core.model.RecordHistoryRequest
import com.chee.videos.core.model.RefreshPayload
import com.chee.videos.core.model.SearchPayload
import com.chee.videos.core.model.SessionTokens
import com.chee.videos.core.model.TvAuthSessionCreatePayload
import com.chee.videos.core.model.TvAuthSessionCreateRequest
import com.chee.videos.core.model.TvAuthSessionStatusPayload
import com.chee.videos.core.model.TvHomePayload
import com.chee.videos.core.model.TvSearchPayload
import com.chee.videos.core.model.TvSeriesDetailDto
import com.chee.videos.core.model.UserProfileDto
import com.chee.videos.core.model.UserStateDto
import com.chee.videos.core.model.VideoDetailDto
import com.chee.videos.core.model.VideoListItemDto
import com.chee.videos.core.network.ApiService
import com.chee.videos.core.player.PlaybackProfileResolver
import com.chee.videos.core.repository.AuthRepository
import com.chee.videos.core.repository.VideoRepository
import com.google.gson.Gson
import java.io.File
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.TestScope
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class ShortSearchViewModelStateTest {
    @Test
    fun normalizeShortSearchQuery_trimsWhitespace() {
        assertEquals("标签", normalizeShortSearchQuery("  标签  "))
        assertEquals("", normalizeShortSearchQuery("   "))
    }

    @Test
    fun resetShortSearchForQuery_clearsPagingAndErrors() {
        val state = ShortSearchUiState(
            activeQuery = "旧词",
            loading = false,
            loaded = true,
            page = 3,
            totalCount = 99,
            items = listOf(VideoListItemDto(id = "a", title = "A", type = "short")),
            errorMessage = "old",
        )

        val next = resetShortSearchForQuery(state, "新词")

        assertEquals("新词", next.activeQuery)
        assertTrue(next.loading)
        assertFalse(next.loaded)
        assertEquals(0, next.page)
        assertEquals(0, next.totalCount)
        assertTrue(next.items.isEmpty())
        assertNull(next.errorMessage)
    }

    @Test
    fun mergeShortSearchItems_deduplicatesById() {
        val merged = mergeShortSearchItems(
            existing = listOf(VideoListItemDto(id = "a", title = "A", type = "short")),
            incoming = listOf(
                VideoListItemDto(id = "a", title = "A2", type = "short"),
                VideoListItemDto(id = "b", title = "B", type = "short"),
            ),
        )

        assertEquals(listOf("a", "b"), merged.map { it.id })
    }

    @Test
    fun toggleLike_loadsDetailThenUpdatesUserState() = runTest {
        withMainDispatcher {
            val api = FakeShortSearchApiService(
                detailStates = mapOf("v1" to UserStateDto(isLiked = false, isFavorited = false)),
                likeResults = mapOf("v1" to true),
            )
            val viewModel = buildViewModel(api)

            viewModel.toggleLike("v1")
            awaitUntil { viewModel.uiState.value.detailByVideoId["v1"]?.userState?.isLiked == true }

            val state = viewModel.uiState.value
            assertEquals(listOf("v1"), api.detailCalls)
            assertEquals(listOf("v1"), api.likeCalls)
            assertTrue(state.detailByVideoId["v1"]?.userState?.isLiked == true)
            assertFalse(state.detailByVideoId["v1"]?.userState?.isDisliked == true)
            assertFalse("操作完成后应释放忙碌状态", "v1" in state.actionBusyVideoIds)
        }
    }

    @Test
    fun toggleFavorite_updatesLoadedDetailUserState() = runTest {
        withMainDispatcher {
            val api = FakeShortSearchApiService(
                detailStates = mapOf("v1" to UserStateDto(isLiked = true, isFavorited = false)),
                favoriteResults = mapOf("v1" to true),
            )
            val viewModel = buildViewModel(api)

            viewModel.ensureDetailLoaded("v1")
            awaitUntil { viewModel.uiState.value.detailByVideoId.containsKey("v1") }
            viewModel.toggleFavorite("v1")
            awaitUntil { viewModel.uiState.value.detailByVideoId["v1"]?.userState?.isFavorited == true }

            val state = viewModel.uiState.value
            assertEquals(listOf("v1"), api.detailCalls)
            assertEquals(listOf("v1"), api.favoriteCalls)
            assertTrue(state.detailByVideoId["v1"]?.userState?.isLiked == true)
            assertTrue(state.detailByVideoId["v1"]?.userState?.isFavorited == true)
        }
    }

    @Test
    fun authExpiredWhileLoadingDetail_clearsLocalToken() = runTest {
        withMainDispatcher {
            val api = FakeShortSearchApiService(expireDetail = true)
            val store = buildStore()
            val viewModel = buildViewModel(api, store)

            viewModel.ensureDetailLoaded("v1")
            awaitUntil { "v1" !in viewModel.uiState.value.detailLoadingVideoIds }

            assertNull(store.readAccessToken())
        }
    }

    private suspend fun buildViewModel(
        api: FakeShortSearchApiService,
        store: AppPreferencesStore? = null,
    ): ShortSearchViewModel {
        val resolvedStore = store ?: buildStore()
        val authRepository = AuthRepository(api = api, store = resolvedStore)
        val videoRepository = VideoRepository(
            api = api,
            store = resolvedStore,
            authRepository = authRepository,
            playbackProfileResolver = PlaybackProfileResolver(),
        )
        return ShortSearchViewModel(
            videoRepository = videoRepository,
            store = resolvedStore,
            authRepository = authRepository,
        )
    }

    private suspend fun buildStore(): AppPreferencesStore {
        val store = AppPreferencesStore(
            dataStore = PreferenceDataStoreFactory.create(
                produceFile = {
                    File.createTempFile("short-search-view-model", ".preferences_pb").apply {
                        deleteOnExit()
                    }
                },
            ),
            gson = Gson(),
        )
        store.setActiveBaseUrl("https://example.com")
        store.saveTokens(SessionTokens(accessToken = "access-token", refreshToken = "refresh-token"))
        return store
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

@OptIn(ExperimentalCoroutinesApi::class)
private suspend fun TestScope.awaitUntil(
    timeoutMs: Long = 1_000L,
    condition: () -> Boolean,
) {
    val deadline = System.currentTimeMillis() + timeoutMs
    while (System.currentTimeMillis() < deadline) {
        advanceUntilIdle()
        if (condition()) {
            return
        }
        Thread.sleep(10)
    }
    advanceUntilIdle()
}

private class FakeShortSearchApiService(
    private val detailStates: Map<String, UserStateDto> = emptyMap(),
    private val likeResults: Map<String, Boolean> = emptyMap(),
    private val favoriteResults: Map<String, Boolean> = emptyMap(),
    private val expireDetail: Boolean = false,
) : ApiService {
    val detailCalls = mutableListOf<String>()
    val likeCalls = mutableListOf<String>()
    val favoriteCalls = mutableListOf<String>()

    override suspend fun detail(
        url: String,
        authorization: String,
    ): ApiEnvelope<VideoDetailDto> {
        val videoId = url.substringAfterLast("/")
        detailCalls += videoId
        if (expireDetail) {
            return ApiEnvelope(code = 401, msg = "expired")
        }
        return ApiEnvelope(
            code = 0,
            data = VideoDetailDto(
                id = videoId,
                title = "测试短视频",
                userState = detailStates[videoId] ?: UserStateDto(),
            ),
        )
    }

    override suspend fun toggleLike(
        url: String,
        authorization: String,
    ): ApiEnvelope<ActionTogglePayload> {
        val videoId = url.substringBeforeLast("/like").substringAfterLast("/")
        likeCalls += videoId
        return ApiEnvelope(code = 0, data = ActionTogglePayload(action = "like", enabled = likeResults[videoId] == true))
    }

    override suspend fun toggleFavorite(
        url: String,
        authorization: String,
    ): ApiEnvelope<ActionTogglePayload> {
        val videoId = url.substringBeforeLast("/favorite").substringAfterLast("/")
        favoriteCalls += videoId
        return ApiEnvelope(code = 0, data = ActionTogglePayload(action = "favorite", enabled = favoriteResults[videoId] == true))
    }

    override suspend fun refresh(url: String, authorization: String): ApiEnvelope<RefreshPayload> =
        ApiEnvelope(code = 401, msg = "expired")

    override suspend fun login(url: String, body: LoginRequest): ApiEnvelope<LoginPayload> =
        error("unused")

    override suspend fun recommend(
        url: String,
        authorization: String,
        pageSize: Int,
    ): ApiEnvelope<FeedPayload> = error("unused")

    override suspend fun randomShort(
        url: String,
        pageSize: Int,
        excludeIds: String?,
    ): ApiEnvelope<FeedPayload> = error("unused")

    override suspend fun shortDiscover(
        url: String,
        authorization: String,
        mode: String,
        tag: String?,
        collectionID: String?,
        page: Int,
        pageSize: Int,
    ): ApiEnvelope<SearchPayload> = error("unused")

    override suspend fun tvHome(
        url: String,
        authorization: String,
        keyword: String?,
        page: Int,
        pageSize: Int,
    ): ApiEnvelope<TvHomePayload> = error("unused")

    override suspend fun tvSearch(
        url: String,
        authorization: String,
        keyword: String,
        page: Int,
        pageSize: Int,
    ): ApiEnvelope<TvSearchPayload> = error("unused")

    override suspend fun tvSeriesDetail(
        url: String,
        authorization: String,
    ): ApiEnvelope<TvSeriesDetailDto> = error("unused")

    override suspend fun createTvAuthSession(
        url: String,
        body: TvAuthSessionCreateRequest,
    ): ApiEnvelope<TvAuthSessionCreatePayload> = error("unused")

    override suspend fun getTvAuthSession(
        url: String,
    ): ApiEnvelope<TvAuthSessionStatusPayload> = error("unused")

    override suspend fun approveTvAuthSession(
        url: String,
        authorization: String,
    ): ApiEnvelope<Map<String, Boolean>> = error("unused")

    override suspend fun denyTvAuthSession(
        url: String,
        authorization: String,
    ): ApiEnvelope<Map<String, Boolean>> = error("unused")

    override suspend fun imageCollections(
        url: String,
        authorization: String,
        keyword: String?,
        page: Int,
        pageSize: Int,
    ): ApiEnvelope<ImageCollectionsPayload> = error("unused")

    override suspend fun imageCollectionDetail(
        url: String,
        authorization: String,
    ): ApiEnvelope<ImageCollectionDetailDto> = error("unused")

    override suspend fun search(
        url: String,
        authorization: String,
        keyword: String,
        type: String,
        page: Int,
        pageSize: Int,
    ): ApiEnvelope<SearchPayload> = error("unused")

    override suspend fun actorDetail(
        url: String,
        authorization: String,
        page: Int,
        pageSize: Int,
    ): ApiEnvelope<ActorWorksPayload> = error("unused")

    override suspend fun toggleDislike(
        url: String,
        authorization: String,
    ): ApiEnvelope<ActionTogglePayload> = error("unused")

    override suspend fun recordHistory(
        url: String,
        authorization: String,
        body: RecordHistoryRequest,
    ): ApiEnvelope<Map<String, Boolean>> = error("unused")

    override suspend fun continueHistory(
        url: String,
        authorization: String,
        page: Int,
        limit: Int,
    ): ApiEnvelope<ContinueHistoryPayload> = error("unused")

    override suspend fun likedVideos(
        url: String,
        authorization: String,
        page: Int,
        pageSize: Int,
    ): ApiEnvelope<SearchPayload> = error("unused")

    override suspend fun favoritedVideos(
        url: String,
        authorization: String,
        page: Int,
        pageSize: Int,
    ): ApiEnvelope<SearchPayload> = error("unused")

    override suspend fun userProfile(
        url: String,
        authorization: String,
    ): ApiEnvelope<UserProfileDto> = error("unused")
}
