package com.chee.videos.core.model

enum class VideoFitMode(val rawValue: String) {
    FILL("fill"),
    FIT("fit");

    companion object {
        fun fromRaw(raw: String?): VideoFitMode {
            return entries.firstOrNull { it.rawValue == raw } ?: FILL
        }
    }
}
