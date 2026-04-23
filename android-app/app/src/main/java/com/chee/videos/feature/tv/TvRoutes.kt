package com.chee.videos.feature.tv

const val TvSeriesIdArg = "seriesId"
const val TvSeasonArg = "season"
const val TvEpisodeArg = "episode"

const val TvSeriesRoutePattern = "tv/series/{$TvSeriesIdArg}"
const val TvPlayerRoutePattern = "tv/player/{$TvSeriesIdArg}?$TvSeasonArg={$TvSeasonArg}&$TvEpisodeArg={$TvEpisodeArg}"

fun buildTvSeriesRoute(seriesId: String): String = "tv/series/$seriesId"

fun buildTvPlayerRoute(seriesId: String, season: Int, episode: Int): String {
    val safeSeason = season.coerceAtLeast(1)
    val safeEpisode = episode.coerceAtLeast(1)
    return "tv/player/$seriesId?$TvSeasonArg=$safeSeason&$TvEpisodeArg=$safeEpisode"
}
