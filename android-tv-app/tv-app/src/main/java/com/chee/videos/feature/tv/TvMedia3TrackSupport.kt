package com.chee.videos.feature.tv

import android.net.Uri
import androidx.media3.common.C
import androidx.media3.common.Format
import androidx.media3.common.MediaItem
import androidx.media3.common.MimeTypes
import androidx.media3.common.TrackGroup
import androidx.media3.common.TrackSelectionOverride
import androidx.media3.common.TrackSelectionParameters
import androidx.media3.common.Tracks
import com.chee.videos.core.model.SubtitleTrackDto
import com.chee.videos.core.ui.LongFormAudioTrack
import com.chee.videos.core.ui.subtitleTrackDisplayLabel
import java.util.Locale

internal fun buildTvMedia3SubtitleConfigurations(
    tracks: List<SubtitleTrackDto>,
    baseUrl: String,
): List<MediaItem.SubtitleConfiguration> =
    tracks
        .filter { it.available && it.url.isNotBlank() && !it.isEmbedded }
        .mapNotNull { track ->
            val uri = resolvePlaybackAssetUrlForMedia3Track(baseUrl, track.url)
            if (uri.isBlank()) {
                return@mapNotNull null
            }
            MediaItem.SubtitleConfiguration.Builder(Uri.parse(uri))
                .setMimeType(resolveTvMedia3SubtitleMimeType(track))
                .setLanguage(track.languageCode.ifBlank { null })
                .setLabel(subtitleTrackDisplayLabel(track))
                .setId(track.id)
                .setSelectionFlags(if (track.isDefault) C.SELECTION_FLAG_DEFAULT else 0)
                .build()
        }

internal fun resolveTvMedia3SubtitleMimeType(track: SubtitleTrackDto): String {
    return resolveTvMedia3SubtitleMimeType(
        mimeType = track.mimeType,
        format = track.format,
    )
}

internal fun resolveTvMedia3SubtitleMimeType(mimeType: String, format: String): String {
    val normalizedMime = mimeType.trim().lowercase(Locale.ROOT)
    if (normalizedMime.isNotBlank()) {
        return when (normalizedMime) {
            "application/x-subrip", "text/srt" -> MimeTypes.APPLICATION_SUBRIP
            "text/vtt", "text/webvtt" -> MimeTypes.TEXT_VTT
            "text/x-ssa", "text/x-ass", "application/x-ass", "application/ass" -> MimeTypes.TEXT_SSA
            else -> normalizedMime
        }
    }
    return when (format.trim().lowercase(Locale.ROOT)) {
        "srt", "subrip" -> MimeTypes.APPLICATION_SUBRIP
        "vtt", "webvtt" -> MimeTypes.TEXT_VTT
        "ass", "ssa" -> MimeTypes.TEXT_SSA
        else -> MimeTypes.APPLICATION_SUBRIP
    }
}

internal fun buildTvMedia3AudioTracks(tracks: Tracks): List<LongFormAudioTrack> =
    tracks.groups
        .filter { group -> group.type == C.TRACK_TYPE_AUDIO }
        .flatMapIndexed { groupIndex, group ->
            (0 until group.length)
                .filter { group.isTrackSupported(it, true) }
                .map { trackIndex ->
                    val format = group.getTrackFormat(trackIndex)
                    val language = format.language.orEmpty()
                    val channelDetail = audioChannelLabelForMedia3(format.channelCount)
                    val codecDetail = audioCodecLabelForMedia3(format.sampleMimeType)
                    val detail = listOf(channelDetail, codecDetail)
                        .filter { it.isNotBlank() }
                        .joinToString(" · ")
                    LongFormAudioTrack(
                        id = buildTvMedia3TrackId("audio", groupIndex, trackIndex, format),
                        vlcTrackId = buildTvMedia3AudioPreferenceTrackId(groupIndex, trackIndex),
                        label = format.label?.trim()
                            ?.takeIf { it.isNotBlank() }
                            ?: media3LanguageDisplayName(language)
                            ?: "音轨 ${groupIndex + 1}-${trackIndex + 1}",
                        detail = detail,
                        groupIndex = groupIndex,
                        trackIndex = trackIndex,
                        selected = group.isTrackSelected(trackIndex),
                        languageCode = language,
                        preferenceType = media3PreferenceType(format),
                    )
                }
        }

internal fun findTvMedia3TrackGroup(
    tracks: Tracks,
    type: @C.TrackType Int,
    groupIndex: Int,
): TrackGroup? =
    tracks.groups
        .filter { it.type == type }
        .getOrNull(groupIndex)
        ?.getMediaTrackGroup()

