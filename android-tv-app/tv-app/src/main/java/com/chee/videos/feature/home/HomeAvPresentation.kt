package com.chee.videos.feature.home

import com.chee.videos.core.model.VideoListItemDto
import java.util.Locale

internal data class AvCatalogCardModel(
    val primaryText: String,
    val secondaryText: String?,
    val metaText: String,
)

internal fun buildAvCatalogCardModel(item: VideoListItemDto): AvCatalogCardModel {
    val primaryText = resolveAvCatalogPrimaryText(item)
    val secondaryText = item.title.trim().takeIf {
        it.isNotBlank() && !it.equals(primaryText, ignoreCase = true)
    }
    val metaText = buildList {
        if (item.duration > 0) {
            add(formatAvCatalogDuration(item.duration))
        }
        item.createdAt?.trim()?.takeIf { it.isNotBlank() }?.take(10)?.let(::add)
    }.ifEmpty { listOf("AV 作品") }.joinToString(" · ")
    return AvCatalogCardModel(
        primaryText = primaryText,
        secondaryText = secondaryText,
        metaText = metaText,
    )
}

internal fun resolveAvCatalogPrimaryText(item: VideoListItemDto): String {
    return anyString(item.metadata?.get("av_code"))
        ?: extractNormalizedAvCode(item.title)
        ?: item.title.trim().ifBlank { "未命名作品" }
}

internal fun extractNormalizedAvCode(rawText: String): String? {
    val match = AV_CODE_REGEX.find(rawText.trim()) ?: return null
    val prefix = match.groupValues[1].uppercase(Locale.ROOT)
    val number = match.groupValues[2]
    return "$prefix-$number"
}

private fun anyString(value: Any?): String? {
    return (value as? String)?.trim()?.takeIf { it.isNotBlank() }
}

private fun formatAvCatalogDuration(totalSeconds: Int): String {
    val safeSeconds = totalSeconds.coerceAtLeast(0)
    val hours = safeSeconds / 3600
    val minutes = (safeSeconds % 3600) / 60
    val seconds = safeSeconds % 60
    return String.format("%02d:%02d:%02d", hours, minutes, seconds)
}

private val AV_CODE_REGEX = Regex("""(?i)\b([a-z]{2,10})[-_\s]?(\d{2,5})\b""")
