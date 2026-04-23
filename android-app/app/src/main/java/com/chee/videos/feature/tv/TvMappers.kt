package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvContinueWatchingDto
import com.chee.videos.core.model.TvEpisodeDto
import com.chee.videos.core.model.TvSectionDto
import com.chee.videos.core.model.TvSeriesDetailDto
import com.chee.videos.core.model.TvSeriesSummaryDto

internal fun tvSeriesSummaryToUiModel(dto: TvSeriesSummaryDto): TvSeriesUiModel {
    val yearText = dto.firstAirDate?.take(4).orEmpty().ifBlank { "待定档" }
    val subtitle = buildString {
        append(yearText)
        if (dto.totalSeasons > 0) {
            append(" · ${dto.totalSeasons} 季")
        }
        if (dto.totalEpisodes > 0) {
            append(" · ${dto.totalEpisodes} 集")
        }
    }
    val updateText = if (dto.playableEpisodes > 0) "${dto.playableEpisodes} 集可播" else "待绑定视频"
    return TvSeriesUiModel(
        id = dto.id,
        title = dto.title,
        subtitle = subtitle,
        ratingText = if (dto.playableEpisodes > 0) "可播 ${dto.playableEpisodes}" else "待更新",
        updateText = updateText,
        description = dto.overview.orEmpty().ifBlank { "暂无简介" },
        tags = emptyList(),
        cast = emptyList(),
        seasons = emptyList(),
        posterSeed = dto.title.hashCode(),
        posterUrl = dto.posterUrl,
        backdropUrl = dto.backdropUrl,
        playableEpisodes = dto.playableEpisodes,
        totalEpisodes = dto.totalEpisodes,
    )
}

internal fun tvSectionToUiModel(dto: TvSectionDto): TvCatalogSectionUiModel =
    TvCatalogSectionUiModel(
        title = dto.title,
        subtitle = dto.subtitle,
        items = dto.items.map(::tvSeriesSummaryToUiModel),
    )

internal fun tvContinueWatchingToUiModel(dto: TvContinueWatchingDto): TvContinueWatchingUiModel =
    TvContinueWatchingUiModel(
        seriesId = dto.seriesId,
        seriesTitle = dto.seriesTitle,
        seasonNumber = dto.seasonNumber,
        episodeNumber = dto.episodeNumber,
        episodeTitle = dto.episodeTitle,
        progressPercent = dto.progressPercent,
    )

internal fun tvEpisodeToUiModel(dto: TvEpisodeDto): TvEpisodeUiModel =
    TvEpisodeUiModel(
        id = dto.id,
        number = dto.episodeNumber,
        title = dto.title.ifBlank { "第${dto.episodeNumber}集" },
        durationLabel = if (dto.runtime > 0) "${dto.runtime} 分钟" else "时长待更新",
        summary = dto.overview.orEmpty().ifBlank { "暂无剧情简介" },
        progressPercent = dto.progressPercent,
        videoId = dto.videoId,
        videoStatus = dto.videoStatus.orEmpty(),
        playable = dto.playable || (dto.videoId.isNotBlank() && dto.videoStatus == "ready"),
    )

internal fun tvSeriesDetailToUiModel(dto: TvSeriesDetailDto): TvSeriesUiModel {
    val seasons = dto.seasons.map { season ->
        TvSeasonUiModel(
            id = season.id,
            number = season.seasonNumber,
            title = season.title.ifBlank { "第 ${season.seasonNumber} 季" },
            overview = season.overview.orEmpty(),
            episodes = season.episodes.map(::tvEpisodeToUiModel),
        )
    }
    val yearText = dto.firstAirDate?.take(4).orEmpty().ifBlank { "待定档" }
    return TvSeriesUiModel(
        id = dto.id,
        title = dto.title,
        subtitle = buildString {
            append(yearText)
            if (dto.totalSeasons > 0) append(" · ${dto.totalSeasons} 季")
            if (dto.totalEpisodes > 0) append(" · ${dto.totalEpisodes} 集")
        },
        ratingText = if (dto.playableEpisodes > 0) "可播 ${dto.playableEpisodes}" else "待更新",
        updateText = if (dto.playableEpisodes > 0) "${dto.playableEpisodes} 集可播" else "待绑定视频",
        description = dto.overview.orEmpty().ifBlank { "暂无简介" },
        tags = dto.tags,
        cast = dto.cast,
        seasons = seasons,
        posterSeed = dto.title.hashCode(),
        posterUrl = dto.posterUrl,
        backdropUrl = dto.backdropUrl,
        playableEpisodes = dto.playableEpisodes,
        totalEpisodes = dto.totalEpisodes,
    )
}

internal fun findPreferredEpisodeNumber(season: TvSeasonUiModel): Int {
    return season.episodes.firstOrNull { it.playable }?.number
        ?: season.episodes.firstOrNull()?.number
        ?: 1
}
