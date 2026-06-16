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
import androidx.media3.common.TrackSelectionParameters
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.source.DefaultMediaSourceFactory
import androidx.media3.ui.PlayerView
import com.chee.videos.core.ui.LongFormAudioTrack
import com.chee.videos.core.player.friendlyLongFormPlaybackErrorMessage
import kotlinx.coroutines.delay

internal const val TvDolbyVisionMedia3StartupTimeoutMillis = 15_000L
internal const val TvDolbyVisionMedia3StartupTimeoutMessage = "杜比视界专用播放链路启动超时，请重试"

internal data class TvMedia3PlaybackSnapshot(
    val positionMs: Long,
    val durationMs: Long,
)

internal fun shouldReportTvDolbyVisionMedia3StartupTimeout(
    preparedSourceKey: String,
    shouldPlay: Boolean,
    playbackState: Int,
    isPlaying: Boolean,
): Boolean =
    preparedSourceKey.isNotBlank() &&
        shouldPlay &&
        !isPlaying &&
        playbackState != Player.STATE_READY &&
        playbackState != Player.STATE_ENDED

@Composable
internal fun TvDolbyVisionMedia3Player(
    sourceUrl: String,
    mediaId: String,
    title: String,
    accessToken: String,
    retryKey: Int,
    shouldPlay: Boolean,
    initialPositionMs: Long,
    seekPositionMs: Long? = null,
    seekRequestKey: Int = 0,
    subtitleConfigurations: List<MediaItem.SubtitleConfiguration> = emptyList(),
    selectedSubtitleTrackId: String? = null,
    selectedAudioTrackId: String? = null,
    modifier: Modifier = Modifier,
    onPlayingChanged: (Boolean) -> Unit = {},
    onError: (String) -> Unit = {},
    onEnded: () -> Unit = {},
    onSnapshotChanged: (TvMedia3PlaybackSnapshot) -> Unit = {},
    onAudioTracksChanged: (List<LongFormAudioTrack>) -> Unit = {},
) {
    val context = LocalContext.current
    val lifecycleOwner = LocalLifecycleOwner.current
    val latestOnPlayingChanged by rememberUpdatedState(onPlayingChanged)
    val latestOnError by rememberUpdatedState(onError)
    val latestOnEnded by rememberUpdatedState(onEnded)
    val latestOnSnapshotChanged by rememberUpdatedState(onSnapshotChanged)
    val latestOnAudioTracksChanged by rememberUpdatedState(onAudioTracksChanged)
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
    var playbackState by remember { mutableStateOf(Player.STATE_IDLE) }
    var isMedia3Playing by remember { mutableStateOf(false) }
    var trackSelectionRevision by remember { mutableStateOf(0) }

    DisposableEffect(player) {
        val listener = object : Player.Listener {
            override fun onIsPlayingChanged(playing: Boolean) {
                isMedia3Playing = playing
                latestOnPlayingChanged(playing)
            }

            override fun onPlaybackStateChanged(nextPlaybackState: Int) {
                playbackState = nextPlaybackState
                if (nextPlaybackState == Player.STATE_ENDED) {
                    latestOnPlayingChanged(false)
                    latestOnEnded()
                }
            }

            override fun onTracksChanged(tracks: androidx.media3.common.Tracks) {
                trackSelectionRevision += 1
                latestOnAudioTracksChanged(buildTvMedia3AudioTracks(tracks))
            }

            override fun onPlayerError(error: androidx.media3.common.PlaybackException) {
                latestOnPlayingChanged(false)
                latestOnError(friendlyLongFormPlaybackErrorMessage(error))
            }
        }
        player.addListener(listener)
        onDispose {
            player.removeListener(listener)
            latestOnAudioTracksChanged(emptyList())
            latestOnSnapshotChanged(player.readTvMedia3PlaybackSnapshot())
            latestOnPlayingChanged(false)
            player.release()
        }
    }

    LaunchedEffect(player, sourceUrl, mediaId, title, dataSourceFactory, subtitleConfigurations) {
        val subtitleKey = subtitleConfigurations.joinToString("|") {
            listOf(it.id.orEmpty(), it.uri.toString(), it.mimeType.orEmpty(), it.language.orEmpty()).joinToString(",")
        }
        val sourceKey = "$mediaId|$sourceUrl|$subtitleKey"
        if (sourceUrl.isBlank() || mediaId.isBlank()) {
            player.pause()
            player.clearMediaItems()
            preparedSourceKey = ""
            resumeAppliedSourceKey = ""
            playbackState = Player.STATE_IDLE
            isMedia3Playing = false
            trackSelectionRevision += 1
            latestOnAudioTracksChanged(emptyList())
            return@LaunchedEffect
        }
        if (preparedSourceKey != sourceKey) {
            val mediaItem = MediaItem.Builder()
                .setUri(sourceUrl)
                .setMediaId(mediaId)
                .setSubtitleConfigurations(subtitleConfigurations)
                .setMediaMetadata(
                    androidx.media3.common.MediaMetadata.Builder()
                        .setTitle(title)
                        .build(),
                )
                .build()
            val mediaSource = DefaultMediaSourceFactory(dataSourceFactory).createMediaSource(mediaItem)
            player.setMediaSource(mediaSource, true)
            player.prepare()
            playbackState = Player.STATE_IDLE
            isMedia3Playing = false
            trackSelectionRevision += 1
            latestOnAudioTracksChanged(emptyList())
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

    LaunchedEffect(player, preparedSourceKey, seekPositionMs, seekRequestKey) {
        val target = seekPositionMs ?: return@LaunchedEffect
        if (preparedSourceKey.isBlank() || seekRequestKey <= 0) {
            return@LaunchedEffect
        }
        player.seekTo(target.coerceAtLeast(0L))
        latestOnSnapshotChanged(player.readTvMedia3PlaybackSnapshot())
    }

    LaunchedEffect(player, preparedSourceKey, trackSelectionRevision, selectedSubtitleTrackId) {
        if (preparedSourceKey.isBlank()) {
            return@LaunchedEffect
        }
        val nextParameters = resolveTvMedia3SubtitleSelectionParameters(
            currentParameters = player.trackSelectionParameters,
            tracks = player.currentTracks,
            selectedSubtitleTrackId = selectedSubtitleTrackId,
        )
        if (nextParameters != player.trackSelectionParameters) {
            player.trackSelectionParameters = nextParameters
        }
    }

    LaunchedEffect(player, preparedSourceKey, trackSelectionRevision, selectedAudioTrackId) {
        if (preparedSourceKey.isBlank()) {
            return@LaunchedEffect
        }
        val nextParameters = resolveTvMedia3AudioSelectionParameters(
            currentParameters = player.trackSelectionParameters,
            tracks = player.currentTracks,
            selectedAudioTrackId = selectedAudioTrackId,
        )
        if (nextParameters != player.trackSelectionParameters) {
            player.trackSelectionParameters = nextParameters
        }
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

    LaunchedEffect(preparedSourceKey, shouldPlay, playbackState, isMedia3Playing, retryKey) {
        if (!shouldReportTvDolbyVisionMedia3StartupTimeout(preparedSourceKey, shouldPlay, playbackState, isMedia3Playing)) {
            return@LaunchedEffect
        }
        delay(TvDolbyVisionMedia3StartupTimeoutMillis)
        if (shouldReportTvDolbyVisionMedia3StartupTimeout(preparedSourceKey, shouldPlay, playbackState, isMedia3Playing)) {
            player.pause()
            latestOnPlayingChanged(false)
            latestOnError(TvDolbyVisionMedia3StartupTimeoutMessage)
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

internal fun resolveTvMedia3SubtitleSelectionParameters(
    currentParameters: TrackSelectionParameters,
    tracks: androidx.media3.common.Tracks,
    selectedSubtitleTrackId: String?,
): TrackSelectionParameters {
    val normalizedSelection = selectedSubtitleTrackId?.trim().orEmpty()
    if (normalizedSelection.isBlank()) {
        return buildTvMedia3SelectionParametersForDisabledText(currentParameters)
    }
    val groupAndTrack = findTvMedia3TrackById(
        tracks = tracks,
        type = androidx.media3.common.C.TRACK_TYPE_TEXT,
        trackId = normalizedSelection,
    ) ?: return currentParameters
    return buildTvMedia3SelectionParametersForTrack(
        currentParameters = currentParameters,
        tracks = tracks,
        type = androidx.media3.common.C.TRACK_TYPE_TEXT,
        groupIndex = groupAndTrack.groupIndex,
        trackIndex = groupAndTrack.trackIndex,
    )
}

internal fun resolveTvMedia3AudioSelectionParameters(
    currentParameters: TrackSelectionParameters,
    tracks: androidx.media3.common.Tracks,
    selectedAudioTrackId: String?,
): TrackSelectionParameters {
    val normalizedSelection = selectedAudioTrackId?.trim().orEmpty()
    if (normalizedSelection.isBlank()) {
        return buildTvMedia3SelectionParametersForAuto(currentParameters, androidx.media3.common.C.TRACK_TYPE_AUDIO)
    }
    val groupAndTrack = findTvMedia3TrackById(
        tracks = tracks,
        type = androidx.media3.common.C.TRACK_TYPE_AUDIO,
        trackId = normalizedSelection,
    ) ?: return currentParameters
    return buildTvMedia3SelectionParametersForTrack(
        currentParameters = currentParameters,
        tracks = tracks,
        type = androidx.media3.common.C.TRACK_TYPE_AUDIO,
        groupIndex = groupAndTrack.groupIndex,
        trackIndex = groupAndTrack.trackIndex,
    )
}

private data class TvMedia3TrackPosition(
    val groupIndex: Int,
    val trackIndex: Int,
)

private fun findTvMedia3TrackById(
    tracks: androidx.media3.common.Tracks,
    type: Int,
    trackId: String,
): TvMedia3TrackPosition? {
    tracks.groups
        .filter { it.type == type }
        .forEachIndexed { groupIndex, group ->
            for (trackIndex in 0 until group.length) {
                val format = group.getTrackFormat(trackIndex)
                val generatedId = if (type == androidx.media3.common.C.TRACK_TYPE_AUDIO) {
                    buildTvMedia3TrackId("audio", groupIndex, trackIndex, format)
                } else {
                    format.id.orEmpty()
                }
                if (generatedId == trackId || format.id == trackId) {
                    return TvMedia3TrackPosition(groupIndex, trackIndex)
                }
            }
        }
    return null
}
