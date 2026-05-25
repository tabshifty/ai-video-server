package com.chee.videos.core.ui

import com.chee.videos.core.model.TvTrackPreference
import com.chee.videos.core.player.TvLongFormLanguagePreference
import com.chee.videos.core.player.TvLongFormVlcTrack
import com.chee.videos.core.player.normalizeLongFormLanguageCode
import com.chee.videos.core.player.resolveLongFormTrackByLanguage
import java.util.Locale
import org.videolan.libvlc.MediaPlayer

data class LongFormAudioTrack(
    val id: String,
    val vlcTrackId: Int = -1,
    val label: String,
    val detail: String,
    val groupIndex: Int,
    val trackIndex: Int,
    val selected: Boolean,
    val languageCode: String = "",
    val preferenceType: String = "",
)

internal data class AudioTrackPickerItem(
    val trackId: String?,
    val label: String,
    val detail: String,
    val selected: Boolean,
)

fun buildLongFormAudioTracksFromVlc(
    tracks: Array<MediaPlayer.TrackDescription>?,
    selectedTrackId: Int,
): List<LongFormAudioTrack> {
    if (tracks == null) {
        return emptyList()
    }
    return tracks
        .filter { it.id >= 0 }
        .mapIndexed { index, track ->
            val name = track.name?.trim().orEmpty()
            val language = inferLongFormTrackLanguageCode(name)
            LongFormAudioTrack(
                id = buildAudioTrackId(track.id, name, index),
                vlcTrackId = track.id,
                label = name.ifBlank { "音轨 ${index + 1}" },
                detail = "",
                groupIndex = 0,
                trackIndex = index,
                selected = track.id == selectedTrackId,
                languageCode = language,
                preferenceType = inferLongFormTrackPreferenceType(name),
            )
        }
}

fun buildLongFormSpuTracksFromVlc(
    tracks: Array<MediaPlayer.TrackDescription>?,
    selectedTrackId: Int,
): List<LongFormAudioTrack> {
    if (tracks == null) {
        return emptyList()
    }
    return tracks
        .filter { it.id >= 0 }
        .mapIndexed { index, track ->
            val name = track.name?.trim().orEmpty()
            val language = inferLongFormTrackLanguageCode(name)
            LongFormAudioTrack(
                id = buildAudioTrackId(track.id, name, index),
                vlcTrackId = track.id,
                label = name.ifBlank { "字幕 ${index + 1}" },
                detail = "",
                groupIndex = 0,
                trackIndex = index,
                selected = track.id == selectedTrackId,
                languageCode = language,
                preferenceType = inferLongFormTrackPreferenceType(name),
            )
        }
}

internal fun buildAudioTrackPickerItems(
    tracks: List<LongFormAudioTrack>,
    selectedAudioTrackId: String?,
): List<AudioTrackPickerItem> {
    val normalizedSelection = selectedAudioTrackId?.trim().orEmpty()
    val selectedTrackId = if (normalizedSelection.isBlank()) {
        ""
    } else {
        tracks.firstOrNull { it.selected }?.id ?: normalizedSelection
    }
    return buildList {
        add(
            AudioTrackPickerItem(
                trackId = null,
                label = "自动选择",
                detail = "跟随视频默认音轨",
                selected = selectedTrackId.isBlank(),
            ),
        )
        tracks.forEach { track ->
            add(
                AudioTrackPickerItem(
                    trackId = track.id,
                    label = track.label,
                    detail = track.detail,
                    selected = selectedTrackId == track.id,
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

fun resolveAudioSelectionOnTrackLoad(
    currentSelection: String?,
    storedPreference: TvTrackPreference?,
    tracks: List<LongFormAudioTrack>,
): String? {
    val normalizedSelection = currentSelection?.trim().orEmpty()
    if (normalizedSelection.isNotBlank() && tracks.any { it.id == normalizedSelection }) {
        return normalizedSelection
    }
    val preference = storedPreference?.takeUnless { it.isBlank() } ?: return null
    val match = resolveLongFormTrackByLanguage(
        tracks = tracks.map {
            TvLongFormVlcTrack(
                id = it.vlcTrackId,
                name = it.label,
                language = it.languageCode,
                type = it.preferenceType,
                selected = it.selected,
            )
        },
        preference = TvLongFormLanguagePreference(
            language = preference.language,
            type = preference.type,
        ),
    ) ?: return null
    return tracks.firstOrNull { it.vlcTrackId == match.id }?.id
}

fun buildAudioTrackPreference(track: LongFormAudioTrack?): TvTrackPreference? {
    val safeTrack = track ?: return null
    val preference = TvTrackPreference(
        language = normalizeLongFormLanguageCode(safeTrack.languageCode),
        type = safeTrack.preferenceType.trim().lowercase(Locale.ROOT),
    )
    return preference.takeUnless { it.isBlank() }
}

internal fun inferLongFormTrackLanguageCode(name: String): String {
    val normalized = name.lowercase(Locale.ROOT)
    return when {
        Regex("""\b(jpn|jp|ja)\b""").containsMatchIn(normalized) ||
            "japanese" in normalized || "日本語" in normalized || "日语" in normalized || "日文" in normalized -> "ja"
        Regex("""\b(eng|en)\b""").containsMatchIn(normalized) ||
            "english" in normalized || "英语" in normalized || "英文" in normalized -> "en"
        Regex("""\b(zho|chi|cmn|zh)\b""").containsMatchIn(normalized) ||
            "chinese" in normalized || "mandarin" in normalized || "中文" in normalized ||
            "国语" in normalized || "普通话" in normalized || "简中" in normalized || "繁中" in normalized -> "zh"
        Regex("""\b(yue|cant|cantonese)\b""").containsMatchIn(normalized) ||
            "粤语" in normalized || "粤文" in normalized -> "yue"
        Regex("""\b(kor|ko)\b""").containsMatchIn(normalized) ||
            "korean" in normalized || "韩语" in normalized || "韩文" in normalized -> "ko"
        else -> normalizeLongFormLanguageCode(name)
            .takeIf { it.matches(Regex("""[a-z]{2,8}(-[a-z0-9]{2,8})?""")) }
            ?: ""
    }
}

internal fun inferLongFormTrackPreferenceType(name: String): String {
    val normalized = name.lowercase(Locale.ROOT)
    return when {
        "forced" in normalized || "强制" in normalized -> "forced"
        "commentary" in normalized || "comment" in normalized || "解说" in normalized || "评论" in normalized -> "commentary"
        "default" in normalized || "默认" in normalized -> "default"
        else -> ""
    }
}

private fun buildAudioTrackId(trackId: Int, name: String, index: Int): String {
    val rawParts = listOf(
        "audio",
        trackId.toString(),
        name,
    )
    return rawParts
        .filter { it.isNotBlank() }
        .joinToString("-")
        .replace(Regex("[^A-Za-z0-9_-]+"), "-")
        .trim('-')
        .ifBlank { "audio-$index" }
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

private fun audioCodecLabel(sampleMimeType: String?): String {
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
