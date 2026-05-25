package com.chee.videos.core.player

import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.viewinterop.AndroidView
import org.videolan.libvlc.MediaPlayer
import org.videolan.libvlc.util.VLCVideoLayout

@Composable
fun VlcLongFormSurface(
    mediaPlayer: MediaPlayer,
    modifier: Modifier = Modifier,
) {
    AndroidView(
        factory = { context ->
            VLCVideoLayout(context).also { layout ->
                mediaPlayer.attachViews(layout, null, true, true)
            }
        },
        modifier = modifier,
        update = { },
    )
    val currentPlayer = remember(mediaPlayer) { mediaPlayer }
    DisposableEffect(currentPlayer) {
        onDispose {
            currentPlayer.detachViews()
        }
    }
}
