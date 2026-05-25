package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvHomePayload
import com.chee.videos.core.model.TvCatalogWallPayload
import com.chee.videos.core.model.TvIptvPayload
import com.chee.videos.core.model.TvSearchPayload
import com.chee.videos.core.model.TvSeriesDetailDto
import com.chee.videos.core.model.TvTrackPreference
import com.chee.videos.core.repository.VideoRepository
import javax.inject.Inject
import javax.inject.Singleton

interface TvRepository {
    suspend fun fetchHome(kind: String = "tv", query: String = "", page: Int = 1, pageSize: Int = 20): Result<TvHomePayload>
    suspend fun fetchSearch(query: String, page: Int = 1, pageSize: Int = 20): Result<TvSearchPayload>
    suspend fun fetchCatalogWall(
        kind: String,
        page: Int = 1,
        pageSize: Int = 24,
        sortBy: String = "added",
        sortOrder: String = "desc",
    ): Result<TvCatalogWallPayload>
    suspend fun fetchIptvChannels(): Result<TvIptvPayload>
    suspend fun fetchSeriesDetail(seriesId: String): Result<TvSeriesDetailDto>
    suspend fun readActiveBaseUrl(): String?
    suspend fun buildSourceUrl(videoId: String): String
    suspend fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean)
    suspend fun readTvSubtitlePreference(videoId: String): TvTrackPreference?
    suspend fun saveTvSubtitlePreference(videoId: String, preference: TvTrackPreference?)
    suspend fun readTvAudioPreference(videoId: String): TvTrackPreference?
    suspend fun saveTvAudioPreference(videoId: String, preference: TvTrackPreference?)
    suspend fun readTvSeekStepSeconds(): Int
    suspend fun saveTvSeekStepSeconds(seconds: Int)
    suspend fun readTvSeriesAutoplayEnabled(): Boolean?
    suspend fun saveTvSeriesAutoplayEnabled(enabled: Boolean)
}

@Singleton
class NetworkTvRepository @Inject constructor(
    private val videoRepository: VideoRepository,
) : TvRepository {
    override suspend fun fetchHome(kind: String, query: String, page: Int, pageSize: Int): Result<TvHomePayload> =
        videoRepository.fetchTvHome(kind, query, page, pageSize)

    override suspend fun fetchSearch(query: String, page: Int, pageSize: Int): Result<TvSearchPayload> =
        videoRepository.fetchTvSearch(query, page, pageSize)

    override suspend fun fetchCatalogWall(
        kind: String,
        page: Int,
        pageSize: Int,
        sortBy: String,
        sortOrder: String,
    ): Result<TvCatalogWallPayload> =
        videoRepository.fetchTvCatalogWall(kind, page, pageSize, sortBy, sortOrder)

    override suspend fun fetchIptvChannels(): Result<TvIptvPayload> =
        videoRepository.fetchTvIptvChannels()

    override suspend fun fetchSeriesDetail(seriesId: String): Result<TvSeriesDetailDto> =
        videoRepository.fetchTvSeriesDetail(seriesId)

    override suspend fun readActiveBaseUrl(): String? =
        videoRepository.readActiveBaseUrl()

    override suspend fun buildSourceUrl(videoId: String): String =
        videoRepository.buildSourceUrl(videoId)

    override suspend fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) {
        videoRepository.reportHistory(videoId, watchSeconds, completed)
    }

    override suspend fun readTvSubtitlePreference(videoId: String): TvTrackPreference? =
        videoRepository.readTvSubtitlePreference(videoId)

    override suspend fun saveTvSubtitlePreference(videoId: String, preference: TvTrackPreference?) {
        videoRepository.saveTvSubtitlePreference(videoId, preference)
    }

    override suspend fun readTvAudioPreference(videoId: String): TvTrackPreference? =
        videoRepository.readTvAudioPreference(videoId)

    override suspend fun saveTvAudioPreference(videoId: String, preference: TvTrackPreference?) {
        videoRepository.saveTvAudioPreference(videoId, preference)
    }

    override suspend fun readTvSeekStepSeconds(): Int =
        videoRepository.readTvSeekStepSeconds()

    override suspend fun saveTvSeekStepSeconds(seconds: Int) {
        videoRepository.saveTvSeekStepSeconds(seconds)
    }

    override suspend fun readTvSeriesAutoplayEnabled(): Boolean? =
        videoRepository.readTvSeriesAutoplayEnabled()

    override suspend fun saveTvSeriesAutoplayEnabled(enabled: Boolean) {
        videoRepository.saveTvSeriesAutoplayEnabled(enabled)
    }
}
