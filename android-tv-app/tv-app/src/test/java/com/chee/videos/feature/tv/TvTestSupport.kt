package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvContinueWatchingDto
import com.chee.videos.core.model.TvCatalogWallItemDto
import com.chee.videos.core.model.TvCatalogWallPayload
import com.chee.videos.core.model.TvEpisodeDto
import com.chee.videos.core.model.FeedVideoDto
import com.chee.videos.core.model.TvHomePayload
import com.chee.videos.core.model.TvIptvPayload
import com.chee.videos.core.model.TvSearchPayload
import com.chee.videos.core.model.TvSearchResultDto
import com.chee.videos.core.model.TvSeasonDto
import com.chee.videos.core.model.TvSeriesDetailDto
import com.chee.videos.core.model.TvSeriesSummaryDto
import com.chee.videos.core.model.VideoActorDto
import com.chee.videos.core.model.TvTrackPreference
import com.chee.videos.core.model.SubtitleTrackDto
import kotlinx.coroutines.CompletableDeferred

typealias MainDispatcherRule = com.chee.videos.core.testing.MainDispatcherRule

class FakeTvRepository(
    private val baseUrl: String = "https://example.com",
    private val homePayload: TvHomePayload = TvHomePayload(),
    private val homePayloads: Map<String, TvHomePayload> = emptyMap(),
    private val searchPayload: TvSearchPayload = TvSearchPayload(),
    private val iptvPayload: TvIptvPayload = TvIptvPayload(),
    private val shortFeedItems: List<FeedVideoDto> = emptyList(),
    private val shortFeedPages: List<List<FeedVideoDto>> = emptyList(),
    private val posterWallPages: List<TvCatalogWallPayload> = emptyList(),
    private val detailPayload: TvSeriesDetailDto = tvSeriesDetail(),
    private val homeError: Throwable? = null,
    private val detailError: Throwable? = null,
    private val sourceUrlError: Throwable? = null,
    private var tvSeekStepSeconds: Int = 10,
    private var tvSeriesAutoplayEnabled: Boolean? = null,
    subtitlePreferences: Map<String, TvTrackPreference> = emptyMap(),
    audioPreferences: Map<String, TvTrackPreference> = emptyMap(),
) : TvRepository {
    private val storedSubtitlePreferences = subtitlePreferences.toMutableMap()
    private val storedAudioPreferences = audioPreferences.toMutableMap()
    val historyReports = mutableListOf<TvHistoryReport>()
    val homeRequests = mutableListOf<TvHomeRequest>()
    val searchRequests = mutableListOf<TvSearchRequest>()
    val shortFeedRequests = mutableListOf<TvShortFeedRequest>()
    val posterWallRequests = mutableListOf<TvPosterWallRequest>()
    val detailRequests = mutableListOf<String>()
    val sourceUrlRequests = mutableListOf<TvSourceUrlRequest>()
    private var shortFeedRequestIndex = 0

    override suspend fun fetchHome(kind: String, query: String, page: Int, pageSize: Int): Result<TvHomePayload> {
        homeRequests += TvHomeRequest(kind = kind, query = query, page = page, pageSize = pageSize)
        homeError?.let { return Result.failure(it) }
        return Result.success(homePayloads[kind] ?: homePayload)
    }

    override suspend fun fetchSearch(query: String, page: Int, pageSize: Int): Result<TvSearchPayload> {
        searchRequests += TvSearchRequest(query = query, page = page, pageSize = pageSize)
        homeError?.let { return Result.failure(it) }
        return Result.success(searchPayload)
    }

    override suspend fun fetchCatalogWall(
        kind: String,
        page: Int,
        pageSize: Int,
        sortBy: String,
        sortOrder: String,
    ): Result<TvCatalogWallPayload> {
        posterWallRequests += TvPosterWallRequest(
            kind = kind,
            page = page,
            pageSize = pageSize,
            sortBy = sortBy,
            sortOrder = sortOrder,
        )
        homeError?.let { return Result.failure(it) }
        val payload = posterWallPages.firstOrNull { it.page == page }
            ?: TvCatalogWallPayload(page = page, pageSize = pageSize)
        return Result.success(payload)
    }

    override suspend fun fetchIptvChannels(): Result<TvIptvPayload> {
        homeError?.let { return Result.failure(it) }
        return Result.success(iptvPayload)
    }

    override suspend fun fetchShortFeed(pageSize: Int, excludeIds: List<String>): Result<List<FeedVideoDto>> {
        shortFeedRequests += TvShortFeedRequest(pageSize = pageSize, excludeIds = excludeIds)
        homeError?.let { return Result.failure(it) }
        val response = shortFeedPages.getOrNull(shortFeedRequestIndex) ?: shortFeedItems
        shortFeedRequestIndex += 1
        return Result.success(response.take(pageSize))
    }

    override suspend fun fetchSeriesDetail(seriesId: String): Result<TvSeriesDetailDto> {
        detailRequests += seriesId
        detailError?.let { return Result.failure(it) }
        return Result.success(detailPayload)
    }

    override suspend fun readActiveBaseUrl(): String = baseUrl

    override suspend fun buildSourceUrl(videoId: String, profile: String?): String {
        sourceUrlRequests += TvSourceUrlRequest(videoId = videoId, profile = profile)
        sourceUrlError?.let { throw it }
        return "https://example.com/$videoId.m3u8"
    }

    override suspend fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) {
        historyReports += TvHistoryReport(videoId, watchSeconds, completed)
    }

    override suspend fun readTvSubtitlePreference(videoId: String): TvTrackPreference? = storedSubtitlePreferences[videoId]

    override suspend fun saveTvSubtitlePreference(videoId: String, preference: TvTrackPreference?) {
        if (preference == null) {
            storedSubtitlePreferences.remove(videoId)
        } else {
            storedSubtitlePreferences[videoId] = preference
        }
    }

    override suspend fun readTvAudioPreference(videoId: String): TvTrackPreference? = storedAudioPreferences[videoId]

    override suspend fun saveTvAudioPreference(videoId: String, preference: TvTrackPreference?) {
        if (preference == null) {
            storedAudioPreferences.remove(videoId)
        } else {
            storedAudioPreferences[videoId] = preference
        }
    }

    override suspend fun readTvSeekStepSeconds(): Int = tvSeekStepSeconds

    override suspend fun saveTvSeekStepSeconds(seconds: Int) {
        tvSeekStepSeconds = seconds
    }

    override suspend fun readTvSeriesAutoplayEnabled(): Boolean? = tvSeriesAutoplayEnabled

    override suspend fun saveTvSeriesAutoplayEnabled(enabled: Boolean) {
        tvSeriesAutoplayEnabled = enabled
    }
}

