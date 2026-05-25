package com.chee.videos.core.player

fun friendlyLongFormPlaybackErrorMessage(error: Throwable?): String {
    val raw = buildString {
        append(error?.message.orEmpty())
        val cause = error?.cause
        if (cause != null) {
            append(" ")
            append(cause.message.orEmpty())
        }
    }.lowercase()
    if (raw.contains("video/hevc") || raw.contains("hevc") || raw.contains("mediacodec")) {
        return "当前设备不支持该视频编码，已建议使用兼容播放档；如仍失败，请在管理端确认兼容转码已生成。"
    }
    return error?.message?.takeIf { it.isNotBlank() } ?: "视频播放失败，请稍后重试。"
}
