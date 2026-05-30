package com.chee.videos.core.ui

import android.view.KeyEvent as AndroidKeyEvent
import androidx.compose.runtime.Immutable

internal sealed interface TvRemoteKeyAction {
    data class Seek(val deltaMs: Long) : TvRemoteKeyAction
    data object EnterFocus : TvRemoteKeyAction
    data object ExitFocus : TvRemoteKeyAction
    data object EnterEpisodeRail : TvRemoteKeyAction
    data object ExitEpisodeRail : TvRemoteKeyAction
    data object TogglePlayPause : TvRemoteKeyAction
    data object DismissUi : TvRemoteKeyAction
    data object PassThrough : TvRemoteKeyAction
}

internal enum class TvPlayerFocusLayer {
    Root,
    Controls,
    EpisodeRail,
}

@Immutable
internal data class TvLongFormTitleOverlayData(
    val primary: String,
    val secondary: String?,
)

internal fun resolveTvRemoteKeyAction(
    visible: Boolean,
    focusLayer: TvPlayerFocusLayer,
    episodeRailEnabled: Boolean,
    keyCode: Int,
    repeatCount: Int,
    seekStepSec: Int,
): TvRemoteKeyAction? {
    val stepMs = normalizeTvSeekStepSeconds(seekStepSec) * 1_000L
    val seekDelta = if (repeatCount > 0) stepMs * 3 else stepMs
    return when (keyCode) {
        AndroidKeyEvent.KEYCODE_DPAD_LEFT,
        AndroidKeyEvent.KEYCODE_MEDIA_REWIND,
        -> when {
            focusLayer != TvPlayerFocusLayer.Root -> TvRemoteKeyAction.PassThrough
            else -> TvRemoteKeyAction.Seek(-seekDelta)
        }

        AndroidKeyEvent.KEYCODE_DPAD_RIGHT,
        AndroidKeyEvent.KEYCODE_MEDIA_FAST_FORWARD,
        -> when {
            focusLayer != TvPlayerFocusLayer.Root -> TvRemoteKeyAction.PassThrough
            else -> TvRemoteKeyAction.Seek(seekDelta)
        }

        AndroidKeyEvent.KEYCODE_DPAD_DOWN -> when {
            focusLayer == TvPlayerFocusLayer.Root -> TvRemoteKeyAction.EnterFocus
            focusLayer == TvPlayerFocusLayer.Controls && episodeRailEnabled -> TvRemoteKeyAction.EnterEpisodeRail
            else -> null
        }

        AndroidKeyEvent.KEYCODE_DPAD_UP -> when {
            focusLayer == TvPlayerFocusLayer.EpisodeRail -> TvRemoteKeyAction.ExitEpisodeRail
            focusLayer == TvPlayerFocusLayer.Controls -> TvRemoteKeyAction.ExitFocus
            else -> null
        }

        AndroidKeyEvent.KEYCODE_DPAD_CENTER,
        AndroidKeyEvent.KEYCODE_ENTER,
        AndroidKeyEvent.KEYCODE_NUMPAD_ENTER,
        AndroidKeyEvent.KEYCODE_MEDIA_PLAY_PAUSE,
        AndroidKeyEvent.KEYCODE_MEDIA_PLAY,
        AndroidKeyEvent.KEYCODE_MEDIA_PAUSE,
        -> when {
            focusLayer != TvPlayerFocusLayer.Root -> TvRemoteKeyAction.PassThrough
            else -> TvRemoteKeyAction.TogglePlayPause
        }

        AndroidKeyEvent.KEYCODE_BACK,
        AndroidKeyEvent.KEYCODE_ESCAPE,
        -> when {
            visible -> TvRemoteKeyAction.DismissUi
            else -> null
        }

        else -> null
    }
}

internal fun shouldResetAutoHideTimer(action: TvRemoteKeyAction): Boolean {
    return when (action) {
        is TvRemoteKeyAction.Seek,
        TvRemoteKeyAction.EnterFocus,
        TvRemoteKeyAction.TogglePlayPause,
        -> true

        TvRemoteKeyAction.ExitFocus,
        TvRemoteKeyAction.EnterEpisodeRail,
        TvRemoteKeyAction.ExitEpisodeRail,
        TvRemoteKeyAction.DismissUi,
        TvRemoteKeyAction.PassThrough,
        -> false
    }
}

internal fun buildTvLongFormTitleOverlayData(
    primaryFallback: String,
    seriesTitle: String?,
    seasonNumber: Int?,
    episodeNumber: Int?,
    episodeTitle: String?,
): TvLongFormTitleOverlayData {
    val normalizedFallback = primaryFallback.trim()
    val primary = seriesTitle?.trim()
        ?.ifBlank { null }
        ?: normalizedFallback
    val secondary = if (seasonNumber != null && episodeNumber != null) {
        buildString {
            append("第 ")
            append(seasonNumber)
            append(" 季 · 第 ")
            append(episodeNumber)
            append(" 集")
            val normalizedEpisodeTitle = episodeTitle?.trim().orEmpty()
            if (normalizedEpisodeTitle.isNotBlank()) {
                append(' ')
                append(normalizedEpisodeTitle)
            }
        }
    } else {
        null
    }
    return TvLongFormTitleOverlayData(
        primary = primary,
        secondary = secondary,
    )
}