data class TvHistoryReport(
    val videoId: String,
    val watchSeconds: Int,
    val completed: Boolean,
)

data class TvHomeRequest(
    val kind: String,
    val query: String,
    val page: Int,
    val pageSize: Int,
)

data class TvSearchRequest(
    val query: String,
    val page: Int,
    val pageSize: Int,
)

data class TvShortFeedRequest(
    val pageSize: Int,
    val excludeIds: List<String>,
)

data class TvPosterWallRequest(
    val kind: String,
    val page: Int,
    val pageSize: Int,
    val sortBy: String,
    val sortOrder: String,
)

data class TvSourceUrlRequest(
    val videoId: String,
    val profile: String?,
)

class DelayedSourceTvRepository(
    private val baseUrl: String = "https://example.com",
    private val detailPayload: TvSeriesDetailDto = tvSeriesDetail(),
) : TvRepository {
    private val pendingSourceUrls = linkedMapOf<String, CompletableDeferred<String>>()

    override suspend fun fetchHome(kind: String, query: String, page: Int, pageSize: Int): Result<TvHomePayload> =
        Result.success(TvHomePayload())

    override suspend fun fetchSearch(query: String, page: Int, pageSize: Int): Result<TvSearchPayload> =
        Result.success(TvSearchPayload())

    override suspend fun fetchCatalogWall(
        kind: String,
        page: Int,
        pageSize: Int,
        sortBy: String,
        sortOrder: String,
    ): Result<TvCatalogWallPayload> =
        Result.success(TvCatalogWallPayload(page = page, pageSize = pageSize))

    override suspend fun fetchIptvChannels(): Result<TvIptvPayload> =
        Result.success(TvIptvPayload())

    override suspend fun fetchShortFeed(pageSize: Int, excludeIds: List<String>): Result<List<FeedVideoDto>> =
        Result.success(emptyList())

    override suspend fun fetchSeriesDetail(seriesId: String): Result<TvSeriesDetailDto> =
        Result.success(detailPayload)

    override suspend fun readActiveBaseUrl(): String = baseUrl

    override suspend fun buildSourceUrl(videoId: String, profile: String?): String =
        pendingSourceUrls.getOrPut(videoId) { CompletableDeferred() }.await()

    override suspend fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) = Unit

    override suspend fun readTvSubtitlePreference(videoId: String): TvTrackPreference? = null

    override suspend fun saveTvSubtitlePreference(videoId: String, preference: TvTrackPreference?) = Unit

    override suspend fun readTvAudioPreference(videoId: String): TvTrackPreference? = null

    override suspend fun saveTvAudioPreference(videoId: String, preference: TvTrackPreference?) = Unit

    override suspend fun readTvSeekStepSeconds(): Int = 10

    override suspend fun saveTvSeekStepSeconds(seconds: Int) = Unit

    override suspend fun readTvSeriesAutoplayEnabled(): Boolean? = null

    override suspend fun saveTvSeriesAutoplayEnabled(enabled: Boolean) = Unit

    fun completeSourceUrl(videoId: String, url: String = "https://example.com/$videoId.m3u8") {
        pendingSourceUrls.getOrPut(videoId) { CompletableDeferred() }.complete(url)
    }

    fun failSourceUrl(videoId: String, message: String) {
        pendingSourceUrls.getOrPut(videoId) { CompletableDeferred() }.completeExceptionally(IllegalStateException(message))
    }
}

