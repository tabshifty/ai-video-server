@file:androidx.annotation.OptIn(markerClass = [androidx.media3.common.util.UnstableApi::class])

package com.chee.videos.feature.tv

import android.util.Log
import androidx.media3.common.Format
import androidx.media3.common.Timeline
import androidx.media3.exoplayer.DecoderReuseEvaluation
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.analytics.AnalyticsListener
import com.chee.videos.core.util.UrlBuilder
import java.net.URI
import java.util.Locale

private const val TvShortPlaybackDiagnosticLogTag = "TvShortPlayback"

internal enum class TvShortPlaybackEntry(val wireValue: String) {
    Local("local"),
    Remote("remote"),
}

internal data class TvShortPlaybackAttempt(
    val entry: TvShortPlaybackEntry,
    val attemptId: Long,
    val videoId: String,
    val sourcePath: String,
    val profile: String,
) {
    fun diagnosticPrefix(event: String): String =
        "event=$event entry=${entry.wireValue} attempt=$attemptId videoId=$videoId " +
            "source=$sourcePath profile=$profile"
}

internal fun buildTvShortPlaybackSourceUrl(baseUrl: String, videoId: String): String =
    UrlBuilder.source(baseUrl, videoId)

internal fun createTvShortPlaybackAttempt(
    entry: TvShortPlaybackEntry,
    attemptId: Long,
    videoId: String,
    sourceUrl: String,
): TvShortPlaybackAttempt {
    val source = sanitizeTvShortPlaybackSource(sourceUrl)
    return TvShortPlaybackAttempt(
        entry = entry,
        attemptId = attemptId,
        videoId = videoId,
        sourcePath = source.path,
        profile = source.profile,
    )
}

private data class SanitizedTvShortPlaybackSource(
    val path: String,
    val profile: String,
)

private fun sanitizeTvShortPlaybackSource(sourceUrl: String): SanitizedTvShortPlaybackSource {
    val uri = runCatching { URI(sourceUrl) }.getOrNull()
    val path = uri?.rawPath
        ?.takeIf { it.startsWith("/") && it.isNotBlank() }
        ?: "unavailable"
    val requestedProfile = uri?.rawQuery
        .orEmpty()
        .split('&')
        .asSequence()
        .mapNotNull { pair ->
            val separator = pair.indexOf('=')
            if (separator <= 0) return@mapNotNull null
            val key = pair.substring(0, separator)
            val value = pair.substring(separator + 1)
            value.takeIf { key.equals("profile", ignoreCase = true) && it.isNotBlank() }
        }
        .firstOrNull()
        ?.lowercase(Locale.ROOT)
    val profile = when (requestedProfile) {
        null -> "primary"
        "primary", "compat", "dv_source" -> requestedProfile
        else -> "unknown"
    }
    return SanitizedTvShortPlaybackSource(path = path, profile = profile)
}

internal class TvShortPlaybackDiagnostics {
    private val analyticsListener = object : AnalyticsListener {
        override fun onVideoDecoderInitialized(
            eventTime: AnalyticsListener.EventTime,
            decoderName: String,
            initializedTimestampMs: Long,
            initializationDurationMs: Long,
        ) {
            val attempt = eventTime.playbackAttempt() ?: return
            Log.i(
                TvShortPlaybackDiagnosticLogTag,
                "${attempt.diagnosticPrefix("decoder_initialized")} " +
                    "decoder=$decoderName initializationMs=$initializationDurationMs",
            )
        }

        override fun onVideoInputFormatChanged(
            eventTime: AnalyticsListener.EventTime,
            format: Format,
            decoderReuseEvaluation: DecoderReuseEvaluation?,
        ) {
            val attempt = eventTime.playbackAttempt() ?: return
            Log.i(
                TvShortPlaybackDiagnosticLogTag,
                "${attempt.diagnosticPrefix("video_input_format")} " +
                    "mime=${format.sampleMimeType.orEmpty()} codecs=${format.codecs.orEmpty()} " +
                    "size=${format.width}x${format.height} frameRate=${format.frameRate} " +
                    "rotation=${format.rotationDegrees} color=${format.colorInfo}",
            )
        }

        override fun onRenderedFirstFrame(
            eventTime: AnalyticsListener.EventTime,
            output: Any,
            renderTimeMs: Long,
        ) {
            val attempt = eventTime.playbackAttempt() ?: return
            Log.i(
                TvShortPlaybackDiagnosticLogTag,
                "${attempt.diagnosticPrefix("first_frame")} renderTimeMs=$renderTimeMs",
            )
        }

        override fun onPlayerError(
            eventTime: AnalyticsListener.EventTime,
            error: androidx.media3.common.PlaybackException,
        ) {
            val attempt = eventTime.playbackAttempt() ?: return
            Log.e(
                TvShortPlaybackDiagnosticLogTag,
                "${attempt.diagnosticPrefix("player_error")} " +
                    "error=${error.errorCodeName} cause=${error.cause?.javaClass?.simpleName.orEmpty()}",
            )
        }
    }

    fun attach(player: ExoPlayer) {
        player.addAnalyticsListener(analyticsListener)
    }

    fun detach(player: ExoPlayer) {
        player.removeAnalyticsListener(analyticsListener)
    }

    fun sourcePrepared(attempt: TvShortPlaybackAttempt) {
        Log.i(TvShortPlaybackDiagnosticLogTag, attempt.diagnosticPrefix("source_prepared"))
    }
}

private fun AnalyticsListener.EventTime.playbackAttempt(): TvShortPlaybackAttempt? {
    if (!timeline.isEmpty && windowIndex in 0 until timeline.windowCount) {
        val attempt = timeline
            .getWindow(windowIndex, Timeline.Window())
            .mediaItem
            .localConfiguration
            ?.tag as? TvShortPlaybackAttempt
        if (attempt != null) return attempt
    }
    if (!currentTimeline.isEmpty && currentWindowIndex in 0 until currentTimeline.windowCount) {
        return currentTimeline
            .getWindow(currentWindowIndex, Timeline.Window())
            .mediaItem
            .localConfiguration
            ?.tag as? TvShortPlaybackAttempt
    }
    return null
}
