package com.chee.videos.core.ui

import androidx.media3.common.C
import androidx.media3.common.Format
import androidx.media3.common.TrackSelectionOverride
import androidx.media3.common.TrackSelectionParameters
import androidx.media3.common.Tracks
import java.util.Locale

data class LongFormAudioTrack(
    val id: String,
    val label: String,
    val groupIndex: Int,
    val trackIndex: Int,
    val selected: Boolean,
)

internal data class AudioTrackPickerItem(
    val trackId: String?,
    val label: String,
    val selected: Boolean,
)

fun buildLongFormAudioTracks(tracks: Tracks): List<LongFormAudioTrack> {
    return buildList {
        tracks.groups.forEachIndexed { groupIndex, group ->
            if (group.type != C.TRACK_TYPE_AUDIO) {
                return@forEachIndexed
            }
            for (trackIndex in 0 until group.length) {
                if (!group.isTrackSupported(trackIndex, true)) {
                    continue
                }
                val format = group.getTrackFormat(trackIndex)
                add(
                    LongFormAudioTrack(
                        id = buildAudioTrackId(groupIndex, trackIndex, format),
                        label = audioTrackDisplayLabel(format, size + 1),
                        groupIndex = groupIndex,
                        trackIndex = trackIndex,
                        selected = group.isTrackSelected(trackIndex),
                    ),
                )
            }
        }
    }
}

internal fun buildAudioTrackPickerItems(
    tracks: List<LongFormAudioTrack>,
    selectedAudioTrackId: String?,
): List<AudioTrackPickerItem> {
    val normalizedSelection = selectedAudioTrackId?.trim().orEmpty()
    return buildList {
        add(
            AudioTrackPickerItem(
                trackId = null,
                label = "默认音轨",
                selected = normalizedSelection.isBlank(),
            ),
        )
        tracks.forEach { track ->
            add(
                AudioTrackPickerItem(
                    trackId = track.id,
                    label = track.label,
                    selected = normalizedSelection == track.id,
                ),
            )
        }
    }
}

fun resolveAudioSelectionOnTrackLoad(
    storedSelection: String?,
    tracks: List<LongFormAudioTrack>,
): String? {
    val normalizedSelection = storedSelection?.trim().orEmpty()
    if (normalizedSelection.isBlank()) {
        return null
    }
    return normalizedSelection.takeIf { selected -> tracks.any { it.id == selected } }
}

fun buildAudioTrackSelectionParameters(
    currentParameters: TrackSelectionParameters,
    currentTracks: Tracks,
    audioTracks: List<LongFormAudioTrack>,
    selectedAudioTrackId: String?,
): TrackSelectionParameters {
    val builder = currentParameters
        .buildUpon()
        .clearOverridesOfType(C.TRACK_TYPE_AUDIO)
        .setTrackTypeDisabled(C.TRACK_TYPE_AUDIO, false)
    val normalizedSelection = selectedAudioTrackId?.trim().orEmpty()
    if (normalizedSelection.isBlank()) {
        return builder.build()
    }
    val selected = audioTracks.firstOrNull { it.id == normalizedSelection } ?: return builder.build()
    val group = currentTracks.groups.getOrNull(selected.groupIndex) ?: return builder.build()
    builder.setOverrideForType(
        TrackSelectionOverride(
            group.mediaTrackGroup,
            selected.trackIndex,
        ),
    )
    return builder.build()
}

private fun buildAudioTrackId(groupIndex: Int, trackIndex: Int, format: Format): String {
    val rawParts = listOf(
        "audio",
        groupIndex.toString(),
        trackIndex.toString(),
        format.id.orEmpty(),
        format.language.orEmpty(),
        format.channelCount.takeIf { it > 0 }?.toString().orEmpty(),
    )
    return rawParts
        .filter { it.isNotBlank() }
        .joinToString("-")
        .replace(Regex("[^A-Za-z0-9_-]+"), "-")
        .trim('-')
        .ifBlank { "audio-$groupIndex-$trackIndex" }
}

private fun audioTrackDisplayLabel(format: Format, fallbackIndex: Int): String {
    val base = format.label?.trim()
        ?.takeIf { it.isNotBlank() }
        ?: languageDisplayName(format.language)
        ?: "音轨 $fallbackIndex"
    val channelLabel = audioChannelLabel(format.channelCount)
    return if (channelLabel.isBlank() || base.contains(channelLabel)) {
        base
    } else {
        "$base $channelLabel"
    }
}

private fun languageDisplayName(language: String?): String? {
    val normalized = language?.trim().orEmpty()
    if (normalized.isBlank()) {
        return null
    }
    return when (normalized.lowercase(Locale.ROOT)) {
        "zh", "zho", "chi", "cmn" -> "中文"
        "en", "eng" -> "英语"
        "ja", "jpn", "jp" -> "日语"
        "ko", "kor" -> "韩语"
        "yue" -> "粤语"
        else -> Locale.forLanguageTag(normalized).getDisplayLanguage(Locale.SIMPLIFIED_CHINESE)
            .takeIf { it.isNotBlank() && it != normalized }
            ?: normalized
    }
}

private fun audioChannelLabel(channelCount: Int): String {
    return when (channelCount) {
        1 -> "单声道"
        2 -> "2.0"
        6 -> "5.1"
        8 -> "7.1"
        else -> if (channelCount > 0) "${channelCount}声道" else ""
    }
}