internal fun buildTvMedia3SelectionParametersForTrack(
    currentParameters: TrackSelectionParameters,
    tracks: Tracks,
    type: @C.TrackType Int,
    groupIndex: Int,
    trackIndex: Int,
): TrackSelectionParameters {
    val group = findTvMedia3TrackGroup(tracks, type, groupIndex) ?: return currentParameters
    return currentParameters
        .buildUpon()
        .setTrackTypeDisabled(type, false)
        .setOverrideForType(TrackSelectionOverride(group, trackIndex))
        .build()
}

internal fun buildTvMedia3SelectionParametersForAuto(
    currentParameters: TrackSelectionParameters,
    type: @C.TrackType Int,
): TrackSelectionParameters =
    currentParameters
        .buildUpon()
        .setTrackTypeDisabled(type, false)
        .clearOverridesOfType(type)
        .build()

internal fun buildTvMedia3SelectionParametersForDisabledText(
    currentParameters: TrackSelectionParameters,
): TrackSelectionParameters =
    currentParameters
        .buildUpon()
        .clearOverridesOfType(C.TRACK_TYPE_TEXT)
        .setTrackTypeDisabled(C.TRACK_TYPE_TEXT, true)
        .build()

private fun resolvePlaybackAssetUrlForMedia3Track(baseUrl: String, url: String): String {
    val trimmed = url.trim()
    if (trimmed.isBlank()) {
        return ""
    }
    if (trimmed.startsWith("http://") || trimmed.startsWith("https://")) {
        return trimmed
    }
    val normalizedBase = baseUrl.trimEnd('/')
    val normalizedPath = trimmed.trimStart('/')
    return if (normalizedBase.isBlank()) "/$normalizedPath" else "$normalizedBase/$normalizedPath"
}

internal fun buildTvMedia3TrackId(prefix: String, groupIndex: Int, trackIndex: Int, format: Format): String {
    val rawParts = listOf(
        prefix,
        groupIndex.toString(),
        trackIndex.toString(),
        format.id.orEmpty(),
        format.language.orEmpty(),
        format.label.orEmpty(),
    )
    return rawParts
        .filter { it.isNotBlank() }
        .joinToString("-")
        .replace(Regex("[^A-Za-z0-9_-]+"), "-")
        .trim('-')
        .ifBlank { "$prefix-$groupIndex-$trackIndex" }
}

private fun buildTvMedia3AudioPreferenceTrackId(groupIndex: Int, trackIndex: Int): Int =
    (groupIndex.coerceAtLeast(0) * 1000) + trackIndex.coerceAtLeast(0)

private fun media3PreferenceType(format: Format): String {
    return when {
        format.roleFlags and C.ROLE_FLAG_COMMENTARY != 0 -> "commentary"
        format.selectionFlags and C.SELECTION_FLAG_DEFAULT != 0 -> "default"
        else -> ""
    }
}

private fun media3LanguageDisplayName(language: String?): String? {
    val normalized = language?.trim().orEmpty()
    if (normalized.isBlank()) {
        return null
    }
    return when (normalized.lowercase(Locale.ROOT)) {
        "zh", "zho", "chi", "cmn", "zh-cn", "zh-hans" -> "中文"
        "en", "eng" -> "英语"
        "ja", "jpn", "jp" -> "日语"
        "ko", "kor" -> "韩语"
        "yue" -> "粤语"
        else -> Locale.forLanguageTag(normalized).getDisplayLanguage(Locale.SIMPLIFIED_CHINESE)
            .takeIf { it.isNotBlank() && it != normalized }
            ?: normalized
    }
}

private fun audioChannelLabelForMedia3(channelCount: Int): String =
    when (channelCount) {
        1 -> "单声道"
        2 -> "2.0"
        6 -> "5.1"
        8 -> "7.1"
        else -> if (channelCount > 0) "${channelCount}声道" else ""
    }

private fun audioCodecLabelForMedia3(sampleMimeType: String?): String {
    return when (sampleMimeType?.lowercase(Locale.ROOT)) {
        "audio/mp4a-latm", "audio/aac" -> "AAC"
        "audio/ac3" -> "AC-3"
        "audio/eac3" -> "E-AC-3"
        "audio/eac3-joc" -> "Atmos"
        "audio/vnd.dts", "audio/dts" -> "DTS"
        "audio/vnd.dts.hd", "audio/dts-hd" -> "DTS-HD"
        "audio/true-hd" -> "TrueHD"
        "audio/flac" -> "FLAC"
        else -> ""
    }
}
