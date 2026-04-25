package com.chee.videos.core.ui

import android.net.Uri
import androidx.media3.common.C
import androidx.media3.common.MediaItem
import androidx.media3.common.MediaMetadata
import androidx.media3.common.MimeTypes
import com.chee.videos.core.model.SubtitleTrackDto
import com.chee.videos.core.util.UrlBuilder
import java.net.URI

fun resolveInitialSubtitleTrackId(tracks: List<SubtitleTrackDto>): String? {
    return tracks.firstOrNull { it.sourceType == "uploaded" && it.isDefault }?.id
        ?: tracks.firstOrNull { it.isDefault }?.id
}

fun resolveSubtitleSelectionOnTrackLoad(
    currentSelection: String?,
    tracks: List<SubtitleTrackDto>,
    hasStartedPlayback: Boolean,
): String? {
    val normalizedSelection = currentSelection?.trim().orEmpty()
    val selectedStillValid = normalizedSelection.isNotBlank() && tracks.any { it.id == normalizedSelection && it.available }
    if (selectedStillValid) {
        return normalizedSelection
    }
    if (hasStartedPlayback) {
        return null
    }
    return resolveInitialSubtitleTrackId(tracks)
}

fun resolveSelectedSubtitleTrack(tracks: List<SubtitleTrackDto>, selectedTrackId: String?): SubtitleTrackDto? {
    val normalizedId = selectedTrackId?.trim().orEmpty()
    if (normalizedId.isBlank()) {
        return null
    }
    return tracks.firstOrNull { it.id == normalizedId && it.available && it.url.isNotBlank() }
}

fun buildLongFormMediaItem(
    sourceUrl: String,
    mediaId: String,
    title: String,
    baseUrl: String,
    selectedSubtitleTrack: SubtitleTrackDto?,
): MediaItem {
    val subtitleConfigurations = buildList {
        val track = selectedSubtitleTrack ?: return@buildList
        val resolvedUrl = resolvePlaybackAssetUrl(baseUrl, track.url) ?: return@buildList
        val builder = MediaItem.SubtitleConfiguration.Builder(Uri.parse(resolvedUrl))
            .setRoleFlags(C.ROLE_FLAG_SUBTITLE)
            .setSelectionFlags(if (track.isDefault) C.SELECTION_FLAG_DEFAULT else 0)
        val mimeType = resolveSubtitleMimeType(track)
        if (mimeType.isNotBlank()) {
            builder.setMimeType(mimeType)
        }
        if (track.languageCode.isNotBlank()) {
            builder.setLanguage(track.languageCode)
        }
        if (track.label.isNotBlank()) {
            builder.setLabel(track.label)
        }
        add(builder.build())
    }
    return MediaItem.Builder()
        .setUri(sourceUrl)
        .setMediaId(mediaId)
        .setMediaMetadata(
            MediaMetadata.Builder()
                .setTitle(title)
                .build(),
        )
        .setSubtitleConfigurations(subtitleConfigurations)
        .build()
}

fun subtitleTrackDisplayLabel(track: SubtitleTrackDto): String {
    val label = track.label.trim().ifBlank { track.languageLabel.trim() }.ifBlank { track.languageCode.trim() }
    if (label.isNotBlank()) {
        return label
    }
    return if (track.sourceType == "embedded") "内嵌字幕" else "外挂字幕"
}

private fun resolvePlaybackAssetUrl(baseUrl: String, rawUrl: String): String? {
    val path = rawUrl.trim()
    if (path.isBlank()) {
        return null
    }
    if (path.startsWith("http://") || path.startsWith("https://")) {
        return path
    }
    val candidate = baseUrl.trim()
    if (candidate.startsWith("http://") || candidate.startsWith("https://")) {
        return runCatching { URI.create(candidate).resolve(path).toString() }.getOrNull()
    }
    val normalized = UrlBuilder.normalizeBaseUrl(candidate)
    if (normalized.isBlank()) {
        return null
    }
    return if (path.startsWith('/')) "$normalized$path" else "$normalized/$path"
}

private fun resolveSubtitleMimeType(track: SubtitleTrackDto): String {
    val raw = track.mimeType.trim()
    if (raw.isNotBlank()) {
        return raw
    }
    return when (track.format.trim().lowercase()) {
        "srt" -> MimeTypes.APPLICATION_SUBRIP
        "vtt", "webvtt" -> MimeTypes.TEXT_VTT
        else -> ""
    }
}
