package com.chee.videos.feature.tv

import java.net.URLDecoder
import java.net.URLEncoder
import java.nio.charset.StandardCharsets

const val TvSeriesIdArg = "seriesId"
const val TvSeasonArg = "season"
const val TvEpisodeArg = "episode"
const val TvLongFormVideoIdArg = "videoId"
const val TvLongFormVideoTypeArg = "videoType"
const val TvLongFormStartFromBeginningArg = "startFromBeginning"
const val TvCatalogWallKindArg = "kind"
const val TvCatalogWallTitleArg = "title"
const val TvShortFeedRoute = "tv/shorts"

const val TvSeriesRoutePattern = "tv/series/{$TvSeriesIdArg}"
const val TvPlayerRoutePattern = "tv/player/{$TvSeriesIdArg}?$TvSeasonArg={$TvSeasonArg}&$TvEpisodeArg={$TvEpisodeArg}&$TvLongFormStartFromBeginningArg={$TvLongFormStartFromBeginningArg}"
const val TvLongFormDetailRoutePattern = "tv/detail/{$TvLongFormVideoIdArg}?$TvLongFormVideoTypeArg={$TvLongFormVideoTypeArg}"
const val TvLongFormPlayerRoutePattern = "tv/long-form-player/{$TvLongFormVideoIdArg}?$TvLongFormVideoTypeArg={$TvLongFormVideoTypeArg}&$TvLongFormStartFromBeginningArg={$TvLongFormStartFromBeginningArg}"
const val TvCatalogWallRoutePattern = "tv/wall/{$TvCatalogWallKindArg}?$TvCatalogWallTitleArg={$TvCatalogWallTitleArg}"
const val TvIptvRoute = "tv/iptv"

fun buildTvSeriesRoute(seriesId: String): String = "tv/series/${encodeTvRouteSegment(seriesId)}"

fun buildTvPlayerRoute(
    seriesId: String,
    season: Int,
    episode: Int,
    startFromBeginning: Boolean = false,
): String {
    val safeSeason = season.coerceAtLeast(1)
    val safeEpisode = episode.coerceAtLeast(1)
    return "tv/player/${encodeTvRouteSegment(seriesId)}?$TvSeasonArg=$safeSeason&$TvEpisodeArg=$safeEpisode" +
        "&$TvLongFormStartFromBeginningArg=$startFromBeginning"
}

fun buildTvLongFormDetailRoute(videoId: String, videoType: String): String {
    return "tv/detail/${encodeTvRouteSegment(videoId)}?$TvLongFormVideoTypeArg=${normalizeTvLongFormVideoType(videoType)}"
}

fun buildTvLongFormPlayerRoute(
    videoId: String,
    videoType: String,
    startFromBeginning: Boolean = false,
): String {
    return "tv/long-form-player/${encodeTvRouteSegment(videoId)}" +
        "?$TvLongFormVideoTypeArg=${normalizeTvLongFormVideoType(videoType)}" +
        "&$TvLongFormStartFromBeginningArg=$startFromBeginning"
}

fun buildTvCatalogWallRoute(kind: String, title: String = ""): String {
    val encodedTitle = title.trim().takeIf { it.isNotBlank() }
        ?.let { encodeTvRouteSegment(it) }
        .orEmpty()
    return "tv/wall/${encodeTvRouteSegment(kind)}?$TvCatalogWallTitleArg=$encodedTitle"
}

internal fun decodeTvRouteArg(value: String?): String {
    val raw = value.orEmpty().trim()
    if (raw.isBlank()) {
        return ""
    }
    return URLDecoder.decode(raw, StandardCharsets.UTF_8.toString()).trim()
}

internal fun encodeTvRouteSegment(value: String): String {
    return URLEncoder.encode(value.trim(), StandardCharsets.UTF_8.toString())
        .replace("+", "%20")
}
