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
    val playbackProfile: String? = null,
    val outputSurface: TvPlaybackOutputSurface = TvPlaybackOutputSurface.TEXTURE_VIEW,
    val blockMessage: String? = null,
)

internal enum class TvPlaybackRouteKind {
    EXOPLAYER,
    BLOCKED,
}

internal enum class TvPlaybackOutputSurface {
    TEXTURE_VIEW,
    SURFACE_VIEW,
}

private const val PlaybackCompatibilityKey = "playback_compat"
private const val PlaybackCompatibilityIncompleteMessage = "播放兼容信息不完整，暂不能确认安全播放"
private const val DolbyVisionTranscodeOutputMessage = "该视频来源为杜比视界，当前压缩结果可能无法安全播放"
private const val DolbyVisionUnsupportedMessage = "该视频可能为杜比视界，当前设备或播放链路不支持安全播放"
private const val DolbyVisionDisplayUnsupportedMessage = "当前电视或显示链路未声明支持杜比视界，无法安全播放该视频"
private const val DolbyVisionDisplayUnknownMessage = "无法确认当前电视或显示链路是否支持杜比视界，暂不能安全播放"
private const val LongFormExoPlayerRouteUnavailableMessage = "长视频 ExoPlayer 播放链路暂不可用，无法播放该视频"

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
            if (compatibility.sourceDolbyVision && !compatibility.outputDolbyVision) {
                if (!compatibility.sourcePlaybackPath.isNullOrBlank()) {
                    TvPlaybackCandidateDecision(allowed = true)
                } else {
                    TvPlaybackCandidateDecision(
                        allowed = false,
                        blockMessage = DolbyVisionTranscodeOutputMessage,
                    )
                }
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
        PlaybackCompatibilityPayload.Historical -> resolveUnifiedExoPlayerRoute(
            playbackUrl = playbackUrl,
            media3Available = media3Available,
            playbackProfile = null,
        )
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
    if (compatibility.sourceDolbyVision) {
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
        return if (compatibility.sourceDolbyVision) {
            if (!compatibility.sourcePlaybackPath.isNullOrBlank()) {
                resolveDolbyVisionGatedExoPlayerRoute(
                    displayCapability = displayCapability,
                    playbackUrl = playbackUrl,
                    media3Available = media3Available,
                    playbackProfile = "dv_source",
                    outputSurface = TvPlaybackOutputSurface.SURFACE_VIEW,
                )
            } else {
                blockPlaybackRoute(DolbyVisionTranscodeOutputMessage)
            }
        } else {
            resolveUnifiedExoPlayerRoute(
                playbackUrl = playbackUrl,
                media3Available = media3Available,
                playbackProfile = null,
            )
        }
    }

    return resolveDolbyVisionGatedExoPlayerRoute(
        displayCapability = displayCapability,
        playbackUrl = playbackUrl,
        media3Available = media3Available,
        playbackProfile = null,
        outputSurface = TvPlaybackOutputSurface.SURFACE_VIEW,
    )
}

private fun resolveDolbyVisionGatedExoPlayerRoute(
    displayCapability: DolbyVisionDisplayCapability,
    playbackUrl: String?,
    media3Available: Boolean,
    playbackProfile: String?,
    outputSurface: TvPlaybackOutputSurface,
): TvPlaybackRoute {
    when (displayCapability.status) {
        DolbyVisionDisplayCapabilityStatus.SUPPORTED -> Unit
        DolbyVisionDisplayCapabilityStatus.UNSUPPORTED -> {
            return blockPlaybackRoute(DolbyVisionDisplayUnsupportedMessage)
        }
        DolbyVisionDisplayCapabilityStatus.UNKNOWN -> {
            return blockPlaybackRoute(DolbyVisionDisplayUnknownMessage)
        }
    }

    return resolveUnifiedExoPlayerRoute(
        playbackUrl = playbackUrl,
        media3Available = media3Available,
        playbackProfile = playbackProfile,
        outputSurface = outputSurface,
    )
}

private fun resolveUnifiedExoPlayerRoute(
    playbackUrl: String?,
    media3Available: Boolean,
    playbackProfile: String?,
    outputSurface: TvPlaybackOutputSurface = TvPlaybackOutputSurface.TEXTURE_VIEW,
): TvPlaybackRoute {
    if (!media3Available || playbackUrl.isNullOrBlank()) {
        return blockPlaybackRoute(LongFormExoPlayerRouteUnavailableMessage)
    }
    return TvPlaybackRoute(
        kind = TvPlaybackRouteKind.EXOPLAYER,
        playbackProfile = playbackProfile,
        outputSurface = outputSurface,
    )
}

internal fun isTvEpisodePlayableForPlayback(episode: TvEpisodeUiModel): Boolean =
    episode.playable &&
        episode.videoId.isNotBlank() &&
        isPlaybackCompatibilityCandidate(episode.metadata)

internal fun resolveTvPlaybackSourceProfile(
    metadata: Map<String, Any?>?,
): String? {
    return when (val compatibility = parsePlaybackCompatibility(metadata)) {
        is PlaybackCompatibilityPayload.Ok -> {
            if (compatibility.sourceDolbyVision && !compatibility.outputDolbyVision && !compatibility.sourcePlaybackPath.isNullOrBlank()) {
                "dv_source"
            } else {
                null
            }
        }
        else -> null
    }
}

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
        val sourcePlaybackPath: String?,
    ) : PlaybackCompatibilityPayload()
}

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
    val sourcePlaybackPath = stringValue(block["source_playback_path"])
    return PlaybackCompatibilityPayload.Ok(
        sourceDolbyVision = sourceDolbyVision,
        outputDolbyVision = outputDolbyVision,
        sourcePlaybackPath = sourcePlaybackPath,
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
