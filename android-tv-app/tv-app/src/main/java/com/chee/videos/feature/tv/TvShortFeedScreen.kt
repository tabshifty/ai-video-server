package com.chee.videos.feature.tv

import android.app.Activity
import android.graphics.Color as AndroidColor
import android.view.KeyEvent as AndroidKeyEvent
import androidx.activity.compose.BackHandler
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.focusable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Pause
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Tv
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableLongStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.key.onPreviewKeyEvent
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.media3.common.MediaItem
import androidx.media3.common.PlaybackException
import androidx.media3.common.Player
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.source.ProgressiveMediaSource
import androidx.media3.ui.AspectRatioFrameLayout
import androidx.media3.ui.PlayerView
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.ui.LaunchedTvInitialFocus
import com.chee.videos.core.ui.tryRequestFocus
import com.chee.videos.core.ui.tvFocusableGlow
import com.chee.videos.core.util.UrlBuilder
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

@Composable
fun TvShortFeedScreen(
    baseUrl: String,
    accessToken: String,
    onBack: () -> Unit,
    viewModel: TvShortFeedViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val context = LocalContext.current
    val activity = context as? Activity
    val lifecycleOwner = LocalLifecycleOwner.current
    val rootFocusRequester = remember { FocusRequester() }
    val coroutineScope = rememberCoroutineScope()
    val dataSourceFactory = remember(accessToken) {
        DefaultHttpDataSource.Factory().setAllowCrossProtocolRedirects(true).apply {
            if (accessToken.isNotBlank()) {
                setDefaultRequestProperties(mapOf("Authorization" to "Bearer $accessToken"))
            }
        }
    }
    val sharedPlayer = remember(accessToken) {
        ExoPlayer.Builder(context).build().apply {
            repeatMode = Player.REPEAT_MODE_OFF
        }
    }

    var renderedVideoId by remember { mutableStateOf<String?>(null) }
    var lastHistoryVideoId by remember { mutableStateOf("") }
    var isPlayerActuallyPlaying by remember { mutableStateOf(false) }
    var hasEndedAtCurrentVideo by remember { mutableStateOf(false) }
    var playerPlaybackState by remember { mutableIntStateOf(Player.STATE_IDLE) }
    var playbackErrorMessage by remember { mutableStateOf<String?>(null) }
    var showCenterIndicator by remember { mutableStateOf(false) }
    var centerIndicatorIsPause by remember { mutableStateOf(false) }
    var centerIndicatorHideJob by remember { mutableStateOf<Job?>(null) }
    var showSeekOverlay by remember { mutableStateOf(false) }
    var seekOverlayPositionMs by remember { mutableLongStateOf(0L) }
    var seekOverlayDurationMs by remember { mutableLongStateOf(0L) }
    var seekOverlayHideJob by remember { mutableStateOf<Job?>(null) }
    var playbackRetryNonce by remember { mutableIntStateOf(0) }

    val currentItem = uiState.items.getOrNull(uiState.currentIndex)
    val currentVideoId = currentItem?.id.orEmpty()
    val currentPausedByUser = currentVideoId.isNotBlank() && currentVideoId in uiState.pausedByUserVideoIds
    val latestCurrentVideoId by rememberUpdatedState(currentVideoId)
    val latestCurrentIndex by rememberUpdatedState(uiState.currentIndex)
    val latestLastIndex by rememberUpdatedState(uiState.items.lastIndex)

    LaunchedEffect(Unit) {
        viewModel.load()
    }

    BackHandler(onBack = onBack)

    DisposableEffect(activity) {
        if (activity == null) {
            onDispose { }
        } else {
            val window = activity.window
            val controller = WindowCompat.getInsetsController(window, window.decorView)
            controller.hide(WindowInsetsCompat.Type.systemBars())
            controller.systemBarsBehavior = WindowInsetsControllerCompat.BEHAVIOR_SHOW_TRANSIENT_BARS_BY_SWIPE
            onDispose {
                controller.show(WindowInsetsCompat.Type.systemBars())
            }
        }
    }

    KeepScreenOnEffect(enabled = isPlayerActuallyPlaying)

    DisposableEffect(sharedPlayer) {
        val listener = object : Player.Listener {
            override fun onRenderedFirstFrame() {
                renderedVideoId = sharedPlayer.currentMediaItem?.mediaId
            }

            override fun onIsPlayingChanged(isPlaying: Boolean) {
                isPlayerActuallyPlaying = isPlaying
            }

            override fun onPlaybackStateChanged(playbackState: Int) {
                playerPlaybackState = playbackState
                if (playbackState == Player.STATE_ENDED) {
                    if (latestCurrentVideoId.isBlank()) {
                        return
                    }
                    if (latestCurrentIndex < latestLastIndex && viewModel.moveNext()) {
                        hasEndedAtCurrentVideo = false
                    } else {
                        hasEndedAtCurrentVideo = true
                    }
                }
            }

            override fun onPlayerError(error: PlaybackException) {
                if (latestCurrentVideoId.isNotBlank()) {
                    playbackErrorMessage = error.message?.trim().takeUnless { it.isNullOrBlank() }
                        ?: "短视频播放失败，请重试"
                }
            }
        }
        sharedPlayer.addListener(listener)
        playerPlaybackState = sharedPlayer.playbackState
        isPlayerActuallyPlaying = sharedPlayer.isPlaying
        onDispose {
            val finalVideoId = latestCurrentVideoId.ifBlank { sharedPlayer.currentMediaItem?.mediaId.orEmpty() }
            if (finalVideoId.isNotBlank()) {
                val (watchSeconds, completed) = readTvShortWatchSnapshot(sharedPlayer)
                viewModel.reportHistory(finalVideoId, watchSeconds, completed)
            }
            centerIndicatorHideJob?.cancel()
            seekOverlayHideJob?.cancel()
            sharedPlayer.removeListener(listener)
            isPlayerActuallyPlaying = false
            playerPlaybackState = Player.STATE_IDLE
            sharedPlayer.release()
        }
    }

    LaunchedEffect(currentVideoId) {
        val previousVideoId = lastHistoryVideoId
        if (previousVideoId.isNotBlank() && previousVideoId != currentVideoId) {
            val (watchSeconds, completed) = readTvShortWatchSnapshot(sharedPlayer)
            viewModel.reportHistory(previousVideoId, watchSeconds, completed)
        }
        lastHistoryVideoId = currentVideoId
        renderedVideoId = null
        playbackErrorMessage = null
        hasEndedAtCurrentVideo = false
        showCenterIndicator = false
        centerIndicatorHideJob?.cancel()
        showSeekOverlay = false
        seekOverlayHideJob?.cancel()
        playbackRetryNonce = 0
    }

    LaunchedEffect(currentVideoId, baseUrl, dataSourceFactory, playbackRetryNonce) {
        if (currentVideoId.isBlank()) {
            sharedPlayer.stop()
            sharedPlayer.clearMediaItems()
            renderedVideoId = null
            return@LaunchedEffect
        }
        val normalizedBase = UrlBuilder.normalizeBaseUrl(baseUrl)
        if (normalizedBase.isBlank()) {
            playbackErrorMessage = "服务器地址无效，无法播放短视频"
            return@LaunchedEffect
        }
        playbackErrorMessage = null
        renderedVideoId = null
        hasEndedAtCurrentVideo = false
        sharedPlayer.stop()
        sharedPlayer.clearMediaItems()
        val mediaItem = MediaItem.Builder()
            .setUri(UrlBuilder.source(normalizedBase, currentVideoId))
            .setMediaId(currentVideoId)
            .build()
        val mediaSource = ProgressiveMediaSource.Factory(dataSourceFactory).createMediaSource(mediaItem)
        sharedPlayer.setMediaSource(mediaSource, true)
        sharedPlayer.prepare()
    }

    LaunchedEffect(uiState.currentIndex, uiState.items.size) {
        if (uiState.items.isNotEmpty()) {
            viewModel.ensureMoreLoaded(uiState.currentIndex)
        }
    }

    LaunchedEffect(uiState.items.size, uiState.currentIndex, currentVideoId, hasEndedAtCurrentVideo) {
        if (hasEndedAtCurrentVideo && currentVideoId.isNotBlank() && uiState.currentIndex < uiState.items.lastIndex) {
            hasEndedAtCurrentVideo = false
            viewModel.moveNext()
        }
    }

    LaunchedEffect(currentVideoId, currentPausedByUser, playbackErrorMessage) {
        val shouldPlay = currentVideoId.isNotBlank() && !currentPausedByUser && playbackErrorMessage == null
        sharedPlayer.playWhenReady = shouldPlay
        if (shouldPlay) {
            sharedPlayer.play()
        } else {
            sharedPlayer.pause()
        }
    }

    DisposableEffect(lifecycleOwner, sharedPlayer, currentVideoId, currentPausedByUser, playbackErrorMessage) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_PAUSE -> sharedPlayer.pause()
                Lifecycle.Event.ON_RESUME -> {
                    if (currentVideoId.isNotBlank() && !currentPausedByUser && playbackErrorMessage == null) {
                        sharedPlayer.play()
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

    val showInitialLoading = uiState.loading && uiState.items.isEmpty()
    val showInitialError = !uiState.loadErrorMessage.isNullOrBlank() && uiState.items.isEmpty()
    val showEmptyState = !showInitialLoading && !showInitialError && uiState.items.isEmpty()
    val showPlaybackLoading = currentVideoId.isNotBlank() &&
        playbackErrorMessage == null &&
        (renderedVideoId != currentVideoId || playerPlaybackState == Player.STATE_BUFFERING)

    LaunchedTvInitialFocus(currentVideoId, playbackErrorMessage, uiState.loading, uiState.items.size) {
        if (currentVideoId.isNotBlank() && playbackErrorMessage == null && !showInitialLoading && !showEmptyState) {
            rootFocusRequester.tryRequestFocus()
        }
    }

    when {
        showInitialLoading -> {
            TvShortFeedLoadingState()
        }

        showInitialError -> {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(Color.Black),
                contentAlignment = Alignment.Center,
            ) {
                TvShortFeedProblemState(
                    title = "短视频加载失败",
                    message = uiState.loadErrorMessage.orEmpty(),
                    primaryActionLabel = "重试",
                    onPrimaryAction = { viewModel.load(force = true) },
                    secondaryActionLabel = "返回首页",
                    onSecondaryAction = onBack,
                )
            }
        }

        showEmptyState -> {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(Color.Black),
                contentAlignment = Alignment.Center,
            ) {
                TvShortFeedProblemState(
                    title = "暂无短视频",
                    message = "当前没有可播放的短视频",
                    primaryActionLabel = "返回首页",
                    onPrimaryAction = onBack,
                )
            }
        }

        else -> {
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .background(Color.Black)
                    .focusRequester(rootFocusRequester)
                    .focusable()
                    .onPreviewKeyEvent { event ->
                        if (event.nativeKeyEvent.action != AndroidKeyEvent.ACTION_DOWN) {
                            return@onPreviewKeyEvent false
                        }
                        if (playbackErrorMessage != null) {
                            return@onPreviewKeyEvent false
                        }
                        when (event.nativeKeyEvent.keyCode) {
                            AndroidKeyEvent.KEYCODE_DPAD_UP -> {
                                if (event.nativeKeyEvent.repeatCount > 0) {
                                    return@onPreviewKeyEvent true
                                }
                                viewModel.movePrevious()
                            }

                            AndroidKeyEvent.KEYCODE_DPAD_DOWN -> {
                                if (event.nativeKeyEvent.repeatCount > 0) {
                                    return@onPreviewKeyEvent true
                                }
                                viewModel.moveNext()
                            }

                            AndroidKeyEvent.KEYCODE_DPAD_LEFT,
                            AndroidKeyEvent.KEYCODE_MEDIA_REWIND,
                            -> {
                                seekCurrentShortVideo(
                                    player = sharedPlayer,
                                    stepSeconds = uiState.seekStepSeconds,
                                    repeatCount = event.nativeKeyEvent.repeatCount,
                                    forward = false,
                                    onSeekApplied = { targetMs, durationMs ->
                                        seekOverlayPositionMs = targetMs
                                        seekOverlayDurationMs = durationMs
                                        showSeekOverlay = true
                                        seekOverlayHideJob?.cancel()
                                        seekOverlayHideJob = coroutineScope.launch {
                                            delay(TvShortSeekOverlayDurationMillis)
                                            showSeekOverlay = false
                                        }
                                    },
                                )
                            }

                            AndroidKeyEvent.KEYCODE_DPAD_RIGHT,
                            AndroidKeyEvent.KEYCODE_MEDIA_FAST_FORWARD,
                            -> {
                                seekCurrentShortVideo(
                                    player = sharedPlayer,
                                    stepSeconds = uiState.seekStepSeconds,
                                    repeatCount = event.nativeKeyEvent.repeatCount,
                                    forward = true,
                                    onSeekApplied = { targetMs, durationMs ->
                                        seekOverlayPositionMs = targetMs
                                        seekOverlayDurationMs = durationMs
                                        showSeekOverlay = true
                                        seekOverlayHideJob?.cancel()
                                        seekOverlayHideJob = coroutineScope.launch {
                                            delay(TvShortSeekOverlayDurationMillis)
                                            showSeekOverlay = false
                                        }
                                    },
                                )
                            }

                            AndroidKeyEvent.KEYCODE_DPAD_CENTER,
                            AndroidKeyEvent.KEYCODE_ENTER,
                            AndroidKeyEvent.KEYCODE_NUMPAD_ENTER,
                            AndroidKeyEvent.KEYCODE_MEDIA_PLAY_PAUSE,
                            AndroidKeyEvent.KEYCODE_MEDIA_PLAY,
                            AndroidKeyEvent.KEYCODE_MEDIA_PAUSE,
                            -> {
                                if (event.nativeKeyEvent.repeatCount > 0 || currentVideoId.isBlank()) {
                                    return@onPreviewKeyEvent true
                                }
                                val pausedNow = viewModel.togglePauseCurrent(currentVideoId)
                                centerIndicatorIsPause = pausedNow
                                showCenterIndicator = true
                                centerIndicatorHideJob?.cancel()
                                centerIndicatorHideJob = coroutineScope.launch {
                                    delay(TvShortCenterIndicatorDurationMillis)
                                    showCenterIndicator = false
                                }
                            }

                            AndroidKeyEvent.KEYCODE_BACK,
                            AndroidKeyEvent.KEYCODE_ESCAPE,
                            -> {
                                onBack()
                            }

                            else -> return@onPreviewKeyEvent false
                        }
                        true
                    },
            ) {
                AndroidView(
                    factory = {
                        PlayerView(it).apply {
                            useController = false
                            setShutterBackgroundColor(AndroidColor.BLACK)
                            setKeepContentOnPlayerReset(false)
                            resizeMode = AspectRatioFrameLayout.RESIZE_MODE_FIT
                            player = sharedPlayer
                        }
                    },
                    update = { view ->
                        view.player = sharedPlayer
                        view.resizeMode = AspectRatioFrameLayout.RESIZE_MODE_FIT
                    },
                    modifier = Modifier.fillMaxSize(),
                )

                if (showPlaybackLoading) {
                    TvShortFeedLoadingOverlay(
                        modifier = Modifier.align(Alignment.Center),
                    )
                }

                if (uiState.loadingMore) {
                    TvShortFeedLoadingOverlay(
                        modifier = Modifier
                            .align(Alignment.BottomCenter)
                            .navigationBarsPadding()
                            .padding(bottom = 28.dp),
                        indicatorSize = 18.dp,
                    )
                }

                AnimatedVisibility(
                    visible = showCenterIndicator,
                    enter = fadeIn(),
                    exit = fadeOut(),
                    modifier = Modifier.align(Alignment.Center),
                ) {
                    Surface(
                        color = AppChrome.Surface.copy(alpha = 0.82f),
                        shape = CircleShape,
                    ) {
                        Row(
                            modifier = Modifier.padding(horizontal = 18.dp, vertical = 14.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(8.dp),
                        ) {
                            Icon(
                                imageVector = if (centerIndicatorIsPause) Icons.Filled.Pause else Icons.Filled.PlayArrow,
                                contentDescription = null,
                                tint = Color.White,
                            )
                            Text(
                                text = if (centerIndicatorIsPause) "已暂停" else "继续播放",
                                color = Color.White,
                                style = MaterialTheme.typography.bodyMedium,
                            )
                        }
                    }
                }

                AnimatedVisibility(
                    visible = showSeekOverlay && seekOverlayDurationMs > 0L,
                    enter = fadeIn(),
                    exit = fadeOut(),
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .navigationBarsPadding()
                        .padding(bottom = 24.dp),
                ) {
                    TvShortFeedBottomProgressBar(
                        positionMs = seekOverlayPositionMs,
                        durationMs = seekOverlayDurationMs,
                    )
                }

                if (!playbackErrorMessage.isNullOrBlank()) {
                    Box(
                        modifier = Modifier
                            .fillMaxSize()
                            .background(Color.Black.copy(alpha = 0.56f)),
                    ) {
                        TvShortFeedProblemState(
                            title = "播放失败",
                            message = playbackErrorMessage.orEmpty(),
                            primaryActionLabel = "重试",
                            onPrimaryAction = {
                                playbackErrorMessage = null
                                renderedVideoId = null
                                hasEndedAtCurrentVideo = false
                                playbackRetryNonce += 1
                            },
                            secondaryActionLabel = "返回首页",
                            onSecondaryAction = onBack,
                            modifier = Modifier.align(Alignment.Center),
                        )
                    }
                }
            }
        }
    }
}

