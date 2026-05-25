package com.chee.videos.core.player

import java.util.Locale

internal data class TvLongFormLanguagePreference(
    val language: String?,
    val type: String?,
)

internal data class TvLongFormVlcTrack(
    val id: Int,
    val name: String,
    val language: String? = null,
    val type: String? = null,
    val selected: Boolean = false,
)

internal fun normalizeLongFormLanguageCode(raw: String?): String {
    val normalized = raw
        ?.trim()
        ?.replace('_', '-')
        ?.lowercase(Locale.ROOT)
        .orEmpty()
    return when (normalized) {
        "jpn", "jp" -> "ja"
        "eng" -> "en"
        "zho", "chi", "cmn" -> "zh"
        else -> normalized
    }
}

internal fun resolveLongFormTrackByLanguage(
    tracks: List<TvLongFormVlcTrack>,
    preference: TvLongFormLanguagePreference,
): TvLongFormVlcTrack? {
    val language = normalizeLongFormLanguageCode(preference.language)
    val type = preference.type?.trim()?.lowercase(Locale.ROOT).orEmpty()

    if (language.isBlank()) {
        // type-only fallback：language 缺失但 type 不空时按 type 在全表里找第一条同 type 的 track。
        // 修复 [[Type-only preference fallback]]：之前直接 return null 会让"isDefault 但无 language"
        // 的偏好永远 resolve 不回任何 track。
        if (type.isBlank()) {
            return null
        }
        tracks.firstOrNull { it.type?.trim()?.lowercase(Locale.ROOT) == type }?.let { return it }
        return tracks.firstOrNull { it.name.lowercase(Locale.ROOT).contains(type) }
    }

    val languageMatches = tracks.filter { track ->
        val trackLanguage = normalizeLongFormLanguageCode(track.language)
        trackLanguage == language ||
            trackLanguage.startsWith("$language-") ||
            language.startsWith("$trackLanguage-") ||
            track.name.lowercase(Locale.ROOT).contains(language)
    }
    if (languageMatches.isEmpty()) {
        return null
    }
    if (type.isNotBlank()) {
        languageMatches.firstOrNull { it.type?.trim()?.lowercase(Locale.ROOT) == type }?.let { return it }
        languageMatches.firstOrNull { it.name.lowercase(Locale.ROOT).contains(type) }?.let { return it }
    }
    return languageMatches.first()
}
