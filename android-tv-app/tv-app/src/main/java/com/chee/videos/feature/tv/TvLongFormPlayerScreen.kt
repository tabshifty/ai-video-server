package com.chee.videos.feature.tv

import android.app.Activity
import androidx.activity.compose.BackHandler
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
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.runtime.withFrameNanos
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.core.view.WindowCompat
import androidx.core.view.WindowInsetsCompat
import androidx.core.view.WindowInsetsControllerCompat
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import com.chee.videos.core.model.TvSubtitlePreferenceMode
import com.chee.videos.core.model.TvTrackPreference
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.ui.LongFormAudioTrack
import com.chee.videos.core.ui.TvErrorState
import com.chee.videos.core.ui.buildAudioTrackPreference
import com.chee.videos.core.ui.buildTvSubtitlePreference
import com.chee.videos.core.ui.resolveAudioSelectionOnTrackLoad
import com.chee.videos.core.ui.resolveTvSubtitleSelection
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
    var showDolbyVisionDiagnostics by remember { mutableStateOf(false) }

    if (uiState.loading) {
        BackHandler(onBack = onBack)
        TvLongFormStartupLoadingCanvas()
        return
    }

    val detail = uiState.detail
    if (detail == null) {
        BackHandler(onBack = onBack)
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black),
        ) {
            TvErrorState(
                message = uiState.errorMessage ?: "播放器数据不存在",
                onAction = viewModel::load,
            )
        }
        return
    }

    LaunchedEffect(detail.id) {
        showDolbyVisionDiagnostics = false
    }

    var routeRetryNonce by remember(detail.id) { mutableStateOf(0) }
    val currentPlaybackIdentity = remember(detail.id, routeRetryNonce) {
        TvLongFormMedia3EventIdentity(detail.id, routeRetryNonce)
    }
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
    var resumePromptDismissed by remember(detail.id, uiState.accessToken) { mutableStateOf(false) }
    var selectedSubtitleTrackId by rememberSaveable(detail.id) { mutableStateOf<String?>(null) }
    var subtitlePreferenceMode by rememberSaveable(detail.id) { mutableStateOf(TvSubtitlePreferenceMode.AUTO) }
    var storedSubtitlePreference by remember(detail.id) { mutableStateOf<TvTrackPreference?>(null) }
    var storedAudioPreference by remember(detail.id) { mutableStateOf<TvTrackPreference?>(null) }
    var selectedAudioTrackId by rememberSaveable(detail.id) { mutableStateOf<String?>(null) }
    var media3AudioTracks by remember(detail.id) { mutableStateOf(emptyList<LongFormAudioTrack>()) }
    var media3SeekPositionMs by remember(detail.id) { mutableStateOf<Long?>(null) }
    var media3SeekRequestKey by remember(detail.id) { mutableStateOf(0) }
    var isPlayerActuallyPlaying by remember(detail.id, uiState.accessToken) { mutableStateOf(false) }
    var playerErrorMessage by remember(detail.id, uiState.accessToken) { mutableStateOf<String?>(null) }
    var hasRenderedFirstFrame by rememberSaveable(detail.id) { mutableStateOf(false) }
    var playbackStatus by remember(detail.id) { mutableStateOf(TvLongFormPlaybackStatus.Idle) }
    var interactionMode by remember(detail.id) { mutableStateOf(TvLongFormInteractionMode.Hidden) }
    var loadingFeedback by remember(detail.id) { mutableStateOf(TvLongFormLoadingFeedback.Hidden) }
    var completedAutomaticRetries by remember(detail.id) { mutableStateOf(0) }
    var scheduledRetry by remember(detail.id) { mutableStateOf<TvLongFormErrorAction.ScheduleRetry?>(null) }
    var completionVisible by rememberSaveable(detail.id) { mutableStateOf(false) }
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
    }

    fun requestPlaybackRetry() {
        showDolbyVisionDiagnostics = false
        playerErrorMessage = null
        completedAutomaticRetries = 0
        scheduledRetry = null
        completionVisible = false
        routeRetryNonce += 1
        hasStartedPlayback = true
        isPausedByUser = false
    }

    BackHandler(enabled = showDolbyVisionDiagnostics) {
        closeDiagnosticsPanel()
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
        storedSubtitlePreference = viewModel.readTvSubtitlePreference()
        storedAudioPreference = viewModel.readTvAudioPreference()
    }

    LaunchedEffect(detail.id, detail.subtitleTracks, storedSubtitlePreference) {
        val selection = resolveTvSubtitleSelection(
            tracks = detail.subtitleTracks,
            preference = storedSubtitlePreference,
        )
        subtitlePreferenceMode = selection.mode
        selectedSubtitleTrackId = selection.trackId
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

    LaunchedEffect(playerErrorMessage) {
        if (playerErrorMessage != null) {
            resumePromptDismissed = true
        }
    }

    LaunchedEffect(scheduledRetry) {
        val retry = scheduledRetry ?: return@LaunchedEffect
        delay(retry.delayMs)
        if (scheduledRetry == retry) {
            scheduledRetry = null
            playerErrorMessage = null
            routeRetryNonce += 1
            hasStartedPlayback = true
            isPausedByUser = false
        }
    }

    val playIntent = playbackSession.hasStartedPlayback && !playbackSession.isPausedByUser && !completionVisible
    LaunchedEffect(playbackStatus, playIntent, hasRenderedFirstFrame, routeRetryNonce) {
        loadingFeedback = TvLongFormLoadingFeedback.Hidden
        if (!playIntent || playbackStatus !in setOf(TvLongFormPlaybackStatus.Preparing, TvLongFormPlaybackStatus.Buffering)) {
            return@LaunchedEffect
        }
        val delayMs = if (hasRenderedFirstFrame) {
            TvLongFormBufferingFeedbackDelayMillis
        } else {
            TvLongFormStartupFeedbackDelayMillis
        }
        delay(delayMs)
        loadingFeedback = if (hasRenderedFirstFrame) {
            TvLongFormLoadingFeedback.Buffering
        } else {
            TvLongFormLoadingFeedback.Startup
        }
    }

    LaunchedEffect(detail.id, resumePromptLastPositionMs, resumePromptDismissed) {
        if (resumePromptLastPositionMs <= 0L || resumePromptDismissed) return@LaunchedEffect
        delay(3_000L)
        resumePromptDismissed = true
    }

    DisposableEffect(isMedia3Route, latestDetailId) {
        onDispose {
            if (isMedia3Route) {
                reportTvLongFormMedia3History(viewModel, latestDetailId, latestMedia3Snapshot)
            }
        }
    }

    val screenOnPlaybackStatus = if (scheduledRetry != null) {
        TvLongFormPlaybackStatus.Preparing
    } else {
        playbackStatus
    }
    KeepScreenOnEffect(enabled = shouldKeepTvLongFormScreenOn(playIntent, screenOnPlaybackStatus))
    BackHandler(enabled = completionVisible, onBack = onBack)

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
                shouldPlay = playIntent,
                initialPositionMs = resolveTvMedia3ResumePositionMs(
                    historyPositionMs = if (uiState.startFromBeginning) {
                        0L
                    } else {
                        detail.userState.watchSeconds.coerceAtLeast(0).times(1000L)
                    },
                    currentSnapshotPositionMs = latestMedia3Snapshot.positionMs,
                    hasCurrentPlaybackSnapshot = resumedFromHistoryVideoId == detail.id,
                ),
                tvSeekStepSeconds = uiState.tvSeekStepSeconds,
                seekPositionMs = media3SeekPositionMs,
                seekRequestKey = media3SeekRequestKey,
                outputSurface = playbackRoute.outputSurface,
                subtitleConfigurations = media3SubtitleConfigurations,
                selectedSubtitleTrackId = selectedSubtitleTrackId?.takeIf { it.isNotBlank() },
                subtitlePreferenceMode = subtitlePreferenceMode,
                selectedAudioTrackId = selectedAudioTrackId,
                interactionMode = interactionMode,
                modifier = Modifier.fillMaxSize(),
                onRenderedFirstFrame = { eventIdentity ->
                    if (eventIdentity == currentPlaybackIdentity) {
                        hasRenderedFirstFrame = true
                    }
                },
                onPlayingChanged = { playing, eventIdentity ->
                    if (eventIdentity == currentPlaybackIdentity) {
                        isPlayerActuallyPlaying = playing
                        if (playing) {
                            playerErrorMessage = null
                            completedAutomaticRetries = 0
                            scheduledRetry = null
                            if (detail.id.isNotBlank() && resumedFromHistoryVideoId != detail.id) {
                                val resumePositionMs = if (uiState.startFromBeginning) {
                                    0L
                                } else {
                                    detail.userState.watchSeconds.coerceAtLeast(0).times(1000L)
                                }
                                resumedFromHistoryVideoId = detail.id
                                if (shouldTriggerResumePrompt(resumePositionMs)) {
                                    resumePromptLastPositionMs = resumePositionMs
                                    resumePromptDismissed = false
                                }
                            }
                        }
                    }
                },
                onError = { message, eventIdentity, retryable ->
                    if (eventIdentity == currentPlaybackIdentity) {
                        isPlayerActuallyPlaying = false
                        when (val action = resolveTvLongFormErrorAction(retryable, completedAutomaticRetries)) {
                            is TvLongFormErrorAction.ScheduleRetry -> {
                                completedAutomaticRetries = action.attempt
                                scheduledRetry = action
                                playerErrorMessage = null
                            }

                            TvLongFormErrorAction.ShowFinalError -> {
                                scheduledRetry = null
                                playerErrorMessage = message.ifBlank { "播放失败，请重试" }
                                updatePlaybackSession(playbackSession.copy(hasStartedPlayback = false))
                            }
                        }
                    }
                },
                onEnded = { eventIdentity ->
                    if (eventIdentity == currentPlaybackIdentity) {
                        reportTvLongFormMedia3History(
                            viewModel,
                            detail.id,
                            latestMedia3Snapshot,
                            completedOverride = true,
                        )
                        updatePlaybackSession(playbackSession.copy(hasStartedPlayback = false))
                        completionVisible = true
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
                onLifecyclePaused = {
                    isPausedByUser = true
                    isPlayerActuallyPlaying = false
                },
                onPlaybackIntentChanged = { shouldPlay ->
                    updatePlaybackSession(playbackSession.setPlayIntent(shouldPlay = shouldPlay, canPlay = canPlay))
                },
                onPlaybackStatusChanged = { status ->
                    playbackStatus = status
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
            TvLongFormPlaybackChrome(
                title = detail.title,
                secondaryTitle = null,
                isPlaying = playbackSession.hasStartedPlayback && !playbackSession.isPausedByUser && isPlayerActuallyPlaying,
                positionMs = screenPositionMs,
                durationMs = screenDurationMs,
                tvSeekStepSeconds = uiState.tvSeekStepSeconds,
                subtitleTracks = detail.subtitleTracks.filter { it.available && it.url.isNotBlank() && !it.isEmbedded },
                selectedSubtitleTrackId = selectedSubtitleTrackId?.takeIf { it.isNotBlank() },
                subtitlePreferenceMode = subtitlePreferenceMode,
                audioTracks = media3AudioTracks,
                selectedAudioTrackId = selectedAudioTrackId?.takeIf { it.isNotBlank() },
                seasons = emptyList(),
                currentEpisodeId = null,
                onTogglePlayPause = {
                    updatePlaybackSession(playbackSession.togglePlayPause(canPlay = canPlay))
                },
                onSetPlaybackIntent = { shouldPlay ->
                    updatePlaybackSession(playbackSession.setPlayIntent(shouldPlay = shouldPlay, canPlay = canPlay))
                },
                onSeekTo = { targetMs ->
                    media3SeekPositionMs = targetMs
                    media3SeekRequestKey += 1
                },
                onSelectSubtitleTrack = { mode, trackId ->
                    subtitlePreferenceMode = mode
                    selectedSubtitleTrackId = trackId
                    val preference = buildTvSubtitlePreference(
                        mode = mode,
                        track = detail.subtitleTracks.firstOrNull { it.id == trackId },
                    )
                    storedSubtitlePreference = preference
                    viewModel.saveTvSubtitlePreference(preference)
                },
                onSelectAudioTrack = { trackId ->
                    selectedAudioTrackId = trackId ?: ""
                    val preference = buildAudioTrackPreference(media3AudioTracks.firstOrNull { it.id == trackId })
                    storedAudioPreference = preference
                    viewModel.saveTvAudioPreference(preference)
                },
                onSelectEpisode = {},
                onPlayNextEpisode = null,
                onExitPlayback = onBack,
                resumeNoticeText = if (!resumePromptDismissed && resumePromptLastPositionMs > 0L) {
                    "已从 ${formatTvLongFormTime(resumePromptLastPositionMs)} 继续播放"
                } else null,
                loadingFeedback = if (scheduledRetry != null) {
                    TvLongFormLoadingFeedback.Startup
                } else {
                    loadingFeedback
                },
                blockingUiVisible = showDolbyVisionDiagnostics || !playerErrorMessage.isNullOrBlank() || completionVisible,
                onInteractionModeChanged = { mode -> interactionMode = mode },
                modifier = Modifier.fillMaxSize(),
            )
            if (showDolbyVisionDiagnostics && !playerErrorMessage.isNullOrBlank()) {
                TvErrorState(
                    title = "诊断信息",
                    message = playbackDiagnosticMessage,
                    onAction = {
                        closeDiagnosticsPanel()
                    },
                    actionLabel = "返回",
                )
            } else if (!playerErrorMessage.isNullOrBlank()) {
                TvErrorState(
                    title = "暂不能播放",
                    message = playerErrorMessage.orEmpty(),
                    actionLabel = "重试播放",
                    onAction = ::requestPlaybackRetry,
                    secondaryActionLabel = "返回详情",
                    onSecondaryAction = onBack,
                    tertiaryActionLabel = if (showDolbyVisionDiagnosticsButton) "诊断信息" else null,
                    onTertiaryAction = if (showDolbyVisionDiagnosticsButton) {
                        {
                            showDolbyVisionDiagnostics = true
                        }
                    } else {
                        null
                    },
                )
            }
            if (completionVisible) {
                TvLongFormCompletionOverlay(
                    headline = "播放结束",
                    title = detail.title,
                    replayLabel = "重新播放",
                    onBackToDetail = onBack,
                    onReplay = {
                        completionVisible = false
                        completedAutomaticRetries = 0
                        scheduledRetry = null
                        media3SeekPositionMs = 0L
                        media3SeekRequestKey += 1
                        hasStartedPlayback = true
                        isPausedByUser = false
                    },
                    modifier = Modifier.fillMaxSize(),
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
        }
    }
}

private fun reportTvLongFormMedia3History(
    viewModel: DetailViewModel,
    videoId: String,
    playerSnapshot: TvMedia3PlaybackSnapshot,
    completedOverride: Boolean? = null,
) {
    val snapshot = tvPlaybackHistorySnapshot(
        positionMs = playerSnapshot.positionMs,
        durationMs = playerSnapshot.durationMs,
    )
    val completed = completedOverride ?: snapshot.completed
    if (shouldSubmitTvPlaybackHistory(videoId, snapshot.watchSeconds, completed)) {
        viewModel.reportHistory(videoId, snapshot.watchSeconds, completed)
    }
}
