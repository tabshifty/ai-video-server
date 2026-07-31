@file:androidx.annotation.OptIn(markerClass = [androidx.media3.common.util.UnstableApi::class])

package com.chee.videos.feature.tv

import android.content.Context
import android.view.LayoutInflater
import android.view.ViewGroup
import android.view.accessibility.CaptioningManager
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
import androidx.media3.common.AudioAttributes
import androidx.media3.common.MediaItem
import androidx.media3.common.PlaybackException
import androidx.media3.common.Player
import androidx.media3.common.Timeline
import androidx.media3.common.TrackSelectionParameters
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.datasource.HttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.analytics.AnalyticsListener
import androidx.media3.exoplayer.source.DefaultMediaSourceFactory
import androidx.media3.session.MediaSession
import androidx.media3.ui.CaptionStyleCompat
import androidx.media3.ui.PlayerView
import androidx.media3.ui.SubtitleView
import com.chee.videos.core.model.TvSubtitlePreferenceMode
import com.chee.videos.core.ui.LongFormAudioTrack
import com.chee.videos.core.player.friendlyLongFormPlaybackErrorMessage
import com.chee.videos.tv.R
import kotlinx.coroutines.delay

internal const val TvLongFormMedia3StartupTimeoutMillis = TvLongFormLoadingTimeoutMillis
internal const val TvLongFormMedia3StartupTimeoutMessage = "视频加载超时，请重试"
private const val TvLongFormMedia3RetryKeySeparator = "|retry:"

internal fun buildTvLongFormMedia3ItemId(mediaId: String, retryKey: Int): String =
    "$mediaId$TvLongFormMedia3RetryKeySeparator$retryKey"

internal data class TvLongFormMedia3EventIdentity(
    val mediaId: String,
    val retryKey: Int,
)

internal fun isCurrentTvLongFormMedia3Event(
    eventIdentity: TvLongFormMedia3EventIdentity,
    preparedIdentity: TvLongFormMedia3EventIdentity?,
): Boolean = eventIdentity == preparedIdentity

internal fun parseTvLongFormMedia3EventIdentity(mediaItemId: String): TvLongFormMedia3EventIdentity? {
    val separatorIndex = mediaItemId.lastIndexOf(TvLongFormMedia3RetryKeySeparator)
    if (separatorIndex <= 0) {
        return null
    }
    val retryKey = mediaItemId
        .substring(separatorIndex + TvLongFormMedia3RetryKeySeparator.length)
        .toIntOrNull()
        ?: return null
    return TvLongFormMedia3EventIdentity(
        mediaId = mediaItemId.substring(0, separatorIndex),
        retryKey = retryKey,
    )
}

private fun AnalyticsListener.EventTime.eventIdentity(): TvLongFormMedia3EventIdentity? {
    if (!timeline.isEmpty && windowIndex in 0 until timeline.windowCount) {
        return parseTvLongFormMedia3EventIdentity(
            timeline.getWindow(windowIndex, Timeline.Window()).mediaItem.mediaId,
        )
    }
    if (!currentTimeline.isEmpty && currentWindowIndex in 0 until currentTimeline.windowCount) {
        return parseTvLongFormMedia3EventIdentity(
            currentTimeline.getWindow(currentWindowIndex, Timeline.Window()).mediaItem.mediaId,
        )
    }
    return null
}

internal data class TvMedia3PlaybackSnapshot(
    val positionMs: Long,
    val durationMs: Long,
)

