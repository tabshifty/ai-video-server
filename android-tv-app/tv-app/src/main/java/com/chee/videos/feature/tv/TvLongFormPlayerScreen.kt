package com.chee.videos.feature.tv

import android.app.Activity
import android.os.SystemClock
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Info
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
import com.chee.videos.core.model.TvTrackPreference
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.ui.LaunchedTvInitialFocus
import com.chee.videos.core.ui.LongFormAudioTrack
import com.chee.videos.core.ui.PlayerGlassSurface
import com.chee.videos.core.ui.PlayerGlassSurfaceStrong
import com.chee.videos.core.ui.TvErrorState
import com.chee.videos.core.ui.TvPageLoadingState
import com.chee.videos.core.ui.TvSeriesCorePlaybackOverlay
import com.chee.videos.core.ui.buildAudioTrackPreference
import com.chee.videos.core.ui.buildSubtitleTrackPreference
import com.chee.videos.core.ui.resolveInitialSubtitleTrackId
import com.chee.videos.core.ui.resolveAudioSelectionOnTrackLoad
import com.chee.videos.core.ui.resolveSelectedSubtitleTrackByPreference
import com.chee.videos.core.ui.resolveSubtitleSelectionOnTrackLoad
import com.chee.videos.core.ui.tryRequestFocus
import com.chee.videos.core.ui.tvFocusableGlow
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
        BackHandler { handlePlaybackBack() }
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
    val isMedia3Route = playbackRoute.kind == TvPlaybackRouteKind.EXOPLAYER && !playableUrl.isNullOrBlank()
    val canPlay = isMedia3Route
    val latestDetailId by rememberUpdatedState(detail.id)
    var media3Snapshot by remember(detail.id) { mutableStateOf(TvMedia3PlaybackSnapshot(positionMs = 0L, durationMs = 0L)) }
    val latestMedia3Snapshot by rememberUpdatedState(media3Snapshot)
    var hasStartedPlayback by rememberSaveable(detail.id) { mutableStateOf(true) }
    var isPausedByUser by rememberSaveable(detail.id) { mutableStateOf(false) }
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
    var hasRenderedFirstFrame by rememberSaveable(detail.id) { mutableStateOf(false) }
    var activeSoftRetryAttemptKey by remember(detail.id) { mutableStateOf<Int?>(null) }
    var ignoredRetryAttemptKey by remember(detail.id) { mutableStateOf<Int?>(null) }
    var cancelPrepareRequestKey by remember(detail.id) { mutableStateOf(0) }
    var softRetryUiState by remember(detail.id) { mutableStateOf<TvLongFormSoftRetryUiState?>(null) }
    var retryActionFocusRequestKey by remember(detail.id) { mutableStateOf(0) }
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

    fun closeDiagnosticsPanel() {
        showDolbyVisionDiagnostics = false
        if (softRetryUiState is TvLongFormSoftRetryUiState.Failed) {
            retryActionFocusRequestKey += 1
        }
    }

    fun requestPlaybackRetry() {
        val nextRetryKey = routeRetryNonce + 1
        showDolbyVisionDiagnostics = false
        ignoredRetryAttemptKey = null
        playerErrorMessage = null
        if (hasRenderedFirstFrame) {
            activeSoftRetryAttemptKey = nextRetryKey
            softRetryUiState = TvLongFormSoftRetryUiState.Preparing(nextRetryKey)
        } else {
            activeSoftRetryAttemptKey = null
            softRetryUiState = null
        }
        routeRetryNonce = nextRetryKey
        hasStartedPlayback = true
        isPausedByUser = false
    }

    fun cancelCurrentPlaybackRetry() {
        val preparingState = softRetryUiState as? TvLongFormSoftRetryUiState.Preparing ?: return
        showDolbyVisionDiagnostics = false
        activeSoftRetryAttemptKey = null
        ignoredRetryAttemptKey = preparingState.retryKey
        cancelPrepareRequestKey += 1
        playerErrorMessage = null
        softRetryUiState = TvLongFormSoftRetryUiState.Canceled(preparingState.retryKey, "已取消重试")
        hasStartedPlayback = false
        isPausedByUser = playbackSession.isPausedByUser
    }

    fun dismissSoftRetryFailure() {
        if (softRetryUiState is TvLongFormSoftRetryUiState.Failed) {
            softRetryUiState = null
        }
        showDolbyVisionDiagnostics = false
    }

    BackHandler(enabled = showDolbyVisionDiagnostics) {
        closeDiagnosticsPanel()
    }
    BackHandler(enabled = !showDolbyVisionDiagnostics) {
        when {
            softRetryUiState is TvLongFormSoftRetryUiState.Preparing -> cancelCurrentPlaybackRetry()
            softRetryUiState is TvLongFormSoftRetryUiState.Failed -> dismissSoftRetryFailure()
            else -> handlePlaybackBack()
        }
    }
    val media3SubtitleConfigurations = remember(detail.subtitleTracks, uiState.baseUrl, uiState.accessToken) {
        buildTvMedia3SubtitleConfigurations(
            tracks = detail.subtitleTracks,
            baseUrl = uiState.baseUrl,
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

    LaunchedEffect(softRetryUiState) {
        when (softRetryUiState) {
            is TvLongFormSoftRetryUiState.Succeeded,
            is TvLongFormSoftRetryUiState.Canceled,
            -> {
                val activeState = softRetryUiState
                delay(900L)
                if (softRetryUiState == activeState) {
                    softRetryUiState = null
                }
            }

            else -> Unit
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

    DisposableEffect(isMedia3Route, latestDetailId) {
        onDispose {
            if (isMedia3Route) {
                reportTvLongFormMedia3History(viewModel, latestDetailId, latestMedia3Snapshot)
            }
        }
    }

    KeepScreenOnEffect(enabled = isPlayerActuallyPlaying)
    val softRetryFailureState = softRetryUiState as? TvLongFormSoftRetryUiState.Failed
    val overlayPlayerErrorVisible = softRetryFailureState != null || showDolbyVisionDiagnostics

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black),
    ) {
        if (isMedia3Route && playableUrl != null) {
            TvLongFormMedia3Player(
                sourceUrl = playableUrl,
                mediaId = detail.id,
                title = detail.title,
                accessToken = uiState.accessToken,
                retryKey = routeRetryNonce,
                cancelPrepareRequestKey = cancelPrepareRequestKey,
                shouldPlay = playbackSession.hasStartedPlayback && !playbackSession.isPausedByUser,
                initialPositionMs = resolveTvMedia3ResumePositionMs(
                    historyPositionMs = detail.userState.watchSeconds.coerceAtLeast(0).times(1000L),
                    currentSnapshotPositionMs = latestMedia3Snapshot.positionMs,
                    hasCurrentPlaybackSnapshot = resumedFromHistoryVideoId == detail.id,
                ),
                seekPositionMs = media3SeekPositionMs,
                seekRequestKey = media3SeekRequestKey,
                outputSurface = playbackRoute.outputSurface,
                subtitleConfigurations = media3SubtitleConfigurations,
                selectedSubtitleTrackId = selectedSubtitleTrackId?.takeIf { it.isNotBlank() },
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
                    if (hasRenderedFirstFrame) {
                        when {
                            activeSoftRetryAttemptKey != null -> {
                                val retryKey = activeSoftRetryAttemptKey ?: routeRetryNonce
                                activeSoftRetryAttemptKey = null
                                softRetryUiState = TvLongFormSoftRetryUiState.Failed(retryKey, message)
                                retryActionFocusRequestKey += 1
                            }

                            shouldIgnoreTvLongFormRetryError(ignoredRetryAttemptKey, routeRetryNonce) -> {
                                ignoredRetryAttemptKey = null
                            }

                            else -> {
                                softRetryUiState = TvLongFormSoftRetryUiState.Failed(routeRetryNonce, message)
                                retryActionFocusRequestKey += 1
                            }
                        }
                    } else {
                        updatePlaybackSession(playbackSession.copy(hasStartedPlayback = false))
                    }
                },
                onSnapshotChanged = { snapshot ->
                    media3Snapshot = snapshot
                    screenPositionMs = snapshot.positionMs
                    screenDurationMs = snapshot.durationMs
                },
                onLifecyclePauseSnapshot = { snapshot ->
                    media3Snapshot = snapshot
                    screenPositionMs = snapshot.positionMs
                    screenDurationMs = snapshot.durationMs
                    reportTvLongFormMedia3History(viewModel, detail.id, snapshot)
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
                playerErrorVisible = overlayPlayerErrorVisible,
                modifier = Modifier.fillMaxSize(),
            )
            val currentSoftRetryUiState = softRetryUiState
            if (currentSoftRetryUiState != null &&
                shouldUseTvLongFormSoftRetryFeedback(hasRenderedFirstFrame, currentSoftRetryUiState, showDolbyVisionDiagnostics)
            ) {
                TvLongFormSoftRetryFeedback(
                    state = currentSoftRetryUiState,
                    showDiagnosticsAction = showDolbyVisionDiagnosticsButton && softRetryFailureState != null,
                    diagnosticsButtonFocusRequestKey = retryActionFocusRequestKey,
                    onRetry = ::requestPlaybackRetry,
                    onOpenDiagnostics = { showDolbyVisionDiagnostics = true },
                    modifier = Modifier.align(Alignment.Center),
                )
            }
            if (!playerErrorMessage.isNullOrBlank() && !hasRenderedFirstFrame) {
                if (showDolbyVisionDiagnostics) {
                    TvErrorState(
                        title = "诊断信息",
                        message = playbackDiagnosticMessage,
                        onAction = {
                            closeDiagnosticsPanel()
                        },
                        actionLabel = "返回",
                    )
                } else {
                    TvErrorState(
                        title = "暂不能播放",
                        message = playerErrorMessage.orEmpty(),
                        actionLabel = "重试播放",
                        onAction = ::requestPlaybackRetry,
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
            if (isTrackSheetVisible && softRetryFailureState == null && !showDolbyVisionDiagnostics && media3TrackPickerKind != null) {
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
                        closeDiagnosticsPanel()
                    },
                    actionLabel = "返回",
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
    }
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

internal fun shouldUseTvLongFormSoftRetryFeedback(
    hasRenderedFirstFrame: Boolean,
    softRetryUiState: TvLongFormSoftRetryUiState?,
    showDiagnostics: Boolean,
): Boolean = hasRenderedFirstFrame && softRetryUiState != null && !showDiagnostics

internal fun shouldIgnoreTvLongFormRetryError(
    ignoredRetryAttemptKey: Int?,
    routeRetryNonce: Int,
): Boolean = ignoredRetryAttemptKey != null && ignoredRetryAttemptKey == routeRetryNonce

internal sealed interface TvLongFormSoftRetryUiState {
    val retryKey: Int

    data class Preparing(
        override val retryKey: Int,
        val message: String = "正在重试播放",
    ) : TvLongFormSoftRetryUiState

    data class Succeeded(
        override val retryKey: Int,
        val message: String,
    ) : TvLongFormSoftRetryUiState

    data class Canceled(
        override val retryKey: Int,
        val message: String,
    ) : TvLongFormSoftRetryUiState

    data class Failed(
        override val retryKey: Int,
        val message: String,
    ) : TvLongFormSoftRetryUiState
}

@Composable
private fun TvLongFormSoftRetryFeedback(
    state: TvLongFormSoftRetryUiState,
    showDiagnosticsAction: Boolean,
    diagnosticsButtonFocusRequestKey: Int,
    onRetry: () -> Unit,
    onOpenDiagnostics: () -> Unit,
    modifier: Modifier = Modifier,
) {
    when (state) {
        is TvLongFormSoftRetryUiState.Preparing -> {
            TvLongFormTransientCenterFeedback(
                icon = Icons.Filled.Refresh,
                text = state.message,
                modifier = modifier,
            )
        }

        is TvLongFormSoftRetryUiState.Succeeded -> {
            TvLongFormTransientCenterFeedback(
                icon = Icons.Filled.PlayArrow,
                text = state.message,
                modifier = modifier,
            )
        }

        is TvLongFormSoftRetryUiState.Canceled -> {
            TvLongFormTransientCenterFeedback(
                icon = Icons.Filled.Pause,
                text = state.message,
                modifier = modifier,
            )
        }

        is TvLongFormSoftRetryUiState.Failed -> {
            TvLongFormPersistentFailureFeedback(
                message = state.message,
                showDiagnosticsAction = showDiagnosticsAction,
                focusRequestKey = diagnosticsButtonFocusRequestKey,
                onRetry = onRetry,
                onOpenDiagnostics = onOpenDiagnostics,
                modifier = modifier,
            )
        }
    }
}

@Composable
private fun TvLongFormTransientCenterFeedback(
    icon: ImageVector,
    text: String,
    modifier: Modifier = Modifier,
) {
    Surface(
        color = PlayerGlassSurface,
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
                text = text,
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.bodyMedium,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }
    }
}

@Composable
private fun TvLongFormPersistentFailureFeedback(
    message: String,
    showDiagnosticsAction: Boolean,
    focusRequestKey: Int,
    onRetry: () -> Unit,
    onOpenDiagnostics: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val retryFocusRequester = remember { FocusRequester() }

    LaunchedTvInitialFocus(focusRequestKey, showDiagnosticsAction, message) {
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
                    text = message,
                    color = AppChrome.TextPrimary,
                    style = MaterialTheme.typography.bodyMedium,
                    maxLines = 2,
                    overflow = TextOverflow.Ellipsis,
                    modifier = Modifier.widthIn(max = 420.dp),
                )
            }
            Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                TvLongFormFeedbackActionButton(
                    text = "重试播放",
                    icon = Icons.Filled.Refresh,
                    onClick = onRetry,
                    modifier = Modifier.focusRequester(retryFocusRequester),
                )
                if (showDiagnosticsAction) {
                    TvLongFormFeedbackActionButton(
                        text = "诊断信息",
                        icon = Icons.Filled.Info,
                        onClick = onOpenDiagnostics,
                    )
                }
            }
        }
    }
}

@Composable
private fun TvLongFormFeedbackActionButton(
    text: String,
    icon: ImageVector,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Surface(
        color = AppChrome.AccentSoft,
        shape = AppChrome.ChipShape,
        modifier = modifier
            .tvFocusableGlow(shape = AppChrome.ChipShape, focusedScale = 1.04f)
            .clickable(onClick = onClick),
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 10.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Icon(
                imageVector = icon,
                contentDescription = null,
                tint = AppChrome.TextPrimary,
                modifier = Modifier.padding(end = 8.dp),
            )
            Text(
                text = text,
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.labelLarge,
            )
        }
    }
}