class DelayedCatalogTvRepository(
    private val baseUrl: String = "https://example.com",
) : TvRepository {
    private val pendingHomes = linkedMapOf<String, CompletableDeferred<Result<TvHomePayload>>>()
    private val pendingSearches = linkedMapOf<String, CompletableDeferred<Result<TvSearchPayload>>>()
    private val pendingPosterWalls = linkedMapOf<TvPosterWallRequestKey, ArrayDeque<CompletableDeferred<Result<TvCatalogWallPayload>>>>()
    val homeRequests = mutableListOf<TvHomeRequest>()
    val searchRequests = mutableListOf<TvSearchRequest>()
    val posterWallRequests = mutableListOf<TvPosterWallRequest>()

    override suspend fun fetchHome(kind: String, query: String, page: Int, pageSize: Int): Result<TvHomePayload> {
        homeRequests += TvHomeRequest(kind = kind, query = query, page = page, pageSize = pageSize)
        return pendingHomes.getOrPut(kind) { CompletableDeferred() }.await()
    }

    override suspend fun fetchSearch(query: String, page: Int, pageSize: Int): Result<TvSearchPayload> {
        searchRequests += TvSearchRequest(query = query, page = page, pageSize = pageSize)
        return pendingSearches.getOrPut(query) { CompletableDeferred() }.await()
    }

    override suspend fun fetchCatalogWall(
        kind: String,
        page: Int,
        pageSize: Int,
        sortBy: String,
        sortOrder: String,
    ): Result<TvCatalogWallPayload> {
        posterWallRequests += TvPosterWallRequest(
            kind = kind,
            page = page,
            pageSize = pageSize,
            sortBy = sortBy,
            sortOrder = sortOrder,
        )
        val key = TvPosterWallRequestKey(kind = kind, page = page, sortBy = sortBy, sortOrder = sortOrder)
        val deferred = CompletableDeferred<Result<TvCatalogWallPayload>>()
        pendingPosterWalls.getOrPut(key) { ArrayDeque() }.addLast(deferred)
        return deferred.await()
    }

    override suspend fun fetchIptvChannels(): Result<TvIptvPayload> =
        Result.success(TvIptvPayload())

    override suspend fun fetchShortFeed(pageSize: Int, excludeIds: List<String>): Result<List<FeedVideoDto>> =
        Result.success(emptyList())

    override suspend fun fetchSeriesDetail(seriesId: String): Result<TvSeriesDetailDto> =
        Result.success(tvSeriesDetail(id = seriesId))

    override suspend fun readActiveBaseUrl(): String = baseUrl

    override suspend fun buildSourceUrl(videoId: String, profile: String?): String =
        "https://example.com/$videoId.m3u8"

    override suspend fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) = Unit

    override suspend fun readTvSubtitlePreference(videoId: String): TvTrackPreference? = null

    override suspend fun saveTvSubtitlePreference(videoId: String, preference: TvTrackPreference?) = Unit

    override suspend fun readTvAudioPreference(videoId: String): TvTrackPreference? = null

    override suspend fun saveTvAudioPreference(videoId: String, preference: TvTrackPreference?) = Unit

    override suspend fun readTvSeekStepSeconds(): Int = 10

    override suspend fun saveTvSeekStepSeconds(seconds: Int) = Unit

    override suspend fun readTvSeriesAutoplayEnabled(): Boolean? = null

    override suspend fun saveTvSeriesAutoplayEnabled(enabled: Boolean) = Unit

    fun completeHome(kind: String = "tv", payload: TvHomePayload = TvHomePayload()) {
        pendingHomes.getOrPut(kind) { CompletableDeferred() }.complete(Result.success(payload))
    }

    fun completeHomeFailure(kind: String = "tv", error: Throwable = IllegalStateException("TV 首页加载失败")) {
        pendingHomes.getOrPut(kind) { CompletableDeferred() }.complete(Result.failure(error))
    }

    fun completeSearch(query: String, payload: TvSearchPayload = TvSearchPayload()) {
        pendingSearches.getOrPut(query) { CompletableDeferred() }.complete(Result.success(payload))
    }

    fun completeSearchFailure(query: String, error: Throwable = IllegalStateException("TV 搜索失败")) {
        pendingSearches.getOrPut(query) { CompletableDeferred() }.complete(Result.failure(error))
    }

    fun completePosterWall(
        kind: String = "tv",
        page: Int = 1,
        sortBy: String = "added",
        sortOrder: String = "desc",
        payload: TvCatalogWallPayload = TvCatalogWallPayload(page = page, pageSize = 24),
    ) {
        val key = TvPosterWallRequestKey(kind = kind, page = page, sortBy = sortBy, sortOrder = sortOrder)
        pendingPosterWalls.getOrPut(key) { ArrayDeque() }
            .removeFirst()
            .complete(Result.success(payload))
    }

    fun completePosterWallFailure(
        kind: String = "tv",
        page: Int = 1,
        sortBy: String = "added",
        sortOrder: String = "desc",
        error: Throwable = IllegalStateException("海报墙加载失败"),
    ) {
        val key = TvPosterWallRequestKey(kind = kind, page = page, sortBy = sortBy, sortOrder = sortOrder)
        pendingPosterWalls.getOrPut(key) { ArrayDeque() }
            .removeFirst()
            .complete(Result.failure(error))
    }
}

