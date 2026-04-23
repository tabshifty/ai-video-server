package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvContinueWatchingDto
import com.chee.videos.core.model.TvEpisodeDto
import com.chee.videos.core.model.TvHomePayload
import com.chee.videos.core.model.TvSeasonDto
import com.chee.videos.core.model.TvSeriesDetailDto
import com.chee.videos.core.model.TvSeriesSummaryDto
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.test.UnconfinedTestDispatcher
import kotlinx.coroutines.test.resetMain
import kotlinx.coroutines.test.setMain
import org.junit.rules.TestWatcher
import org.junit.runner.Description

class FakeTvRepository(
    private val homePayload: TvHomePayload = TvHomePayload(),
    private val searchPayload: TvHomePayload = homePayload,
    private val detailPayload: TvSeriesDetailDto = tvSeriesDetail(),
    private val homeError: Throwable? = null,
    private val detailError: Throwable? = null,
) : TvRepository {
    override suspend fun fetchHome(query: String, page: Int, pageSize: Int): Result<TvHomePayload> {
        homeError?.let { return Result.failure(it) }
        return Result.success(if (query.isBlank()) homePayload else searchPayload)
    }

    override suspend fun fetchSeriesDetail(seriesId: String): Result<TvSeriesDetailDto> {
        detailError?.let { return Result.failure(it) }
        return Result.success(detailPayload)
    }

    override suspend fun buildSourceUrl(videoId: String): String = "https://example.com/$videoId.m3u8"

    override suspend fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) = Unit
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
): TvEpisodeDto = TvEpisodeDto(
    id = id,
    episodeNumber = number,
    title = title,
    videoId = videoId,
    videoStatus = videoStatus,
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
