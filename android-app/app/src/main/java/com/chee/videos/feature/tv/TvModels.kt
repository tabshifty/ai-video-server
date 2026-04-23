package com.chee.videos.feature.tv

data class TvEpisodeUiModel(
    val id: String,
    val number: Int,
    val title: String,
    val durationLabel: String,
    val summary: String,
    val progressPercent: Int = 0,
)

data class TvSeasonUiModel(
    val number: Int,
    val title: String,
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
    val progressPercent: Int,
)
