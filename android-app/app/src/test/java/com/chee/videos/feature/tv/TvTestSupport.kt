package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvContinueWatchingDto
import com.chee.videos.core.model.TvEpisodeDto
import com.chee.videos.core.model.TvHomePayload
import com.chee.videos.core.model.TvSeasonDto
import com.chee.videos.core.model.TvSeriesDetailDto
import com.chee.videos.core.model.TvSeriesSummaryDto
import kotlinx.coroutines.CompletableDeferred
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.setMain
import org.junit.rules.TestWatcher
import org.junit.runner.Description

class FakeTvRepository(
    private val baseUrl: String = "https://example.com",
    private val homePayload: TvHomePayload = TvHomePayload(),
    private val searchPayload: TvHomePayload = homePayload,
    private val detailPayload: TvSeriesDetailDto = tvSeriesDetail(),
    private val homeError: Throwable? = null,
    private val detailError: Throwable? = null,
    subtitlePreferences: Map<String, String> = emptyMap(),
) : TvRepository {
    private val storedSubtitlePreferences = subtitlePreferences.toMutableMap()

    override suspend fun fetchHome(query: String, page: Int, pageSize: Int): Result<TvHomePayload> {
        homeError?.let { return Result.failure(it) }
        return Result.success(if (query.isBlank()) homePayload else searchPayload)
    }

    override suspend fun fetchSeriesDetail(seriesId: String): Result<TvSeriesDetailDto> {
        detailError?.let { return Result.failure(it) }
        return Result.success(detailPayload)
    }

    override suspend fun readActiveBaseUrl(): String = baseUrl

    override suspend fun buildSourceUrl(videoId: String): String = "https://example.com/$videoId.m3u8"

    override suspend fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) = Unit

    override suspend fun readTvSubtitlePreference(videoId: String): String? = storedSubtitlePreferences[videoId]

    override suspend fun saveTvSubtitlePreference(videoId: String, subtitleTrackId: String?) {
        if (subtitleTrackId == null) {
            storedSubtitlePreferences.remove(videoId)
        } else {
            storedSubtitlePreferences[videoId] = subtitleTrackId
        }
    }
}

class DelayedSourceTvRepository(
    private val baseUrl: String = "https://example.com",
    private val detailPayload: TvSeriesDetailDto = tvSeriesDetail(),
) : TvRepository {
    private val pendingSourceUrls = linkedMapOf<String, CompletableDeferred<String>>()

    override suspend fun fetchHome(query: String, page: Int, pageSize: Int): Result<TvHomePayload> =
        Result.success(TvHomePayload())

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
suspend fun TvSeriesDetailViewModel.awaitIdle() = Unit
suspend fun TvSeriesPlayerViewModel.awaitIdle() = Unit

@OptIn(ExperimentalCoroutinesApi::class)
class MainDispatcherRule : TestWatcher() {
    override fun starting(description: Description) {
        Dispatchers.setMain(UnconfinedTestDispatcher())
    }

    override fun finished(description: Description) {
        Dispatchers.resetMain()
    }
}