@Composable
private fun TvShortFeedLoadingState() {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black),
        contentAlignment = Alignment.Center,
    ) {
        TvShortFeedLoadingOverlay()
    }
}

@Composable
private fun TvShortFeedLoadingOverlay(
    modifier: Modifier = Modifier,
    indicatorSize: androidx.compose.ui.unit.Dp = 28.dp,
) {
    CircularProgressIndicator(
        modifier = modifier.size(indicatorSize),
        color = AppChrome.Accent,
        strokeWidth = 2.5.dp,
    )
}

@Composable
private fun TvShortFeedProblemState(
    title: String,
    message: String,
    primaryActionLabel: String,
    onPrimaryAction: () -> Unit,
    secondaryActionLabel: String? = null,
    onSecondaryAction: (() -> Unit)? = null,
    modifier: Modifier = Modifier,
) {
    val primaryFocusRequester = remember { FocusRequester() }

    LaunchedTvInitialFocus(title, message, primaryActionLabel, secondaryActionLabel) {
        primaryFocusRequester.tryRequestFocus()
    }

    Surface(
        modifier = modifier.widthIn(max = 560.dp),
        color = AppChrome.SurfaceElevated,
        shape = AppChrome.SurfaceShape,
    ) {
        Column(
            modifier = Modifier.padding(horizontal = 28.dp, vertical = 24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Text(
                text = title,
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.titleMedium,
            )
            Text(
                text = message,
                color = AppChrome.TextSecondary,
                style = MaterialTheme.typography.bodyMedium,
            )
            Row(
                horizontalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                TvShortFeedStateButton(
                    text = primaryActionLabel,
                    icon = if (primaryActionLabel == "重试") Icons.Filled.Refresh else Icons.Filled.Tv,
                    onClick = onPrimaryAction,
                    modifier = Modifier.focusRequester(primaryFocusRequester),
                )
                if (secondaryActionLabel != null && onSecondaryAction != null) {
                    TvShortFeedStateButton(
                        text = secondaryActionLabel,
                        icon = Icons.Filled.Tv,
                        onClick = onSecondaryAction,
                    )
                }
            }
        }
    }
}

@Composable
private fun TvShortFeedStateButton(
    text: String,
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Surface(
        color = AppChrome.AccentSoft,
        shape = AppChrome.SurfaceShape,
        modifier = modifier
            .tvFocusableGlow(shape = AppChrome.SurfaceShape, focusedScale = 1.04f)
            .clickable(onClick = onClick),
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Icon(
                imageVector = icon,
                contentDescription = null,
                tint = AppChrome.TextPrimary,
                modifier = Modifier.size(20.dp),
            )
            Text(
                text = text,
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.bodyMedium,
            )
        }
    }
}

