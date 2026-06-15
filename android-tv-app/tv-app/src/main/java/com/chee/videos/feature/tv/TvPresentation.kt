package com.chee.videos.feature.tv

import com.chee.videos.core.model.VideoDetailDto
import com.chee.videos.core.model.TvCatalogWallItemDto
import com.chee.videos.core.model.resolveAvPoster
import com.chee.videos.core.model.resolveAvPosterUrl
import com.chee.videos.core.util.UrlBuilder
import java.util.Locale

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
    val usesPosterAsBackdropFallback: Boolean = false,
    val actors: List<TvLongFormDetailActorUiModel> = emptyList(),
    val primaryActionLabel: String,
    val secondaryActionLabel: String,
)

internal data class TvLongFormDetailActorUiModel(
    val name: String,
    val avatarUrl: String?,
    val hasAvatar: Boolean,
)

internal data class TvCatalogWallSpec(
    val kind: String,
    val title: String,
    val subtitle: String,
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
                "av" -> "继续看18+"
                else -> "继续追剧"
            },
            title = item.seriesTitle,
            subtitle = when (item.type) {
                "movie" -> item.episodeTitle.ifBlank { "继续播放" }
                "av" -> item.episodeTitle.ifBlank { "继续播放" }
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
        "18+精选" to av.firstOrNull(),
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
    detail: VideoDetailDto,
    videoType: String = "movie",
): TvLongFormDetailHeroUiModel {
    val posterUrl = resolveTvResourceUrl(baseUrl, detail.thumbnailPath)
    val normalizedVideoType = normalizeTvLongFormVideoType(videoType)
    val preferredBackdrop = if (normalizedVideoType == "av") {
        val avPoster = resolveAvPoster(detail)
        val url = resolveAvPosterUrl(baseUrl, detail)
        TvLongFormBackdropCandidate(
            url = url,
            isPosterFallback = url != null && !avPoster.isScrapedPoster,
        )
    } else {
        TvLongFormBackdropCandidate(url = resolveTvBackdropUrl(baseUrl, detail))
    }
    val preferredBackdropUrl = preferredBackdrop.url
    val backdropUrl = preferredBackdropUrl ?: posterUrl
    return TvLongFormDetailHeroUiModel(
        eyebrow = tvLongFormTypeLabel(normalizedVideoType),
        title = detail.title,
        metaLine = buildTvLongFormMetaLine(detail),
        summary = detail.description.orEmpty().ifBlank { "暂无简介" },
        posterUrl = posterUrl,
        backdropUrl = backdropUrl,
        usesPosterAsBackdropFallback = preferredBackdrop.isPosterFallback ||
            (preferredBackdropUrl.isNullOrBlank() && !posterUrl.isNullOrBlank()),
        actors = buildTvLongFormDetailActors(baseUrl, detail),
        primaryActionLabel = "播放",
        secondaryActionLabel = if (detail.userState.isFavorited) "取消收藏" else "收藏",
    )
}

private data class TvLongFormBackdropCandidate(
    val url: String?,
    val isPosterFallback: Boolean = false,
)

internal fun resolveTvCatalogWallSpec(kind: String, fallbackTitle: String = ""): TvCatalogWallSpec {
    val normalizedKind = kind.trim().lowercase()
    val title = fallbackTitle.trim().takeIf { it.isNotBlank() } ?: when (normalizedKind) {
        "recent" -> "最近更新"
        "binge" -> "高能连播"
        "classic" -> "经典补档"
        "tv" -> "电视剧"
        "movie" -> "电影"
        "av" -> "18+"
        else -> "海报墙"
    }
    val subtitle = when (normalizedKind) {
        "recent" -> "按最近播出时间排序"
        "binge" -> "优先展示可直接播放的剧集"
        "classic" -> "从较早首播的系列开始补看"
        "tv" -> "全部电视剧"
        "movie" -> "全部电影"
        "av" -> "全部18+"
        else -> "滚动到底部自动加载下一页"
    }
    return TvCatalogWallSpec(
        kind = normalizedKind,
        title = title,
        subtitle = subtitle,
    )
}

internal fun tvCatalogWallItemToUiModel(item: TvCatalogWallItemDto): TvCatalogWallItemUiModel {
    val normalizedType = normalizeTvWallItemType(item.type)
    return TvCatalogWallItemUiModel(
        id = item.id,
        type = normalizedType,
        title = item.title,
        description = item.overview.ifBlank { tvCatalogWallTypeLabel(normalizedType) },
        posterUrl = item.posterUrl,
        backdropUrl = item.backdropUrl,
        videoId = item.videoId,
        seasonNumber = item.seasonNumber,
        episodeNumber = item.episodeNumber,
        progressPercent = item.progressPercent,
    )
}

private fun normalizeTvWallItemType(raw: String): String {
    return when (raw.trim().lowercase()) {
        "episode" -> "tv"
        else -> raw.trim().lowercase()
    }
}

private fun tvCatalogWallTypeLabel(type: String): String {
    return when (type) {
        "tv" -> "电视剧"
        "movie" -> "电影"
        "av" -> "18+"
        else -> type.ifBlank { "长视频" }
    }
}