class DelayedIptvTvRepository(
    private val baseUrl: String = "https://example.com",
) : TvRepository {
    private val pendingIptvRequests = ArrayDeque<CompletableDeferred<Result<TvIptvPayload>>>()
    val iptvRequestCount: Int
        get() = pendingIptvRequests.size

    override suspend fun fetchHome(kind: String, query: String, page: Int, pageSize: Int): Result<TvHomePayload> =
        Result.success(TvHomePayload())

    override suspend fun fetchSearch(query: String, page: Int, pageSize: Int): Result<TvSearchPayload> =
        Result.success(TvSearchPayload())

    override suspend fun fetchCatalogWall(
        kind: String,
        page: Int,
        pageSize: Int,
        sortBy: String,
        sortOrder: String,
    ): Result<TvCatalogWallPayload> =
        Result.success(TvCatalogWallPayload(page = page, pageSize = pageSize))

    override suspend fun fetchIptvChannels(): Result<TvIptvPayload> {
        val deferred = CompletableDeferred<Result<TvIptvPayload>>()
        pendingIptvRequests.addLast(deferred)
        return deferred.await()
    }

    override suspend fun fetchSeriesDetail(seriesId: String): Result<TvSeriesDetailDto> =
        Result.success(tvSeriesDetail(id = seriesId))

    override suspend fun fetchShortFeed(pageSize: Int, excludeIds: List<String>): Result<List<FeedVideoDto>> =
        Result.success(emptyList())

    override suspend fun readActiveBaseUrl(): String = baseUrl

    override suspend fun buildSourceUrl(videoId: String, profile: String?): String =
        "https://example.com/$videoId.m3u8"

    override suspend fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) = Unit

    override suspend fun readTvSubtitlePreference(videoId: String): TvTrackPreference? = null

    override suspend fun saveTvSubtitlePreference(videoId: String, preference: TvTrackPreference?) = Unit

    override suspend fun readTvAudioPreference(videoId: String): TvTrackPreference? = null

    override suspend fun saveTvAudioPreference(videoId: String, preference: TvTrackPreference?) = Unit

    override suspend fun readTvSeekStepSeconds(): Int = 10

    override suspend fun saveTvSeekStepSeconds(seconds: Int) = Unit

    override suspend fun readTvSeriesAutoplayEnabled(): Boolean? = null

    override suspend fun saveTvSeriesAutoplayEnabled(enabled: Boolean) = Unit

    fun completeIptv(payload: TvIptvPayload = TvIptvPayload()) {
        pendingIptvRequests.removeFirst().complete(Result.success(payload))
    }

    fun completeLatestIptv(payload: TvIptvPayload = TvIptvPayload()) {
        pendingIptvRequests.removeLast().complete(Result.success(payload))
    }

    fun completeIptvFailure(error: Throwable = IllegalStateException("IPTV 频道加载失败，请重试")) {
        pendingIptvRequests.removeFirst().complete(Result.failure(error))
    }
}

