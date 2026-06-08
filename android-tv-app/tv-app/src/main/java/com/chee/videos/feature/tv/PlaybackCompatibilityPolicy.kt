package com.chee.videos.feature.tv

internal data class TvPlaybackCompatibilityDecision(
    val allowed: Boolean,
    val blockMessage: String? = null,
)

private const val PlaybackCompatibilityKey = "playback_compat"
private const val ProbeUnconfirmedMessage = "该视频播放兼容性未确认，当前 TV 端暂不自动播放"
private const val DolbyVisionTranscodeOutputMessage = "该视频来源为杜比视界，当前压缩结果可能无法安全播放"
private const val DolbyVisionUnsupportedMessage = "该视频可能为杜比视界，当前设备或播放链路不支持安全播放"

internal fun resolveTvPlaybackCompatibilityDecision(
    metadata: Map<String, Any?>?,
): TvPlaybackCompatibilityDecision {
    val compatibility = metadata?.get(PlaybackCompatibilityKey) ?: return TvPlaybackCompatibilityDecision(allowed = true)
    val block = compatibility as? Map<*, *> ?: return blockProbeUnconfirmed()
    val version = numberValue(block["version"])?.toInt() ?: return blockProbeUnconfirmed()
    val status = stringValue(block["status"]).orEmpty()
    if (version != 1 || status != "ok") {
        return blockProbeUnconfirmed()
    }
    val source = block["source"] as? Map<*, *> ?: return blockProbeUnconfirmed()
    val output = block["output"] as? Map<*, *> ?: return blockProbeUnconfirmed()
    val sourceDolbyVision = optionalBooleanValue(source["dolby_vision"]) ?: return blockProbeUnconfirmed()
    val outputDolbyVision = optionalBooleanValue(output["dolby_vision"]) ?: return blockProbeUnconfirmed()
    if (outputDolbyVision) {
        return TvPlaybackCompatibilityDecision(
            allowed = false,
            blockMessage = DolbyVisionUnsupportedMessage,
        )
    }
    if (sourceDolbyVision) {
        return TvPlaybackCompatibilityDecision(
            allowed = false,
            blockMessage = DolbyVisionTranscodeOutputMessage,
        )
    }
    return TvPlaybackCompatibilityDecision(allowed = true)
}

internal fun isTvEpisodePlayableForPlayback(episode: TvEpisodeUiModel): Boolean =
    episode.playable &&
        episode.videoId.isNotBlank() &&
        resolveTvPlaybackCompatibilityDecision(episode.metadata).allowed

private fun blockProbeUnconfirmed(): TvPlaybackCompatibilityDecision =
    TvPlaybackCompatibilityDecision(
        allowed = false,
        blockMessage = ProbeUnconfirmedMessage,
    )

private fun stringValue(value: Any?): String? =
    when (value) {
        is String -> value.trim()
        else -> null
    }?.takeIf { it.isNotBlank() }

private fun numberValue(value: Any?): Number? =
    when (value) {
        is Number -> value
        is String -> value.trim().toIntOrNull()
        else -> null
    }

private fun optionalBooleanValue(value: Any?): Boolean? =
    when (value) {
        is Boolean -> value
        is String -> when (value.trim().lowercase()) {
            "true" -> true
            "false" -> false
            else -> null
        }
        else -> null
    }
