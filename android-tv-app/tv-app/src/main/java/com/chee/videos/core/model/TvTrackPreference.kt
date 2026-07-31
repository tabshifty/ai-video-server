package com.chee.videos.core.model

enum class TvSubtitlePreferenceMode(val storageValue: String) {
    AUTO("auto"),
    OFF("off"),
    SPECIFIC("specific"),
    ;

    companion object {
        fun fromPreference(preference: TvTrackPreference?): TvSubtitlePreferenceMode {
            return when (preference?.subtitleMode?.trim()?.lowercase()) {
                AUTO.storageValue -> AUTO
                OFF.storageValue -> OFF
                SPECIFIC.storageValue -> SPECIFIC
                else -> if (preference != null && (preference.language.isNotBlank() || preference.type.isNotBlank())) {
                    SPECIFIC
                } else {
                    AUTO
                }
            }
        }
    }
}

data class TvTrackPreference(
    val language: String = "",
    val type: String = "",
    val subtitleMode: String? = null,
) {
    fun isBlank(): Boolean = language.isBlank() && type.isBlank() && subtitleMode.isNullOrBlank()
}
