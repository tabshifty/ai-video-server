package com.chee.videos.feature.tv

internal data class TvPlaybackCompatibilityDecision(
    val allowed: Boolean,
    val blockMessage: String? = null,
)

internal data class TvPlaybackCandidateDecision(
    val allowed: Boolean,
    val blockMessage: String? = null,
)

internal data class TvPlaybackRoute(
    val kind: TvPlaybackRouteKind,
    val blockMessage: String? = null,
)

internal enum class TvPlaybackRouteKind {
    VLC,
    MEDIA3_DOLBY_VISION,
    BLOCKED,
}

private const val PlaybackCompatibilityKey = "playback_compat"
private const val TrustedToneMappedSdrOutput = "dv_sdr_bt709"
private const val PlaybackCompatibilityIncompleteMessage = "播放兼容信息不完整，暂不能确认安全播放"
private const val DolbyVisionTranscodeOutputMessage = "该视频来源为杜比视界，当前压缩结果可能无法安全播放"
private const val DolbyVisionUnsupportedMessage = "该视频可能为杜比视界，当前设备或播放链路不支持安全播放"
private const val DolbyVisionDisplayUnsupportedMessage = "当前电视或显示链路未声明支持杜比视界，无法安全播放该视频"
private const val DolbyVisionDisplayUnknownMessage = "无法确认当前电视或显示链路是否支持杜比视界，暂不能安全播放"
private const val DolbyVisionDedicatedRouteUnavailableMessage = "杜比视界专用播放链路暂不可用，无法安全播放该视频"

internal fun resolveTvPlaybackCompatibilityDecision(
    metadata: Map<String, Any?>?,
): TvPlaybackCompatibilityDecision {
    return when (val compatibility = parsePlaybackCompatibility(metadata)) {
        PlaybackCompatibilityPayload.Historical -> TvPlaybackCompatibilityDecision(allowed = true)
        PlaybackCompatibilityPayload.Incomplete -> blockPlaybackCompatibilityIncomplete()
        is PlaybackCompatibilityPayload.Ok -> resolveCompatibilityDecision(compatibility)
    }
}

internal fun resolveTvPlaybackCandidateDecision(
    metadata: Map<String, Any?>?,
): TvPlaybackCandidateDecision {
    return when (val compatibility = parsePlaybackCompatibility(metadata)) {
        PlaybackCompatibilityPayload.Historical -> TvPlaybackCandidateDecision(allowed = true)
        PlaybackCompatibilityPayload.Incomplete -> TvPlaybackCandidateDecision(
            allowed = false,
            blockMessage = PlaybackCompatibilityIncompleteMessage,
        )
        is PlaybackCompatibilityPayload.Ok -> {
            if (compatibility.hasUnsafeDolbyVisionTranscodeOutput()) {
                TvPlaybackCandidateDecision(
                    allowed = false,
                    blockMessage = DolbyVisionTranscodeOutputMessage,
                )
            } else {
                TvPlaybackCandidateDecision(allowed = true)
            }
        }
    }
}

internal fun resolveTvPlaybackRoute(
    metadata: Map<String, Any?>?,
    displayCapability: DolbyVisionDisplayCapability,
    playbackUrl: String?,
    media3Available: Boolean,
): TvPlaybackRoute {
    return when (val compatibility = parsePlaybackCompatibility(metadata)) {
        PlaybackCompatibilityPayload.Historical -> TvPlaybackRoute(kind = TvPlaybackRouteKind.VLC)
        PlaybackCompatibilityPayload.Incomplete -> blockPlaybackRoute(PlaybackCompatibilityIncompleteMessage)
        is PlaybackCompatibilityPayload.Ok -> resolvePlaybackRoute(
            compatibility = compatibility,
            displayCapability = displayCapability,
            playbackUrl = playbackUrl,
            media3Available = media3Available,
        )
    }
}

private fun resolveCompatibilityDecision(
    compatibility: PlaybackCompatibilityPayload.Ok,
): TvPlaybackCompatibilityDecision {
    if (compatibility.outputDolbyVision) {
        return TvPlaybackCompatibilityDecision(
            allowed = false,
            blockMessage = DolbyVisionUnsupportedMessage,
        )
    }
    if (compatibility.hasUnsafeDolbyVisionTranscodeOutput()) {
        return TvPlaybackCompatibilityDecision(
            allowed = false,
            blockMessage = DolbyVisionTranscodeOutputMessage,
        )
    }
    return TvPlaybackCompatibilityDecision(allowed = true)
}

