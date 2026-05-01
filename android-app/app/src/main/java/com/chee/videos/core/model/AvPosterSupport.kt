package com.chee.videos.core.model

import com.chee.videos.core.util.UrlBuilder

internal data class AvPoster(
    val rawUrl: String?,
    val isScrapedPoster: Boolean,
    val posterSource: String?,
    val posterQuality: String?,
    val posterDecision: String?,
)

internal fun resolveAvPoster(item: VideoListItemDto): AvPoster {
    return buildAvPoster(item.metadata, item.thumbnailPath, preferOriginalPoster = false)
}

internal fun resolveAvPoster(detail: VideoDetailDto): AvPoster {
    return buildAvPoster(detail.metadata, detail.thumbnailPath, preferOriginalPoster = true)
}

internal fun resolveAvPosterUrl(baseUrl: String, item: VideoListItemDto): String? {
    return resolveAvPosterUrl(baseUrl, resolveAvPoster(item))
}

internal fun resolveAvPosterUrl(baseUrl: String, detail: VideoDetailDto): String? {
    return resolveAvPosterUrl(baseUrl, resolveAvPoster(detail))
}

private fun resolveAvPosterUrl(baseUrl: String, poster: AvPoster): String? {
    val path = poster.rawUrl?.trim().orEmpty()
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

private fun buildAvPoster(
    metadata: Map<String, Any?>?,
    thumbnailPath: String?,
    preferOriginalPoster: Boolean,
): AvPoster {
    val posterDecision = anyString(metadata?.get("poster_decision"))
    val posterSource = anyString(metadata?.get("poster_source"))
    val posterQuality = anyString(metadata?.get("poster_quality"))
    val posterVariant = anyString(metadata?.get("poster_variant"))
    val croppedPosterPath = anyString(metadata?.get("poster_cropped_path"))
    val originalPosterPath = anyString(metadata?.get("poster_original_path"))
    val scrapeSource = anyString(metadata?.get("scrape_source"))
    val sourceBlock = metadata?.get(scrapeSource) as? Map<*, *>
    val sourcePosterUrl = anyString(sourceBlock?.get("poster_url"))
    val sourcePosterPath = anyString(sourceBlock?.get("poster_path"))
    val localizedPosterPath = selectLocalizedPosterPath(
        posterVariant = posterVariant,
        croppedPosterPath = croppedPosterPath,
        originalPosterPath = originalPosterPath,
        preferOriginalPoster = preferOriginalPoster,
    )
    val rawScrapedPoster = listOf(
        localizedPosterPath,
        anyString(metadata?.get("poster_url")),
        anyString(metadata?.get("poster_path")),
        sourcePosterUrl,
        sourcePosterPath,
    ).firstOrNull { !it.isNullOrBlank() }

    if (!posterDecision.isNullOrBlank() && posterDecision.startsWith("invalid")) {
        return AvPoster(
            rawUrl = anyString(thumbnailPath),
            isScrapedPoster = false,
            posterSource = posterSource,
            posterQuality = posterQuality,
            posterDecision = posterDecision,
        )
    }

    if (!rawScrapedPoster.isNullOrBlank()) {
        return AvPoster(
            rawUrl = rawScrapedPoster,
            isScrapedPoster = true,
            posterSource = posterSource ?: anyString(sourceBlock?.get("poster_source")),
            posterQuality = posterQuality ?: anyString(sourceBlock?.get("poster_quality")),
            posterDecision = posterDecision,
        )
    }

    return AvPoster(
        rawUrl = anyString(thumbnailPath),
        isScrapedPoster = false,
        posterSource = posterSource,
        posterQuality = posterQuality,
        posterDecision = posterDecision,
    )
}

private fun selectLocalizedPosterPath(
    posterVariant: String?,
    croppedPosterPath: String?,
    originalPosterPath: String?,
    preferOriginalPoster: Boolean,
): String? {
    if (preferOriginalPoster) {
        return when (posterVariant) {
            "original" -> originalPosterPath ?: croppedPosterPath
            "cropped" -> originalPosterPath ?: croppedPosterPath
            else -> originalPosterPath ?: croppedPosterPath
        }
    }
    return when (posterVariant) {
        "cropped" -> croppedPosterPath ?: originalPosterPath
        "original" -> croppedPosterPath ?: originalPosterPath
        else -> croppedPosterPath ?: originalPosterPath
    }
}

private fun anyString(value: Any?): String? {
    return (value as? String)?.trim()?.takeIf { it.isNotBlank() }
}
