package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvContinueWatchingDto
import com.chee.videos.core.model.TvEpisodeDto
import com.chee.videos.core.model.TvSectionDto
import com.chee.videos.core.model.TvSeasonDto
import com.chee.videos.core.model.TvSeriesDetailDto
import com.chee.videos.core.model.TvSeriesSummaryDto

@Suppress("UNCHECKED_CAST")
internal fun <T> coerceListOrEmpty(value: Any?): List<T> = value as? List<T> ?: emptyList()

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
        items = coerceListOrEmpty<TvSeriesSummaryDto>(dto.items).map(::tvSeriesSummaryToUiModel),
    )

internal fun tvContinueWatchingToUiModel(dto: TvContinueWatchingDto): TvContinueWatchingUiModel =
    TvContinueWatchingUiModel(
        seriesId = dto.seriesId,
        seriesTitle = dto.seriesTitle,
        seasonNumber = dto.seasonNumber,
        episodeNumber = dto.episodeNumber,
        episodeTitle = dto.episodeTitle,
        posterUrl = dto.posterUrl,
        backdropUrl = dto.backdropUrl,
        watchSeconds = dto.watchSeconds,
        progressPercent = dto.progressPercent,
    )

internal fun tvEpisodeToUiModel(dto: TvEpisodeDto): TvEpisodeUiModel =
    TvEpisodeUiModel(
        id = dto.id,
        number = dto.episodeNumber,
        title = dto.title.ifBlank { "第${dto.episodeNumber}集" },
        durationLabel = if (dto.runtime > 0) "${dto.runtime} 分钟" else "时长待更新",
        summary = dto.overview.orEmpty().ifBlank { "暂无剧情简介" },
        watchSeconds = dto.watchSeconds,
        lastWatchedAt = dto.lastWatchedAt.orEmpty(),
        progressPercent = dto.progressPercent,
        videoId = dto.videoId,
        videoStatus = dto.videoStatus.orEmpty(),
        playable = dto.playable || (dto.videoId.isNotBlank() && dto.videoStatus == "ready"),
        subtitleTracks = coerceListOrEmpty(dto.subtitleTracks),
    )

internal fun tvSeriesDetailToUiModel(dto: TvSeriesDetailDto): TvSeriesUiModel {
    val seasons = coerceListOrEmpty<TvSeasonDto>(dto.seasons).map { season ->
        TvSeasonUiModel(
            id = season.id,
            number = season.seasonNumber,
            title = season.title.ifBlank { "第 ${season.seasonNumber} 季" },
            overview = season.overview.orEmpty(),
            episodes = coerceListOrEmpty<TvEpisodeDto>(season.episodes).map(::tvEpisodeToUiModel),
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
        tags = coerceListOrEmpty<String>(dto.tags),
        cast = coerceListOrEmpty<String>(dto.cast),
        seasons = seasons,
        posterSeed = dto.title.hashCode(),
        posterUrl = dto.posterUrl,
        backdropUrl = dto.backdropUrl,
        playableEpisodes = dto.playableEpisodes,
        totalEpisodes = dto.totalEpisodes,
    )
}

internal fun findPreferredEpisodeNumber(season: TvSeasonUiModel): Int {
    return season.episodes.maxWithOrNull(tvEpisodePreferenceComparator)?.number
        ?: season.episodes.firstOrNull { it.playable }?.number
        ?: season.episodes.firstOrNull()?.number
        ?: 1
}

internal data class TvPreferredEpisodeSelection(
    val seasonNumber: Int,
    val episodeNumber: Int,
)

internal fun findPreferredSeriesSelection(series: TvSeriesUiModel): TvPreferredEpisodeSelection? {
    return series.seasons
        .flatMap { season ->
            season.episodes.map { episode ->
                TvPreferredEpisodeSelection(
                    seasonNumber = season.number,
                    episodeNumber = episode.number,
                ) to episode
            }
        }
        .maxWithOrNull(compareBy<Pair<TvPreferredEpisodeSelection, TvEpisodeUiModel>> { preferredEpisodeTimestamp(it.second) }
            .thenBy { it.second.watchSeconds }
            .thenBy { if (it.second.playable) 1 else 0 })
        ?.first
        ?: series.seasons.firstOrNull()?.let { season ->
            TvPreferredEpisodeSelection(season.number, findPreferredEpisodeNumber(season))
        }
}

private val tvEpisodePreferenceComparator = compareBy<TvEpisodeUiModel> { preferredEpisodeTimestamp(it) }
    .thenBy { it.watchSeconds }
    .thenBy { if (it.playable) 1 else 0 }

private fun preferredEpisodeTimestamp(episode: TvEpisodeUiModel): Long =
    parseEpisodeTimestamp(episode.lastWatchedAt)

private fun parseEpisodeTimestamp(raw: String): Long {
    val value = raw.trim()
    if (value.isBlank()) {
        return Long.MIN_VALUE
    }
    return runCatching { java.time.OffsetDateTime.parse(value).toInstant().toEpochMilli() }.getOrDefault(Long.MIN_VALUE)
}