internal fun normalizeTvLongFormVideoType(videoType: String): String {
    return when (videoType.trim().lowercase()) {
        "movie", "av" -> videoType.trim().lowercase()
        else -> "movie"
    }
}

internal fun resolveTvLongFormPlayUrl(
    baseUrl: String,
    detail: VideoDetailDto,
    preferredPlaybackProfile: String,
    overridePlaybackProfile: String? = null,
): String? {
    val playbackProfile = overridePlaybackProfile?.trim().takeUnless { it.isNullOrBlank() } ?: preferredPlaybackProfile
    val raw = detail.playUrl?.trim().orEmpty()
    if (raw.isNotBlank()) {
        val resolved = resolveTvResourceUrl(baseUrl, raw) ?: return null
        return appendTvPlaybackProfileQuery(resolved, playbackProfile)
    }
    val normalizedBase = UrlBuilder.normalizeBaseUrl(baseUrl)
    if (normalizedBase.isBlank()) {
        return null
    }
    return UrlBuilder.source(normalizedBase, detail.id, playbackProfile)
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
    if (hasLocalMovieBackdrop(detail.metadata)) {
        return UrlBuilder.thumbnail(baseUrl, detail.id, variant = "backdrop")
    }
    val candidates = listOf(
        anyString(detail.metadata?.get("backdrop_url")),
        anyString(detail.metadata?.get("backdrop_path")),
        anyString(detail.metadata?.get("fanart_url")),
        anyString(detail.metadata?.get("fanart_path")),
    )
    return candidates.firstNotNullOfOrNull { resolveTvResourceUrl(baseUrl, it) }
}

private fun hasLocalMovieBackdrop(metadata: Map<String, Any?>?): Boolean {
    if (metadata.isNullOrEmpty()) {
        return false
    }
    val direct = listOf(
        anyString(metadata["backdrop_url"]),
        anyString(metadata["backdrop_path"]),
    )
    if (direct.any(::isLocalMovieBackdropPath)) {
        return true
    }
    val tmdb = metadata["tmdb"] as? Map<*, *> ?: return false
    return listOf(
        anyString(tmdb["backdrop_url"]),
        anyString(tmdb["backdrop_path"]),
    ).any(::isLocalMovieBackdropPath)
}

private fun isLocalMovieBackdropPath(raw: String?): Boolean {
    val value = raw?.trim().orEmpty()
    if (value.isBlank()) {
        return false
    }
    if (value.startsWith("http://") || value.startsWith("https://") || value.startsWith("/api/")) {
        return false
    }
    return value.replace('\\', '/').contains("/videos/") && value.contains("backdrop", ignoreCase = true)
}

private fun buildTvLongFormMetaLine(detail: VideoDetailDto): String {
    val pieces = buildList {
        extractTvLongFormYear(detail.metadata)?.let(::add)
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

private fun buildTvLongFormDetailActors(
    baseUrl: String,
    detail: VideoDetailDto,
): List<TvLongFormDetailActorUiModel> {
    val apiActors = detail.actors.orEmpty()
        .mapNotNull { actor ->
            val name = actor.name.trim()
            if (name.isBlank()) {
                return@mapNotNull null
            }
            val avatarUrl = resolveTvResourceUrl(baseUrl, actor.avatarUrl)
            TvLongFormDetailActorUiModel(
                name = name,
                avatarUrl = avatarUrl,
                hasAvatar = !avatarUrl.isNullOrBlank(),
            )
        }
        .distinctBy { it.name.lowercase(Locale.ROOT) }
    val actors = apiActors.ifEmpty {
        anyStringList(detail.metadata?.get("actors")).map { name ->
            TvLongFormDetailActorUiModel(
                name = name,
                avatarUrl = null,
                hasAvatar = false,
            )
        }
    }
    return actors.take(5)
}

private fun extractTvLongFormYear(metadata: Map<String, Any?>?): String? {
    val directYear = metadata?.get("year")
    when (directYear) {
        is Number -> return directYear.toInt().takeIf { it > 0 }?.toString()
        is String -> directYear.trim().takeIf { it.matches(Regex("""\d{4}""")) }?.let { return it }
    }
    return listOf("release_date", "released_at", "published_at", "premiered", "air_date")
        .firstNotNullOfOrNull { key ->
            anyString(metadata?.get(key))?.let { YEAR_REGEX.find(it)?.value }
        }
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

private fun tvLongFormTypeLabel(videoType: String): String {
    return when (normalizeTvLongFormVideoType(videoType)) {
        "av" -> "18+"
        else -> "电影"
    }
}

private fun anyString(value: Any?): String? {
    return (value as? String)?.trim()?.takeIf { it.isNotBlank() }
}

private fun anyStringList(value: Any?): List<String> {
    return when (value) {
        is List<*> -> value.mapNotNull(::anyString).distinct()
        is String -> value.split(',', '、', '/', '|')
            .map { it.trim() }
            .filter { it.isNotBlank() }
            .distinct()
        else -> emptyList()
    }
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
    else -> type.ifBlank { "长视频" }
}

private val YEAR_REGEX = Regex("""\b(19|20)\d{2}\b""")
