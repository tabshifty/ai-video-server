package com.chee.videos.feature.actor

import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.ActionTogglePayload
import com.chee.videos.core.model.ApiEnvelope
import com.chee.videos.core.model.ActorDetailDto
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
class ActorDetailViewModelTest {
    @Test
    fun initialize_loadsFirstPage() = runTest {
        withMainDispatcher {
            val api = FakeActorApiService(
                pages = mapOf(1 to listOf(actorVideo("v1", "SSIS-101"))),
                totalCount = 2,
            )
            val viewModel = buildViewModel(api)

            viewModel.initialize("actor-1")
            awaitUntil { viewModel.uiState.value.loaded && !viewModel.uiState.value.loading }

            val state = viewModel.uiState.value
            assertEquals("actor-1", state.actorId)
            assertEquals("演员甲", state.actor?.name)
            assertEquals(1, state.items.size)
            assertTrue(state.hasMore)
            assertEquals(listOf(ActorCall("actor-1", 1, 24)), api.calls)
        }
    }

    @Test
    fun loadMoreIfNeeded_appendsNextPageAndStopsAtTotal() = runTest {
        withMainDispatcher {
            val api = FakeActorApiService(
                pages = mapOf(
                    1 to listOf(actorVideo("v1", "SSIS-101")),
                    2 to listOf(actorVideo("v2", "SSIS-102")),
                ),
                totalCount = 2,
            )
            val viewModel = buildViewModel(api)

            viewModel.initialize("actor-1")
            awaitUntil { viewModel.uiState.value.loaded && !viewModel.uiState.value.loading }
            viewModel.loadMoreIfNeeded(0)
            awaitUntil { viewModel.uiState.value.page == 2 && !viewModel.uiState.value.loadingMore }

            val state = viewModel.uiState.value
            assertEquals(listOf("v1", "v2"), state.items.map { it.id })
            assertFalse(state.hasMore)
            assertEquals(
                listOf(
                    ActorCall("actor-1", 1, 24),
                    ActorCall("actor-1", 2, 24),
                ),
                api.calls,
            )
        }
    }

    @Test
    fun initialize_failureShowsError() = runTest {
        withMainDispatcher {
            val api = FakeActorApiService(errors = mapOf(1 to IllegalStateException("加载失败")))
            val viewModel = buildViewModel(api)

            viewModel.initialize("actor-1")
            awaitUntil { viewModel.uiState.value.loaded && !viewModel.uiState.value.loading }

            assertEquals("加载失败", viewModel.uiState.value.errorMessage)
            assertTrue(viewModel.uiState.value.items.isEmpty())
        }
    }

    @Test
    fun authExpired_clearsLocalToken() = runTest {
        withMainDispatcher {
            val api = FakeActorApiService(expireFirstPage = true)
            val store = buildStore()
            val viewModel = buildViewModel(api, store)

            viewModel.initialize("actor-1")
            awaitUntil { viewModel.uiState.value.loaded && !viewModel.uiState.value.loading }

            assertNull(store.readAccessToken())
            assertEquals("登录已失效，请重新登录", viewModel.uiState.value.errorMessage)
        }
    }

    private suspend fun buildViewModel(
        api: FakeActorApiService,
        store: AppPreferencesStore? = null,
    ): ActorDetailViewModel {
        val resolvedStore = store ?: buildStore()
        val authRepository = AuthRepository(api = api, store = resolvedStore)
        val videoRepository = VideoRepository(
            api = api,
            store = resolvedStore,
            authRepository = authRepository,
            playbackProfileResolver = PlaybackProfileResolver(),
        )
        return ActorDetailViewModel(videoRepository = videoRepository, authRepository = authRepository)
    }

