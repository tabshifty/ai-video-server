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
import com.chee.videos.core.ui.LongFormAudioTrack
import com.chee.videos.core.ui.LongFormVideoPlayer
import com.chee.videos.core.ui.TvErrorState
import com.chee.videos.core.ui.TvPageLoadingState
import com.chee.videos.core.ui.TvSeriesCorePlaybackOverlay
import com.chee.videos.core.ui.appendAccessTokenQuery
import com.chee.videos.core.ui.applyLongFormMediaSource
import com.chee.videos.core.ui.resolveInitialSubtitleTrackId
import com.chee.videos.core.ui.buildSubtitleTrackPreference
import com.chee.videos.core.ui.buildAudioTrackPreference
import com.chee.videos.core.ui.resolveLongFormPlayerUpdate
import com.chee.videos.core.ui.resolvePlaybackAssetUrl
import com.chee.videos.core.ui.resolveAudioSelectionOnTrackLoad
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
    var showDolbyVisionDiagnostics by remember { mutableStateOf(false) }
    var pendingDvExitToDetail by remember { mutableStateOf(false) }

    fun exitToDetailWithDvCover(isMedia3DolbyVisionRoute: Boolean) {
        if (!isMedia3DolbyVisionRoute) {
            onBack()
            return
        }
        pendingDvExitToDetail = true
    }

    fun handlePlaybackBack(isMedia3DolbyVisionRoute: Boolean) {
        val now = SystemClock.uptimeMillis()
        when (resolveTvPlayerBackAction(backPromptAtMillis, now)) {
            TvPlayerBackAction.ShowPrompt -> {
                backPromptAtMillis = now
                showBackConfirmPrompt = true
            }

            TvPlayerBackAction.Exit -> {
                backPromptAtMillis = null
                showBackConfirmPrompt = false
                exitToDetailWithDvCover(isMedia3DolbyVisionRoute)
            }
        }
    }

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
        BackHandler { handlePlaybackBack(isMedia3DolbyVisionRoute = false) }
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
        BackHandler { handlePlaybackBack(isMedia3DolbyVisionRoute = false) }
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

    LaunchedEffect(detail.id) {
        showDolbyVisionDiagnostics = false
    }

    var routeRetryNonce by remember(detail.id) { mutableStateOf(0) }
    val displayCapability = remember(context, detail.id, routeRetryNonce) {
        evaluateDolbyVisionDisplayCapability(AndroidDisplayHdrCapabilityReader(context))
    }
    val playbackRoute = remember(detail.metadata, displayCapability, uiState.baseUrl, uiState.preferredPlaybackProfile) {
        resolveTvPlaybackRoute(
            metadata = detail.metadata,
            displayCapability = displayCapability,
            playbackUrl = resolveTvLongFormPlayUrl(
                baseUrl = uiState.baseUrl,
                detail = detail,
                preferredPlaybackProfile = uiState.preferredPlaybackProfile,
                overridePlaybackProfile = null,
            ),
            media3Available = true,
        )
    }
    val playUrl = remember(uiState.baseUrl, detail, uiState.preferredPlaybackProfile, playbackRoute.playbackProfile) {
        resolveTvLongFormPlayUrl(
            baseUrl = uiState.baseUrl,
            detail = detail,
            preferredPlaybackProfile = uiState.preferredPlaybackProfile,
            overridePlaybackProfile = playbackRoute.playbackProfile,
        )
    }
    val showDolbyVisionDiagnosticsButton = remember(detail.metadata, playbackRoute) {
        isTvDolbyVisionDiagnosticsAvailable(playbackRoute)
    }
    val playerBlockMessage = playbackRoute.blockMessage
    val playableUrl = playUrl?.takeIf { playbackRoute.kind != TvPlaybackRouteKind.BLOCKED }
    val isVlcRoute = playbackRoute.kind == TvPlaybackRouteKind.VLC && !playableUrl.isNullOrBlank()
    val isMedia3Route = playbackRoute.kind == TvPlaybackRouteKind.MEDIA3_DOLBY_VISION && !playableUrl.isNullOrBlank()
    val canPlay = isVlcRoute || isMedia3Route
    val canPlayVlc = isVlcRoute
    val libVLC = remember { TvVlcLibrary.shared(context) }
    val mediaPlayer = remember(uiState.accessToken) { newLongFormMediaPlayer(libVLC) }
    val latestDetailId by rememberUpdatedState(detail.id)
    var media3Snapshot by remember(detail.id) { mutableStateOf(TvMedia3PlaybackSnapshot(positionMs = 0L, durationMs = 0L)) }
    val latestMedia3Snapshot by rememberUpdatedState(media3Snapshot)
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
    var media3AudioTracks by remember(detail.id) { mutableStateOf(emptyList<LongFormAudioTrack>()) }
    var media3TrackPickerKind by remember(detail.id) { mutableStateOf<TvMedia3TrackPickerKind?>(null) }
    var media3SeekPositionMs by remember(detail.id) { mutableStateOf<Long?>(null) }
    var media3SeekRequestKey by remember(detail.id) { mutableStateOf(0) }
    var isPlayerActuallyPlaying by remember(detail.id, uiState.accessToken) { mutableStateOf(false) }
    var playerErrorMessage by remember(detail.id, uiState.accessToken) { mutableStateOf<String?>(null) }
    var screenPositionMs by remember(detail.id) { mutableStateOf(0L) }
    var screenDurationMs by remember(detail.id) { mutableStateOf(0L) }
    val playbackDiagnosticMessage = remember(playbackRoute, displayCapability, playUrl, playerErrorMessage) {
        buildTvDolbyVisionDiagnosticMessage(
            route = playbackRoute,
            displayCapability = displayCapability,
            playbackUrl = playUrl,
            media3Available = true,
            failureMessage = playerErrorMessage,
        )
    }

    val playbackSession = remember(hasStartedPlayback, isPausedByUser) {
        LongFormPlaybackSession(
            hasStartedPlayback = hasStartedPlayback,
            isPausedByUser = isPausedByUser,
        )
    }

    BackHandler(enabled = showDolbyVisionDiagnostics) {
        showDolbyVisionDiagnostics = false
    }
    BackHandler(enabled = !showDolbyVisionDiagnostics) {
        handlePlaybackBack(isMedia3DolbyVisionRoute = isMedia3Route)
    }

    LaunchedEffect(pendingDvExitToDetail) {
        if (!pendingDvExitToDetail) {
            return@LaunchedEffect
        }
        delay(TvDolbyVisionExitToDetailCoverDelayMillis)
        onBack()
    }
    val media3SubtitleConfigurations = remember(detail.subtitleTracks, uiState.baseUrl, uiState.accessToken) {
        buildTvMedia3SubtitleConfigurations(
            tracks = detail.subtitleTracks,
            baseUrl = uiState.baseUrl,
            accessToken = uiState.accessToken,
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

    LaunchedEffect(playableUrl, canPlayVlc, playbackSession.hasStartedPlayback) {
        val updateDecision = resolveLongFormPlayerUpdate(
            preparedUrl = preparedUrl,
            nextUrl = playableUrl.takeIf { canPlayVlc },
        )
        if (!canPlayVlc || !playbackSession.hasStartedPlayback || playableUrl.isNullOrBlank()) {
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

    LaunchedEffect(playbackSession.hasStartedPlayback, playbackSession.isPausedByUser, playableUrl, canPlayVlc) {
        if (!canPlayVlc || !playbackSession.hasStartedPlayback || playableUrl.isNullOrBlank()) {
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
        canPlayVlc,
        playbackSession.hasStartedPlayback,
        playbackSession.isPausedByUser,
    ) {
        if (!shouldStartPeriodicHistoryReport(
                videoId = detail.id,
                canPlay = canPlayVlc,
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

    LaunchedEffect(
        detail.id,
        isMedia3Route,
        playbackSession.hasStartedPlayback,
        playbackSession.isPausedByUser,
    ) {
        if (!shouldStartPeriodicHistoryReport(
                videoId = detail.id,
                canPlay = isMedia3Route,
                hasStartedPlayback = playbackSession.hasStartedPlayback,
                isPausedByUser = playbackSession.isPausedByUser,
            )
        ) {
            return@LaunchedEffect
        }
        while (true) {
            delay(TvPeriodicHistoryReportIntervalMillis)
            reportTvLongFormMedia3History(viewModel, detail.id, latestMedia3Snapshot)
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

    DisposableEffect(lifecycleOwner, mediaPlayer, playbackSession.hasStartedPlayback, playbackSession.isPausedByUser, canPlayVlc) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_PAUSE -> {
                    reportTvLongFormHistory(viewModel, latestDetailId, mediaPlayer)
                    mediaPlayer.pause()
                }

                Lifecycle.Event.ON_RESUME -> {
                    if (playbackSession.shouldResumeOnLifecycle() && canPlayVlc) {
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

    DisposableEffect(isMedia3Route, latestDetailId) {
        onDispose {
            if (isMedia3Route) {
                reportTvLongFormMedia3History(viewModel, latestDetailId, latestMedia3Snapshot)
            }
        }
    }

    KeepScreenOnEffect(enabled = isPlayerActuallyPlaying)

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black),
    ) {
        if (isVlcRoute) {
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
                onRequestExitPlayback = { handlePlaybackBack(isMedia3DolbyVisionRoute = false) },
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
        } else if (isMedia3Route && playableUrl != null) {
            TvDolbyVisionMedia3Player(
                sourceUrl = playableUrl,
                mediaId = detail.id,
                title = detail.title,
                accessToken = uiState.accessToken,
                retryKey = routeRetryNonce,
                shouldPlay = playbackSession.hasStartedPlayback && !playbackSession.isPausedByUser,
                initialPositionMs = resolveTvMedia3ResumePositionMs(
                    historyPositionMs = detail.userState.watchSeconds.coerceAtLeast(0).times(1000L),
                    currentSnapshotPositionMs = latestMedia3Snapshot.positionMs,
                    hasCurrentPlaybackSnapshot = resumedFromHistoryVideoId == detail.id,
                ),
                seekPositionMs = media3SeekPositionMs,
                seekRequestKey = media3SeekRequestKey,
                subtitleConfigurations = media3SubtitleConfigurations,
                selectedSubtitleTrackId = selectedSubtitleTrackId?.takeIf { it.isNotBlank() },
                selectedAudioTrackId = selectedAudioTrackId,
                modifier = Modifier.fillMaxSize(),
                onPlayingChanged = { playing ->
                    isPlayerActuallyPlaying = playing
                    if (playing) {
                        playerErrorMessage = null
                        if (detail.id.isNotBlank() && resumedFromHistoryVideoId != detail.id) {
                            val resumePositionMs = detail.userState.watchSeconds.coerceAtLeast(0).times(1000L)
                            resumedFromHistoryVideoId = detail.id
                            if (shouldTriggerResumePrompt(resumePositionMs)) {
                                resumePromptLastPositionMs = resumePositionMs
                                resumePromptRemainingMs = TvResumePromptTokens.CountdownDurationMs
                                resumePromptDismissed = false
                            }
                        }
                    }
                },
                onError = { message ->
                    playerErrorMessage = message
                    isPlayerActuallyPlaying = false
                    updatePlaybackSession(playbackSession.copy(hasStartedPlayback = false))
                },
                onSnapshotChanged = { snapshot ->
                    media3Snapshot = snapshot
                    screenPositionMs = snapshot.positionMs
                    screenDurationMs = snapshot.durationMs
                },
                onAudioTracksChanged = { tracks ->
                    media3AudioTracks = tracks
                    val resolvedSelection = resolveAudioSelectionOnTrackLoad(
                        currentSelection = selectedAudioTrackId,
                        storedPreference = storedAudioPreference,
                        tracks = tracks,
                    )
                    if (resolvedSelection != selectedAudioTrackId?.takeIf { it.isNotBlank() }) {
                        selectedAudioTrackId = resolvedSelection ?: ""
                    }
                },
            )
            TvSeriesCorePlaybackOverlay(
                title = detail.title,
                isPlaying = playbackSession.hasStartedPlayback && !playbackSession.isPausedByUser && isPlayerActuallyPlaying,
                positionMs = screenPositionMs,
                durationMs = screenDurationMs,
                tvSeekStepSeconds = uiState.tvSeekStepSeconds,
                seriesTitleForOverlay = null,
                seasonNumber = null,
                episodeNumber = null,
                episodeTitle = null,
                episodeRailItems = emptyList(),
                currentEpisodeRailItemId = null,
                onTogglePlayPause = {
                    updatePlaybackSession(playbackSession.togglePlayPause(canPlay = canPlay))
                },
                onSeekTo = { targetMs ->
                    media3SeekPositionMs = targetMs
                    media3SeekRequestKey += 1
                },
                showTrackActions = true,
                onOpenSubtitle = {
                    media3TrackPickerKind = TvMedia3TrackPickerKind.Subtitle
                    isTrackSheetVisible = true
                },
                onOpenAudioTrack = {
                    media3TrackPickerKind = TvMedia3TrackPickerKind.Audio
                    isTrackSheetVisible = true
                },
                resumePromptVisible = shouldShowResumePromptCard,
                resumePromptSlot = {
                    TvResumePromptCard(
                        lastPositionMs = resumePromptLastPositionMs,
                        visible = shouldShowResumePromptCard,
                        remainingSeconds = resumePromptCountdownTickRemaining(resumePromptRemainingMs),
                        onContinue = { resumePromptDismissed = true },
                        onStartFromBeginning = {
                            media3SeekPositionMs = 0L
                            media3SeekRequestKey += 1
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
                modifier = Modifier.fillMaxSize(),
            )
            if (!playerErrorMessage.isNullOrBlank()) {
                if (showDolbyVisionDiagnostics) {
                    TvErrorState(
                        title = "诊断信息",
                        message = playbackDiagnosticMessage,
                        onAction = {
                            showDolbyVisionDiagnostics = false
                            playerErrorMessage = null
                            routeRetryNonce += 1
                            updatePlaybackSession(LongFormPlaybackSession(hasStartedPlayback = true, isPausedByUser = false))
                        },
                    )
                } else {
                    TvErrorState(
                        title = "暂不能播放",
                        message = playerErrorMessage.orEmpty(),
                        onAction = {
                            playerErrorMessage = null
                            routeRetryNonce += 1
                            updatePlaybackSession(LongFormPlaybackSession(hasStartedPlayback = true, isPausedByUser = false))
                        },
                        secondaryActionLabel = if (showDolbyVisionDiagnosticsButton) "诊断信息" else null,
                        onSecondaryAction = if (showDolbyVisionDiagnosticsButton) {
                            {
                                showDolbyVisionDiagnostics = true
                            }
                        } else {
                            null
                        },
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
            if (isTrackSheetVisible && playerErrorMessage == null && media3TrackPickerKind != null) {
                TvMedia3TrackPickerLayer(
                    kind = media3TrackPickerKind,
                    subtitleTracks = detail.subtitleTracks.filter { it.available && it.url.isNotBlank() && !it.isEmbedded },
                    selectedSubtitleTrackId = selectedSubtitleTrackId?.takeIf { it.isNotBlank() },
                    onSelectSubtitleTrack = { trackId ->
                        selectedSubtitleTrackId = trackId ?: ""
                        val preference = buildSubtitleTrackPreference(detail.subtitleTracks.firstOrNull { it.id == trackId })
                        storedSubtitlePreference = preference
                        viewModel.saveTvSubtitlePreference(detail.id, preference)
                    },
                    audioTracks = media3AudioTracks,
                    selectedAudioTrackId = selectedAudioTrackId,
                    onSelectAudioTrack = { trackId ->
                        selectedAudioTrackId = trackId ?: ""
                        val preference = buildAudioTrackPreference(media3AudioTracks.firstOrNull { it.id == trackId })
                        storedAudioPreference = preference
                        viewModel.saveTvAudioPreference(detail.id, preference)
                    },
                    onDismissRequest = {
                        isTrackSheetVisible = false
                        media3TrackPickerKind = null
                    },
                )
            }
        } else {
            if (showDolbyVisionDiagnostics) {
                TvErrorState(
                    title = "诊断信息",
                    message = playbackDiagnosticMessage,
                    onAction = {
                        showDolbyVisionDiagnostics = false
                        routeRetryNonce += 1
                        viewModel.load()
                    },
                )
            } else {
                TvErrorState(
                    title = "暂不能播放",
                    message = playerBlockMessage ?: "暂无可播放视频",
                    onAction = {
                        routeRetryNonce += 1
                        viewModel.load()
                    },
                    secondaryActionLabel = if (showDolbyVisionDiagnosticsButton) "诊断信息" else null,
                    onSecondaryAction = if (showDolbyVisionDiagnosticsButton) {
                        {
                            showDolbyVisionDiagnostics = true
                        }
                    } else {
                        null
                    },
                )
            }
            if (showBackConfirmPrompt) {
                TvPlayerBackConfirmPrompt(
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .padding(bottom = 48.dp),
                )
            }
        }
        if (pendingDvExitToDetail) {
            TvDolbyVisionExitToDetailCover()
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

private fun reportTvLongFormMedia3History(
    viewModel: DetailViewModel,
    videoId: String,
    playerSnapshot: TvMedia3PlaybackSnapshot,
) {
    val snapshot = tvPlaybackHistorySnapshot(
        positionMs = playerSnapshot.positionMs,
        durationMs = playerSnapshot.durationMs,
    )
    viewModel.reportHistory(videoId, snapshot.watchSeconds, snapshot.completed)
}