private fun resolvePlaybackRoute(
    compatibility: PlaybackCompatibilityPayload.Ok,
    displayCapability: DolbyVisionDisplayCapability,
    playbackUrl: String?,
    media3Available: Boolean,
): TvPlaybackRoute {
    if (!compatibility.outputDolbyVision) {
        return if (compatibility.hasUnsafeDolbyVisionTranscodeOutput()) {
            blockPlaybackRoute(DolbyVisionTranscodeOutputMessage)
        } else {
            TvPlaybackRoute(kind = TvPlaybackRouteKind.VLC)
        }
    }

    when (displayCapability.status) {
        DolbyVisionDisplayCapabilityStatus.SUPPORTED -> Unit
        DolbyVisionDisplayCapabilityStatus.UNSUPPORTED -> {
            return blockPlaybackRoute(DolbyVisionDisplayUnsupportedMessage)
        }
        DolbyVisionDisplayCapabilityStatus.UNKNOWN -> {
            return blockPlaybackRoute(DolbyVisionDisplayUnknownMessage)
        }
    }

    if (!media3Available || playbackUrl.isNullOrBlank()) {
        return blockPlaybackRoute(DolbyVisionDedicatedRouteUnavailableMessage)
    }
    return TvPlaybackRoute(kind = TvPlaybackRouteKind.MEDIA3_DOLBY_VISION)
}

internal fun isTvEpisodePlayableForPlayback(episode: TvEpisodeUiModel): Boolean =
    episode.playable &&
        episode.videoId.isNotBlank() &&
        isPlaybackCompatibilityCandidate(episode.metadata)

private fun blockPlaybackCompatibilityIncomplete(): TvPlaybackCompatibilityDecision =
    TvPlaybackCompatibilityDecision(
        allowed = false,
        blockMessage = PlaybackCompatibilityIncompleteMessage,
    )

private fun blockPlaybackRoute(message: String): TvPlaybackRoute =
    TvPlaybackRoute(
        kind = TvPlaybackRouteKind.BLOCKED,
        blockMessage = message,
    )

private sealed class PlaybackCompatibilityPayload {
    object Historical : PlaybackCompatibilityPayload()
    object Incomplete : PlaybackCompatibilityPayload()

    data class Ok(
        val sourceDolbyVision: Boolean,
        val outputDolbyVision: Boolean,
        val trustedToneMappedSdr: Boolean,
    ) : PlaybackCompatibilityPayload()
}

private fun PlaybackCompatibilityPayload.Ok.hasUnsafeDolbyVisionTranscodeOutput(): Boolean =
    sourceDolbyVision && !outputDolbyVision && !trustedToneMappedSdr

private fun parsePlaybackCompatibility(metadata: Map<String, Any?>?): PlaybackCompatibilityPayload {
    val compatibility = metadata?.get(PlaybackCompatibilityKey) ?: return PlaybackCompatibilityPayload.Historical
    val block = compatibility as? Map<*, *> ?: return PlaybackCompatibilityPayload.Incomplete
    val version = numberValue(block["version"])?.toInt() ?: return PlaybackCompatibilityPayload.Incomplete
    if (version < 1) {
        return PlaybackCompatibilityPayload.Historical
    }
    val status = stringValue(block["status"]).orEmpty()
    if (version != 1 || status != "ok") {
        return PlaybackCompatibilityPayload.Incomplete
    }
    val source = block["source"] as? Map<*, *> ?: return PlaybackCompatibilityPayload.Incomplete
    val output = block["output"] as? Map<*, *> ?: return PlaybackCompatibilityPayload.Incomplete
    val sourceDolbyVision = optionalBooleanValue(source["dolby_vision"])
        ?: return PlaybackCompatibilityPayload.Incomplete
    val outputDolbyVision = optionalBooleanValue(output["dolby_vision"])
        ?: return PlaybackCompatibilityPayload.Incomplete
    val trustedToneMappedSdr = stringValue(block["trusted_compat_output"]) == TrustedToneMappedSdrOutput &&
        optionalBooleanValue(block["tone_mapped_sdr"]) == true
    return PlaybackCompatibilityPayload.Ok(
        sourceDolbyVision = sourceDolbyVision,
        outputDolbyVision = outputDolbyVision,
        trustedToneMappedSdr = trustedToneMappedSdr,
    )
}

private fun isPlaybackCompatibilityCandidate(metadata: Map<String, Any?>?): Boolean =
    resolveTvPlaybackCandidateDecision(metadata).allowed

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
