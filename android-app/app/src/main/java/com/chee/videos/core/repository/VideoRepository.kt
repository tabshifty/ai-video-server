package com.chee.videos.core.repository

import com.chee.videos.core.data.AppPreferencesStore
import com.chee.videos.core.model.ActionTogglePayload
import com.chee.videos.core.model.ApiEnvelope
import com.chee.videos.core.model.AppException
import com.chee.videos.core.model.AuthExpiredException
import com.chee.videos.core.model.ContinueHistoryPayload
import com.chee.videos.core.model.FeedVideoDto
import com.chee.videos.core.model.ImageCollectionDetailDto
import com.chee.videos.core.model.ImageCollectionsPayload
import com.chee.videos.core.model.RecordHistoryRequest
import com.chee.videos.core.model.SearchPayload
import com.chee.videos.core.model.TvHomePayload
import com.chee.videos.core.model.TvSearchPayload
import com.chee.videos.core.model.TvSeriesDetailDto
import com.chee.videos.core.model.UserProfileDto
import com.chee.videos.core.model.VideoDetailDto
import com.chee.videos.core.model.VideoListItemDto
import com.chee.videos.core.network.ApiService
import com.chee.videos.core.player.PlaybackProfile
import com.chee.videos.core.player.PlaybackProfileResolver
import com.chee.videos.core.util.UrlBuilder
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class VideoRepository @Inject constructor(
    private val api: ApiService,
    private val store: AppPreferencesStore,
    private val authRepository: AuthRepository,
    private val playbackProfileResolver: PlaybackProfileResolver,
) {
    suspend fun fetchShortFeed(
        pageSize: Int = 20,
        excludeIds: List<String> = emptyList(),
    ): Result<List<FeedVideoDto>> {
        val baseUrl = store.readActiveBaseUrl()
            ?: return Result.failure(AppException("请先配置服务器地址"))

        return runCatching {
            val resp = api.randomShort(
                url = UrlBuilder.randomShort(baseUrl),
                pageSize = pageSize,
                excludeIds = excludeIds.distinct().filter { it.isNotBlank() }.takeIf { it.isNotEmpty() }?.joinToString(","),
            )
            if (resp.code != 0 || resp.data == null) {
                throw AppException(resp.msg.ifBlank { "短视频加载失败(code=${resp.code})" })
            }
            resp.data.items
        }
    }

    suspend fun fetchCategory(type: String, page: Int = 1, pageSize: Int = 30): Result<SearchPayload> {
        return callWithAuth { baseUrl, bearer ->
            api.search(
                url = UrlBuilder.search(baseUrl),
                authorization = bearer,
                keyword = "",
                type = type,
                page = page,
                pageSize = pageSize,
            )
        }
    }

    suspend fun searchAv(query: String, page: Int = 1, pageSize: Int = 30): Result<SearchPayload> {
        return callWithAuth { baseUrl, bearer ->
            api.search(
                url = UrlBuilder.search(baseUrl),
                authorization = bearer,
                keyword = query.trim(),
                type = "av",
                page = page,
                pageSize = pageSize,
            )
        }
    }

    suspend fun searchShort(query: String, page: Int = 1, pageSize: Int = 30): Result<SearchPayload> {
        return callWithAuth { baseUrl, bearer ->
            api.search(
                url = UrlBuilder.search(baseUrl),
                authorization = bearer,
                keyword = query.trim(),
                type = "short",
                page = page,
                pageSize = pageSize,
            )
        }
    }

    suspend fun fetchDetail(videoId: String): Result<VideoDetailDto> {
        return callWithAuth { baseUrl, bearer ->
            api.detail(UrlBuilder.detail(baseUrl, videoId), bearer)
        }
    }

    suspend fun fetchShortDiscover(
        mode: String,
        value: String,
        page: Int = 1,
        pageSize: Int = 30,
    ): Result<SearchPayload> {
        val normalizedMode = mode.trim().lowercase()
        val normalizedValue = value.trim()
        return callWithAuth { baseUrl, bearer ->
            when (normalizedMode) {
                "tag" -> api.shortDiscover(
                    url = UrlBuilder.shortDiscover(baseUrl),
                    authorization = bearer,
                    mode = "tag",
                    tag = normalizedValue,
                    collectionID = null,
                    page = page,
                    pageSize = pageSize,
                )

                "collection" -> api.shortDiscover(
                    url = UrlBuilder.shortDiscover(baseUrl),
                    authorization = bearer,
                    mode = "collection",
                    tag = null,
                    collectionID = normalizedValue,
                    page = page,
                    pageSize = pageSize,
                )

                else -> throw AppException("不支持的发现模式: $mode")
            }
        }
    }

    suspend fun fetchTvHome(
        query: String = "",
        page: Int = 1,
        pageSize: Int = 20,
    ): Result<TvHomePayload> {
        return callWithAuth { baseUrl, bearer ->
            api.tvHome(
                url = UrlBuilder.tvHome(baseUrl),
                authorization = bearer,
                keyword = query.trim().takeIf { it.isNotBlank() },
                page = page,
                pageSize = pageSize,
            )
        }
    }

    suspend fun fetchTvSeriesDetail(seriesId: String): Result<TvSeriesDetailDto> {
        return callWithAuth { baseUrl, bearer ->
            api.tvSeriesDetail(
                url = UrlBuilder.tvSeriesDetail(baseUrl, seriesId),
                authorization = bearer,
            )
        }
    }

    suspend fun fetchTvSearch(
        query: String,
        page: Int = 1,
        pageSize: Int = 20,
    ): Result<TvSearchPayload> {
        return callWithAuth { baseUrl, bearer ->
            api.tvSearch(
                url = UrlBuilder.tvSearch(baseUrl),
                authorization = bearer,
                keyword = query.trim(),
                page = page,
                pageSize = pageSize,
            )
        }
    }

    suspend fun fetchImageCollections(
        query: String? = null,
        page: Int = 1,
        pageSize: Int = 20,
    ): Result<ImageCollectionsPayload> {
        return callWithAuth { baseUrl, bearer ->
            api.imageCollections(
                url = UrlBuilder.imageCollections(baseUrl),
                authorization = bearer,
                keyword = query,
                page = page,
                pageSize = pageSize,
            )
        }
    }

    suspend fun fetchImageCollectionDetail(collectionId: String): Result<ImageCollectionDetailDto> {
        return callWithAuth { baseUrl, bearer ->
            api.imageCollectionDetail(
                url = UrlBuilder.imageCollectionDetail(baseUrl, collectionId),
                authorization = bearer,
            )
        }
    }

    suspend fun toggleLike(videoId: String): Result<ActionTogglePayload> {
        return callWithAuth { baseUrl, bearer ->
            api.toggleLike(UrlBuilder.toggleLike(baseUrl, videoId), bearer)
        }
    }

    suspend fun toggleFavorite(videoId: String): Result<ActionTogglePayload> {
        return callWithAuth { baseUrl, bearer ->
            api.toggleFavorite(UrlBuilder.toggleFavorite(baseUrl, videoId), bearer)
        }
    }

    suspend fun toggleDislike(videoId: String): Result<ActionTogglePayload> {
        return callWithAuth { baseUrl, bearer ->
            api.toggleDislike(UrlBuilder.toggleDislike(baseUrl, videoId), bearer)
        }
    }

    suspend fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) {
        callWithAuth { baseUrl, bearer ->
            api.recordHistory(
                url = UrlBuilder.history(baseUrl),
                authorization = bearer,
                body = RecordHistoryRequest(videoId, watchSeconds, completed),
            )
        }
    }

    suspend fun fetchContinueHistory(page: Int = 1, limit: Int = 30): Result<ContinueHistoryPayload> {
        return callWithAuth { baseUrl, bearer ->
            api.continueHistory(
                url = UrlBuilder.historyContinue(baseUrl),
                authorization = bearer,
                page = page,
                limit = limit,
            )
        }
    }

    suspend fun fetchLikedVideos(page: Int = 1, pageSize: Int = 30): Result<List<VideoListItemDto>> {
        return callWithAuth { baseUrl, bearer ->
            api.likedVideos(
                url = UrlBuilder.likedVideos(baseUrl),
                authorization = bearer,
                page = page,
                pageSize = pageSize,
            )
        }.map { it.items }
    }

    suspend fun fetchFavoritedVideos(page: Int = 1, pageSize: Int = 30): Result<List<VideoListItemDto>> {
        return callWithAuth { baseUrl, bearer ->
            api.favoritedVideos(
                url = UrlBuilder.favoritedVideos(baseUrl),
                authorization = bearer,
                page = page,
                pageSize = pageSize,
            )
        }.map { it.items }
    }

    suspend fun fetchUserProfile(): Result<UserProfileDto> {
        return callWithAuth { baseUrl, bearer ->
            api.userProfile(
                url = UrlBuilder.userProfile(baseUrl),
                authorization = bearer,
            )
        }
    }

    suspend fun readActiveBaseUrl(): String? = store.readActiveBaseUrl()

    suspend fun readAccessToken(): String? = store.readAccessToken()

    suspend fun readTvSubtitlePreference(videoId: String): String? =
        store.readTvSubtitlePreference(videoId)

    suspend fun saveTvSubtitlePreference(videoId: String, subtitleTrackId: String?) {
        store.saveTvSubtitlePreference(videoId, subtitleTrackId)
    }

    fun preferredLongFormPlaybackProfile(): PlaybackProfile =
        playbackProfileResolver.preferredLongFormProfile()

    suspend fun buildSourceUrl(videoId: String): String {
        val baseUrl = store.readActiveBaseUrl().orEmpty()
        return UrlBuilder.source(baseUrl, videoId, preferredLongFormPlaybackProfile().wireValue)
    }

    private suspend fun <T> callWithAuth(
        block: suspend (baseUrl: String, authorization: String) -> ApiEnvelope<T>,
    ): Result<T> {
        val baseUrl = store.readActiveBaseUrl()
            ?: return Result.failure(AppException("请先配置服务器地址"))

        var accessToken = store.readAccessToken()
            ?: return Result.failure(AuthExpiredException())

        var resp = runCatching {
            block(baseUrl, "Bearer $accessToken")
        }.getOrElse { return Result.failure(it) }

        if (resp.code == 401) {
            val refreshed = authRepository.refreshTokenIfPossible()
            if (!refreshed) {
                return Result.failure(AuthExpiredException())
            }
            accessToken = store.readAccessToken()
                ?: return Result.failure(AuthExpiredException())
            resp = runCatching {
                block(baseUrl, "Bearer $accessToken")
            }.getOrElse { return Result.failure(it) }
        }

        val data = resp.data
        if (resp.code != 0 || data == null) {
            return Result.failure(AppException(resp.msg.ifBlank { "请求失败(code=${resp.code})" }))
        }
        return Result.success(data)
    }
}
