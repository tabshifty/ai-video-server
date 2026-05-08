package com.chee.videos.feature.tv

import com.chee.videos.core.model.VideoDetailDto
import com.chee.videos.core.model.resolveAvPosterUrl
import com.chee.videos.core.util.UrlBuilder

internal enum class TvFeaturedContentSource {
    CONTINUE_WATCHING,
    SECTION,
    SHELF,
}

internal data class TvFeaturedContentUiModel(
    val source: TvFeaturedContentSource,
    val targetId: String,
    val targetType: String,
    val eyebrow: String,
    val title: String,
    val subtitle: String,
    val description: String,
    val posterUrl: String? = null,
    val backdropUrl: String? = null,
    val progressPercent: Int = 0,
    val seasonNumber: Int = 0,
    val episodeNumber: Int = 0,
    val videoId: String? = null,
)

internal data class TvLongFormDetailHeroUiModel(
    val eyebrow: String,
    val title: String,
    val metaLine: String,
    val summary: String,
    val posterUrl: String? = null,
    val backdropUrl: String? = null,
    val primaryActionLabel: String,
    val secondaryActionLabel: String,
)

internal fun resolveTvContinueWatchingPlaybackTargetId(item: TvContinueWatchingUiModel): String {
    return if (item.type == "tv") {
        item.seriesId
    } else {
        item.videoId?.trim()?.takeIf { it.isNotBlank() } ?: item.seriesId
    }
}

internal fun resolveTvFeaturedContent(
    continueWatching: TvContinueWatchingUiModel?,
    sections: List<TvCatalogSectionUiModel>,
    tvSeries: List<TvHomeShelfItemUiModel>,
    movies: List<TvHomeShelfItemUiModel>,
    av: List<TvHomeShelfItemUiModel>,
): TvFeaturedContentUiModel? {
    continueWatching?.let { item ->
        return TvFeaturedContentUiModel(
            source = TvFeaturedContentSource.CONTINUE_WATCHING,
            targetId = resolveTvContinueWatchingPlaybackTargetId(item),
            targetType = item.type,
            eyebrow = when (item.type) {
                "movie" -> "继续看电影"
                "av" -> "继续播放 AV"
                else -> "继续追剧"
            },
            title = item.seriesTitle,
            subtitle = when (item.type) {
                "movie", "av" -> item.episodeTitle.ifBlank { "继续播放" }
                else -> "S${item.seasonNumber} · E${item.episodeNumber}  ${item.episodeTitle}".trim()
            },
            description = "已观看 ${item.progressPercent.coerceIn(0, 100)}%",
            posterUrl = item.posterUrl,
            backdropUrl = item.backdropUrl ?: item.posterUrl,
            progressPercent = item.progressPercent,
            seasonNumber = item.seasonNumber,
            episodeNumber = item.episodeNumber,
            videoId = item.videoId,
        )
    }
    sections.firstNotNullOfOrNull { section ->
        section.items.firstOrNull()?.let { series ->
            TvFeaturedContentUiModel(
                source = TvFeaturedContentSource.SECTION,
                targetId = series.id,
                targetType = "tv",
                eyebrow = section.title.ifBlank { "电视剧推荐" },
                title = series.title,
                subtitle = series.subtitle,
                description = series.description,
                posterUrl = series.posterUrl,
                backdropUrl = series.backdropUrl ?: series.posterUrl,
            )
        }
    }?.let { return it }
    listOf(
        "电视剧精选" to tvSeries.firstOrNull(),
        "电影精选" to movies.firstOrNull(),
        "AV 精选" to av.firstOrNull(),
    ).firstOrNull { it.second != null }?.let { (eyebrow, item) ->
        item ?: return null
        return TvFeaturedContentUiModel(
            source = TvFeaturedContentSource.SHELF,
            targetId = item.id,
            targetType = item.type,
            eyebrow = eyebrow,
            title = item.title,
            subtitle = tvTypeLabel(item.type),
            description = item.description,
            posterUrl = item.posterUrl,
            backdropUrl = item.backdropUrl ?: item.posterUrl,
            progressPercent = item.progressPercent,
        )
    }
    return null
}

