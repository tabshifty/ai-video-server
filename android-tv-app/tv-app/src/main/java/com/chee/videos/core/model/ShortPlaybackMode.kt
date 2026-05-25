package com.chee.videos.core.model

enum class ShortPlaybackMode(val rawValue: String) {
    LOOP_ONE("loop_one"),
    AUTO_NEXT("auto_next");

    companion object {
        fun fromRaw(raw: String?): ShortPlaybackMode {
            return entries.firstOrNull { it.rawValue == raw } ?: LOOP_ONE
        }
    }
}

fun ShortPlaybackMode.toPlayerRepeatMode(): Int {
    return when (this) {
        ShortPlaybackMode.LOOP_ONE -> 1
        ShortPlaybackMode.AUTO_NEXT -> 0
    }
}