internal fun shouldReportTvLongFormMedia3StartupTimeout(
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

internal fun shouldPrepareTvLongFormMedia3Source(
    preparedSourceKey: String,
    currentSourceKey: String,
    isPreparedPlayerCurrent: Boolean,
): Boolean =
    currentSourceKey.isNotBlank() &&
        (!isPreparedPlayerCurrent || preparedSourceKey != currentSourceKey)

@Composable
internal fun TvLongFormMedia3Player(
    sourceUrl: String,
    mediaId: String,
    title: String,
    accessToken: String,
    retryKey: Int,
    shouldPlay: Boolean,
    initialPositionMs: Long,
    tvSeekStepSeconds: Int,
    seekPositionMs: Long? = null,
    seekRequestKey: Int = 0,
    outputSurface: TvPlaybackOutputSurface = TvPlaybackOutputSurface.TEXTURE_VIEW,
    subtitleConfigurations: List<MediaItem.SubtitleConfiguration> = emptyList(),
    selectedSubtitleTrackId: String? = null,
    subtitlePreferenceMode: TvSubtitlePreferenceMode = TvSubtitlePreferenceMode.AUTO,
    selectedAudioTrackId: String? = null,
    interactionMode: TvLongFormInteractionMode = TvLongFormInteractionMode.Hidden,
    modifier: Modifier = Modifier,
    onPlayingChanged: (Boolean, TvLongFormMedia3EventIdentity) -> Unit = { _, _ -> },
    onRenderedFirstFrame: (TvLongFormMedia3EventIdentity) -> Unit = {},
    onError: (String, TvLongFormMedia3EventIdentity, Boolean) -> Unit = { _, _, _ -> },
    onEnded: (TvLongFormMedia3EventIdentity) -> Unit = {},
    onSnapshotChanged: (TvMedia3PlaybackSnapshot) -> Unit = {},
    onLifecyclePauseSnapshot: (TvMedia3PlaybackSnapshot) -> Unit = {},
    onLifecyclePaused: () -> Unit = {},
    onPlaybackIntentChanged: (Boolean) -> Unit = {},
    onPlaybackStatusChanged: (TvLongFormPlaybackStatus) -> Unit = {},
    onAudioTracksChanged: (List<LongFormAudioTrack>) -> Unit = {},
) {
    val context = LocalContext.current
    val lifecycleOwner = LocalLifecycleOwner.current
    val latestOnPlayingChanged by rememberUpdatedState(onPlayingChanged)
    val latestOnRenderedFirstFrame by rememberUpdatedState(onRenderedFirstFrame)
    val latestOnError by rememberUpdatedState(onError)
    val latestOnEnded by rememberUpdatedState(onEnded)
    val latestOnSnapshotChanged by rememberUpdatedState(onSnapshotChanged)
    val latestOnLifecyclePauseSnapshot by rememberUpdatedState(onLifecyclePauseSnapshot)
    val latestOnLifecyclePaused by rememberUpdatedState(onLifecyclePaused)
    val latestOnPlaybackIntentChanged by rememberUpdatedState(onPlaybackIntentChanged)
    val latestOnPlaybackStatusChanged by rememberUpdatedState(onPlaybackStatusChanged)
    val latestOnAudioTracksChanged by rememberUpdatedState(onAudioTracksChanged)
    val dataSourceFactory = remember(accessToken) {
        DefaultHttpDataSource.Factory().setAllowCrossProtocolRedirects(true).apply {
            if (accessToken.isNotBlank()) {
                setDefaultRequestProperties(mapOf("Authorization" to "Bearer $accessToken"))
            }
        }
    }
    val seekIncrementMs = tvSeekStepSeconds.coerceAtLeast(1).toLong() * 1_000L
    val player = remember(accessToken, seekIncrementMs) {
        ExoPlayer.Builder(context)
            .setSeekBackIncrementMs(seekIncrementMs)
            .setSeekForwardIncrementMs(seekIncrementMs)
            .build()
            .apply {
            setAudioAttributes(AudioAttributes.DEFAULT, true)
            setHandleAudioBecomingNoisy(true)
        }
    }
    val playerSourceKey = remember(player) { Any() }
    var preparedPlayerSourceKey by remember { mutableStateOf<Any?>(null) }
    var preparedSourceKey by remember { mutableStateOf("") }
    var preparedIdentity by remember { mutableStateOf<TvLongFormMedia3EventIdentity?>(null) }
    var resumeAppliedSourceKey by remember { mutableStateOf("") }
    var playbackState by remember { mutableStateOf(Player.STATE_IDLE) }
    var isMedia3Playing by remember { mutableStateOf(false) }
    var trackSelectionRevision by remember { mutableStateOf(0) }
    var pendingInternalPlaybackIntent by remember(player) { mutableStateOf<Boolean?>(null) }

    fun applyInternalPlaybackIntent(shouldPlay: Boolean) {
        if (player.playWhenReady != shouldPlay) {
            pendingInternalPlaybackIntent = shouldPlay
        }
        if (shouldPlay) {
            player.play()
        } else {
            player.pause()
        }
    }

    DisposableEffect(player) {
        val listener = object : Player.Listener {
            override fun onTracksChanged(tracks: androidx.media3.common.Tracks) {
                trackSelectionRevision += 1
                latestOnAudioTracksChanged(buildTvMedia3AudioTracks(tracks))
            }

            override fun onPlayWhenReadyChanged(playWhenReady: Boolean, reason: Int) {
                val consumedInternalIntent = pendingInternalPlaybackIntent == playWhenReady
                if (consumedInternalIntent) {
                    pendingInternalPlaybackIntent = null
                } else if (reason == Player.PLAY_WHEN_READY_CHANGE_REASON_USER_REQUEST) {
                    latestOnPlaybackIntentChanged(playWhenReady)
                }
                if (
                    !playWhenReady &&
                    reason in setOf(
                        Player.PLAY_WHEN_READY_CHANGE_REASON_AUDIO_FOCUS_LOSS,
                        Player.PLAY_WHEN_READY_CHANGE_REASON_AUDIO_BECOMING_NOISY,
                    )
                ) {
                    latestOnLifecyclePaused()
                }
            }
        }
        val analyticsListener = object : AnalyticsListener {
            override fun onIsPlayingChanged(eventTime: AnalyticsListener.EventTime, isPlaying: Boolean) {
                val eventIdentity = eventTime.eventIdentity() ?: return
                if (isCurrentTvLongFormMedia3Event(eventIdentity, preparedIdentity)) {
                    isMedia3Playing = isPlaying
                    latestOnPlaybackStatusChanged(
                        if (isPlaying) TvLongFormPlaybackStatus.Playing else resolveTvLongFormPlaybackStatus(
                            playbackState = playbackState,
                            shouldPlay = player.playWhenReady,
                        ),
                    )
                }
                latestOnPlayingChanged(isPlaying, eventIdentity)
            }

            override fun onPlaybackStateChanged(
                eventTime: AnalyticsListener.EventTime,
                nextPlaybackState: Int,
            ) {
                val eventIdentity = eventTime.eventIdentity() ?: return
                if (!isCurrentTvLongFormMedia3Event(eventIdentity, preparedIdentity)) {
                    return
                }
                playbackState = nextPlaybackState
                latestOnPlaybackStatusChanged(
                    resolveTvLongFormPlaybackStatus(
                        playbackState = nextPlaybackState,
                        shouldPlay = player.playWhenReady,
                    ),
                )
                if (nextPlaybackState == Player.STATE_ENDED) {
                    latestOnEnded(eventIdentity)
                }
            }

            override fun onRenderedFirstFrame(
                eventTime: AnalyticsListener.EventTime,
                output: Any,
                renderTimeMs: Long,
            ) {
                val eventIdentity = eventTime.eventIdentity() ?: return
                latestOnRenderedFirstFrame(eventIdentity)
            }

            override fun onPlayerError(
                eventTime: AnalyticsListener.EventTime,
                error: androidx.media3.common.PlaybackException,
            ) {
                val eventIdentity = eventTime.eventIdentity() ?: return
                if (!isCurrentTvLongFormMedia3Event(eventIdentity, preparedIdentity)) {
                    return
                }
                isMedia3Playing = false
                latestOnPlaybackStatusChanged(TvLongFormPlaybackStatus.Error)
                latestOnPlayingChanged(false, eventIdentity)
                latestOnError(
                    friendlyLongFormPlaybackErrorMessage(error),
                    eventIdentity,
                    error.isTvLongFormRetryable(),
                )
            }
        }
        player.addListener(listener)
        player.addAnalyticsListener(analyticsListener)
        onDispose {
            player.removeListener(listener)
            player.removeAnalyticsListener(analyticsListener)
            latestOnAudioTracksChanged(emptyList())
            latestOnPlaybackStatusChanged(TvLongFormPlaybackStatus.Idle)
            latestOnSnapshotChanged(player.readTvMedia3PlaybackSnapshot())
            latestOnPlayingChanged(
                false,
                preparedIdentity ?: TvLongFormMedia3EventIdentity(mediaId = mediaId, retryKey = retryKey),
            )
            player.release()
        }
    }

    LaunchedEffect(player, sourceUrl, mediaId, title, dataSourceFactory, subtitleConfigurations, retryKey) {
        val subtitleKey = subtitleConfigurations.joinToString("|") {
            listOf(it.id.orEmpty(), it.uri.toString(), it.mimeType.orEmpty(), it.language.orEmpty()).joinToString(",")
        }
        val sourceKey = "$mediaId|$sourceUrl|$subtitleKey|retry:$retryKey"
        if (sourceUrl.isBlank() || mediaId.isBlank()) {
            applyInternalPlaybackIntent(false)
            player.clearMediaItems()
            preparedPlayerSourceKey = null
            preparedSourceKey = ""
            preparedIdentity = null
            resumeAppliedSourceKey = ""
            playbackState = Player.STATE_IDLE
            isMedia3Playing = false
            latestOnPlaybackStatusChanged(TvLongFormPlaybackStatus.Idle)
            trackSelectionRevision += 1
            latestOnAudioTracksChanged(emptyList())
            return@LaunchedEffect
        }
        if (
            shouldPrepareTvLongFormMedia3Source(
                preparedSourceKey = preparedSourceKey,
                currentSourceKey = sourceKey,
                isPreparedPlayerCurrent = preparedPlayerSourceKey === playerSourceKey,
            )
        ) {
            val nextIdentity = TvLongFormMedia3EventIdentity(mediaId = mediaId, retryKey = retryKey)
            preparedPlayerSourceKey = playerSourceKey
            preparedSourceKey = sourceKey
            preparedIdentity = nextIdentity
            resumeAppliedSourceKey = ""
            val mediaItem = MediaItem.Builder()
                .setUri(sourceUrl)
                .setMediaId(buildTvLongFormMedia3ItemId(mediaId, retryKey))
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
            latestOnPlaybackStatusChanged(TvLongFormPlaybackStatus.Preparing)
            trackSelectionRevision += 1
            latestOnAudioTracksChanged(emptyList())
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

    LaunchedEffect(player, preparedSourceKey, trackSelectionRevision, selectedSubtitleTrackId, subtitlePreferenceMode) {
        if (preparedSourceKey.isBlank()) {
            return@LaunchedEffect
        }
        val nextParameters = resolveTvMedia3SubtitleSelectionParameters(
            currentParameters = player.trackSelectionParameters,
            tracks = player.currentTracks,
            selectedSubtitleTrackId = selectedSubtitleTrackId,
            subtitlePreferenceMode = subtitlePreferenceMode,
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

    LaunchedEffect(player, preparedSourceKey, preparedIdentity, shouldPlay) {
        if (preparedSourceKey.isBlank()) {
            return@LaunchedEffect
        }
        applyInternalPlaybackIntent(shouldPlay)
    }

    LaunchedEffect(preparedSourceKey, preparedIdentity, shouldPlay, playbackState, isMedia3Playing, retryKey) {
        if (
            !shouldReportTvLongFormMedia3StartupTimeout(preparedSourceKey, shouldPlay, playbackState, isMedia3Playing)
        ) {
            return@LaunchedEffect
        }
        delay(TvLongFormMedia3StartupTimeoutMillis)
        if (
            shouldReportTvLongFormMedia3StartupTimeout(preparedSourceKey, shouldPlay, playbackState, isMedia3Playing)
        ) {
            applyInternalPlaybackIntent(false)
            val timeoutIdentity = TvLongFormMedia3EventIdentity(mediaId, retryKey)
            latestOnPlayingChanged(false, timeoutIdentity)
            latestOnError(TvLongFormMedia3StartupTimeoutMessage, timeoutIdentity, true)
        }
    }

    DisposableEffect(lifecycleOwner, player, preparedSourceKey) {
        var mediaSession: MediaSession? = null
        fun activateMediaSession() {
            if (mediaSession == null) {
                mediaSession = MediaSession.Builder(context, player).build()
            }
        }
        fun deactivateMediaSession() {
            mediaSession?.release()
            mediaSession = null
        }
        if (lifecycleOwner.lifecycle.currentState.isAtLeast(Lifecycle.State.RESUMED)) {
            activateMediaSession()
        }
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_PAUSE -> {
                    if (preparedSourceKey.isNotBlank()) {
                        val snapshot = player.readTvMedia3PlaybackSnapshot()
                        latestOnSnapshotChanged(snapshot)
                        latestOnLifecyclePauseSnapshot(snapshot)
                    }
                    applyInternalPlaybackIntent(false)
                    latestOnPlaybackStatusChanged(TvLongFormPlaybackStatus.Paused)
                    latestOnLifecyclePaused()
                    deactivateMediaSession()
                }
                Lifecycle.Event.ON_RESUME -> {
                    activateMediaSession()
                }
                else -> Unit
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose {
            lifecycleOwner.lifecycle.removeObserver(observer)
            deactivateMediaSession()
        }
    }

    LaunchedEffect(player, preparedSourceKey) {
        // 250ms 是进度条刷新时基，不能降频；但暂停/缓冲态下快照恒定，
        // 在发射端做结构相等去重，省掉下游逐 tick 的 State 写入与比较。
        var lastEmitted: TvMedia3PlaybackSnapshot? = null
        while (preparedSourceKey.isNotBlank()) {
            val snapshot = player.readTvMedia3PlaybackSnapshot()
            if (snapshot != lastEmitted) {
                lastEmitted = snapshot
                latestOnSnapshotChanged(snapshot)
            }
            delay(250L)
        }
    }

    val playerModifier = modifier.fillMaxSize()
    Box(modifier = playerModifier.background(Color.Black)) {
        AndroidView(
            modifier = Modifier.fillMaxSize(),
            factory = {
                (LayoutInflater.from(it).inflate(
                    outputSurface.playerViewLayoutResId(),
                    null,
                ) as PlayerView).apply {
                    layoutParams = ViewGroup.LayoutParams(
                        ViewGroup.LayoutParams.MATCH_PARENT,
                        ViewGroup.LayoutParams.MATCH_PARENT,
                    )
                    setKeepContentOnPlayerReset(true)
                    this.player = player
                    applyTvLongFormSubtitlePresentation(context, interactionMode)
                }
            },
            update = { view ->
                view.player = player
                view.applyTvLongFormSubtitlePresentation(context, interactionMode)
            },
        )
    }
}

internal fun resolveTvLongFormPlaybackStatus(
    playbackState: Int,
    shouldPlay: Boolean,
): TvLongFormPlaybackStatus = when (playbackState) {
    Player.STATE_IDLE -> TvLongFormPlaybackStatus.Preparing
    Player.STATE_BUFFERING -> TvLongFormPlaybackStatus.Buffering
    Player.STATE_READY -> if (shouldPlay) TvLongFormPlaybackStatus.Playing else TvLongFormPlaybackStatus.Paused
    Player.STATE_ENDED -> TvLongFormPlaybackStatus.Ended
    else -> TvLongFormPlaybackStatus.Idle
}

internal fun isTvLongFormMedia3ErrorCodeRetryable(errorCode: Int): Boolean = errorCode in setOf(
    PlaybackException.ERROR_CODE_TIMEOUT,
    PlaybackException.ERROR_CODE_IO_UNSPECIFIED,
    PlaybackException.ERROR_CODE_IO_NETWORK_CONNECTION_FAILED,
    PlaybackException.ERROR_CODE_IO_NETWORK_CONNECTION_TIMEOUT,
)

internal fun isTvLongFormHttpStatusRetryable(statusCode: Int): Boolean =
    statusCode == 408 || statusCode == 429 || statusCode in 500..599

private fun PlaybackException.isTvLongFormRetryable(): Boolean {
    if (isTvLongFormMedia3ErrorCodeRetryable(errorCode)) {
        return true
    }
    if (errorCode != PlaybackException.ERROR_CODE_IO_BAD_HTTP_STATUS) {
        return false
    }
    val responseCode = generateSequence(cause) { it.cause }
        .filterIsInstance<HttpDataSource.InvalidResponseCodeException>()
        .firstOrNull()
        ?.responseCode
        ?: return false
    return isTvLongFormHttpStatusRetryable(responseCode)
}

private fun PlayerView.applyTvLongFormSubtitlePresentation(
    context: Context,
    interactionMode: TvLongFormInteractionMode,
) {
    val captions = context.getSystemService(Context.CAPTIONING_SERVICE) as? CaptioningManager
    subtitleView?.apply {
        if (captions?.isEnabled == true) {
            setStyle(CaptionStyleCompat.createFromCaptionStyle(captions.userStyle))
            setFractionalTextSize(
                SubtitleView.DEFAULT_TEXT_SIZE_FRACTION * captions.fontScale.coerceIn(0.75f, 2f),
            )
        } else {
            setStyle(CaptionStyleCompat.DEFAULT)
            setFractionalTextSize(SubtitleView.DEFAULT_TEXT_SIZE_FRACTION)
        }
        setBottomPaddingFraction(
            when (interactionMode) {
                TvLongFormInteractionMode.Controls,
                TvLongFormInteractionMode.PrecisionSeek,
                -> 0.24f
                else -> 0.08f
            },
        )
        val endPadding = if (
            interactionMode == TvLongFormInteractionMode.SubtitlePanel ||
            interactionMode == TvLongFormInteractionMode.AudioPanel ||
            interactionMode == TvLongFormInteractionMode.EpisodePanel
        ) {
            (440 * context.resources.displayMetrics.density).toInt()
        } else {
            0
        }
        setPadding(0, 0, endPadding, 0)
    }
}

private fun TvPlaybackOutputSurface.playerViewLayoutResId(): Int =
    when (this) {
        TvPlaybackOutputSurface.TEXTURE_VIEW -> R.layout.tv_long_form_media3_player_view
        TvPlaybackOutputSurface.SURFACE_VIEW -> R.layout.tv_long_form_media3_surface_player_view
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
    subtitlePreferenceMode: TvSubtitlePreferenceMode,
): TrackSelectionParameters {
    if (subtitlePreferenceMode == TvSubtitlePreferenceMode.AUTO) {
        return buildTvMedia3SelectionParametersForAuto(currentParameters, androidx.media3.common.C.TRACK_TYPE_TEXT)
    }
    if (subtitlePreferenceMode == TvSubtitlePreferenceMode.OFF) {
        return buildTvMedia3SelectionParametersForDisabledText(currentParameters)
    }
    val normalizedSelection = selectedSubtitleTrackId?.trim().orEmpty()
    if (normalizedSelection.isBlank()) {
        return buildTvMedia3SelectionParametersForAuto(currentParameters, androidx.media3.common.C.TRACK_TYPE_TEXT)
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
