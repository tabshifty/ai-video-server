package com.chee.videos.core.ui

import android.app.Activity
import android.content.Context
import android.content.ContextWrapper
import android.content.pm.ActivityInfo
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Fullscreen
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import androidx.media3.common.Player
import androidx.media3.exoplayer.ExoPlayer
import com.chee.videos.core.model.ShortPlaybackMode
import com.chee.videos.core.model.SubtitleTrackDto
import com.chee.videos.core.model.toPlayerRepeatMode

@Composable
fun ShortOverlayFullscreenHost(
    isFullscreen: Boolean,
    onFullscreenChange: (Boolean) -> Unit,
    player: ExoPlayer,
    title: String,
    subtitleTracks: List<SubtitleTrackDto>,
    fallbackPlaybackMode: ShortPlaybackMode,
    modifier: Modifier = Modifier,
) {
    val activity = LocalContext.current.findShortOverlayActivity()

    DisposableEffect(activity, player, isFullscreen, fallbackPlaybackMode) {
        if (activity != null && isFullscreen) {
            val window = activity.window
            val controller = WindowCompat.getInsetsController(window, window.decorView)
            activity.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_LANDSCAPE
            controller.systemBarsBehavior = WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
            controller.hide(WindowInsetsCompat.Type.systemBars())
            player.repeatMode = shortOverlayRepeatModeWhileFullscreen(player.repeatMode)
        }

        onDispose {
            if (activity != null) {
                val window = activity.window
                WindowCompat.getInsetsController(window, window.decorView)
                    .show(WindowInsetsCompat.Type.systemBars())
                activity.requestedOrientation = ActivityInfo.SCREEN_ORIENTATION_UNSPECIFIED
            }
            player.repeatMode = shortOverlayRepeatModeAfterFullscreen(fallbackPlaybackMode.toPlayerRepeatMode())
        }
    }

    if (!isFullscreen) {
        return
    }

    Dialog(
        onDismissRequest = { onFullscreenChange(false) },
        properties = DialogProperties(
            usePlatformDefaultWidth = false,
            decorFitsSystemWindows = false,
            dismissOnBackPress = false,
            dismissOnClickOutside = false,
        ),
    ) {
        BackHandler { onFullscreenChange(false) }
        Box(
            modifier = modifier
                .fillMaxSize()
                .background(Color.Black),
        ) {
            LongFormVideoPlayer(
                title = title,
                player = player,
                isFullscreen = true,
                onBack = { onFullscreenChange(false) },
                onTogglePlayPause = { player.playWhenReady = !player.playWhenReady },
                onToggleFullscreen = { onFullscreenChange(false) },
                subtitleTracks = subtitleTracks,
                selectedSubtitleTrackId = null,
                onSelectSubtitleTrack = {},
                modifier = Modifier.fillMaxSize(),
            )
        }
    }
}

@Composable
fun ShortOverlayFullscreenButton(
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    ShortVideoOverlayActionButton(
        icon = Icons.Filled.Fullscreen,
        active = false,
        enabled = true,
        onClick = onClick,
        contentDescription = "全屏播放",
        modifier = modifier,
    )
}

internal fun shortOverlayRepeatModeWhileFullscreen(@Suppress("UNUSED_PARAMETER") fallback: Int): Int =
    Player.REPEAT_MODE_ONE

internal fun shortOverlayRepeatModeAfterFullscreen(fallback: Int): Int = fallback

private tailrec fun Context.findShortOverlayActivity(): Activity? {
    return when (this) {
        is Activity -> this
        is ContextWrapper -> baseContext.findShortOverlayActivity()
        else -> null
    }
}
