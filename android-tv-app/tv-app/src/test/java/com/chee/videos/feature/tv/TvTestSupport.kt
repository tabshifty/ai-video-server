package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvContinueWatchingDto
import com.chee.videos.core.model.TvCatalogWallItemDto
import com.chee.videos.core.model.TvCatalogWallPayload
import com.chee.videos.core.model.TvEpisodeDto
import com.chee.videos.core.model.TvHomePayload
import com.chee.videos.core.model.TvSearchPayload
import com.chee.videos.core.model.TvSearchResultDto
import com.chee.videos.core.model.TvSeasonDto
import com.chee.videos.core.model.TvSeriesDetailDto
import com.chee.videos.core.model.TvSeriesSummaryDto
import kotlinx.coroutines.CompletableDeferred

typealias MainDispatcherRule = com.chee.videos.core.testing.MainDispatcherRule

class FakeTvRepository(
    private val baseUrl: String = "https://example.com",
    private val homePayload: TvHomePayload = TvHomePayload(),
    private val searchPayload: TvSearchPayload = TvSearchPayload(),
    private val posterWallPages: List<TvCatalogWallPayload> = emptyList(),
    private val detailPayload: TvSeriesDetailDto = tvSeriesDetail(),
    private val homeError: Throwable? = null,
    private val detailError: Throwable? = null,
    subtitlePreferences: Map<String, String> = emptyMap(),
) : TvRepository {
    private val storedSubtitlePreferences = subtitlePreferences.toMutableMap()
    val historyReports = mutableListOf<TvHistoryReport>()

    override suspend fun fetchHome(query: String, page: Int, pageSize: Int): Result<TvHomePayload> {
        homeError?.let { return Result.failure(it) }
        return Result.success(homePayload)
    }

    override suspend fun fetchSearch(query: String, page: Int, pageSize: Int): Result<TvSearchPayload> {
        homeError?.let { return Result.failure(it) }
        return Result.success(searchPayload)
    }

    override suspend fun fetchCatalogWall(kind: String, page: Int, pageSize: Int): Result<TvCatalogWallPayload> {
        homeError?.let { return Result.failure(it) }
        val payload = posterWallPages.firstOrNull { it.page == page }
            ?: TvCatalogWallPayload(page = page, pageSize = pageSize)
        return Result.success(payload)
    }

    override suspend fun fetchSeriesDetail(seriesId: String): Result<TvSeriesDetailDto> {
        detailError?.let { return Result.failure(it) }
        return Result.success(detailPayload)
    }

    override suspend fun readActiveBaseUrl(): String = baseUrl

    override suspend fun buildSourceUrl(videoId: String): String = "https://example.com/$videoId.m3u8"

    override suspend fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) {
        historyReports += TvHistoryReport(videoId, watchSeconds, completed)
    }

    override suspend fun readTvSubtitlePreference(videoId: String): String? = storedSubtitlePreferences[videoId]

    override suspend fun saveTvSubtitlePreference(videoId: String, subtitleTrackId: String?) {
        if (subtitleTrackId == null) {
            storedSubtitlePreferences.remove(videoId)
        } else {
            storedSubtitlePreferences[videoId] = subtitleTrackId
        }
    }
}

data class TvHistoryReport(
    val videoId: String,
    val watchSeconds: Int,
    val completed: Boolean,
)

class DelayedSourceTvRepository(
    private val baseUrl: String = "https://example.com",
    private val detailPayload: TvSeriesDetailDto = tvSeriesDetail(),
) : TvRepository {
    private val pendingSourceUrls = linkedMapOf<String, CompletableDeferred<String>>()

    override suspend fun fetchHome(query: String, page: Int, pageSize: Int): Result<TvHomePayload> =
        Result.success(TvHomePayload())

    override suspend fun fetchSearch(query: String, page: Int, pageSize: Int): Result<TvSearchPayload> =
        Result.success(TvSearchPayload())

    override suspend fun fetchCatalogWall(kind: String, page: Int, pageSize: Int): Result<TvCatalogWallPayload> =
        Result.success(TvCatalogWallPayload(page = page, pageSize = pageSize))

    override suspend fun fetchSeriesDetail(seriesId: String): Result<TvSeriesDetailDto> =
        Result.success(detailPayload)

    override suspend fun readActiveBaseUrl(): String = baseUrl

    override suspend fun buildSourceUrl(videoId: String): String =
        pendingSourceUrls.getOrPut(videoId) { CompletableDeferred() }.await()

    override suspend fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) = Unit

    override suspend fun readTvSubtitlePreference(videoId: String): String? = null

    override suspend fun saveTvSubtitlePreference(videoId: String, subtitleTrackId: String?) = Unit

    fun completeSourceUrl(videoId: String, url: String = "https://example.com/$videoId.m3u8") {
        pendingSourceUrls.getOrPut(videoId) { CompletableDeferred() }.complete(url)
    }
}

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
): TvSeriesDetailDto = TvSeriesDetailDto(
    id = id,
    title = title,
    overview = "调查悬案",
    seasons = seasons,
    tags = emptyList(),
    cast = emptyList(),
)

fun tvEpisode(
    id: String,
    number: Int,
    title: String,
    videoId: String = "",
    videoStatus: String = "",
    watchSeconds: Int = 0,
    lastWatchedAt: String? = null,
): TvEpisodeDto = TvEpisodeDto(
    id = id,
    episodeNumber = number,
    title = title,
    videoId = videoId,
    videoStatus = videoStatus,
    watchSeconds = watchSeconds,
    lastWatchedAt = lastWatchedAt,
    playable = videoId.isNotBlank() && videoStatus == "ready",
)

suspend fun TvCatalogViewModel.awaitIdle() = Unit
suspend fun TvPosterWallViewModel.awaitIdle() = Unit
suspend fun TvSeriesDetailViewModel.awaitIdle() = Unit
suspend fun TvSeriesPlayerViewModel.awaitIdle() = Unit

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
