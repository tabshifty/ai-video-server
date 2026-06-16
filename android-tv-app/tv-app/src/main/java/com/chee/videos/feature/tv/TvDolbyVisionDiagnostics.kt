package com.chee.videos.feature.tv

internal fun isTvDolbyVisionDiagnosticsAvailable(
    route: TvPlaybackRoute,
): Boolean {
    return when (route.kind) {
        TvPlaybackRouteKind.EXOPLAYER -> true
        TvPlaybackRouteKind.BLOCKED -> route.blockMessage?.contains("杜比视界") == true
    }
}

internal fun buildTvDolbyVisionDiagnosticMessage(
    route: TvPlaybackRoute,
    displayCapability: DolbyVisionDisplayCapability,
    playbackUrl: String?,
    media3Available: Boolean,
    failureMessage: String? = null,
): String {
    val lines = buildList {
        add("当前路径：${resolveTvDolbyVisionDiagnosticPathLabel(route.kind)}")
        add("当前原因：${resolveTvDolbyVisionDiagnosticReasonLabel(route, failureMessage)}")
        add(
            "兼容状态：${resolveTvDolbyVisionDiagnosticDisplayLabel(displayCapability.status)}，" +
                "${resolveTvDolbyVisionDiagnosticSourceLabel(playbackUrl)}，" +
                "${resolveTvDolbyVisionDiagnosticRouteLabel(route.kind, playbackUrl, media3Available)}",
        )
        failureMessage
            ?.trim()
            ?.takeIf { it.isNotBlank() }
            ?.let { add("错误摘要：$it") }
    }
    return lines.joinToString(separator = "\n")
}

private fun resolveTvDolbyVisionDiagnosticPathLabel(kind: TvPlaybackRouteKind): String =
    when (kind) {
        TvPlaybackRouteKind.EXOPLAYER -> "ExoPlayer 播放失败页"
        TvPlaybackRouteKind.BLOCKED -> "播放阻断页"
    }

private fun resolveTvDolbyVisionDiagnosticReasonLabel(
    route: TvPlaybackRoute,
    failureMessage: String?,
): String {
    if (!route.blockMessage.isNullOrBlank()) {
        return route.blockMessage.trim()
    }
    if (!failureMessage.isNullOrBlank()) {
        return "长视频 ExoPlayer 播放链路发生错误"
    }
    return when (route.kind) {
        TvPlaybackRouteKind.EXOPLAYER -> "长视频 ExoPlayer 播放链路"
        TvPlaybackRouteKind.BLOCKED -> "播放被阻断"
    }
}

private fun resolveTvDolbyVisionDiagnosticDisplayLabel(
    status: DolbyVisionDisplayCapabilityStatus,
): String =
    when (status) {
        DolbyVisionDisplayCapabilityStatus.SUPPORTED -> "显示支持杜比视界"
        DolbyVisionDisplayCapabilityStatus.UNSUPPORTED -> "显示不支持杜比视界"
        DolbyVisionDisplayCapabilityStatus.UNKNOWN -> "显示状态未知"
    }

private fun resolveTvDolbyVisionDiagnosticSourceLabel(playbackUrl: String?): String =
    if (playbackUrl.isNullOrBlank()) {
        "播放源不可用"
    } else {
        "播放源可用"
    }

private fun resolveTvDolbyVisionDiagnosticRouteLabel(
    kind: TvPlaybackRouteKind,
    playbackUrl: String?,
    media3Available: Boolean,
): String {
    val routeAvailable = when (kind) {
        TvPlaybackRouteKind.EXOPLAYER -> media3Available && !playbackUrl.isNullOrBlank()
        TvPlaybackRouteKind.BLOCKED -> false
    }
    return if (routeAvailable) {
        when (kind) {
            TvPlaybackRouteKind.EXOPLAYER -> "ExoPlayer 链路可用"
            TvPlaybackRouteKind.BLOCKED -> "ExoPlayer 链路不可用"
        }
    } else {
        when (kind) {
            TvPlaybackRouteKind.EXOPLAYER -> "ExoPlayer 链路不可用"
            TvPlaybackRouteKind.BLOCKED -> "ExoPlayer 链路不可用"
        }
    }
}
