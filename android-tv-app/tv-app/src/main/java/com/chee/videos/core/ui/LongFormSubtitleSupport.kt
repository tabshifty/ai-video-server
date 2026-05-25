package com.chee.videos.core.ui

import com.chee.videos.core.model.SubtitleTrackDto
import com.chee.videos.core.model.TvTrackPreference
import com.chee.videos.core.player.TvLongFormVlcMediaSpec
import com.chee.videos.core.player.buildLongFormMedia
import com.chee.videos.core.player.normalizeLongFormLanguageCode
import com.chee.videos.core.util.UrlBuilder
import java.net.URI
import org.videolan.libvlc.LibVLC
import org.videolan.libvlc.MediaPlayer

internal data class LongFormPlayerUpdateDecision(
    val shouldClear: Boolean,
    val shouldReplaceSource: Boolean,
    val preservePosition: Boolean,
)

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

internal fun resolveLongFormPlayerUpdate(
    preparedUrl: String?,
    nextUrl: String?,
    preparedSubtitleTrackId: String?,
    nextSubtitleTrackId: String?,
): LongFormPlayerUpdateDecision {
    val normalizedPreparedUrl = preparedUrl?.trim().orEmpty()
    val normalizedNextUrl = nextUrl?.trim().orEmpty()
    if (normalizedNextUrl.isBlank()) {
        return LongFormPlayerUpdateDecision(
            shouldClear = normalizedPreparedUrl.isNotBlank(),
            shouldReplaceSource = false,
            preservePosition = false,
        )
    }
    if (normalizedPreparedUrl.isBlank()) {
        return LongFormPlayerUpdateDecision(
            shouldClear = false,
            shouldReplaceSource = true,
            preservePosition = false,
        )
    }
    if (normalizedPreparedUrl != normalizedNextUrl) {
        return LongFormPlayerUpdateDecision(
            shouldClear = false,
            shouldReplaceSource = true,
            preservePosition = false,
        )
    }
    if (preparedSubtitleTrackId != nextSubtitleTrackId) {
        return LongFormPlayerUpdateDecision(
            shouldClear = false,
            shouldReplaceSource = true,
            preservePosition = true,
        )
    }
    return LongFormPlayerUpdateDecision(
        shouldClear = false,
        shouldReplaceSource = false,
        preservePosition = false,
    )
}

fun resolveSelectedSubtitleTrack(tracks: List<SubtitleTrackDto>, selectedTrackId: String?): SubtitleTrackDto? {
    val normalizedId = selectedTrackId?.trim().orEmpty()
    if (normalizedId.isBlank()) {
        return null
    }
    return tracks.firstOrNull { it.id == normalizedId && it.available && it.url.isNotBlank() }
}

fun subtitleTrackDisplayLabel(track: SubtitleTrackDto): String {
    val label = track.label.trim().ifBlank { track.languageLabel.trim() }.ifBlank { track.languageCode.trim() }
    if (label.isNotBlank()) {
        return label
    }
    return if (track.sourceType == "embedded") "内嵌字幕" else "外挂字幕"
}

fun applyLongFormMediaSource(
    libVLC: LibVLC,
    mediaPlayer: MediaPlayer,
    sourceUrl: String,
    baseUrl: String,
    selectedSubtitleTrack: SubtitleTrackDto?,
) {
    val subtitleUrl = selectedSubtitleTrack
        ?.takeIf { it.available && it.url.isNotBlank() }
        ?.let { resolvePlaybackAssetUrl(baseUrl, it.url) }
    val media = buildLongFormMedia(
        libVLC = libVLC,
        spec = TvLongFormVlcMediaSpec(
            sourceUrl = sourceUrl,
            subtitleUrl = subtitleUrl,
        ),
    )
    mediaPlayer.media = media
    media.release()
}

fun resolveSelectedSubtitleTrackByLanguage(
    tracks: List<SubtitleTrackDto>,
    preferredLanguage: String?,
): SubtitleTrackDto? {
    return resolveSelectedSubtitleTrackByPreference(
        tracks = tracks,
        preference = TvTrackPreference(language = preferredLanguage.orEmpty()),
    )
}

fun resolveSelectedSubtitleTrackByPreference(
    tracks: List<SubtitleTrackDto>,
    preference: TvTrackPreference?,
): SubtitleTrackDto? {
    val safePreference = preference ?: return null
    val normalizedPreference = normalizeLongFormLanguageCode(safePreference.language)
    if (normalizedPreference.isBlank()) {
        return null
    }
    val typePreference = safePreference.type.trim().lowercase()
    val languageMatches = tracks.filter { track ->
        val language = normalizeLongFormLanguageCode(track.languageCode)
        track.available &&
            (
                language == normalizedPreference ||
                    language.startsWith("$normalizedPreference-") ||
                    normalizedPreference.startsWith("$language-")
                )
    }
    if (languageMatches.isEmpty()) {
        return null
    }
    if (typePreference.isNotBlank()) {
        languageMatches.firstOrNull { subtitlePreferenceType(it) == typePreference }?.let { return it }
    }
    return languageMatches.first()
}

fun buildSubtitleTrackPreference(track: SubtitleTrackDto?): TvTrackPreference? {
    val safeTrack = track ?: return null
    val language = normalizeLongFormLanguageCode(safeTrack.languageCode)
    val type = subtitlePreferenceType(safeTrack)
    val preference = TvTrackPreference(language = language, type = type)
    return preference.takeUnless { it.isBlank() }
}

private fun subtitlePreferenceType(track: SubtitleTrackDto): String {
    val label = listOf(track.label, track.languageLabel, track.sourceType)
        .joinToString(" ")
        .lowercase()
    return when {
        "forced" in label || "强制" in label -> "forced"
        "commentary" in label || "comment" in label || "解说" in label || "评论" in label -> "commentary"
        track.isDefault -> "default"
        else -> ""
    }
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