class DelayedSeriesDetailTvRepository(
    private val baseUrl: String = "https://example.com",
) : TvRepository {
    private val pendingDetailRequests = ArrayDeque<CompletableDeferred<Result<TvSeriesDetailDto>>>()
    val detailRequests = mutableListOf<String>()

    override suspend fun fetchHome(kind: String, query: String, page: Int, pageSize: Int): Result<TvHomePayload> =
        Result.success(TvHomePayload())

    override suspend fun fetchSearch(query: String, page: Int, pageSize: Int): Result<TvSearchPayload> =
        Result.success(TvSearchPayload())

    override suspend fun fetchCatalogWall(
        kind: String,
        page: Int,
        pageSize: Int,
        sortBy: String,
        sortOrder: String,
    ): Result<TvCatalogWallPayload> =
        Result.success(TvCatalogWallPayload(page = page, pageSize = pageSize))

    override suspend fun fetchIptvChannels(): Result<TvIptvPayload> =
        Result.success(TvIptvPayload())

    override suspend fun fetchShortFeed(pageSize: Int, excludeIds: List<String>): Result<List<FeedVideoDto>> =
        Result.success(emptyList())

    override suspend fun fetchSeriesDetail(seriesId: String): Result<TvSeriesDetailDto> {
        detailRequests += seriesId
        val deferred = CompletableDeferred<Result<TvSeriesDetailDto>>()
        pendingDetailRequests.addLast(deferred)
        return deferred.await()
    }

    override suspend fun readActiveBaseUrl(): String = baseUrl

    override suspend fun buildSourceUrl(videoId: String, profile: String?): String =
        "https://example.com/$videoId.m3u8"

    override suspend fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) = Unit

    override suspend fun readTvSubtitlePreference(videoId: String): TvTrackPreference? = null

    override suspend fun saveTvSubtitlePreference(videoId: String, preference: TvTrackPreference?) = Unit

    override suspend fun readTvAudioPreference(videoId: String): TvTrackPreference? = null

    override suspend fun saveTvAudioPreference(videoId: String, preference: TvTrackPreference?) = Unit

    override suspend fun readTvSeekStepSeconds(): Int = 10

    override suspend fun saveTvSeekStepSeconds(seconds: Int) = Unit

    override suspend fun readTvSeriesAutoplayEnabled(): Boolean? = null

    override suspend fun saveTvSeriesAutoplayEnabled(enabled: Boolean) = Unit

    fun completeDetail(payload: TvSeriesDetailDto = tvSeriesDetail()) {
        pendingDetailRequests.removeFirst().complete(Result.success(payload))
    }

    fun completeLatestDetail(payload: TvSeriesDetailDto = tvSeriesDetail()) {
        pendingDetailRequests.removeLast().complete(Result.success(payload))
    }

    fun completeDetailFailure(error: Throwable = IllegalStateException("电视剧详情加载失败")) {
        pendingDetailRequests.removeFirst().complete(Result.failure(error))
    }
}

