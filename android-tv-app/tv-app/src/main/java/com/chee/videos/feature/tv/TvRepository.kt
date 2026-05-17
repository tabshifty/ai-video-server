package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvHomePayload
import com.chee.videos.core.model.TvCatalogWallPayload
import com.chee.videos.core.model.TvSearchPayload
import com.chee.videos.core.model.TvSeriesDetailDto
import com.chee.videos.core.repository.VideoRepository
import javax.inject.Inject
import javax.inject.Singleton

interface TvRepository {
    suspend fun fetchHome(query: String = "", page: Int = 1, pageSize: Int = 20): Result<TvHomePayload>
    suspend fun fetchSearch(query: String, page: Int = 1, pageSize: Int = 20): Result<TvSearchPayload>
    suspend fun fetchCatalogWall(kind: String, page: Int = 1, pageSize: Int = 24): Result<TvCatalogWallPayload>
    suspend fun fetchSeriesDetail(seriesId: String): Result<TvSeriesDetailDto>
    suspend fun readActiveBaseUrl(): String?
    suspend fun buildSourceUrl(videoId: String): String
    suspend fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean)
    suspend fun readTvSubtitlePreference(videoId: String): String?
    suspend fun saveTvSubtitlePreference(videoId: String, subtitleTrackId: String?)
    suspend fun readTvAudioPreference(videoId: String): String?
    suspend fun saveTvAudioPreference(videoId: String, audioTrackId: String?)
}

@Singleton
class NetworkTvRepository @Inject constructor(
    private val videoRepository: VideoRepository,
) : TvRepository {
    override suspend fun fetchHome(query: String, page: Int, pageSize: Int): Result<TvHomePayload> =
        videoRepository.fetchTvHome(query, page, pageSize)

    override suspend fun fetchSearch(query: String, page: Int, pageSize: Int): Result<TvSearchPayload> =
        videoRepository.fetchTvSearch(query, page, pageSize)

    override suspend fun fetchCatalogWall(kind: String, page: Int, pageSize: Int): Result<TvCatalogWallPayload> =
        videoRepository.fetchTvCatalogWall(kind, page, pageSize)

    override suspend fun fetchSeriesDetail(seriesId: String): Result<TvSeriesDetailDto> =
        videoRepository.fetchTvSeriesDetail(seriesId)

    override suspend fun readActiveBaseUrl(): String? =
        videoRepository.readActiveBaseUrl()

    override suspend fun buildSourceUrl(videoId: String): String =
        videoRepository.buildSourceUrl(videoId)

    override suspend fun reportHistory(videoId: String, watchSeconds: Int, completed: Boolean) {
        videoRepository.reportHistory(videoId, watchSeconds, completed)
    }

    override suspend fun readTvSubtitlePreference(videoId: String): String? =
        videoRepository.readTvSubtitlePreference(videoId)

    override suspend fun saveTvSubtitlePreference(videoId: String, subtitleTrackId: String?) {
        videoRepository.saveTvSubtitlePreference(videoId, subtitleTrackId)
    }

    override suspend fun readTvAudioPreference(videoId: String): String? =
        videoRepository.readTvAudioPreference(videoId)

    override suspend fun saveTvAudioPreference(videoId: String, audioTrackId: String?) {
        videoRepository.saveTvAudioPreference(videoId, audioTrackId)
    }
}
