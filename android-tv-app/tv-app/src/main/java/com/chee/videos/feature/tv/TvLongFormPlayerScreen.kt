package com.chee.videos.feature.tv

import android.app.Activity
import android.os.SystemClock
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.unit.dp
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.exoplayer.source.DefaultMediaSourceFactory
import com.chee.videos.core.player.friendlyLongFormPlaybackErrorMessage
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.ui.LongFormVideoPlayer
import com.chee.videos.core.ui.buildLongFormMediaItem
import com.chee.videos.core.ui.resolveLongFormPlayerUpdate
import com.chee.videos.core.ui.resolveSelectedSubtitleTrack
import com.chee.videos.core.ui.resolveSubtitleSelectionOnTrackLoad
import com.chee.videos.feature.detail.DetailViewModel
import com.chee.videos.feature.detail.LongFormPlaybackSession
import kotlinx.coroutines.delay

@Composable
fun TvLongFormPlayerScreen(
    onBack: () -> Unit,
    viewModel: DetailViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val context = LocalContext.current
    val activity = context as? Activity
    val lifecycleOwner = LocalLifecycleOwner.current
    var backPromptAtMillis by remember { mutableStateOf<Long?>(null) }
    var showBackConfirmPrompt by remember { mutableStateOf(false) }

    fun handlePlaybackBack() {
        val now = SystemClock.uptimeMillis()
        when (resolveTvPlayerBackAction(backPromptAtMillis, now)) {
            TvPlayerBackAction.ShowPrompt -> {
                backPromptAtMillis = now
                showBackConfirmPrompt = true
            }

            TvPlayerBackAction.Exit -> {
                backPromptAtMillis = null
                showBackConfirmPrompt = false
                onBack()
            }
        }
    }

    BackHandler(onBack = ::handlePlaybackBack)

    LaunchedEffect(showBackConfirmPrompt, backPromptAtMillis) {
        val promptAt = backPromptAtMillis
        if (showBackConfirmPrompt && promptAt != null) {
            delay(TvPlayerBackConfirmWindowMillis)
            if (backPromptAtMillis == promptAt) {
                showBackConfirmPrompt = false
            }
        }
    }

    if (uiState.loading) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black),
            contentAlignment = Alignment.Center,
        ) {
            CircularProgressIndicator(color = AppChrome.AccentStrong)
            if (showBackConfirmPrompt) {
                TvPlayerBackConfirmPrompt(
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .padding(bottom = 48.dp),
                )
            }
        }
        return
    }

    val detail = uiState.detail
    if (detail == null) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black),
            contentAlignment = Alignment.Center,
        ) {
            Text(uiState.errorMessage ?: "播放器数据不存在", color = AppChrome.TextSecondary)
            if (showBackConfirmPrompt) {
                TvPlayerBackConfirmPrompt(
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .padding(bottom = 48.dp),
                )
            }
        }
        return
    }

    val playUrl = resolveTvLongFormPlayUrl(
        baseUrl = uiState.baseUrl,
        detail = detail,
        preferredPlaybackProfile = uiState.preferredPlaybackProfile,
    )
    val canPlay = !playUrl.isNullOrBlank()
    val dataSourceFactory = remember(uiState.accessToken) {
        DefaultHttpDataSource.Factory().setAllowCrossProtocolRedirects(true).apply {
            if (uiState.accessToken.isNotBlank()) {
                setDefaultRequestProperties(mapOf("Authorization" to "Bearer ${uiState.accessToken}"))
            }
        }
    }
    val exoPlayer = remember(uiState.accessToken) { ExoPlayer.Builder(context).build() }
    var hasStartedPlayback by rememberSaveable(detail.id) { mutableStateOf(true) }
    var isPausedByUser by rememberSaveable(detail.id) { mutableStateOf(false) }
    var preparedUrl by remember(detail.id, uiState.accessToken) { mutableStateOf<String?>(null) }
    var preparedSubtitleTrackId by remember(detail.id, uiState.accessToken) { mutableStateOf<String?>(null) }
    var selectedSubtitleTrackId by rememberSaveable(detail.id) { mutableStateOf<String?>(null) }
    var isPlayerActuallyPlaying by remember(detail.id, uiState.accessToken) { mutableStateOf(false) }
    var playerErrorMessage by remember(detail.id, uiState.accessToken) { mutableStateOf<String?>(null) }

    val playbackSession = remember(hasStartedPlayback, isPausedByUser) {
        LongFormPlaybackSession(
            hasStartedPlayback = hasStartedPlayback,
            isPausedByUser = isPausedByUser,
        )
    }

    fun updatePlaybackSession(nextSession: LongFormPlaybackSession) {
        hasStartedPlayback = nextSession.hasStartedPlayback
        isPausedByUser = nextSession.isPausedByUser
    }

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

    LaunchedEffect(detail.id, detail.subtitleTracks, hasStartedPlayback) {
        selectedSubtitleTrackId = resolveSubtitleSelectionOnTrackLoad(
            currentSelection = selectedSubtitleTrackId,
            tracks = detail.subtitleTracks,
            hasStartedPlayback = hasStartedPlayback,
        )
    }

    LaunchedEffect(playUrl, dataSourceFactory, selectedSubtitleTrackId, playbackSession.hasStartedPlayback) {
        val updateDecision = resolveLongFormPlayerUpdate(
            preparedUrl = preparedUrl,
            nextUrl = playUrl,
            preparedSubtitleTrackId = preparedSubtitleTrackId,
            nextSubtitleTrackId = selectedSubtitleTrackId,
        )
        if (!playbackSession.hasStartedPlayback || playUrl.isNullOrBlank()) {
            return@LaunchedEffect
        }
        if (updateDecision.shouldReplaceSource) {
            playerErrorMessage = null
            val mediaItem = buildLongFormMediaItem(
                sourceUrl = playUrl,
                mediaId = detail.id,
                title = detail.title,
                baseUrl = uiState.baseUrl,
                selectedSubtitleTrack = resolveSelectedSubtitleTrack(detail.subtitleTracks, selectedSubtitleTrackId),
            )
            val mediaSource = DefaultMediaSourceFactory(dataSourceFactory).createMediaSource(mediaItem)
            exoPlayer.setMediaSource(mediaSource, true)
            exoPlayer.prepare()
            preparedUrl = playUrl
            preparedSubtitleTrackId = selectedSubtitleTrackId
        }
        if (!playbackSession.isPausedByUser) {
            exoPlayer.playWhenReady = true
            exoPlayer.play()
        }
    }

    LaunchedEffect(playbackSession.hasStartedPlayback, playbackSession.isPausedByUser, playUrl) {
        if (!playbackSession.hasStartedPlayback || playUrl.isNullOrBlank()) {
            exoPlayer.pause()
            return@LaunchedEffect
        }
        if (playbackSession.isPausedByUser) {
            exoPlayer.pause()
            return@LaunchedEffect
        }
        exoPlayer.playWhenReady = true
        exoPlayer.play()
    }

    DisposableEffect(lifecycleOwner, exoPlayer, playbackSession.hasStartedPlayback, playbackSession.isPausedByUser) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_PAUSE -> exoPlayer.pause()
                Lifecycle.Event.ON_RESUME -> {
                    if (playbackSession.shouldResumeOnLifecycle()) {
                        exoPlayer.playWhenReady = true
                        exoPlayer.play()
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

    DisposableEffect(exoPlayer) {
        val listener = object : androidx.media3.common.Player.Listener {
            override fun onIsPlayingChanged(isPlaying: Boolean) {
                isPlayerActuallyPlaying = isPlaying
            }

            override fun onRenderedFirstFrame() {
                playerErrorMessage = null
            }

            override fun onPlayerError(error: androidx.media3.common.PlaybackException) {
                playerErrorMessage = friendlyLongFormPlaybackErrorMessage(error)
                isPlayerActuallyPlaying = false
            }
        }
        exoPlayer.addListener(listener)
        onDispose {
            exoPlayer.removeListener(listener)
            exoPlayer.release()
        }
    }

    KeepScreenOnEffect(enabled = isPlayerActuallyPlaying)

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black),
    ) {
        if (canPlay) {
            LongFormVideoPlayer(
                title = detail.title,
                player = exoPlayer,
                isFullscreen = true,
                onBack = onBack,
                onTogglePlayPause = {
                    updatePlaybackSession(playbackSession.togglePlayPause(canPlay = canPlay))
                },
                onToggleFullscreen = {},
                modifier = Modifier.fillMaxSize(),
                showStatusBarPadding = false,
                subtitleTracks = detail.subtitleTracks,
                selectedSubtitleTrackId = selectedSubtitleTrackId,
                onSelectSubtitleTrack = { selectedSubtitleTrackId = it },
                tvMode = true,
                onExitPlayback = onBack,
            )
            if (!playerErrorMessage.isNullOrBlank()) {
                Surface(
                    color = Color(0xCC2B0F12),
                    shape = RoundedCornerShape(14.dp),
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .padding(14.dp),
                ) {
                    Text(
                        text = playerErrorMessage.orEmpty(),
                        color = Color.White,
                        style = androidx.compose.material3.MaterialTheme.typography.bodySmall,
                        modifier = Modifier.padding(horizontal = 12.dp, vertical = 10.dp),
                    )
                }
            }
            if (showBackConfirmPrompt) {
                TvPlayerBackConfirmPrompt(
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .padding(bottom = 72.dp),
                )
            }
        } else {
            Text(
                text = "暂无可播放视频",
                color = AppChrome.TextSecondary,
                modifier = Modifier.align(Alignment.Center),
            )
            if (showBackConfirmPrompt) {
                TvPlayerBackConfirmPrompt(
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .padding(bottom = 48.dp),
                )
            }
        }
    }
}
