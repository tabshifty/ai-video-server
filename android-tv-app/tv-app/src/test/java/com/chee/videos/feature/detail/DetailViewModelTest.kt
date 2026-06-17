package com.chee.videos.feature.detail

import androidx.datastore.preferences.core.PreferenceDataStoreFactory
import androidx.lifecycle.SavedStateHandle
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
import com.chee.videos.core.model.TvTrackPreference
import com.chee.videos.core.model.RefreshPayload
import com.chee.videos.core.model.SearchPayload
import com.chee.videos.core.model.SessionTokens
import com.chee.videos.core.model.TvAuthCreateEnvelope
import com.chee.videos.core.model.TvAuthSessionCreateRequest
import com.chee.videos.core.model.TvAuthStatusEnvelope
import com.chee.videos.core.model.TvCatalogWallPayload
import com.chee.videos.core.model.TvHomePayload
import com.chee.videos.core.model.TvIptvPayload
import com.chee.videos.core.model.TvSearchPayload
import com.chee.videos.core.model.TvSeriesDetailDto
import com.chee.videos.core.model.UserStateDto
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
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.test.advanceUntilIdle
import kotlinx.coroutines.test.runCurrent
import kotlinx.coroutines.test.runTest
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test

@OptIn(ExperimentalCoroutinesApi::class)
class DetailViewModelTest {
    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    @Test
    fun reportHistory_withPositiveProgress_callsRepository() = runTest {
        val api = FakeDetailApiService()
        val viewModel = buildDetailViewModel(api, backgroundScope)

        viewModel.reportHistory("video-1", watchSeconds = 42, completed = false)
        mainDispatcherRule.scheduler.runCurrent()
        advanceUntilIdle()

        assertEquals(listOf(RecordHistoryRequest("video-1", 42, false)), api.historyRequests)
    }

    @Test
    fun reportHistory_ignoresBlankVideoIdAndZeroProgress() = runTest {
        val api = FakeDetailApiService()
        val viewModel = buildDetailViewModel(api, backgroundScope)

        viewModel.reportHistory("", watchSeconds = 42, completed = false)
        viewModel.reportHistory("video-1", watchSeconds = 0, completed = false)
        mainDispatcherRule.scheduler.runCurrent()
        advanceUntilIdle()

        assertEquals(emptyList<RecordHistoryRequest>(), api.historyRequests)
    }

    @Test
    fun tvAudioPreference_readsAndSavesLanguagePreferenceThroughRepository() = runTest {
        val api = FakeDetailApiService()
        val viewModel = buildDetailViewModel(api, backgroundScope)

        assertEquals(null, viewModel.readTvAudioPreference("video-1"))

        val preference = TvTrackPreference(language = "zh", type = "default")
        viewModel.saveTvAudioPreference("video-1", preference)
        mainDispatcherRule.scheduler.runCurrent()
        advanceUntilIdle()

        assertEquals(preference, viewModel.readTvAudioPreference("video-1"))
    }

