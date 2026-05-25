package com.chee.videos.core.model

data class TvTrackPreference(
    val language: String = "",
    val type: String = "",
) {
    fun isBlank(): Boolean = language.isBlank() && type.isBlank()
}
