package com.chee.videos.feature.home

import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.ActionTogglePayload
import com.chee.videos.core.model.ApiEnvelope
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
import com.chee.videos.core.model.TvCatalogWallPayload
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
import com.chee.videos.core.testing.MainDispatcherRule
import com.google.gson.Gson
import java.io.File
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.StandardTestDispatcher
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import kotlinx.coroutines.test.TestScope
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class HomeViewModelTest {
    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    @Test
    fun `loadCategory loads default av list`() = runTest(mainDispatcherRule.standardDispatcher) {
        withMainDispatcher {
            val api = FakeHomeApiService(
                browseItems = mapOf(
                    "av" to listOf(
                        avItem(id = "av-1", title = "SSIS-101 夜色标本"),
                    ),
                ),
            )
            val viewModel = buildViewModel(api, CoroutineScope(UnconfinedTestDispatcher(testScheduler)))
            viewModel.avSearchDebounceMs = 0

            viewModel.loadCategory("av")
            advanceUntilIdle()

            val state = viewModel.uiState.value
            assertTrue(state.av.loaded)
            assertFalse(state.av.loading)
            assertEquals(1, state.av.items.size)
            assertEquals("SSIS-101 夜色标本", state.av.items.first().title)
        }
    }

    @Test
    fun `updateAvQuery triggers remote av search after debounce`() = runTest(mainDispatcherRule.standardDispatcher) {
        withMainDispatcher {
            val api = FakeHomeApiService(
                browseItems = mapOf("av" to listOf(avItem(id = "av-1", title = "SSIS-101 夜色标本"))),
                searchItems = mapOf(
                    "ssis 222" to listOf(
                        avItem(id = "av-2", title = "SSIS-222 清晨边界"),
                    ),
                ),
            )
            val viewModel = buildViewModel(api, CoroutineScope(UnconfinedTestDispatcher(testScheduler)))
            viewModel.avSearchDebounceMs = 0

            viewModel.loadCategory("av")
            advanceUntilIdle()
            viewModel.updateAvQuery("ssis 222")
            advanceUntilIdle()

            val state = viewModel.uiState.value
            assertEquals("ssis 222", state.avSearch.query)
            assertTrue(state.avSearch.isSearchMode)
            assertFalse(state.avSearch.loading)
            assertEquals(1, state.avSearch.results.size)
            assertEquals("SSIS-222 清晨边界", state.avSearch.results.first().title)
            assertEquals(
                listOf(
                    SearchCall(keyword = "", type = "av"),
                    SearchCall(keyword = "ssis 222", type = "av"),
                ),
                api.searchCalls,
            )
        }
    }

    @Test
    fun `clearing query restores default av browse list`() = runTest(mainDispatcherRule.standardDispatcher) {
        withMainDispatcher {
            val api = FakeHomeApiService(
                browseItems = mapOf(
                    "av" to listOf(
                        avItem(id = "av-1", title = "IPX-901 雾面轮廓"),
                    ),
                ),
                searchItems = mapOf(
                    "ipx" to listOf(
                        avItem(id = "av-2", title = "IPX-902 逆光碎片"),
                    ),
                ),
            )
            val viewModel = buildViewModel(api, CoroutineScope(UnconfinedTestDispatcher(testScheduler)))
            viewModel.avSearchDebounceMs = 0

            viewModel.loadCategory("av")
            advanceUntilIdle()
            viewModel.updateAvQuery("ipx")
            advanceUntilIdle()

            viewModel.updateAvQuery("")
            advanceUntilIdle()

            val state = viewModel.uiState.value
            assertFalse(state.avSearch.isSearchMode)
            assertFalse(state.avSearch.loading)
            assertTrue(state.avSearch.results.isEmpty())
            assertEquals(1, state.av.items.size)
            assertEquals("IPX-901 雾面轮廓", state.av.items.first().title)
        }
    }

    @Test
    fun `search failure does not overwrite default av browse list`() = runTest(mainDispatcherRule.standardDispatcher) {
        withMainDispatcher {
            val api = FakeHomeApiService(
                browseItems = mapOf(
                    "av" to listOf(
                        avItem(id = "av-1", title = "CAWD-808 白昼残响"),
                    ),
                ),
                searchErrors = mapOf("broken" to IllegalStateException("搜索失败")),
            )
            val viewModel = buildViewModel(api, CoroutineScope(UnconfinedTestDispatcher(testScheduler)))
            viewModel.avSearchDebounceMs = 0

            viewModel.loadCategory("av")
            advanceUntilIdle()
            viewModel.updateAvQuery("broken")
            advanceUntilIdle()

            val state = viewModel.uiState.value
            assertEquals(1, state.av.items.size)
            assertEquals("CAWD-808 白昼残响", state.av.items.first().title)
            assertTrue(state.avSearch.isSearchMode)
            assertEquals("搜索失败", state.avSearch.errorMessage)
            assertTrue(state.avSearch.results.isEmpty())
        }
    }

    private suspend fun buildViewModel(
        api: FakeHomeApiService,
        dataStoreScope: CoroutineScope,
    ): HomeViewModel {
        val store = AppPreferencesStore(
            dataStore = PreferenceDataStoreFactory.create(
                scope = dataStoreScope,
                produceFile = {
                    File.createTempFile("home-view-model", ".preferences_pb").apply {
                        deleteOnExit()
                    }
                },
            ),
            gson = Gson(),
        )
        store.setActiveBaseUrl("https://example.com")
        store.saveTokens(
            SessionTokens(
                accessToken = "access-token",
                refreshToken = "refresh-token",
            ),
        )

        val authRepository = AuthRepository(api = api, store = store)
        val videoRepository = VideoRepository(
            api = api,
            store = store,
            authRepository = authRepository,
            playbackProfileResolver = PlaybackProfileResolver(),
        )
        return HomeViewModel(
            videoRepository = videoRepository,
            authRepository = authRepository,
        )
    }
}