data class TvPosterWallRequestKey(
    val kind: String,
    val page: Int,
    val sortBy: String,
    val sortOrder: String,
)

fun tvSeriesSummary(
    id: String = "series-1",
    title: String = "雾城档案",
): TvSeriesSummaryDto = TvSeriesSummaryDto(
    id = id,
    title = title,
    overview = "调查悬案",
    totalSeasons = 2,
    totalEpisodes = 16,
    playableEpisodes = 8,
)

fun tvSeriesDetail(
    id: String = "series-1",
    title: String = "雾城档案",
    seasons: List<TvSeasonDto> = emptyList(),
    cast: List<VideoActorDto> = emptyList(),
): TvSeriesDetailDto = TvSeriesDetailDto(
    id = id,
    title = title,
    overview = "调查悬案",
    seasons = seasons,
    tags = emptyList(),
    cast = cast,
)

fun tvEpisode(
    id: String,
    number: Int,
    title: String,
    videoId: String = "",
    videoStatus: String = "",
    watchSeconds: Int = 0,
    lastWatchedAt: String? = null,
    subtitleTracks: List<SubtitleTrackDto> = emptyList(),
    metadata: Map<String, Any?>? = emptyMap(),
): TvEpisodeDto = TvEpisodeDto(
    id = id,
    episodeNumber = number,
    title = title,
    videoId = videoId,
    videoStatus = videoStatus,
    watchSeconds = watchSeconds,
    lastWatchedAt = lastWatchedAt,
    playable = videoId.isNotBlank() && videoStatus == "ready",
    subtitleTracks = subtitleTracks,
    metadata = metadata,
)

suspend fun TvCatalogViewModel.awaitIdle() = Unit
suspend fun TvPosterWallViewModel.awaitIdle() = Unit
suspend fun TvSeriesDetailViewModel.awaitIdle() = Unit
suspend fun TvSeriesPlayerViewModel.awaitIdle() = Unit
suspend fun TvIptvViewModel.awaitIdle() = Unit

fun tvSearchResult(
    id: String = "movie-1",
    type: String = "movie",
    title: String = "午夜列车",
): TvSearchResultDto = TvSearchResultDto(
    id = id,
    type = type,
    title = title,
    overview = "悬疑长片",
)

fun tvPosterWallItem(
    id: String = "tv-1",
    type: String = "tv",
    title: String = "雾城档案",
): TvCatalogWallItemDto = TvCatalogWallItemDto(
    id = id,
    type = type,
    title = title,
    overview = "海报墙项目",
    posterUrl = "/poster.jpg",
    backdropUrl = "/backdrop.jpg",
)

fun tvPosterWallPage(
    page: Int = 1,
    pageSize: Int = 24,
    totalCount: Int = 0,
    items: List<TvCatalogWallItemDto> = emptyList(),
): TvCatalogWallPayload = TvCatalogWallPayload(
    items = items,
    totalCount = totalCount,
    page = page,
    pageSize = pageSize,
)