internal fun buildTvLongFormDetailHero(
    baseUrl: String,
    videoType: String,
    detail: VideoDetailDto,
): TvLongFormDetailHeroUiModel {
    val normalizedType = normalizeTvLongFormVideoType(videoType)
    val posterUrl = when (normalizedType) {
        "av" -> resolveAvPosterUrl(baseUrl, detail)
        else -> resolveTvResourceUrl(baseUrl, detail.thumbnailPath)
    }
    val backdropUrl = resolveTvBackdropUrl(baseUrl, detail) ?: posterUrl
    return TvLongFormDetailHeroUiModel(
        eyebrow = when (normalizedType) {
            "av" -> "AV"
            else -> "电影"
        },
        title = detail.title,
        metaLine = buildTvLongFormMetaLine(detail),
        summary = detail.description.orEmpty().ifBlank { "暂无简介" },
        posterUrl = posterUrl,
        backdropUrl = backdropUrl,
        primaryActionLabel = "立即播放",
        secondaryActionLabel = "更多信息",
    )
}

internal fun normalizeTvLongFormVideoType(videoType: String): String {
    return when (videoType.trim().lowercase()) {
        "av" -> "av"
        else -> "movie"
    }
}

internal fun resolveTvLongFormPlayUrl(
    baseUrl: String,
    detail: VideoDetailDto,
    preferredPlaybackProfile: String,
): String? {
    val raw = detail.playUrl?.trim().orEmpty()
    if (raw.isNotBlank()) {
        val resolved = resolveTvResourceUrl(baseUrl, raw) ?: return null
        return appendTvPlaybackProfileQuery(resolved, preferredPlaybackProfile)
    }
    val normalizedBase = UrlBuilder.normalizeBaseUrl(baseUrl)
    if (normalizedBase.isBlank()) {
        return null
    }
    return UrlBuilder.source(normalizedBase, detail.id, preferredPlaybackProfile)
}

internal fun resolveTvResourceUrl(baseUrl: String, rawUrl: String?): String? {
    val path = rawUrl?.trim().orEmpty()
    if (path.isBlank()) {
        return null
    }
    if (path.startsWith("http://") || path.startsWith("https://")) {
        return path
    }
    val normalizedBase = UrlBuilder.normalizeBaseUrl(baseUrl)
    if (normalizedBase.isBlank()) {
        return null
    }
    return if (path.startsWith("/")) "$normalizedBase$path" else "$normalizedBase/$path"
}

private fun resolveTvBackdropUrl(baseUrl: String, detail: VideoDetailDto): String? {
    val candidates = listOf(
        anyString(detail.metadata?.get("backdrop_url")),
        anyString(detail.metadata?.get("backdrop_path")),
        anyString(detail.metadata?.get("fanart_url")),
        anyString(detail.metadata?.get("fanart_path")),
        detail.thumbnailPath,
    )
    return candidates.firstNotNullOfOrNull { resolveTvResourceUrl(baseUrl, it) }
}

private fun buildTvLongFormMetaLine(detail: VideoDetailDto): String {
    val pieces = buildList {
        if (detail.duration > 0) {
            add(formatTvDuration(detail.duration))
        }
        detail.tags.orEmpty().take(2).forEach { tag ->
            if (tag.isNotBlank()) {
                add(tag)
            }
        }
    }
    return pieces.joinToString(" · ")
}

private fun formatTvDuration(totalSeconds: Int): String {
    if (totalSeconds <= 0) {
        return ""
    }
    val hours = totalSeconds / 3600
    val minutes = (totalSeconds % 3600) / 60
    return when {
        hours > 0 && minutes > 0 -> "${hours}小时${minutes}分钟"
        hours > 0 -> "${hours}小时"
        else -> "${minutes.coerceAtLeast(1)}分钟"
    }
}

private fun anyString(value: Any?): String? {
    return (value as? String)?.trim()?.takeIf { it.isNotBlank() }
}

private fun appendTvPlaybackProfileQuery(rawUrl: String, preferredPlaybackProfile: String): String {
    val normalizedProfile = preferredPlaybackProfile.trim()
    if (normalizedProfile.isBlank() || !rawUrl.contains("/source")) {
        return rawUrl
    }
    val separator = if (rawUrl.contains("?")) "&" else "?"
    return "$rawUrl${separator}profile=$normalizedProfile"
}

private fun tvTypeLabel(type: String): String = when (type) {
    "tv" -> "电视剧"
    "movie" -> "电影"
    "av" -> "AV"
    else -> type.ifBlank { "长视频" }
}
