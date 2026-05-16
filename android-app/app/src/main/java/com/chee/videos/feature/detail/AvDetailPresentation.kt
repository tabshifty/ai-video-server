package com.chee.videos.feature.detail

import com.chee.videos.core.model.VideoDetailDto
import java.util.Locale

internal data class AvDetailHeroModel(
    val primaryText: String,
    val secondaryTitle: String,
    val actorNames: List<String>,
    val metaItems: List<String>,
    val overviewText: String,
    val releaseDate: String?,
)

internal data class AvDetailActorModel(
    val id: String?,
    val name: String,
    val avatarUrl: String?,
    val hasAvatar: Boolean,
) {
    val canOpenDetail: Boolean = !id.isNullOrBlank()
}

internal fun buildAvDetailHeroModel(detail: VideoDetailDto): AvDetailHeroModel {
    val primaryText = anyString(detail.metadata?.get("av_code"))
        ?: extractNormalizedAvCode(detail.title)
        ?: detail.title.trim().ifBlank { "未命名作品" }
    val actorNames = detail.actors.orEmpty()
        .map { it.name.trim() }
        .filter { it.isNotBlank() }
        .ifEmpty { anyStringList(detail.metadata?.get("actors")) }
    val releaseDate = anyString(detail.metadata?.get("release_date"))
        ?: anyString(detail.metadata?.get("published_at"))
        ?: anyString(detail.metadata?.get("premiered"))
    val metaItems = buildList {
        if (detail.duration > 0) {
            add(formatAvDetailDuration(detail.duration))
        }
        releaseDate?.let(::add)
        detail.tags.orEmpty().firstOrNull()?.trim()?.takeIf { it.isNotBlank() }?.let(::add)
    }
    return AvDetailHeroModel(
        primaryText = primaryText,
        secondaryTitle = detail.title.trim().ifBlank { primaryText },
        actorNames = actorNames,
        metaItems = metaItems,
        overviewText = detail.description.orEmpty().ifBlank { "暂无简介" },
        releaseDate = releaseDate,
    )
}

internal fun buildAvDetailActorModels(baseUrl: String, detail: VideoDetailDto): List<AvDetailActorModel> {
    val apiActors = detail.actors.orEmpty()
        .mapNotNull { actor ->
            val name = actor.name.trim()
            if (name.isBlank()) {
                return@mapNotNull null
            }
            val resolvedAvatarUrl = resolveResourceUrl(baseUrl, actor.avatarUrl)
            AvDetailActorModel(
                id = actor.id.trim().takeIf { it.isNotBlank() },
                name = name,
                avatarUrl = resolvedAvatarUrl,
                hasAvatar = !resolvedAvatarUrl.isNullOrBlank(),
            )
        }
        .distinctBy { it.name.lowercase(Locale.ROOT) }
    if (apiActors.isNotEmpty()) {
        return apiActors
    }
    return anyStringList(detail.metadata?.get("actors"))
        .map { name ->
            AvDetailActorModel(
                id = null,
                name = name,
                avatarUrl = null,
                hasAvatar = false,
            )
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

private fun extractNormalizedAvCode(rawText: String): String? {
    val match = AV_CODE_REGEX.find(rawText.trim()) ?: return null
    val prefix = match.groupValues[1].uppercase(Locale.ROOT)
    val number = match.groupValues[2]
    return "$prefix-$number"
}

private fun formatAvDetailDuration(totalSeconds: Int): String {
    val safeSeconds = totalSeconds.coerceAtLeast(0)
    val hours = safeSeconds / 3600
    val minutes = (safeSeconds % 3600) / 60
    val seconds = safeSeconds % 60
    return String.format("%02d:%02d:%02d", hours, minutes, seconds)
}

private val AV_CODE_REGEX = Regex("""(?i)\b([a-z]{2,10})[-_\s]?(\d{2,5})\b""")
