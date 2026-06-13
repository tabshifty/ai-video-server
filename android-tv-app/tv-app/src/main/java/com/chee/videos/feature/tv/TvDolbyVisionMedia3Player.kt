package com.chee.videos.feature.tv

import android.view.ViewGroup
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.viewinterop.AndroidView
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.media3.common.MediaItem
import androidx.media3.common.Player
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.source.DefaultMediaSourceFactory
import androidx.media3.ui.PlayerView
import com.chee.videos.core.player.friendlyLongFormPlaybackErrorMessage
import kotlinx.coroutines.delay

internal data class TvMedia3PlaybackSnapshot(
    val positionMs: Long,
    val durationMs: Long,
)

@Composable
internal fun TvDolbyVisionMedia3Player(
    sourceUrl: String,
    mediaId: String,
    title: String,
    accessToken: String,
    retryKey: Int,
    shouldPlay: Boolean,
    initialPositionMs: Long,
    modifier: Modifier = Modifier,
    onPlayingChanged: (Boolean) -> Unit = {},
    onError: (String) -> Unit = {},
    onEnded: () -> Unit = {},
    onSnapshotChanged: (TvMedia3PlaybackSnapshot) -> Unit = {},
) {
    val context = LocalContext.current
    val lifecycleOwner = LocalLifecycleOwner.current
    val latestOnPlayingChanged by rememberUpdatedState(onPlayingChanged)
    val latestOnError by rememberUpdatedState(onError)
    val latestOnEnded by rememberUpdatedState(onEnded)
    val latestOnSnapshotChanged by rememberUpdatedState(onSnapshotChanged)
    val dataSourceFactory = remember(accessToken) {
        DefaultHttpDataSource.Factory().setAllowCrossProtocolRedirects(true).apply {
            if (accessToken.isNotBlank()) {
                setDefaultRequestProperties(mapOf("Authorization" to "Bearer $accessToken"))
            }
        }
    }
    val player = remember(accessToken, retryKey) { ExoPlayer.Builder(context).build() }
    var preparedSourceKey by remember { mutableStateOf("") }
    var resumeAppliedSourceKey by remember { mutableStateOf("") }

    DisposableEffect(player) {
        val listener = object : Player.Listener {
            override fun onIsPlayingChanged(isPlaying: Boolean) {
                latestOnPlayingChanged(isPlaying)
            }

            override fun onPlaybackStateChanged(playbackState: Int) {
                if (playbackState == Player.STATE_ENDED) {
                    latestOnPlayingChanged(false)
                    latestOnEnded()
                }
            }

            override fun onPlayerError(error: androidx.media3.common.PlaybackException) {
                latestOnPlayingChanged(false)
                latestOnError(friendlyLongFormPlaybackErrorMessage(error))
            }
        }
        player.addListener(listener)
        onDispose {
            player.removeListener(listener)
            latestOnSnapshotChanged(player.readTvMedia3PlaybackSnapshot())
            latestOnPlayingChanged(false)
            player.release()
        }
    }

    LaunchedEffect(player, sourceUrl, mediaId, title, dataSourceFactory) {
        val sourceKey = "$mediaId|$sourceUrl"
        if (sourceUrl.isBlank() || mediaId.isBlank()) {
            player.pause()
            player.clearMediaItems()
            preparedSourceKey = ""
            resumeAppliedSourceKey = ""
            return@LaunchedEffect
        }
        if (preparedSourceKey != sourceKey) {
            val mediaItem = MediaItem.Builder()
                .setUri(sourceUrl)
                .setMediaId(mediaId)
                .setMediaMetadata(
                    androidx.media3.common.MediaMetadata.Builder()
                        .setTitle(title)
                        .build(),
                )
                .build()
            val mediaSource = DefaultMediaSourceFactory(dataSourceFactory).createMediaSource(mediaItem)
            player.setMediaSource(mediaSource, true)
            player.prepare()
            preparedSourceKey = sourceKey
            resumeAppliedSourceKey = ""
        }
    }

    LaunchedEffect(player, preparedSourceKey, initialPositionMs) {
        if (preparedSourceKey.isBlank() || resumeAppliedSourceKey == preparedSourceKey || initialPositionMs <= 0L) {
            return@LaunchedEffect
        }
        delay(250L)
        player.seekTo(initialPositionMs)
        resumeAppliedSourceKey = preparedSourceKey
    }

    LaunchedEffect(player, preparedSourceKey, shouldPlay) {
        if (preparedSourceKey.isBlank()) {
            return@LaunchedEffect
        }
        player.playWhenReady = shouldPlay
        if (shouldPlay) {
            player.play()
        } else {
            player.pause()
        }
    }

    DisposableEffect(lifecycleOwner, player, shouldPlay, preparedSourceKey) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_PAUSE -> player.pause()
                Lifecycle.Event.ON_RESUME -> {
                    if (shouldPlay && preparedSourceKey.isNotBlank()) {
                        player.play()
                    }
                }
                else -> Unit
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose {
            lifecycleOwner.lifecycle.removeObserver(observer)
        }
    }

    LaunchedEffect(player, preparedSourceKey) {
        while (preparedSourceKey.isNotBlank()) {
            latestOnSnapshotChanged(player.readTvMedia3PlaybackSnapshot())
            delay(250L)
        }
    }

    Box(
        modifier = modifier
            .fillMaxSize()
            .background(Color.Black),
    ) {
        AndroidView(
            modifier = Modifier.fillMaxSize(),
            factory = {
                PlayerView(it).apply {
                    useController = false
                    setShutterBackgroundColor(android.graphics.Color.BLACK)
                    layoutParams = ViewGroup.LayoutParams(
                        ViewGroup.LayoutParams.MATCH_PARENT,
                        ViewGroup.LayoutParams.MATCH_PARENT,
                    )
                    this.player = player
                }
            },
            update = { view ->
                view.player = player
            },
        )
    }
}

internal fun ExoPlayer.readTvMedia3PlaybackSnapshot(): TvMedia3PlaybackSnapshot =
    TvMedia3PlaybackSnapshot(
        positionMs = currentPosition.coerceAtLeast(0L),
        durationMs = duration.takeIf { it > 0L } ?: 0L,
    )