    @Test
    fun reload_withExistingDetailKeepsContentAndUsesRefreshingState() = runTest {
        val api = DelayedDetailApiService()
        val viewModel = buildDetailViewModel(api, backgroundScope)

        runCurrent()
        api.completeDetail(videoDetail(title = "旧详情"))
        advanceUntilIdle()

        viewModel.load()
        runCurrent()

        val refreshingState = viewModel.uiState.value
        assertFalse(refreshingState.loading)
        assertTrue(refreshingState.refreshing)
        assertEquals("旧详情", refreshingState.detail?.title)

        api.completeDetail(videoDetail(title = "新详情"))
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.loading)
        assertFalse(state.refreshing)
        assertEquals("新详情", state.detail?.title)
    }

    @Test
    fun reload_failureWithExistingDetailKeepsOldContent() = runTest {
        val api = DelayedDetailApiService()
        val viewModel = buildDetailViewModel(api, backgroundScope)

        runCurrent()
        api.completeDetail(videoDetail(title = "旧详情"))
        advanceUntilIdle()

        viewModel.load()
        runCurrent()
        api.completeDetailFailure(IllegalStateException("详情刷新失败"))
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertFalse(state.loading)
        assertFalse(state.refreshing)
        assertEquals("旧详情", state.detail?.title)
        assertEquals("详情刷新失败", state.errorMessage)
    }

    @Test
    fun staleReloadResponse_doesNotOverrideNewerDetailState() = runTest {
        val api = DelayedDetailApiService()
        val viewModel = buildDetailViewModel(api, backgroundScope)

        runCurrent()
        api.completeDetail(videoDetail(title = "基础详情"))
        advanceUntilIdle()

        viewModel.load()
        runCurrent()
        viewModel.load()
        runCurrent()

        api.completeLatestDetail(videoDetail(title = "最新详情"))
        advanceUntilIdle()
        api.completeDetail(videoDetail(title = "过期详情"))
        advanceUntilIdle()

        val state = viewModel.uiState.value
        assertEquals("最新详情", state.detail?.title)
    }

    private suspend fun buildDetailViewModel(
        api: ApiService,
        dataStoreScope: CoroutineScope,
    ): DetailViewModel {
        val store = AppPreferencesStore(
            dataStore = PreferenceDataStoreFactory.create(
                scope = dataStoreScope,
                produceFile = {
                    File.createTempFile("detail-view-model", ".preferences_pb").apply {
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
        return DetailViewModel(
            savedStateHandle = SavedStateHandle(mapOf("videoId" to "video-1", "videoType" to "movie")),
            videoRepository = videoRepository,
            authRepository = authRepository,
        )
    }
}

private fun videoDetail(
    id: String = "video-1",
    title: String,
): VideoDetailDto = VideoDetailDto(
    id = id,
    title = title,
    description = "详情描述",
    userState = UserStateDto(),
)

private class FakeDetailApiService : ApiService {
    val historyRequests = mutableListOf<RecordHistoryRequest>()

    override suspend fun detail(
        url: String,
        authorization: String,
    ): ApiEnvelope<VideoDetailDto> = ApiEnvelope(
        code = 0,
        data = VideoDetailDto(
            id = "video-1",
            title = "电影",
        ),
    )

    override suspend fun recordHistory(
        url: String,
        authorization: String,
        body: RecordHistoryRequest,
    ): ApiEnvelope<Map<String, Boolean>> {
        historyRequests += body
        return ApiEnvelope(code = 0, data = mapOf("ok" to true))
    }

    override suspend fun refresh(url: String, authorization: String): ApiEnvelope<RefreshPayload> =
        ApiEnvelope(code = 401, msg = "expired")

    override suspend fun login(url: String, body: LoginRequest): ApiEnvelope<LoginPayload> = error("unused")

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

    override suspend fun search(
        url: String,
        authorization: String,
        keyword: String,
        type: String,
        page: Int,
        pageSize: Int,
    ): ApiEnvelope<SearchPayload> = error("unused")

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
        sortBy: String,
        sortOrder: String,
    ): ApiEnvelope<TvCatalogWallPayload> = error("unused")

    override suspend fun tvIptvChannels(
        url: String,
        authorization: String,
    ): ApiEnvelope<TvIptvPayload> = error("unused")

    override suspend fun tvSeriesDetail(
        url: String,
        authorization: String,
    ): ApiEnvelope<TvSeriesDetailDto> = error("unused")

    override suspend fun createTvAuthSession(
        url: String,
        body: TvAuthSessionCreateRequest,
    ): TvAuthCreateEnvelope = error("unused")

    override suspend fun getTvAuthSession(url: String): TvAuthStatusEnvelope = error("unused")

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

private class DelayedDetailApiService : ApiService {
    private val pendingDetailRequests = ArrayDeque<CompletableDeferred<ApiEnvelope<VideoDetailDto>>>()
    val historyRequests = mutableListOf<RecordHistoryRequest>()

    override suspend fun detail(
        url: String,
        authorization: String,
    ): ApiEnvelope<VideoDetailDto> {
        val deferred = CompletableDeferred<ApiEnvelope<VideoDetailDto>>()
        pendingDetailRequests.addLast(deferred)
        return deferred.await()
    }

    override suspend fun recordHistory(
        url: String,
        authorization: String,
        body: RecordHistoryRequest,
    ): ApiEnvelope<Map<String, Boolean>> {
        historyRequests += body
        return ApiEnvelope(code = 0, data = mapOf("ok" to true))
    }

    override suspend fun refresh(url: String, authorization: String): ApiEnvelope<RefreshPayload> =
        ApiEnvelope(code = 401, msg = "expired")

    override suspend fun login(url: String, body: LoginRequest): ApiEnvelope<LoginPayload> = error("unused")

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

    override suspend fun search(
        url: String,
        authorization: String,
        keyword: String,
        type: String,
        page: Int,
        pageSize: Int,
    ): ApiEnvelope<SearchPayload> = error("unused")

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
        sortBy: String,
        sortOrder: String,
    ): ApiEnvelope<TvCatalogWallPayload> = error("unused")

    override suspend fun tvIptvChannels(
        url: String,
        authorization: String,
    ): ApiEnvelope<TvIptvPayload> = error("unused")

    override suspend fun tvSeriesDetail(
        url: String,
        authorization: String,
    ): ApiEnvelope<TvSeriesDetailDto> = error("unused")

    override suspend fun createTvAuthSession(
        url: String,
        body: TvAuthSessionCreateRequest,
    ): TvAuthCreateEnvelope = error("unused")

    override suspend fun getTvAuthSession(url: String): TvAuthStatusEnvelope = error("unused")

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

    fun completeDetail(detail: VideoDetailDto) {
        pendingDetailRequests.removeFirst().complete(ApiEnvelope(code = 0, data = detail))
    }

    fun completeLatestDetail(detail: VideoDetailDto) {
        pendingDetailRequests.removeLast().complete(ApiEnvelope(code = 0, data = detail))
    }

    fun completeDetailFailure(error: Throwable = IllegalStateException("详情加载失败")) {
        pendingDetailRequests.removeFirst().completeExceptionally(error)
    }
}
