package com.chee.videos.feature.tv

const val TvSeriesIdArg = "seriesId"
const val TvSeasonArg = "season"
const val TvEpisodeArg = "episode"
const val TvLongFormVideoIdArg = "videoId"
const val TvLongFormVideoTypeArg = "videoType"

const val TvSeriesRoutePattern = "tv/series/{$TvSeriesIdArg}"
const val TvPlayerRoutePattern = "tv/player/{$TvSeriesIdArg}?$TvSeasonArg={$TvSeasonArg}&$TvEpisodeArg={$TvEpisodeArg}"
const val TvLongFormDetailRoutePattern = "tv/detail/{$TvLongFormVideoIdArg}?$TvLongFormVideoTypeArg={$TvLongFormVideoTypeArg}"
const val TvLongFormPlayerRoutePattern = "tv/long-form-player/{$TvLongFormVideoIdArg}?$TvLongFormVideoTypeArg={$TvLongFormVideoTypeArg}"

fun buildTvSeriesRoute(seriesId: String): String = "tv/series/$seriesId"

fun buildTvPlayerRoute(seriesId: String, season: Int, episode: Int): String {
    val safeSeason = season.coerceAtLeast(1)
    val safeEpisode = episode.coerceAtLeast(1)
    return "tv/player/$seriesId?$TvSeasonArg=$safeSeason&$TvEpisodeArg=$safeEpisode"
}

fun buildTvLongFormDetailRoute(videoId: String, videoType: String): String {
    return "tv/detail/$videoId?$TvLongFormVideoTypeArg=${normalizeTvLongFormVideoType(videoType)}"
}

fun buildTvLongFormPlayerRoute(videoId: String, videoType: String): String {
    return "tv/long-form-player/$videoId?$TvLongFormVideoTypeArg=${normalizeTvLongFormVideoType(videoType)}"
}
