package com.chee.videos.feature.tv

import android.app.Activity
import android.os.SystemClock
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Pause
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
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
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.ui.LaunchedTvInitialFocus
import com.chee.videos.core.ui.LongFormAudioTrack
import com.chee.videos.core.ui.PlayerGlassSurfaceStrong
import com.chee.videos.core.ui.TvEpisodeRailItem
import com.chee.videos.core.ui.TvErrorState
import com.chee.videos.core.ui.TvPageLoadingState
import com.chee.videos.core.ui.TvSeriesCorePlaybackOverlay
import com.chee.videos.core.ui.buildAudioTrackPreference
import com.chee.videos.core.ui.buildSubtitleTrackPreference
import com.chee.videos.core.ui.resolveAudioSelectionOnTrackLoad
import com.chee.videos.core.ui.resolveSubtitleSelectionOnTrackLoad
import com.chee.videos.core.ui.tvFocusableScaleOnly
import com.chee.videos.core.ui.tryRequestFocus
import com.chee.videos.feature.detail.LongFormPlaybackSession
import kotlinx.coroutines.delay

@Composable
fun TvSeriesPlayerScreen(
    accessToken: String,
    onBack: () -> Unit,
    viewModel: TvSeriesPlayerViewModel = hiltViewModel(),
) {
    val uiState by viewModel.uiState.collectAsStateWithLifecycle()
    val context = LocalContext.current
    val activity = context as? Activity
    var backPromptAtMillis by remember { mutableStateOf<Long?>(null) }
    var showBackConfirmPrompt by remember { mutableStateOf(false) }
    var showDolbyVisionDiagnostics by remember { mutableStateOf(false) }

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
        BackHandler { handlePlaybackBack() }
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black),
        ) {
            TvPageLoadingState(message = "正在加载电视剧播放器")
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

    val series = uiState.series
    if (series == null) {
        BackHandler { handlePlaybackBack() }
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black),
        ) {
            TvErrorState(
                message = uiState.errorMessage ?: "播放器数据不存在",
                onAction = viewModel::retry,
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

    LaunchedEffect(uiState.currentVideoId) {
        showDolbyVisionDiagnostics = false
    }

    val currentEpisode = activeEpisode(uiState)
    var routeRetryNonce by remember(uiState.currentVideoId) { mutableStateOf(0) }
    val displayCapability = remember(context, uiState.currentVideoId, routeRetryNonce) {
        evaluateDolbyVisionDisplayCapability(AndroidDisplayHdrCapabilityReader(context))
    }
    val playbackRoute = remember(currentEpisode?.metadata, displayCapability, uiState.currentSourceUrl, uiState.playbackBlockedMessage) {
        if (uiState.playbackBlockedMessage != null) {
            TvPlaybackRoute(kind = TvPlaybackRouteKind.BLOCKED, blockMessage = uiState.playbackBlockedMessage)
        } else {
            resolveTvPlaybackRoute(
                metadata = currentEpisode?.metadata,
                displayCapability = displayCapability,
                playbackUrl = uiState.currentSourceUrl,
                media3Available = true,
            )
        }
    }
    val showDolbyVisionDiagnosticsButton = remember(currentEpisode?.metadata, playbackRoute) {
        isTvDolbyVisionDiagnosticsAvailable(playbackRoute)
    }
    val playerBlockMessage = playbackRoute.blockMessage
    val isMedia3Route = uiState.canPlayCurrentEpisode &&
        uiState.currentSourceUrl.isNotBlank() &&
        playbackRoute.kind == TvPlaybackRouteKind.EXOPLAYER
    val canPlay = isMedia3Route
    val latestUiState by rememberUpdatedState(uiState)
    var media3Snapshot by remember { mutableStateOf(TvMedia3PlaybackSnapshot(positionMs = 0L, durationMs = 0L)) }
    val latestMedia3Snapshot by rememberUpdatedState(media3Snapshot)
    var hasStartedPlayback by rememberSaveable(uiState.currentVideoId) { mutableStateOf(false) }
    var isPausedByUser by rememberSaveable(uiState.currentVideoId) { mutableStateOf(false) }
    var selectedSubtitleTrackId by rememberSaveable(uiState.currentVideoId) { mutableStateOf<String?>(null) }
    var selectedAudioTrackId by rememberSaveable(uiState.currentVideoId) { mutableStateOf<String?>(null) }
    var media3AudioTracks by remember(uiState.currentVideoId) { mutableStateOf(emptyList<LongFormAudioTrack>()) }
    var isPlayerActuallyPlaying by remember(uiState.currentVideoId) { mutableStateOf(false) }
    var playerErrorMessage by remember(uiState.currentVideoId) { mutableStateOf<String?>(null) }
    // 首帧渲染门槛：对齐单片屏 `TvLongFormPlayerScreen` 的软重试语义。首帧已现后的 onError
    // 不再清 hasStartedPlayback、不再触发全屏 TvErrorState，改为非阻塞中心失败提示 + OK 重试。
    var hasRenderedFirstFrame by rememberSaveable(uiState.currentVideoId) { mutableStateOf(false) }
    var activeSoftRetryAttemptKey by remember(uiState.currentVideoId) { mutableStateOf<Int?>(null) }
    var ignoredRetryAttemptKey by remember(uiState.currentVideoId) { mutableStateOf<Int?>(null) }
    var cancelPrepareRequestKey by remember(uiState.currentVideoId) { mutableStateOf(0) }
    var softRetryUiState by remember(uiState.currentVideoId) { mutableStateOf<TvLongFormSoftRetryUiState?>(null) }
    var retryActionFocusRequestKey by remember(uiState.currentVideoId) { mutableStateOf(0) }
    val softRetryFailureVisible = softRetryUiState is TvLongFormSoftRetryUiState.Failed
    val playbackDiagnosticMessage = remember(playbackRoute, displayCapability, uiState.currentSourceUrl, playerErrorMessage) {
        buildTvDolbyVisionDiagnosticMessage(
            route = playbackRoute,
            displayCapability = displayCapability,
            playbackUrl = uiState.currentSourceUrl,
            media3Available = true,
            failureMessage = playerErrorMessage,
        )
    }
    var screenPositionMs by remember { mutableStateOf(0L) }
    var screenDurationMs by remember { mutableStateOf(0L) }
    var lastHistoryVideoId by remember { mutableStateOf("") }
    var resumedFromHistoryVideoId by remember { mutableStateOf("") }
    var resumePromptLastPositionMs by remember(uiState.currentVideoId) { mutableStateOf(0L) }
    var resumePromptRemainingMs by remember(uiState.currentVideoId) { mutableStateOf(0L) }
    var resumePromptDismissed by remember(uiState.currentVideoId) { mutableStateOf(false) }
    var isTrackSheetVisible by remember(uiState.currentVideoId) { mutableStateOf(false) }
    var media3TrackPickerKind by remember(uiState.currentVideoId) { mutableStateOf<TvMedia3TrackPickerKind?>(null) }
    var lastAutoplaySwitchedVideoId by remember { mutableStateOf("") }
    var openEpisodeRailRequestKey by remember(uiState.currentVideoId) { mutableStateOf(0) }

    val nextEpisodeRef = remember(uiState.series, uiState.activeSeasonNumber, uiState.activeEpisodeNumber) {
        uiState.nextEpisodeRef()
    }
    val episodeRailItems = remember(uiState.series, uiState.activeSeasonNumber, uiState.activeEpisodeNumber) {
        activeSeason(uiState)?.episodes.orEmpty().map { episode ->
            TvEpisodeRailItem(
                id = episode.id,
                number = episode.number,
                title = episode.title,
                playable = episode.playable,
                current = episode.number == uiState.activeEpisodeNumber,
            )
        }
    }
    val hasNextEpisode = nextEpisodeRef != null
    val remainingMs = (screenDurationMs - screenPositionMs).coerceAtLeast(0L)
    val remainingSeconds = autoplayCountdownTickRemaining(remainingMs)
    val shouldShowAutoplayPromptCard = shouldShowAutoplayPromptCard(
        AutoplayPromptGuardInput(
            isPlaying = isPlayerActuallyPlaying,
            autoplayEnabled = uiState.autoplayEnabled,
            hasNextEpisode = hasNextEpisode,
            isPlayerError = playerErrorMessage != null || softRetryFailureVisible,
            isSelectorVisible = uiState.selectorVisible,
            isBackConfirmVisible = showBackConfirmPrompt,
            isEndOverlayVisible = uiState.pendingEndOverlayKind != null,
            isLoading = uiState.loading,
            isCanceledForCurrentEpisode = uiState.autoplayCanceledForCurrentEpisode,
            remainingMs = remainingMs,
            durationMs = screenDurationMs,
        ),
    )

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

    fun requestSoftPlaybackRetry() {
        val nextRetryKey = routeRetryNonce + 1
        showDolbyVisionDiagnostics = false
        ignoredRetryAttemptKey = null
        playerErrorMessage = null
        activeSoftRetryAttemptKey = nextRetryKey
        softRetryUiState = TvLongFormSoftRetryUiState.Preparing(nextRetryKey)
        routeRetryNonce = nextRetryKey
        updatePlaybackSession(LongFormPlaybackSession(hasStartedPlayback = true, isPausedByUser = false))
    }

    fun cancelCurrentPlaybackRetry() {
        val preparingState = softRetryUiState as? TvLongFormSoftRetryUiState.Preparing ?: return
        showDolbyVisionDiagnostics = false
        activeSoftRetryAttemptKey = null
        ignoredRetryAttemptKey = preparingState.retryKey
        cancelPrepareRequestKey += 1
        playerErrorMessage = null
        softRetryUiState = TvLongFormSoftRetryUiState.Canceled(preparingState.retryKey, "已取消重试")
    }

    fun dismissSoftRetryFailure() {
        if (softRetryUiState is TvLongFormSoftRetryUiState.Failed) {
            softRetryUiState = null
        }
        showDolbyVisionDiagnostics = false
    }

    BackHandler(enabled = showDolbyVisionDiagnostics) {
        showDolbyVisionDiagnostics = false
    }
    BackHandler(enabled = !showDolbyVisionDiagnostics) {
        when (resolveSeriesSoftRetryBackAction(softRetryUiState)) {
            SeriesSoftRetryBackAction.CancelPreparing -> cancelCurrentPlaybackRetry()
            SeriesSoftRetryBackAction.DismissFailure -> dismissSoftRetryFailure()
            SeriesSoftRetryBackAction.DelegateToPlayerBack -> when {
                uiState.playbackPreparing && uiState.episodeSwitchState is TvEpisodeSwitchUiState.Preparing -> {
                    viewModel.cancelEpisodeSwitch()
                }

                uiState.episodeSwitchState is TvEpisodeSwitchUiState.Failed -> {
                    viewModel.clearEpisodeSwitchFeedback()
                }

                else -> handlePlaybackBack()
            }
        }
    }

    val resumePromptGuardInput = ResumePromptGuardInput(
        hasResumeSeekTriggered = resumedFromHistoryVideoId == uiState.currentVideoId && resumePromptLastPositionMs > 0L,
        promptPermanentlyDismissed = resumePromptDismissed,
        isPlayerError = playerErrorMessage != null || softRetryFailureVisible,
        isBackConfirmVisible = showBackConfirmPrompt,
        isEpisodeSelectorVisible = uiState.selectorVisible,
        isTrackSheetVisible = isTrackSheetVisible,
        isEndOverlayVisible = uiState.pendingEndOverlayKind != null,
        isAutoplayPromptVisible = shouldShowAutoplayPromptCard,
        isPausedByUser = playbackSession.isPausedByUser,
        remainingMs = resumePromptRemainingMs,
        skipForUserInitiatedEpisodeSwitch = uiState.episodeSwitchState is TvEpisodeSwitchUiState.Succeeded,
    )
    val shouldTickResumePromptCountdown = shouldTickResumePromptCountdown(resumePromptGuardInput)
    val shouldShowResumePromptCard = shouldShowResumePromptCard(resumePromptGuardInput)
    var media3SeekPositionMs by remember(uiState.currentVideoId) { mutableStateOf<Long?>(null) }
    var media3SeekRequestKey by remember(uiState.currentVideoId) { mutableStateOf(0) }
    val media3SubtitleConfigurations = remember(currentEpisode?.subtitleTracks, uiState.baseUrl, accessToken) {
        buildTvMedia3SubtitleConfigurations(
            tracks = currentEpisode?.subtitleTracks.orEmpty(),
            baseUrl = uiState.baseUrl,
        )
    }

    fun reportCurrentEpisodeHistory(completedOverride: Boolean? = null) {
        val videoId = latestUiState.currentVideoId
        if (videoId.isBlank()) {
            return
        }
        reportTvSeriesMedia3History(
            viewModel = viewModel,
            videoId = videoId,
            playerSnapshot = latestMedia3Snapshot,
            completedOverride = completedOverride,
        )
        if (lastHistoryVideoId == videoId) {
            lastHistoryVideoId = ""
        }
    }

    fun selectEpisodeFromPlayer(episodeNumber: Int) {
        reportCurrentEpisodeHistory()
        viewModel.selectEpisode(episodeNumber)
    }

    fun advanceFromAutoplay() {
        val state = latestUiState
        val videoId = state.currentVideoId
        if (videoId.isBlank() || lastAutoplaySwitchedVideoId == videoId || !state.hasNextPlayableEpisode()) {
            return
        }
        lastAutoplaySwitchedVideoId = videoId
        reportCurrentEpisodeHistory(completedOverride = true)
        viewModel.advanceToNextEpisodeFromAutoplay()
    }

    fun handlePlaybackEnded() {
        val state = latestUiState
        if (!shouldHandlePlaybackEnded(state.currentVideoId, lastAutoplaySwitchedVideoId)) {
            return
        }
        val hasNext = state.hasNextPlayableEpisode()
        when {
            state.autoplayEnabled && !state.autoplayCanceledForCurrentEpisode && hasNext -> advanceFromAutoplay()
            hasNext -> {
                reportCurrentEpisodeHistory(completedOverride = true)
                viewModel.showEndOverlay(TvEndOverlayKind.CURRENT_FINISHED)
            }
            else -> {
                reportCurrentEpisodeHistory(completedOverride = true)
                viewModel.showEndOverlay(TvEndOverlayKind.SERIES_FINISHED)
            }
        }
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

    LaunchedEffect(uiState.currentVideoId) {
        if (lastHistoryVideoId.isNotBlank() && lastHistoryVideoId != uiState.currentVideoId) {
            reportTvSeriesMedia3History(
                viewModel = viewModel,
                videoId = lastHistoryVideoId,
                playerSnapshot = latestMedia3Snapshot,
            )
        }
        lastHistoryVideoId = uiState.currentVideoId
        media3Snapshot = TvMedia3PlaybackSnapshot(positionMs = 0L, durationMs = 0L)
        screenPositionMs = 0L
        screenDurationMs = 0L
        resumePromptLastPositionMs = 0L
        resumePromptRemainingMs = 0L
        resumePromptDismissed = false
    }

    var media3AutoStartedSourceUrl by remember(uiState.currentVideoId) { mutableStateOf("") }
    LaunchedEffect(uiState.currentSourceUrl, isMedia3Route) {
        if (
            shouldAutoStartTvLongFormMedia3Playback(
                currentSourceUrl = uiState.currentSourceUrl,
                isMedia3Route = isMedia3Route,
                autoStartedSourceUrl = media3AutoStartedSourceUrl,
            )
        ) {
            playerErrorMessage = null
            media3AutoStartedSourceUrl = uiState.currentSourceUrl
            updatePlaybackSession(LongFormPlaybackSession(hasStartedPlayback = true, isPausedByUser = false))
        }
    }

    LaunchedEffect(uiState.currentVideoId, uiState.selectedSubtitleTrackId, currentEpisode?.subtitleTracks, hasStartedPlayback) {
        selectedSubtitleTrackId = resolveTvSubtitleSelectionOnTrackLoad(
            currentSelection = selectedSubtitleTrackId,
            storedSelection = uiState.selectedSubtitleTrackId,
            tracks = currentEpisode?.subtitleTracks.orEmpty(),
            hasStartedPlayback = hasStartedPlayback,
        )
    }

    LaunchedEffect(uiState.currentVideoId, uiState.selectedAudioTrackId) {
        selectedAudioTrackId = uiState.selectedAudioTrackId
    }

    LaunchedEffect(softRetryUiState) {
        when (softRetryUiState) {
            is TvLongFormSoftRetryUiState.Succeeded,
            is TvLongFormSoftRetryUiState.Canceled,
            -> {
                val transientState = softRetryUiState
                delay(900L)
                if (softRetryUiState == transientState) {
                    softRetryUiState = null
                }
            }

            else -> Unit
        }
    }

    LaunchedEffect(
        uiState.currentVideoId,
        isMedia3Route,
        playbackSession.hasStartedPlayback,
        playbackSession.isPausedByUser,
    ) {
        if (!shouldStartPeriodicHistoryReport(
                videoId = uiState.currentVideoId,
                canPlay = isMedia3Route,
                hasStartedPlayback = playbackSession.hasStartedPlayback,
                isPausedByUser = playbackSession.isPausedByUser,
            )
        ) {
            return@LaunchedEffect
        }
        while (true) {
            delay(TvPeriodicHistoryReportIntervalMillis)
            reportTvSeriesMedia3History(viewModel, uiState.currentVideoId, latestMedia3Snapshot)
        }
    }

    LaunchedEffect(
        uiState.currentVideoId,
        uiState.autoplayEnabled,
        uiState.autoplayCanceledForCurrentEpisode,
        hasNextEpisode,
        isPlayerActuallyPlaying,
        screenDurationMs,
        remainingMs,
        playerErrorMessage,
        softRetryFailureVisible,
        showBackConfirmPrompt,
        uiState.selectorVisible,
        uiState.pendingEndOverlayKind,
    ) {
        if (
            uiState.autoplayEnabled &&
            !uiState.autoplayCanceledForCurrentEpisode &&
            hasNextEpisode &&
            isPlayerActuallyPlaying &&
            playerErrorMessage == null &&
            !softRetryFailureVisible &&
            !showBackConfirmPrompt &&
            !uiState.selectorVisible &&
            uiState.pendingEndOverlayKind == null &&
            screenDurationMs > 0L &&
            remainingMs <= 0L
        ) {
            advanceFromAutoplay()
        }
    }

    LaunchedEffect(
        playerErrorMessage,
        softRetryFailureVisible,
        showBackConfirmPrompt,
        uiState.selectorVisible,
        isTrackSheetVisible,
        uiState.pendingEndOverlayKind,
        shouldShowAutoplayPromptCard,
    ) {
        if (
            playerErrorMessage != null ||
            softRetryFailureVisible ||
            showBackConfirmPrompt ||
            uiState.selectorVisible ||
            isTrackSheetVisible ||
            uiState.pendingEndOverlayKind != null ||
            shouldShowAutoplayPromptCard
        ) {
            resumePromptDismissed = true
        }
    }

    LaunchedEffect(uiState.currentVideoId, shouldTickResumePromptCountdown, resumePromptDismissed) {
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

    DisposableEffect(isMedia3Route, uiState.currentVideoId) {
        val ownedVideoId = uiState.currentVideoId
        val ownedIsMedia3Route = isMedia3Route
        onDispose {
            if (ownedIsMedia3Route) {
                reportTvSeriesMedia3History(viewModel, ownedVideoId, latestMedia3Snapshot)
            }
        }
    }

    KeepScreenOnEffect(enabled = isPlayerActuallyPlaying)
    val overlayPlayerErrorVisible = playerErrorMessage != null ||
        softRetryFailureVisible ||
        showDolbyVisionDiagnostics

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black),
    ) {
        if (isMedia3Route) {
            Box(modifier = Modifier.fillMaxSize()) {
                TvLongFormMedia3Player(
                    sourceUrl = uiState.currentSourceUrl,
                    mediaId = uiState.currentVideoId,
                    title = series.title.ifBlank { currentEpisode?.title.orEmpty() },
                    accessToken = accessToken,
                    retryKey = routeRetryNonce,
                    cancelPrepareRequestKey = cancelPrepareRequestKey,
                    shouldPlay = playbackSession.hasStartedPlayback && !playbackSession.isPausedByUser,
                    initialPositionMs = resolveTvMedia3ResumePositionMs(
                        historyPositionMs = if (!uiState.startCurrentEpisodeFromBeginning) {
                            currentEpisode?.watchSeconds?.coerceAtLeast(0)?.times(1000L) ?: 0L
                        } else {
                            0L
                        },
                        currentSnapshotPositionMs = latestMedia3Snapshot.positionMs,
                        hasCurrentPlaybackSnapshot = resumedFromHistoryVideoId == uiState.currentVideoId,
                    ),
                    seekPositionMs = media3SeekPositionMs,
                    seekRequestKey = media3SeekRequestKey,
                    outputSurface = playbackRoute.outputSurface,
                    subtitleConfigurations = media3SubtitleConfigurations,
                    selectedSubtitleTrackId = normalizeTvSubtitleSelection(selectedSubtitleTrackId),
                    selectedAudioTrackId = selectedAudioTrackId,
                    modifier = Modifier.fillMaxSize(),
                    onRenderedFirstFrame = {
                        hasRenderedFirstFrame = true
                    },
                    onPlayingChanged = { playing ->
                        isPlayerActuallyPlaying = playing
                        if (playing) {
                            playerErrorMessage = null
                            ignoredRetryAttemptKey = null
                            val activeRetryKey = activeSoftRetryAttemptKey
                            if (activeRetryKey != null) {
                                activeSoftRetryAttemptKey = null
                                softRetryUiState = TvLongFormSoftRetryUiState.Succeeded(activeRetryKey, "已恢复播放")
                            }
                            val resumePositionMs = if (!uiState.startCurrentEpisodeFromBeginning) {
                                currentEpisode?.watchSeconds?.coerceAtLeast(0)?.times(1000L) ?: 0L
                            } else {
                                0L
                            }
                            if (uiState.currentVideoId.isNotBlank() && resumedFromHistoryVideoId != uiState.currentVideoId) {
                                resumedFromHistoryVideoId = uiState.currentVideoId
                                if (shouldTriggerResumePrompt(resumePositionMs)) {
                                    resumePromptLastPositionMs = resumePositionMs
                                    resumePromptRemainingMs = TvResumePromptTokens.CountdownDurationMs
                                    resumePromptDismissed = false
                                }
                            }
                        }
                    },
                    onError = { message ->
                        isPlayerActuallyPlaying = false
                        when (val action = resolveSeriesOnErrorAction(hasRenderedFirstFrame, message)) {
                            is SeriesOnErrorAction.SoftRetry -> when {
                                activeSoftRetryAttemptKey != null -> {
                                    val retryKey = activeSoftRetryAttemptKey ?: routeRetryNonce
                                    activeSoftRetryAttemptKey = null
                                    softRetryUiState = TvLongFormSoftRetryUiState.Failed(retryKey, action.message)
                                    retryActionFocusRequestKey += 1
                                }

                                shouldIgnoreTvLongFormRetryError(ignoredRetryAttemptKey, routeRetryNonce) -> {
                                    ignoredRetryAttemptKey = null
                                }

                                else -> {
                                    softRetryUiState = TvLongFormSoftRetryUiState.Failed(routeRetryNonce, action.message)
                                    retryActionFocusRequestKey += 1
                                }
                            }

                            SeriesOnErrorAction.HardError -> {
                                playerErrorMessage = message.ifBlank { "播放失败，请重试" }
                                updatePlaybackSession(playbackSession.copy(hasStartedPlayback = false))
                            }
                        }
                    },
                    onEnded = ::handlePlaybackEnded,
                    onSnapshotChanged = { snapshot ->
                        media3Snapshot = snapshot
                        screenPositionMs = snapshot.positionMs
                        screenDurationMs = snapshot.durationMs
                    },
                    onLifecyclePauseSnapshot = { snapshot ->
                        media3Snapshot = snapshot
                        screenPositionMs = snapshot.positionMs
                        screenDurationMs = snapshot.durationMs
                        reportTvSeriesMedia3History(viewModel, uiState.currentVideoId, snapshot)
                    },
                    onAudioTracksChanged = { tracks ->
                        media3AudioTracks = tracks
                        val resolvedSelection = resolveAudioSelectionOnTrackLoad(
                            currentSelection = selectedAudioTrackId,
                            storedPreference = uiState.selectedAudioPreference,
                            tracks = tracks,
                        )
                        if (resolvedSelection != selectedAudioTrackId?.takeIf { it.isNotBlank() }) {
                            selectedAudioTrackId = resolvedSelection ?: ""
                        }
                    },
                )
                TvSeriesCorePlaybackOverlay(
                    title = series.title.ifBlank { currentEpisode?.title.orEmpty() },
                    isPlaying = playbackSession.hasStartedPlayback && !playbackSession.isPausedByUser && isPlayerActuallyPlaying,
                    positionMs = screenPositionMs,
                    durationMs = screenDurationMs,
                    tvSeekStepSeconds = uiState.tvSeekStepSeconds,
                    seriesTitleForOverlay = series.title,
                    seasonNumber = uiState.activeSeasonNumber,
                    episodeNumber = uiState.activeEpisodeNumber,
                    episodeTitle = currentEpisode?.title,
                    episodeRailItems = episodeRailItems,
                    currentEpisodeRailItemId = currentEpisode?.id,
                    episodeSwitchState = uiState.episodeSwitchState,
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
                    onSelectEpisodeRailItem = { selectedItem ->
                        val episodeNumber = activeSeason(uiState)?.episodes
                            ?.firstOrNull { it.id == selectedItem.id }
                            ?.number
                        if (episodeNumber != null) {
                            selectEpisodeFromPlayer(episodeNumber)
                        }
                    },
                    onEpisodeRailVisibilityChanged = viewModel::setSelectorVisible,
                    onDismissEpisodeSwitchFeedback = viewModel::clearEpisodeSwitchFeedback,
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
                    playerErrorVisible = overlayPlayerErrorVisible,
                    openEpisodeRailRequestKey = openEpisodeRailRequestKey,
                    modifier = Modifier.fillMaxSize(),
                )
                nextEpisodeRef?.let { next ->
                    TvAutoplayPromptCard(
                        nextEpisodeRef = next,
                        visible = shouldShowAutoplayPromptCard,
                        remainingSeconds = remainingSeconds,
                        onPlayNow = ::advanceFromAutoplay,
                        onCancel = viewModel::cancelAutoplayForCurrentEpisode,
                        modifier = Modifier
                            .align(Alignment.BottomEnd)
                            .padding(
                                end = TvAutoplayPromptTokens.HorizontalPaddingDp,
                                bottom = TvAutoplayPromptTokens.BottomPaddingDp,
                            ),
                    )
                }
                TvSeriesEndOverlay(
                    kind = uiState.pendingEndOverlayKind,
                    onPlayNext = viewModel::nextEpisode,
                    onBackToDetail = onBack,
                    modifier = Modifier.fillMaxSize(),
                )
                // 首帧已现后的非阻塞软重试卡片：保留当前分集已渲染画面，OK 键重试当前分集。
                // 不清 hasStartedPlayback、不进全屏 TvErrorState，对齐 CONTEXT.md「TV 长视频播放器软准备」契约。
                TvSeriesPlayerSoftRetryFeedback(
                    state = softRetryUiState,
                    focusRequestKey = retryActionFocusRequestKey,
                    onRetry = ::requestSoftPlaybackRetry,
                    modifier = Modifier.align(Alignment.Center),
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
                            tertiaryActionLabel = if (episodeRailItems.isNotEmpty()) "选集" else null,
                            onTertiaryAction = if (episodeRailItems.isNotEmpty()) {
                                {
                                    showDolbyVisionDiagnostics = false
                                    playerErrorMessage = null
                                    openEpisodeRailRequestKey += 1
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
                        subtitleTracks = currentEpisode?.subtitleTracks.orEmpty().filter { it.available && it.url.isNotBlank() && !it.isEmbedded },
                        selectedSubtitleTrackId = normalizeTvSubtitleSelection(selectedSubtitleTrackId),
                        onSelectSubtitleTrack = { trackId ->
                            selectedSubtitleTrackId = trackId ?: ""
                            viewModel.selectSubtitleTrack(trackId)
                        },
                        audioTracks = media3AudioTracks,
                        selectedAudioTrackId = selectedAudioTrackId,
                        onSelectAudioTrack = { trackId ->
                            selectedAudioTrackId = trackId ?: ""
                            val preference = buildAudioTrackPreference(media3AudioTracks.firstOrNull { it.id == trackId })
                            viewModel.selectAudioTrack(trackId, preference)
                        },
                        onDismissRequest = {
                            isTrackSheetVisible = false
                            media3TrackPickerKind = null
                        },
                    )
                }
            }
        } else if (uiState.playbackPreparing && uiState.currentSourceUrl.isBlank()) {
            Box(modifier = Modifier.fillMaxSize()) {
                val message = when (val switchState = uiState.episodeSwitchState) {
                    is TvEpisodeSwitchUiState.Preparing -> "正在切到第 ${switchState.targetEpisodeNumber} 集"
                    else -> "正在准备当前分集"
                }
                TvPageLoadingState(message = message)
                if (showBackConfirmPrompt) {
                    TvPlayerBackConfirmPrompt(
                        modifier = Modifier
                            .align(Alignment.BottomCenter)
                            .padding(bottom = 48.dp),
                    )
                }
            }
        } else {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                if (showDolbyVisionDiagnostics) {
                    TvErrorState(
                        title = "诊断信息",
                        message = playbackDiagnosticMessage,
                        onAction = {
                            showDolbyVisionDiagnostics = false
                            routeRetryNonce += 1
                            viewModel.retry()
                        },
                    )
                } else {
                    TvErrorState(
                        title = "暂不能播放",
                        message = playerBlockMessage ?: "当前分集暂无可播放视频",
                        onAction = {
                            routeRetryNonce += 1
                            viewModel.retry()
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
        }
    }

}

private fun reportTvSeriesMedia3History(
    viewModel: TvSeriesPlayerViewModel,
    videoId: String,
    playerSnapshot: TvMedia3PlaybackSnapshot,
    completedOverride: Boolean? = null,
) {
    val snapshot = tvPlaybackHistorySnapshot(
        positionMs = playerSnapshot.positionMs,
        durationMs = playerSnapshot.durationMs,
    )
    viewModel.reportHistory(videoId, snapshot.watchSeconds, completedOverride ?: snapshot.completed)
}

private fun TvSeriesPlayerUiState.nextEpisodeRef(): TvNextEpisodeRef? {
    val currentSeries = series ?: return null
    return resolveNextPlayableEpisode(
        series = currentSeries,
        currentSeasonNumber = selectedSeasonNumber,
        currentEpisodeNumber = selectedEpisodeNumber,
    )
}

private fun resolveTvSubtitleSelectionOnTrackLoad(
    currentSelection: String?,
    storedSelection: String?,
    tracks: List<com.chee.videos.core.model.SubtitleTrackDto>,
    hasStartedPlayback: Boolean,
): String? {
    val current = normalizeTvSubtitleSelection(currentSelection)
    if (hasStartedPlayback && !current.isNullOrBlank() && tracks.any { it.id == current }) {
        return current
    }
    val stored = normalizeTvSubtitleSelection(storedSelection)
    if (!stored.isNullOrBlank() && tracks.any { it.id == stored }) {
        return stored
    }
    return resolveSubtitleSelectionOnTrackLoad(
        currentSelection = current,
        tracks = tracks,
        hasStartedPlayback = hasStartedPlayback,
    )
}

private fun normalizeTvSubtitleSelection(selection: String?): String? =
    selection?.trim()?.takeIf { it.isNotBlank() }

internal fun shouldAutoStartTvLongFormMedia3Playback(
    currentSourceUrl: String,
    isMedia3Route: Boolean,
    autoStartedSourceUrl: String,
): Boolean =
    isMedia3Route &&
        currentSourceUrl.isNotBlank() &&
        autoStartedSourceUrl != currentSourceUrl

/**
 * 剧集屏 Media3 onError 的分派决策。对齐单片屏 `TvLongFormPlayerScreen` 的 hasRenderedFirstFrame 门槛：
 * - [SeriesOnErrorAction.SoftRetry]：首帧已现后失败，保留当前分集已渲染画面，走非阻塞中心失败提示 + OK 重试。
 * - [SeriesOnErrorAction.HardError]：首帧未现（首次 prepare 失败）或无有效错误信息，回退原全屏硬错误卡片。
 *
 * 抽成纯函数以便单测覆盖门槛边界，避免内联在 Composable lambda 中难以测试。
 */
internal sealed interface SeriesOnErrorAction {
    data class SoftRetry(val message: String) : SeriesOnErrorAction
    object HardError : SeriesOnErrorAction
}

internal sealed interface SeriesSoftRetryBackAction {
    object CancelPreparing : SeriesSoftRetryBackAction
    object DismissFailure : SeriesSoftRetryBackAction
    object DelegateToPlayerBack : SeriesSoftRetryBackAction
}

internal fun resolveSeriesSoftRetryBackAction(
    state: TvLongFormSoftRetryUiState?,
): SeriesSoftRetryBackAction = when (state) {
    is TvLongFormSoftRetryUiState.Preparing -> SeriesSoftRetryBackAction.CancelPreparing
    is TvLongFormSoftRetryUiState.Failed -> SeriesSoftRetryBackAction.DismissFailure
    else -> SeriesSoftRetryBackAction.DelegateToPlayerBack
}

internal fun resolveSeriesOnErrorAction(
    hasRenderedFirstFrame: Boolean,
    errorMessage: String,
): SeriesOnErrorAction =
    if (hasRenderedFirstFrame) {
        SeriesOnErrorAction.SoftRetry(errorMessage.ifBlank { "播放失败，请重试" })
    } else {
        SeriesOnErrorAction.HardError
    }

@Composable
private fun TvSeriesPlayerSoftRetryFeedback(
    state: TvLongFormSoftRetryUiState?,
    focusRequestKey: Int,
    onRetry: () -> Unit,
    modifier: Modifier = Modifier,
) {
    when (state) {
        is TvLongFormSoftRetryUiState.Preparing -> {
            TvSeriesPlayerSoftRetryTransientFeedback(
                icon = Icons.Filled.Refresh,
                message = state.message,
                modifier = modifier,
            )
        }

        is TvLongFormSoftRetryUiState.Succeeded -> {
            TvSeriesPlayerSoftRetryTransientFeedback(
                icon = Icons.Filled.PlayArrow,
                message = state.message,
                modifier = modifier,
            )
        }

        is TvLongFormSoftRetryUiState.Canceled -> {
            TvSeriesPlayerSoftRetryTransientFeedback(
                icon = Icons.Filled.Pause,
                message = state.message,
                modifier = modifier,
            )
        }

        is TvLongFormSoftRetryUiState.Failed -> {
            TvSeriesPlayerSoftRetryFailureFeedback(
                state = state,
                focusRequestKey = focusRequestKey,
                onRetry = onRetry,
                modifier = modifier,
            )
        }

        null -> Unit
    }
}

@Composable
private fun TvSeriesPlayerSoftRetryTransientFeedback(
    icon: ImageVector,
    message: String,
    modifier: Modifier = Modifier,
) {
    Surface(
        color = PlayerGlassSurfaceStrong,
        shape = AppChrome.SurfaceShape,
        modifier = modifier,
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Icon(
                imageVector = icon,
                contentDescription = null,
                tint = AppChrome.TextPrimary,
                modifier = Modifier.padding(end = 8.dp),
            )
            Text(
                text = message,
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.bodyMedium,
            )
        }
    }
}

@Composable
private fun TvSeriesPlayerSoftRetryFailureFeedback(
    state: TvLongFormSoftRetryUiState.Failed,
    focusRequestKey: Int,
    onRetry: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val retryFocusRequester = remember { FocusRequester() }

    LaunchedTvInitialFocus(true, focusRequestKey, state.retryKey, state.message) {
        if (focusRequestKey > 0) {
            retryFocusRequester.tryRequestFocus()
        }
    }

    Surface(
        color = PlayerGlassSurfaceStrong,
        shape = AppChrome.SurfaceShape,
        modifier = modifier,
    ) {
        Column(
            modifier = Modifier.padding(horizontal = 18.dp, vertical = 14.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Icon(
                    imageVector = Icons.Filled.Warning,
                    contentDescription = null,
                    tint = AppChrome.Error,
                    modifier = Modifier.padding(end = 8.dp),
                )
                Text(
                    text = state.message,
                    color = AppChrome.TextPrimary,
                    style = MaterialTheme.typography.bodyMedium,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier.widthIn(max = 420.dp),
                )
            }
            TvSeriesPlayerSoftRetryActionButton(
                onClick = onRetry,
                modifier = Modifier.focusRequester(retryFocusRequester),
            )
        }
    }
}

@Composable
private fun TvSeriesPlayerSoftRetryActionButton(
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Surface(
        color = AppChrome.AccentSoft,
        shape = AppChrome.ChipShape,
        modifier = modifier
            .tvFocusableScaleOnly(focusedScale = 1.04f)
            .clickable(onClick = onClick),
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 10.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Icon(
                imageVector = Icons.Filled.Refresh,
                contentDescription = null,
                tint = AppChrome.TextPrimary,
                modifier = Modifier.padding(end = 8.dp),
            )
            Text(
                text = "重试播放",
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.labelLarge,
            )
        }
    }
}
