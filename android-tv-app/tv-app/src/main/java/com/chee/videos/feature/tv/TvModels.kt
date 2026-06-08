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
    val stillUrl: String? = null,
    val subtitleTracks: List<SubtitleTrackDto> = emptyList(),
    val metadata: Map<String, Any?>? = emptyMap(),
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

data class TvCatalogWallItemUiModel(
    val id: String,
    val type: String,
    val title: String,
    val description: String,
    val posterUrl: String? = null,
    val backdropUrl: String? = null,
    val videoId: String? = null,
    val seasonNumber: Int = 0,
    val episodeNumber: Int = 0,
    val progressPercent: Int = 0,
)

data class TvHomeShelfItemUiModel(
    val id: String,
    val type: String,
    val title: String,
    val description: String,
    val posterUrl: String? = null,
    val backdropUrl: String? = null,
    val videoId: String? = null,
    val seasonNumber: Int = 0,
    val episodeNumber: Int = 0,
    val progressPercent: Int = 0,
)

data class TvSearchResultUiModel(
    val id: String,
    val type: String,
    val title: String,
    val description: String,
    val posterUrl: String? = null,
    val backdropUrl: String? = null,
)

data class TvContinueWatchingUiModel(
    val type: String,
    val seriesId: String,
    val seriesTitle: String,
    val seasonNumber: Int,
    val episodeNumber: Int,
    val episodeTitle: String,
    val videoId: String? = null,
    val posterUrl: String? = null,
    val backdropUrl: String? = null,
    val watchSeconds: Int = 0,
    val progressPercent: Int,
)

data class TvCatalogWallUiState(
    val loading: Boolean = true,
    val loadingMore: Boolean = false,
    val refreshing: Boolean = false,
    val kind: String = "",
    val title: String = "",
    val subtitle: String = "",
    val page: Int = 0,
    val totalCount: Int = 0,
    val items: List<TvCatalogWallItemUiModel> = emptyList(),
    val sortBy: String = "added",
    val sortOrder: String = "desc",
    val errorMessage: String? = null,
)