@OptIn(ExperimentalCoroutinesApi::class)
private suspend fun TestScope.withMainDispatcher(block: suspend TestScope.() -> Unit) {
    Dispatchers.setMain(StandardTestDispatcher(testScheduler))
    block()
}

private data class SearchCall(
    val keyword: String,
    val type: String,
)

private fun avItem(
    id: String,
    title: String,
): VideoListItemDto = VideoListItemDto(
    id = id,
    title = title,
    type = "av",
    thumbnailPath = "/thumb/$id.jpg",
    duration = 5_400,
)

private class FakeHomeApiService(
    private val browseItems: Map<String, List<VideoListItemDto>> = emptyMap(),
    private val searchItems: Map<String, List<VideoListItemDto>> = emptyMap(),
    private val searchErrors: Map<String, Throwable> = emptyMap(),
) : ApiService {
    val searchCalls = mutableListOf<SearchCall>()

    override suspend fun search(
        url: String,
        authorization: String,
        keyword: String,
        type: String,
        page: Int,
        pageSize: Int,
    ): ApiEnvelope<SearchPayload> {
        val normalizedKeyword = keyword.trim()
        searchCalls += SearchCall(keyword = normalizedKeyword, type = type)
        searchErrors[normalizedKeyword]?.let { throw it }
        val items = if (normalizedKeyword.isBlank()) {
            browseItems[type].orEmpty()
        } else {
            searchItems[normalizedKeyword].orEmpty()
        }
        return ApiEnvelope(
            code = 0,
            data = SearchPayload(
                items = items,
                totalCount = items.size,
                page = page,
                pageSize = pageSize,
            ),
        )
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
        kind: String?,
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

    override suspend fun tvCatalogWall(
        url: String,
        authorization: String,
        kind: String,
        page: Int,
        pageSize: Int,
    ): ApiEnvelope<TvCatalogWallPayload> = error("unused")

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

    override suspend fun detail(
        url: String,
        authorization: String,
    ): ApiEnvelope<VideoDetailDto> = error("unused")

    override suspend fun toggleLike(
        url: String,
        authorization: String,
    ): ApiEnvelope<ActionTogglePayload> = error("unused")

    override suspend fun toggleFavorite(
        url: String,
        authorization: String,
    ): ApiEnvelope<ActionTogglePayload> = error("unused")

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
