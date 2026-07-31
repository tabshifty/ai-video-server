package com.chee.videos.feature.tv

import android.app.Activity
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
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
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.ui.LongFormAudioTrack
import com.chee.videos.core.ui.TvErrorState
import com.chee.videos.core.ui.TvPageLoadingState
import com.chee.videos.core.ui.buildAudioTrackPreference
import com.chee.videos.core.ui.buildTvLongFormTitleOverlayData
import com.chee.videos.core.ui.buildSubtitleTrackPreference
import com.chee.videos.core.ui.resolveAudioSelectionOnTrackLoad
import com.chee.videos.core.ui.resolveSubtitleSelectionOnTrackLoad
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
    var showDolbyVisionDiagnostics by remember { mutableStateOf(false) }

    if (uiState.loading) {
        BackHandler(onBack = onBack)
        TvLongFormStartupLoadingCanvas()
        return
    }

    val series = uiState.series
    if (series == null) {
        BackHandler(onBack = onBack)
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black),
        ) {
            TvErrorState(
                message = uiState.errorMessage ?: "播放器数据不存在",
                onAction = viewModel::retry,
            )
        }
        return
    }

    LaunchedEffect(uiState.currentVideoId) {
        showDolbyVisionDiagnostics = false
    }

    val currentEpisode = activeEpisode(uiState)
    var routeRetryNonce by remember(uiState.currentVideoId) { mutableStateOf(0) }
    val currentPlaybackIdentity = remember(uiState.currentVideoId, routeRetryNonce) {
        TvLongFormMedia3EventIdentity(uiState.currentVideoId, routeRetryNonce)
    }
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
    var hasRenderedFirstFrame by rememberSaveable(uiState.currentVideoId) { mutableStateOf(false) }
    var playbackStatus by remember(uiState.currentVideoId) { mutableStateOf(TvLongFormPlaybackStatus.Idle) }
    var interactionMode by remember(uiState.currentVideoId) { mutableStateOf(TvLongFormInteractionMode.Hidden) }
    var loadingFeedback by remember(uiState.currentVideoId) { mutableStateOf(TvLongFormLoadingFeedback.Hidden) }
    var sourcePreparingFeedbackVisible by remember(uiState.currentVideoId) { mutableStateOf(false) }
    var completedAutomaticRetries by remember(uiState.currentVideoId) { mutableStateOf(0) }
    var scheduledRetry by remember(uiState.currentVideoId) { mutableStateOf<TvLongFormErrorAction.ScheduleRetry?>(null) }
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
    var resumePromptDismissed by remember(uiState.currentVideoId) { mutableStateOf(false) }
    var lastAutoplaySwitchedVideoId by remember { mutableStateOf("") }

    val titleOverlayData = remember(
        series.title,
        currentEpisode?.title,
        uiState.activeSeasonNumber,
        uiState.activeEpisodeNumber,
    ) {
        buildTvLongFormTitleOverlayData(
            primaryFallback = currentEpisode?.title.orEmpty(),
            seriesTitle = series.title,
            seasonNumber = uiState.activeSeasonNumber,
            episodeNumber = uiState.activeEpisodeNumber,
            episodeTitle = currentEpisode?.title,
        )
    }
    val mediaSessionTitle = remember(titleOverlayData) {
        listOfNotNull(titleOverlayData.primary, titleOverlayData.secondary)
            .filter { it.isNotBlank() }
            .joinToString(" · ")
    }

    val nextEpisodeRef = remember(uiState.series, uiState.activeSeasonNumber, uiState.activeEpisodeNumber) {
        uiState.nextEpisodeRef()
    }
    val playbackSeasons = remember(uiState.series, uiState.activeSeasonNumber, uiState.activeEpisodeNumber) {
        series.seasons.map { season ->
            TvLongFormSeasonOption(
                number = season.number,
                title = season.title,
                episodes = season.episodes.map { episode ->
                    TvLongFormEpisodeOption(
                        id = episode.id,
                        seasonNumber = season.number,
                        episodeNumber = episode.number,
                        title = episode.title,
                        progressPercent = episode.progressPercent,
                        playable = episode.playable,
                        current = season.number == uiState.activeSeasonNumber && episode.number == uiState.activeEpisodeNumber,
                    )
                },
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
            isPlayerError = playerErrorMessage != null,
            isSelectorVisible = uiState.selectorVisible,
            isBackConfirmVisible = false,
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

    fun requestPlaybackRetry() {
        showDolbyVisionDiagnostics = false
        playerErrorMessage = null
        completedAutomaticRetries = 0
        scheduledRetry = null
        routeRetryNonce += 1
        updatePlaybackSession(LongFormPlaybackSession(hasStartedPlayback = true, isPausedByUser = false))
    }

    BackHandler(enabled = showDolbyVisionDiagnostics) {
        showDolbyVisionDiagnostics = false
    }
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

    fun selectEpisodeFromPlayer(seasonNumber: Int, episodeNumber: Int) {
        reportCurrentEpisodeHistory()
        viewModel.selectEpisode(seasonNumber, episodeNumber)
    }

    fun advanceToNextEpisodeManually() {
        reportCurrentEpisodeHistory()
        viewModel.nextEpisode()
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

    LaunchedEffect(scheduledRetry) {
        val retry = scheduledRetry ?: return@LaunchedEffect
        delay(retry.delayMs)
        if (scheduledRetry == retry) {
            scheduledRetry = null
            playerErrorMessage = null
            routeRetryNonce += 1
            updatePlaybackSession(LongFormPlaybackSession(hasStartedPlayback = true, isPausedByUser = false))
        }
    }

    LaunchedEffect(uiState.playbackPreparing, uiState.currentSourceUrl, uiState.currentVideoId) {
        sourcePreparingFeedbackVisible = false
        if (uiState.playbackPreparing && uiState.currentSourceUrl.isBlank()) {
            delay(TvLongFormStartupFeedbackDelayMillis)
            sourcePreparingFeedbackVisible = true
        }
    }

    val playIntent = playbackSession.hasStartedPlayback &&
        !playbackSession.isPausedByUser &&
        uiState.pendingEndOverlayKind == null
    LaunchedEffect(playbackStatus, playIntent, hasRenderedFirstFrame, routeRetryNonce) {
        loadingFeedback = TvLongFormLoadingFeedback.Hidden
        if (!playIntent || playbackStatus !in setOf(TvLongFormPlaybackStatus.Preparing, TvLongFormPlaybackStatus.Buffering)) {
            return@LaunchedEffect
        }
        delay(
            if (hasRenderedFirstFrame) {
                TvLongFormBufferingFeedbackDelayMillis
            } else {
                TvLongFormStartupFeedbackDelayMillis
            },
        )
        loadingFeedback = if (hasRenderedFirstFrame) {
            TvLongFormLoadingFeedback.Buffering
        } else {
            TvLongFormLoadingFeedback.Startup
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
        uiState.selectorVisible,
        uiState.pendingEndOverlayKind,
    ) {
        if (
            uiState.autoplayEnabled &&
            !uiState.autoplayCanceledForCurrentEpisode &&
            hasNextEpisode &&
            isPlayerActuallyPlaying &&
            playerErrorMessage == null &&
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
        uiState.selectorVisible,
        uiState.pendingEndOverlayKind,
        shouldShowAutoplayPromptCard,
    ) {
        if (
            playerErrorMessage != null ||
            uiState.selectorVisible ||
            uiState.pendingEndOverlayKind != null ||
            shouldShowAutoplayPromptCard
        ) {
            resumePromptDismissed = true
        }
    }

    LaunchedEffect(uiState.currentVideoId, resumePromptLastPositionMs, resumePromptDismissed) {
        if (resumePromptDismissed || resumePromptLastPositionMs <= 0L) return@LaunchedEffect
        delay(3_000L)
        resumePromptDismissed = true
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

    val screenOnPlaybackStatus = if (scheduledRetry != null) {
        TvLongFormPlaybackStatus.Preparing
    } else {
        playbackStatus
    }
    KeepScreenOnEffect(enabled = shouldKeepTvLongFormScreenOn(playIntent, screenOnPlaybackStatus))
    BackHandler(enabled = uiState.pendingEndOverlayKind != null, onBack = onBack)

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black),
    ) {
        if (currentEpisode != null) {
            Box(modifier = Modifier.fillMaxSize()) {
                TvLongFormMedia3Player(
                    sourceUrl = uiState.currentSourceUrl.takeIf { isMedia3Route }.orEmpty(),
                    mediaId = uiState.currentVideoId,
                    title = mediaSessionTitle,
                    accessToken = accessToken,
                    retryKey = routeRetryNonce,
                    shouldPlay = playIntent,
                    initialPositionMs = resolveTvMedia3ResumePositionMs(
                        historyPositionMs = if (!uiState.startCurrentEpisodeFromBeginning) {
                            currentEpisode.watchSeconds.coerceAtLeast(0).times(1000L)
                        } else {
                            0L
                        },
                        currentSnapshotPositionMs = latestMedia3Snapshot.positionMs,
                        hasCurrentPlaybackSnapshot = resumedFromHistoryVideoId == uiState.currentVideoId,
                    ),
                    tvSeekStepSeconds = uiState.tvSeekStepSeconds,
                    seekPositionMs = media3SeekPositionMs,
                    seekRequestKey = media3SeekRequestKey,
                    outputSurface = playbackRoute.outputSurface,
                    subtitleConfigurations = media3SubtitleConfigurations,
                    selectedSubtitleTrackId = normalizeTvSubtitleSelection(selectedSubtitleTrackId),
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
                                val resumePositionMs = if (!uiState.startCurrentEpisodeFromBeginning) {
                                    currentEpisode.watchSeconds.coerceAtLeast(0).times(1000L)
                                } else {
                                    0L
                                }
                                if (uiState.currentVideoId.isNotBlank() && resumedFromHistoryVideoId != uiState.currentVideoId) {
                                    resumedFromHistoryVideoId = uiState.currentVideoId
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
                    onLifecyclePaused = {
                        isPausedByUser = true
                        isPlayerActuallyPlaying = false
                    },
                    onPlaybackStatusChanged = { status ->
                        playbackStatus = status
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
                if (isMedia3Route) {
                    TvLongFormPlaybackChrome(
                        title = titleOverlayData.primary,
                        secondaryTitle = titleOverlayData.secondary,
                        isPlaying = playbackSession.hasStartedPlayback &&
                            !playbackSession.isPausedByUser &&
                            isPlayerActuallyPlaying,
                        positionMs = screenPositionMs,
                        durationMs = screenDurationMs,
                        tvSeekStepSeconds = uiState.tvSeekStepSeconds,
                        subtitleTracks = currentEpisode.subtitleTracks
                            .filter { it.available && it.url.isNotBlank() && !it.isEmbedded },
                        selectedSubtitleTrackId = normalizeTvSubtitleSelection(selectedSubtitleTrackId),
                        audioTracks = media3AudioTracks,
                        selectedAudioTrackId = selectedAudioTrackId?.takeIf { it.isNotBlank() },
                        seasons = playbackSeasons,
                        currentEpisodeId = currentEpisode.id,
                        onTogglePlayPause = {
                            updatePlaybackSession(playbackSession.togglePlayPause(canPlay = canPlay))
                        },
                        onSeekTo = { targetMs ->
                            media3SeekPositionMs = targetMs
                            media3SeekRequestKey += 1
                        },
                        onSelectSubtitleTrack = { trackId ->
                            selectedSubtitleTrackId = trackId ?: ""
                            viewModel.selectSubtitleTrack(trackId)
                        },
                        onSelectAudioTrack = { trackId ->
                            selectedAudioTrackId = trackId ?: ""
                            val preference = buildAudioTrackPreference(media3AudioTracks.firstOrNull { it.id == trackId })
                            viewModel.selectAudioTrack(trackId, preference)
                        },
                        onSelectEpisode = { option ->
                            selectEpisodeFromPlayer(option.seasonNumber, option.episodeNumber)
                        },
                        onPlayNextEpisode = if (hasNextEpisode) ::advanceToNextEpisodeManually else null,
                        onExitPlayback = onBack,
                        resumeNoticeText = if (!resumePromptDismissed && resumePromptLastPositionMs > 0L) {
                            "已从 ${formatTvLongFormTime(resumePromptLastPositionMs)} 继续播放"
                        } else {
                            null
                        },
                        loadingFeedback = if (scheduledRetry != null) {
                            TvLongFormLoadingFeedback.Startup
                        } else {
                            loadingFeedback
                        },
                        blockingUiVisible = playerErrorMessage != null ||
                            showDolbyVisionDiagnostics ||
                            uiState.pendingEndOverlayKind != null ||
                            shouldShowAutoplayPromptCard,
                        onInteractionModeChanged = { mode ->
                            interactionMode = mode
                            viewModel.setSelectorVisible(
                                mode == TvLongFormInteractionMode.PrecisionSeek ||
                                    mode == TvLongFormInteractionMode.SubtitlePanel ||
                                    mode == TvLongFormInteractionMode.AudioPanel ||
                                    mode == TvLongFormInteractionMode.EpisodePanel,
                            )
                        },
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
                        onReplayCurrent = {
                            viewModel.dismissEndOverlay()
                            media3SeekPositionMs = 0L
                            media3SeekRequestKey += 1
                            updatePlaybackSession(
                                LongFormPlaybackSession(hasStartedPlayback = true, isPausedByUser = false),
                            )
                        },
                        onBackToDetail = onBack,
                        modifier = Modifier.fillMaxSize(),
                    )
                    if (!playerErrorMessage.isNullOrBlank()) {
                        if (showDolbyVisionDiagnostics) {
                            TvErrorState(
                                title = "诊断信息",
                                message = playbackDiagnosticMessage,
                                onAction = {
                                    showDolbyVisionDiagnostics = false
                                    requestPlaybackRetry()
                                },
                            )
                        } else {
                            TvErrorState(
                                title = "暂不能播放",
                                message = playerErrorMessage.orEmpty(),
                                actionLabel = "重试播放",
                                onAction = ::requestPlaybackRetry,
                                secondaryActionLabel = "返回详情",
                                onSecondaryAction = onBack,
                                tertiaryActionLabel = if (showDolbyVisionDiagnosticsButton) "诊断信息" else null,
                                onTertiaryAction = if (showDolbyVisionDiagnosticsButton) {
                                    { showDolbyVisionDiagnostics = true }
                                } else {
                                    null
                                },
                            )
                        }
                    }
                } else if (uiState.playbackPreparing && uiState.currentSourceUrl.isBlank()) {
                    if (sourcePreparingFeedbackVisible) {
                        TvPageLoadingState(message = "")
                    }
                } else {
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
                            secondaryActionLabel = "返回详情",
                            onSecondaryAction = onBack,
                            tertiaryActionLabel = if (showDolbyVisionDiagnosticsButton) "诊断信息" else null,
                            onTertiaryAction = if (showDolbyVisionDiagnosticsButton) {
                                { showDolbyVisionDiagnostics = true }
                            } else {
                                null
                            },
                        )
                    }
                }
            }
        } else {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                TvErrorState(
                    title = "暂不能播放",
                    message = "当前分集不存在",
                    onAction = viewModel::retry,
                    secondaryActionLabel = "返回详情",
                    onSecondaryAction = onBack,
                )
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
    val completed = completedOverride ?: snapshot.completed
    if (shouldSubmitTvPlaybackHistory(videoId, snapshot.watchSeconds, completed)) {
        viewModel.reportHistory(videoId, snapshot.watchSeconds, completed)
    }
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