@Composable
private fun TvShortFeedBottomProgressBar(
    positionMs: Long,
    durationMs: Long,
) {
    val progress = if (durationMs <= 0L) {
        0f
    } else {
        (positionMs.toFloat() / durationMs.toFloat()).coerceIn(0f, 1f)
    }

    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Surface(
            color = AppChrome.Surface.copy(alpha = 0.92f),
            shape = AppChrome.SurfaceShape,
        ) {
            Text(
                text = "${formatTvShortPlaybackTime(positionMs)} / ${formatTvShortPlaybackTime(durationMs)}",
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.labelMedium,
                modifier = Modifier.padding(horizontal = 12.dp, vertical = 7.dp),
            )
        }
        Box(
            modifier = Modifier
                .widthIn(min = 320.dp, max = 560.dp)
                .fillMaxWidth()
                .height(6.dp)
                .clip(CircleShape)
                .background(AppChrome.TextPrimary.copy(alpha = 0.18f)),
        ) {
            Box(
                modifier = Modifier
                    .fillMaxWidth(progress)
                    .fillMaxSize()
                    .clip(CircleShape)
                    .background(AppChrome.AccentStrong, CircleShape),
            )
        }
    }
}

private fun seekCurrentShortVideo(
    player: ExoPlayer,
    stepSeconds: Int,
    repeatCount: Int,
    forward: Boolean,
    onSeekApplied: (targetMs: Long, durationMs: Long) -> Unit,
) {
    val durationMs = player.duration.takeIf { it > 0L } ?: return
    val currentPositionMs = player.currentPosition.coerceAtLeast(0L)
    val stepMs = TvPlaybackSeekStepSetting.normalize(stepSeconds) * 1_000L
    val deltaMs = stepMs * if (repeatCount > 0) 3L else 1L
    val targetMs = if (forward) {
        (currentPositionMs + deltaMs).coerceAtMost(durationMs)
    } else {
        (currentPositionMs - deltaMs).coerceAtLeast(0L)
    }
    player.seekTo(targetMs)
    onSeekApplied(targetMs, durationMs)
}

private fun readTvShortWatchSnapshot(player: ExoPlayer): Pair<Int, Boolean> {
    val watchSeconds = (player.currentPosition.coerceAtLeast(0L) / 1_000L).toInt()
    val durationMs = player.duration
    val durationSeconds = if (durationMs > 0L) (durationMs / 1_000L).toInt() else 0
    val completedThreshold = (durationSeconds - 3).coerceAtLeast(1)
    val completed = durationSeconds > 0 && watchSeconds >= completedThreshold
    return watchSeconds to completed
}

private fun formatTvShortPlaybackTime(ms: Long): String {
    val totalSeconds = (ms.coerceAtLeast(0L) / 1_000L)
    val hours = totalSeconds / 3_600L
    val minutes = (totalSeconds % 3_600L) / 60L
    val seconds = totalSeconds % 60L
    return buildString {
        append(hours.toString().padStart(2, '0'))
        append(':')
        append(minutes.toString().padStart(2, '0'))
        append(':')
        append(seconds.toString().padStart(2, '0'))
    }
}

private const val TvShortCenterIndicatorDurationMillis = 700L
private const val TvShortSeekOverlayDurationMillis = 1_200L
