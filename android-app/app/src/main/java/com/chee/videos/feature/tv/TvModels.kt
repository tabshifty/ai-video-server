package com.chee.videos.feature.tv

import com.chee.videos.core.model.SubtitleTrackDto

data class TvEpisodeUiModel(
    val id: String,
    val number: Int,
    val title: String,
    val durationLabel: String,
    val summary: String,
    val watchSeconds: Int = 0,
    val lastWatchedAt: String = "",
    val progressPercent: Int = 0,
    val videoId: String = "",
    val videoStatus: String = "",
    val playable: Boolean = false,
    val subtitleTracks: List<SubtitleTrackDto> = emptyList(),
)

data class TvSeasonUiModel(
    val id: String,
    val number: Int,
    val title: String,
    val overview: String = "",
    val episodes: List<TvEpisodeUiModel>,
)

data class TvSeriesUiModel(
    val id: String,
    val title: String,
    val subtitle: String,
    val ratingText: String,
    val updateText: String,
    val description: String,
    val tags: List<String>,
    val cast: List<String>,
    val seasons: List<TvSeasonUiModel>,
    val posterSeed: Int,
    val posterUrl: String? = null,
    val backdropUrl: String? = null,
    val playableEpisodes: Int = 0,
    val totalEpisodes: Int = 0,
)

data class TvCatalogSectionUiModel(
    val title: String,
    val subtitle: String,
    val items: List<TvSeriesUiModel>,
)

data class TvContinueWatchingUiModel(
    val seriesId: String,
    val seriesTitle: String,
    val seasonNumber: Int,
    val episodeNumber: Int,
    val episodeTitle: String,
    val posterUrl: String? = null,
    val backdropUrl: String? = null,
    val watchSeconds: Int = 0,
    val progressPercent: Int,
)
