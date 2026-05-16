package com.chee.videos.feature.home

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
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.runTest
import kotlinx.coroutines.test.setMain
import kotlinx.coroutines.test.TestScope
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class HomeViewModelTest {
    @Test
    fun `loadCategory loads default av list`() = runTest {
        withMainDispatcher {
            val api = FakeHomeApiService(
                browsePages = mapOf(
                    "av" to mapOf(
                        1 to listOf(
                        avItem(id = "av-1", title = "SSIS-101 夜色标本"),
                    ),
                    ),
                ),
            )
            val viewModel = buildViewModel(api)
            viewModel.avSearchDebounceMs = 0

            viewModel.loadCategory("av")
            awaitUntil { viewModel.uiState.value.av.loaded && !viewModel.uiState.value.av.loading }

            val state = viewModel.uiState.value
            assertTrue(state.av.loaded)
            assertFalse(state.av.loading)
            assertEquals(1, state.av.items.size)
            assertEquals("SSIS-101 夜色标本", state.av.items.first().title)
        }
    }

    @Test
    fun `updateAvQuery triggers remote av search after debounce`() = runTest {
        withMainDispatcher {
            val api = FakeHomeApiService(
                browsePages = mapOf("av" to mapOf(1 to listOf(avItem(id = "av-1", title = "SSIS-101 夜色标本")))),
                searchPages = mapOf(
                    "ssis 222" to mapOf(
                        1 to listOf(
                            avItem(id = "av-2", title = "SSIS-222 清晨边界"),
                        ),
                    ),
                ),
            )
            val viewModel = buildViewModel(api)
            viewModel.avSearchDebounceMs = 0

            viewModel.loadCategory("av")
            awaitUntil { viewModel.uiState.value.av.loaded && !viewModel.uiState.value.av.loading }
            viewModel.updateAvQuery("ssis 222")
            awaitUntil {
                viewModel.uiState.value.avSearch.isSearchMode &&
                    !viewModel.uiState.value.avSearch.resultState.loading &&
                    viewModel.uiState.value.avSearch.resultState.items.isNotEmpty()
            }

            val state = viewModel.uiState.value
            assertEquals("ssis 222", state.avSearch.query)
            assertTrue(state.avSearch.isSearchMode)
            assertFalse(state.avSearch.resultState.loading)
            assertEquals(1, state.avSearch.resultState.items.size)
            assertEquals("SSIS-222 清晨边界", state.avSearch.resultState.items.first().title)
            assertEquals(
                listOf(
                    SearchCall(keyword = "", type = "av", page = 1, pageSize = 30),
                    SearchCall(keyword = "ssis 222", type = "av", page = 1, pageSize = 30),
                ),
                api.searchCalls,
            )
        }
    }

    @Test
    fun `clearing query restores default av browse list`() = runTest {
        withMainDispatcher {
            val api = FakeHomeApiService(
                browsePages = mapOf(
                    "av" to mapOf(
                        1 to listOf(
                        avItem(id = "av-1", title = "IPX-901 雾面轮廓"),
                    ),
                    ),
                ),
                searchPages = mapOf(
                    "ipx" to mapOf(
                        1 to listOf(
                            avItem(id = "av-2", title = "IPX-902 逆光碎片"),
                        ),
                    ),
                ),
            )
            val viewModel = buildViewModel(api)
            viewModel.avSearchDebounceMs = 0

            viewModel.loadCategory("av")
            awaitUntil { viewModel.uiState.value.av.loaded && !viewModel.uiState.value.av.loading }
            viewModel.updateAvQuery("ipx")
            awaitUntil {
                viewModel.uiState.value.avSearch.isSearchMode &&
                    !viewModel.uiState.value.avSearch.resultState.loading &&
                    viewModel.uiState.value.avSearch.resultState.items.isNotEmpty()
            }

            viewModel.updateAvQuery("")
            awaitUntil { !viewModel.uiState.value.avSearch.isSearchMode }

            val state = viewModel.uiState.value
            assertFalse(state.avSearch.isSearchMode)
            assertFalse(state.avSearch.resultState.loading)
            assertTrue(state.avSearch.resultState.items.isEmpty())
            assertEquals(1, state.av.items.size)
            assertEquals("IPX-901 雾面轮廓", state.av.items.first().title)
        }
    }

    @Test
    fun `search failure does not overwrite default av browse list`() = runTest {
        withMainDispatcher {
            val api = FakeHomeApiService(
                browsePages = mapOf(
                    "av" to mapOf(
                        1 to listOf(
                        avItem(id = "av-1", title = "CAWD-808 白昼残响"),
                    ),
                    ),
                ),
                searchErrors = mapOf(SearchErrorKey(keyword = "broken", page = 1) to IllegalStateException("搜索失败")),
            )
            val viewModel = buildViewModel(api)
            viewModel.avSearchDebounceMs = 0

            viewModel.loadCategory("av")
            awaitUntil { viewModel.uiState.value.av.loaded && !viewModel.uiState.value.av.loading }
            viewModel.updateAvQuery("broken")
            awaitUntil {
                viewModel.uiState.value.avSearch.isSearchMode &&
                    !viewModel.uiState.value.avSearch.resultState.loading &&
                    !viewModel.uiState.value.avSearch.resultState.errorMessage.isNullOrBlank()
            }

            val state = viewModel.uiState.value
            assertEquals(1, state.av.items.size)
            assertEquals("CAWD-808 白昼残响", state.av.items.first().title)
            assertTrue(state.avSearch.isSearchMode)
            assertEquals("搜索失败", state.avSearch.resultState.errorMessage)
            assertTrue(state.avSearch.resultState.items.isEmpty())
        }
    }

    @Test
    fun `loadMoreAvIfNeeded appends next browse page and marks no more when total reached`() = runTest {
        withMainDispatcher {
            val api = FakeHomeApiService(
                browsePages = mapOf(
                    "av" to mapOf(
                        1 to listOf(
                            avItem(id = "av-1", title = "SSIS-101 夜色标本"),
                            avItem(id = "av-2", title = "SSIS-102 深夜边界"),
                        ),
                        2 to listOf(
                            avItem(id = "av-3", title = "SSIS-103 余温残响"),
                        ),
                    ),
                ),
                browseTotalCounts = mapOf("av" to 3),
            )
            val viewModel = buildViewModel(api)
            viewModel.avSearchDebounceMs = 0

            viewModel.loadCategory("av")
            awaitUntil { viewModel.uiState.value.av.loaded && !viewModel.uiState.value.av.loading }
            viewModel.loadMoreAvIfNeeded(1)
            awaitUntil { viewModel.uiState.value.av.page == 2 && !viewModel.uiState.value.av.loadingMore }

            val state = viewModel.uiState.value
            assertEquals(3, state.av.items.size)
            assertEquals(2, state.av.page)
            assertFalse(state.av.hasMore)
            assertFalse(state.av.loadingMore)
            assertEquals(
                listOf(
                    SearchCall(keyword = "", type = "av", page = 1, pageSize = 30),
                    SearchCall(keyword = "", type = "av", page = 2, pageSize = 30),
                ),
                api.searchCalls,
            )
        }
    }

    @Test
    fun `refreshAvState resets browse list back to first page`() = runTest {
        withMainDispatcher {
            val api = FakeHomeApiService(
                browsePages = mapOf(
                    "av" to mapOf(
                        1 to listOf(
                            avItem(id = "av-1", title = "IPX-901 雾面轮廓"),
                            avItem(id = "av-2", title = "IPX-902 逆光碎片"),
                        ),
                        2 to listOf(
                            avItem(id = "av-3", title = "IPX-903 冷调夜话"),
                        ),
                    ),
                ),
                browseTotalCounts = mapOf("av" to 3),
            )
            val viewModel = buildViewModel(api)
            viewModel.avSearchDebounceMs = 0

            viewModel.loadCategory("av")
            awaitUntil { viewModel.uiState.value.av.loaded && !viewModel.uiState.value.av.loading }
            viewModel.loadMoreAvIfNeeded(1)
            awaitUntil { viewModel.uiState.value.av.page == 2 && !viewModel.uiState.value.av.loadingMore }
            viewModel.refreshAvState()
            awaitUntil { viewModel.uiState.value.av.page == 1 && !viewModel.uiState.value.av.refreshing }

            val state = viewModel.uiState.value
            assertEquals(2, state.av.items.size)
            assertEquals(1, state.av.page)
            assertFalse(state.av.refreshing)
            assertTrue(state.av.hasMore)
            assertEquals(
                listOf(
                    SearchCall(keyword = "", type = "av", page = 1, pageSize = 30),
                    SearchCall(keyword = "", type = "av", page = 2, pageSize = 30),
                    SearchCall(keyword = "", type = "av", page = 1, pageSize = 30),
                ),
                api.searchCalls,
            )
        }
    }

    @Test
    fun `search results support load more and pull to refresh`() = runTest {
        withMainDispatcher {
            val api = FakeHomeApiService(
                browsePages = mapOf("av" to mapOf(1 to listOf(avItem(id = "av-0", title = "默认浏览")))),
                searchPages = mapOf(
                    "ssis" to mapOf(
                        1 to listOf(avItem(id = "av-1", title = "SSIS-111 潮汐边界")),
                        2 to listOf(avItem(id = "av-2", title = "SSIS-112 月相回声")),
                    ),
                ),
                searchTotalCounts = mapOf("ssis" to 2),
            )
            val viewModel = buildViewModel(api)
            viewModel.avSearchDebounceMs = 0

            viewModel.loadCategory("av")
            awaitUntil { viewModel.uiState.value.av.loaded && !viewModel.uiState.value.av.loading }
            viewModel.updateAvQuery("ssis")
            awaitUntil {
                viewModel.uiState.value.avSearch.isSearchMode &&
                    viewModel.uiState.value.avSearch.resultState.page == 1 &&
                    viewModel.uiState.value.avSearch.resultState.items.isNotEmpty()
            }
            viewModel.loadMoreAvIfNeeded(0)
            awaitUntil {
                viewModel.uiState.value.avSearch.resultState.page == 2 &&
                    !viewModel.uiState.value.avSearch.resultState.loadingMore
            }
            viewModel.refreshAvState()
            awaitUntil {
                viewModel.uiState.value.avSearch.resultState.page == 1 &&
                    !viewModel.uiState.value.avSearch.resultState.refreshing
            }

            val state = viewModel.uiState.value
            assertTrue(state.avSearch.isSearchMode)
            assertEquals(1, state.avSearch.resultState.page)
            assertEquals(1, state.avSearch.resultState.items.size)
            assertEquals("SSIS-111 潮汐边界", state.avSearch.resultState.items.first().title)
            assertFalse(state.avSearch.resultState.refreshing)
            assertEquals(
                listOf(
                    SearchCall(keyword = "", type = "av", page = 1, pageSize = 30),
                    SearchCall(keyword = "ssis", type = "av", page = 1, pageSize = 30),
                    SearchCall(keyword = "ssis", type = "av", page = 2, pageSize = 30),
                    SearchCall(keyword = "ssis", type = "av", page = 1, pageSize = 30),
                ),
                api.searchCalls,
            )
        }
    }

    @Test
    fun `browse load more failure keeps loaded items and exposes footer error`() = runTest {
        withMainDispatcher {
            val api = FakeHomeApiService(
                browsePages = mapOf(
                    "av" to mapOf(
                        1 to listOf(avItem(id = "av-1", title = "CAWD-808 白昼残响")),
                    ),
                ),
                browseTotalCounts = mapOf("av" to 2),
                browseErrors = mapOf(BrowseErrorKey(type = "av", page = 2) to IllegalStateException("补货失败")),
            )
            val viewModel = buildViewModel(api)
            viewModel.avSearchDebounceMs = 0

            viewModel.loadCategory("av")
            awaitUntil { viewModel.uiState.value.av.loaded && !viewModel.uiState.value.av.loading }
            viewModel.loadMoreAvIfNeeded(0)
            awaitUntil {
                !viewModel.uiState.value.av.loadingMore &&
                    !viewModel.uiState.value.av.loadMoreErrorMessage.isNullOrBlank()
            }

            val state = viewModel.uiState.value
            assertEquals(1, state.av.items.size)
            assertEquals("补货失败", state.av.loadMoreErrorMessage)
            assertEquals(1, state.av.page)
            assertTrue(state.av.hasMore)
            assertFalse(state.av.loadingMore)
        }
    }

    private suspend fun buildViewModel(api: FakeHomeApiService): HomeViewModel {
        val store = AppPreferencesStore(
            dataStore = PreferenceDataStoreFactory.create(
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

private data class SearchCall(
    val keyword: String,
    val type: String,
    val page: Int,
    val pageSize: Int,
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
    private val browsePages: Map<String, Map<Int, List<VideoListItemDto>>> = emptyMap(),
    private val browseTotalCounts: Map<String, Int> = emptyMap(),
    private val browseErrors: Map<BrowseErrorKey, Throwable> = emptyMap(),
    private val searchPages: Map<String, Map<Int, List<VideoListItemDto>>> = emptyMap(),
    private val searchTotalCounts: Map<String, Int> = emptyMap(),
    private val searchErrors: Map<SearchErrorKey, Throwable> = emptyMap(),
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
        searchCalls += SearchCall(keyword = normalizedKeyword, type = type, page = page, pageSize = pageSize)
        val items = if (normalizedKeyword.isBlank()) {
            browseErrors[BrowseErrorKey(type = type, page = page)]?.let { throw it }
            browsePages[type]?.get(page).orEmpty()
        } else {
            searchErrors[SearchErrorKey(keyword = normalizedKeyword, page = page)]?.let { throw it }
            searchPages[normalizedKeyword]?.get(page).orEmpty()
        }
        val totalCount = if (normalizedKeyword.isBlank()) {
            browseTotalCounts[type] ?: browsePages[type]?.values?.sumOf { it.size } ?: items.size
        } else {
            searchTotalCounts[normalizedKeyword] ?: searchPages[normalizedKeyword]?.values?.sumOf { it.size } ?: items.size
        }
        return ApiEnvelope(
            code = 0,
            data = SearchPayload(
                items = items,
                totalCount = totalCount,
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

    override suspend fun detail(
        url: String,
        authorization: String,
    ): ApiEnvelope<VideoDetailDto> = error("unused")

    override suspend fun actorDetail(
        url: String,
        authorization: String,
        page: Int,
        pageSize: Int,
    ): ApiEnvelope<ActorWorksPayload> = error("unused")

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

private data class BrowseErrorKey(
    val type: String,
    val page: Int,
)

private data class SearchErrorKey(
    val keyword: String,
    val page: Int,
)