    private suspend fun buildStore(): AppPreferencesStore {
        val store = AppPreferencesStore(
            dataStore = PreferenceDataStoreFactory.create(
                produceFile = {
                    File.createTempFile("actor-view-model", ".preferences_pb").apply {
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

private data class ActorCall(val actorId: String, val page: Int, val pageSize: Int)

private fun actorVideo(id: String, title: String): VideoListItemDto =
    VideoListItemDto(id = id, title = title, type = "av", thumbnailPath = "/thumb/$id.jpg")

private class FakeActorApiService(
    private val pages: Map<Int, List<VideoListItemDto>> = emptyMap(),
    private val totalCount: Int = pages.values.sumOf { it.size },
    private val errors: Map<Int, Throwable> = emptyMap(),
    private val expireFirstPage: Boolean = false,
) : ApiService {
    val calls = mutableListOf<ActorCall>()

    override suspend fun actorDetail(
        url: String,
        authorization: String,
        page: Int,
        pageSize: Int,
    ): ApiEnvelope<ActorWorksPayload> {
        val actorId = url.substringAfterLast("/")
        calls += ActorCall(actorId, page, pageSize)
        if (expireFirstPage && page == 1) {
            return ApiEnvelope(code = 401, msg = "expired")
        }
        errors[page]?.let { throw it }
        return ApiEnvelope(
            code = 0,
            data = ActorWorksPayload(
                actor = ActorDetailDto(id = actorId, name = "演员甲"),
                items = pages[page].orEmpty(),
                totalCount = totalCount,
                page = page,
                pageSize = pageSize,
            ),
        )
    }

    override suspend fun refresh(url: String, authorization: String): ApiEnvelope<RefreshPayload> =
        ApiEnvelope(code = 401, msg = "expired")

    override suspend fun login(url: String, body: LoginRequest): ApiEnvelope<LoginPayload> = error("unused")
    override suspend fun recommend(url: String, authorization: String, pageSize: Int): ApiEnvelope<FeedPayload> = error("unused")
    override suspend fun randomShort(url: String, pageSize: Int, excludeIds: String?): ApiEnvelope<FeedPayload> = error("unused")
    override suspend fun shortDiscover(url: String, authorization: String, mode: String, tag: String?, collectionID: String?, page: Int, pageSize: Int): ApiEnvelope<SearchPayload> = error("unused")
    override suspend fun tvHome(url: String, authorization: String, keyword: String?, page: Int, pageSize: Int): ApiEnvelope<TvHomePayload> = error("unused")
    override suspend fun tvSearch(url: String, authorization: String, keyword: String, page: Int, pageSize: Int): ApiEnvelope<TvSearchPayload> = error("unused")
    override suspend fun tvSeriesDetail(url: String, authorization: String): ApiEnvelope<TvSeriesDetailDto> = error("unused")
    override suspend fun createTvAuthSession(url: String, body: TvAuthSessionCreateRequest): ApiEnvelope<TvAuthSessionCreatePayload> = error("unused")
    override suspend fun getTvAuthSession(url: String): ApiEnvelope<TvAuthSessionStatusPayload> = error("unused")
    override suspend fun approveTvAuthSession(url: String, authorization: String): ApiEnvelope<Map<String, Boolean>> = error("unused")
    override suspend fun denyTvAuthSession(url: String, authorization: String): ApiEnvelope<Map<String, Boolean>> = error("unused")
    override suspend fun imageCollections(url: String, authorization: String, keyword: String?, page: Int, pageSize: Int): ApiEnvelope<ImageCollectionsPayload> = error("unused")
    override suspend fun imageCollectionDetail(url: String, authorization: String): ApiEnvelope<ImageCollectionDetailDto> = error("unused")
    override suspend fun search(url: String, authorization: String, keyword: String, type: String, page: Int, pageSize: Int): ApiEnvelope<SearchPayload> = error("unused")
    override suspend fun detail(url: String, authorization: String): ApiEnvelope<VideoDetailDto> = error("unused")
    override suspend fun toggleLike(url: String, authorization: String): ApiEnvelope<ActionTogglePayload> = error("unused")
    override suspend fun toggleFavorite(url: String, authorization: String): ApiEnvelope<ActionTogglePayload> = error("unused")
    override suspend fun toggleDislike(url: String, authorization: String): ApiEnvelope<ActionTogglePayload> = error("unused")
    override suspend fun recordHistory(url: String, authorization: String, body: RecordHistoryRequest): ApiEnvelope<Map<String, Boolean>> = error("unused")
    override suspend fun continueHistory(url: String, authorization: String, page: Int, limit: Int): ApiEnvelope<ContinueHistoryPayload> = error("unused")
    override suspend fun likedVideos(url: String, authorization: String, page: Int, pageSize: Int): ApiEnvelope<SearchPayload> = error("unused")
    override suspend fun favoritedVideos(url: String, authorization: String, page: Int, pageSize: Int): ApiEnvelope<SearchPayload> = error("unused")
    override suspend fun userProfile(url: String, authorization: String): ApiEnvelope<UserProfileDto> = error("unused")
}
