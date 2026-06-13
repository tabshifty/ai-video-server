package com.chee.videos.feature.tv

import android.app.Activity
import android.os.SystemClock
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.runtime.withFrameNanos
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
import com.chee.videos.core.model.TvTrackPreference
import com.chee.videos.core.player.friendlyLongFormPlaybackErrorMessage
import com.chee.videos.core.player.TvVlcLibrary
import com.chee.videos.core.player.newLongFormMediaPlayer
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.ui.LongFormVideoPlayer
import com.chee.videos.core.ui.TvErrorState
import com.chee.videos.core.ui.TvPageLoadingState
import com.chee.videos.core.ui.appendAccessTokenQuery
import com.chee.videos.core.ui.applyLongFormMediaSource
import com.chee.videos.core.ui.resolveInitialSubtitleTrackId
import com.chee.videos.core.ui.buildSubtitleTrackPreference
import com.chee.videos.core.ui.resolveLongFormPlayerUpdate
import com.chee.videos.core.ui.resolvePlaybackAssetUrl
import com.chee.videos.core.ui.resolveSelectedSubtitleTrack
import com.chee.videos.core.ui.resolveSelectedSubtitleTrackByPreference
import com.chee.videos.core.ui.resolveSubtitleSelectionOnTrackLoad
import com.chee.videos.feature.detail.DetailViewModel
import com.chee.videos.feature.detail.LongFormPlaybackSession
import kotlinx.coroutines.delay
import org.videolan.libvlc.MediaPlayer
import org.videolan.libvlc.interfaces.IMedia

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
        ) {
            TvPageLoadingState(message = "正在加载播放器")
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
        ) {
            TvErrorState(
                message = uiState.errorMessage ?: "播放器数据不存在",
                onAction = viewModel::load,
            )
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
    val playbackCompatibilityDecision = resolveTvPlaybackCompatibilityDecision(detail.metadata)
    val playerBlockMessage = playbackCompatibilityDecision.blockMessage
    val playableUrl = playUrl?.takeIf { playbackCompatibilityDecision.allowed }
    val canPlay = !playableUrl.isNullOrBlank()
    val libVLC = remember { TvVlcLibrary.shared(context) }
    val mediaPlayer = remember(uiState.accessToken) { newLongFormMediaPlayer(libVLC) }
    val latestDetailId by rememberUpdatedState(detail.id)
    var hasStartedPlayback by rememberSaveable(detail.id) { mutableStateOf(true) }
    var isPausedByUser by rememberSaveable(detail.id) { mutableStateOf(false) }
    var preparedUrl by remember(detail.id, uiState.accessToken) { mutableStateOf<String?>(null) }
    var appliedSubtitleSlaveUrl by remember(detail.id, uiState.accessToken) { mutableStateOf<String?>(null) }
    var isVlcPlaying by remember(detail.id, uiState.accessToken) { mutableStateOf(false) }
    var resumedFromHistoryVideoId by remember(detail.id, uiState.accessToken) { mutableStateOf("") }
    var resumePromptLastPositionMs by remember(detail.id, uiState.accessToken) { mutableStateOf(0L) }
    var resumePromptRemainingMs by remember(detail.id, uiState.accessToken) { mutableStateOf(0L) }
    var resumePromptDismissed by remember(detail.id, uiState.accessToken) { mutableStateOf(false) }
    var isTrackSheetVisible by remember(detail.id) { mutableStateOf(false) }
    var selectedSubtitleTrackId by rememberSaveable(detail.id) { mutableStateOf<String?>(null) }
    var storedSubtitlePreference by remember(detail.id) { mutableStateOf<TvTrackPreference?>(null) }
    var storedAudioPreference by remember(detail.id) { mutableStateOf<TvTrackPreference?>(null) }
    var selectedAudioTrackId by rememberSaveable(detail.id) { mutableStateOf<String?>(null) }
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

    LaunchedEffect(detail.id) {
        storedSubtitlePreference = viewModel.readTvSubtitlePreference(detail.id)
        storedAudioPreference = viewModel.readTvAudioPreference(detail.id)
    }

    LaunchedEffect(detail.id, detail.subtitleTracks, hasStartedPlayback, storedSubtitlePreference) {
        selectedSubtitleTrackId = resolveSelectedSubtitleTrackByPreference(
            tracks = detail.subtitleTracks,
            preference = storedSubtitlePreference,
        )?.id ?: resolveSubtitleSelectionOnTrackLoad(
            currentSelection = selectedSubtitleTrackId,
            tracks = detail.subtitleTracks,
            hasStartedPlayback = hasStartedPlayback,
        ) ?: resolveInitialSubtitleTrackId(detail.subtitleTracks)
    }

    LaunchedEffect(playableUrl, playbackSession.hasStartedPlayback) {
        val updateDecision = resolveLongFormPlayerUpdate(
            preparedUrl = preparedUrl,
            nextUrl = playableUrl,
        )
        if (!playbackSession.hasStartedPlayback || playableUrl.isNullOrBlank()) {
            mediaPlayer.pause()
            if (updateDecision.shouldClear) {
                mediaPlayer.stop()
            }
            preparedUrl = null
            appliedSubtitleSlaveUrl = null
            isVlcPlaying = false
            return@LaunchedEffect
        }
        if (updateDecision.shouldReplaceSource) {
            playerErrorMessage = null
            val initialResumePositionMs = if (resumedFromHistoryVideoId != detail.id) {
                detail.userState.watchSeconds.coerceAtLeast(0).times(1000L)
            } else {
                0L
            }
            appliedSubtitleSlaveUrl = null
            isVlcPlaying = false
            applyLongFormMediaSource(
                libVLC = libVLC,
                mediaPlayer = mediaPlayer,
                sourceUrl = playableUrl,
                accessToken = uiState.accessToken,
            )
            mediaPlayer.play()
            if (initialResumePositionMs > 0L) {
                delay(250L)
                mediaPlayer.time = initialResumePositionMs
            }
            if (detail.id.isNotBlank()) {
                resumedFromHistoryVideoId = detail.id
                if (shouldTriggerResumePrompt(initialResumePositionMs)) {
                    resumePromptLastPositionMs = initialResumePositionMs
                    resumePromptRemainingMs = TvResumePromptTokens.CountdownDurationMs
                    resumePromptDismissed = false
                }
            }
            preparedUrl = playableUrl
        }
        if (!playbackSession.isPausedByUser) {
            mediaPlayer.play()
        }
    }

    LaunchedEffect(isVlcPlaying, selectedSubtitleTrackId, detail.subtitleTracks, uiState.baseUrl, uiState.accessToken) {
        if (!isVlcPlaying) return@LaunchedEffect
        val track = resolveSelectedSubtitleTrack(detail.subtitleTracks, selectedSubtitleTrackId)
        val subtitleUrl = track
            ?.takeIf { it.available && it.url.isNotBlank() }
            ?.let { resolvePlaybackAssetUrl(uiState.baseUrl, it.url) }
            ?.let { appendAccessTokenQuery(it, uiState.accessToken) }
        if (subtitleUrl.isNullOrBlank()) {
            appliedSubtitleSlaveUrl = null
            return@LaunchedEffect
        }
        if (subtitleUrl == appliedSubtitleSlaveUrl) {
            return@LaunchedEffect
        }
        // 走 Uri 重载：LibVLC 的 String 重载会把 "http://..." 当文件路径，规范化成 "/http:/..." 然后报
        // "No such file or directory"。Uri 重载经 VLCUtil.locationFromUri 转成正确的 MRL。
        mediaPlayer.addSlave(IMedia.Slave.Type.Subtitle, android.net.Uri.parse(subtitleUrl), true)
        appliedSubtitleSlaveUrl = subtitleUrl
    }

    LaunchedEffect(playbackSession.hasStartedPlayback, playbackSession.isPausedByUser, playableUrl) {
        if (!playbackSession.hasStartedPlayback || playableUrl.isNullOrBlank()) {
            mediaPlayer.pause()
            return@LaunchedEffect
        }
        if (playbackSession.isPausedByUser) {
            mediaPlayer.pause()
            return@LaunchedEffect
        }
        mediaPlayer.play()
    }

    LaunchedEffect(
        detail.id,
        canPlay,
        playbackSession.hasStartedPlayback,
        playbackSession.isPausedByUser,
    ) {
        if (!shouldStartPeriodicHistoryReport(
                videoId = detail.id,
                canPlay = canPlay,
                hasStartedPlayback = playbackSession.hasStartedPlayback,
                isPausedByUser = playbackSession.isPausedByUser,
            )
        ) {
            return@LaunchedEffect
        }
        while (true) {
            delay(TvPeriodicHistoryReportIntervalMillis)
            reportTvLongFormHistory(viewModel, detail.id, mediaPlayer)
        }
    }

    LaunchedEffect(playerErrorMessage, showBackConfirmPrompt, isTrackSheetVisible) {
        if (playerErrorMessage != null || showBackConfirmPrompt || isTrackSheetVisible) {
            resumePromptDismissed = true
        }
    }

    val resumePromptGuardInput = ResumePromptGuardInput(
        hasResumeSeekTriggered = resumedFromHistoryVideoId == detail.id && resumePromptLastPositionMs > 0L,
        promptPermanentlyDismissed = resumePromptDismissed,
        isPlayerError = playerErrorMessage != null,
        isBackConfirmVisible = showBackConfirmPrompt,
        isEpisodeSelectorVisible = false,
        isTrackSheetVisible = isTrackSheetVisible,
        isEndOverlayVisible = false,
        isAutoplayPromptVisible = false,
        isPausedByUser = playbackSession.isPausedByUser,
        remainingMs = resumePromptRemainingMs,
    )
    val shouldTickResumePromptCountdown = shouldTickResumePromptCountdown(resumePromptGuardInput)
    val shouldShowResumePromptCard = shouldShowResumePromptCard(resumePromptGuardInput)

    LaunchedEffect(detail.id, shouldTickResumePromptCountdown, resumePromptDismissed) {
        if (resumePromptDismissed || !shouldTickResumePromptCountdown) return@LaunchedEffect
        val startNanos = withFrameNanos { it }
        val initialRemainingMs = resumePromptRemainingMs
        while (resumePromptRemainingMs > 0L) {
            val nowNanos = withFrameNanos { it }
            val elapsedMs = (nowNanos - startNanos) / 1_000_000L
            val next = (initialRemainingMs - elapsedMs).coerceAtLeast(0L)
            if (next != resumePromptRemainingMs) {
                resumePromptRemainingMs = next
            }
            if (next <= 0L) break
        }
        if (resumePromptRemainingMs <= 0L && resumePromptLastPositionMs > 0L) {
            resumePromptDismissed = true
        }
    }

    DisposableEffect(lifecycleOwner, mediaPlayer, playbackSession.hasStartedPlayback, playbackSession.isPausedByUser, canPlay) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_PAUSE -> {
                    reportTvLongFormHistory(viewModel, latestDetailId, mediaPlayer)
                    mediaPlayer.pause()
                }

                Lifecycle.Event.ON_RESUME -> {
                    if (playbackSession.shouldResumeOnLifecycle() && canPlay) {
                        mediaPlayer.play()
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

    DisposableEffect(mediaPlayer) {
        onDispose {
            reportTvLongFormHistory(viewModel, latestDetailId, mediaPlayer)
            mediaPlayer.release()
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
                player = mediaPlayer,
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
                onSelectSubtitleTrack = { trackId ->
                    selectedSubtitleTrackId = trackId
                    val preference = buildSubtitleTrackPreference(
                        detail.subtitleTracks.firstOrNull { it.id == trackId },
                    )
                    storedSubtitlePreference = preference
                    viewModel.saveTvSubtitlePreference(detail.id, preference)
                },
                selectedAudioTrackId = selectedAudioTrackId,
                selectedAudioPreference = storedAudioPreference,
                onSelectAudioTrack = { trackId, preference, isUserAction ->
                    selectedAudioTrackId = trackId ?: ""
                    storedAudioPreference = preference
                    if (isUserAction) {
                        viewModel.saveTvAudioPreference(detail.id, preference)
                    }
                },
                tvMode = true,
                tvSeekStepSeconds = uiState.tvSeekStepSeconds,
                onRequestExitPlayback = ::handlePlaybackBack,
                onExitPlayback = onBack,
                onTrackSheetVisibilityChanged = { isTrackSheetVisible = it },
                onVlcEvent = { event ->
                    when (event.type) {
                        MediaPlayer.Event.Playing -> {
                            isPlayerActuallyPlaying = true
                            playerErrorMessage = null
                            isVlcPlaying = true
                        }
                        MediaPlayer.Event.Paused -> {
                            isPlayerActuallyPlaying = false
                        }
                        MediaPlayer.Event.Stopped,
                        MediaPlayer.Event.EndReached -> {
                            isPlayerActuallyPlaying = false
                            isVlcPlaying = false
                        }
                        MediaPlayer.Event.EncounteredError -> {
                            playerErrorMessage = friendlyLongFormPlaybackErrorMessage(null)
                            isPlayerActuallyPlaying = false
                            isVlcPlaying = false
                        }
                    }
                },
                resumePromptVisible = shouldShowResumePromptCard,
                resumePromptSlot = {
                    TvResumePromptCard(
                        lastPositionMs = resumePromptLastPositionMs,
                        visible = shouldShowResumePromptCard,
                        remainingSeconds = resumePromptCountdownTickRemaining(resumePromptRemainingMs),
                        onContinue = { resumePromptDismissed = true },
                        onStartFromBeginning = {
                            mediaPlayer.time = 0L
                            resumePromptDismissed = true
                        },
                        modifier = Modifier
                            .align(Alignment.BottomStart)
                            .padding(
                                start = TvResumePromptTokens.HorizontalPaddingDp,
                                bottom = TvResumePromptTokens.BottomPaddingDp,
                            ),
                    )
                },
                backConfirmPromptVisible = showBackConfirmPrompt,
                playerErrorVisible = !playerErrorMessage.isNullOrBlank(),
            )
            if (!playerErrorMessage.isNullOrBlank()) {
                Surface(
                    color = Color(0xCC2B0F12),
                    shape = AppChrome.SurfaceShape,
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
            TvErrorState(
                title = "暂不能播放",
                message = playerBlockMessage ?: "暂无可播放视频",
                onAction = viewModel::load,
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

private fun reportTvLongFormHistory(
    viewModel: DetailViewModel,
    videoId: String,
    player: MediaPlayer,
) {
    val snapshot = tvPlaybackHistorySnapshot(
        positionMs = player.time,
        durationMs = player.length,
    )
    viewModel.reportHistory(videoId, snapshot.watchSeconds, snapshot.completed)
}
