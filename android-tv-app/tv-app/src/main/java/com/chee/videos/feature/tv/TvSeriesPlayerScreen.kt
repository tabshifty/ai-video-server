package com.chee.videos.feature.tv

import android.app.Activity
import android.os.SystemClock
import androidx.activity.compose.BackHandler
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
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
import com.chee.videos.core.player.friendlyLongFormPlaybackErrorMessage
import com.chee.videos.core.player.TvVlcLibrary
import com.chee.videos.core.player.newLongFormMediaPlayer
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.core.ui.LongFormAudioTrack
import com.chee.videos.core.ui.LongFormVideoPlayer
import com.chee.videos.core.ui.TvEpisodeRailItem
import com.chee.videos.core.ui.TvErrorState
import com.chee.videos.core.ui.TvLongFormControlsVariant
import com.chee.videos.core.ui.TvPageLoadingState
import com.chee.videos.core.ui.TvSeriesCorePlaybackOverlay
import com.chee.videos.core.ui.appendAccessTokenQuery
import com.chee.videos.core.ui.applyLongFormMediaSource
import com.chee.videos.core.ui.buildAudioTrackPreference
import com.chee.videos.core.ui.buildSubtitleTrackPreference
import com.chee.videos.core.ui.resolveLongFormPlayerUpdate
import com.chee.videos.core.ui.resolvePlaybackAssetUrl
import com.chee.videos.core.ui.resolveAudioSelectionOnTrackLoad
import com.chee.videos.core.ui.resolveSelectedSubtitleTrack
import com.chee.videos.core.ui.resolveSubtitleSelectionOnTrackLoad
import com.chee.videos.feature.detail.LongFormPlaybackSession
import kotlinx.coroutines.delay
import org.videolan.libvlc.MediaPlayer
import org.videolan.libvlc.interfaces.IMedia

@Composable
fun TvSeriesPlayerScreen(
    accessToken: String,
    onBack: () -> Unit,
    viewModel: TvSeriesPlayerViewModel = hiltViewModel(),
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
        BackHandler { handlePlaybackBack(isMedia3DolbyVisionRoute = false) }
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

    val currentEpisode = selectedEpisode(uiState)
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
    val isVlcRoute = uiState.canPlayCurrentEpisode &&
        uiState.currentSourceUrl.isNotBlank() &&
        playbackRoute.kind == TvPlaybackRouteKind.VLC
    val isMedia3Route = uiState.canPlayCurrentEpisode &&
        uiState.currentSourceUrl.isNotBlank() &&
        playbackRoute.kind == TvPlaybackRouteKind.MEDIA3_DOLBY_VISION
    val canPlay = isVlcRoute || isMedia3Route
    val libVLC = remember { TvVlcLibrary.shared(context) }
    val mediaPlayer = remember(accessToken) { newLongFormMediaPlayer(libVLC) }
    val latestUiState by rememberUpdatedState(uiState)
    val latestCurrentVideoId by rememberUpdatedState(uiState.currentVideoId)
    val latestIsMedia3Route by rememberUpdatedState(isMedia3Route)
    var media3Snapshot by remember { mutableStateOf(TvMedia3PlaybackSnapshot(positionMs = 0L, durationMs = 0L)) }
    val latestMedia3Snapshot by rememberUpdatedState(media3Snapshot)
    var hasStartedPlayback by rememberSaveable(uiState.currentVideoId) { mutableStateOf(false) }
    var isPausedByUser by rememberSaveable(uiState.currentVideoId) { mutableStateOf(false) }
    var preparedUrl by remember(uiState.currentVideoId) { mutableStateOf<String?>(null) }
    var appliedSubtitleSlaveUrl by remember(uiState.currentVideoId) { mutableStateOf<String?>(null) }
    var isVlcPlaying by remember(uiState.currentVideoId) { mutableStateOf(false) }
    var selectedSubtitleTrackId by rememberSaveable(uiState.currentVideoId) { mutableStateOf<String?>(null) }
    var selectedAudioTrackId by rememberSaveable(uiState.currentVideoId) { mutableStateOf<String?>(null) }
    var media3AudioTracks by remember(uiState.currentVideoId) { mutableStateOf(emptyList<LongFormAudioTrack>()) }
    var isPlayerActuallyPlaying by remember(uiState.currentVideoId) { mutableStateOf(false) }
    var playerErrorMessage by remember(uiState.currentVideoId) { mutableStateOf<String?>(null) }
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
    var lastHistoryUsesMedia3Snapshot by remember { mutableStateOf(false) }
    var resumedFromHistoryVideoId by remember { mutableStateOf("") }
    var resumePromptLastPositionMs by remember(uiState.currentVideoId) { mutableStateOf(0L) }
    var resumePromptRemainingMs by remember(uiState.currentVideoId) { mutableStateOf(0L) }
    var resumePromptDismissed by remember(uiState.currentVideoId) { mutableStateOf(false) }
    var isTrackSheetVisible by remember(uiState.currentVideoId) { mutableStateOf(false) }
    var media3TrackPickerKind by remember(uiState.currentVideoId) { mutableStateOf<TvMedia3TrackPickerKind?>(null) }
    var lastAutoplaySwitchedVideoId by remember { mutableStateOf("") }
    var openEpisodeRailRequestKey by remember(uiState.currentVideoId) { mutableStateOf(0) }

    val nextEpisodeRef = remember(uiState.series, uiState.selectedSeasonNumber, uiState.selectedEpisodeNumber) {
        uiState.nextEpisodeRef()
    }
    val episodeRailItems = remember(uiState.series, uiState.selectedSeasonNumber, uiState.selectedEpisodeNumber) {
        selectedSeason(uiState)?.episodes.orEmpty().map { episode ->
            TvEpisodeRailItem(
                id = episode.id,
                number = episode.number,
                title = episode.title,
                playable = episode.playable,
                current = episode.number == uiState.selectedEpisodeNumber,
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

    fun updatePlaybackSession(nextSession: LongFormPlaybackSession) {
        hasStartedPlayback = nextSession.hasStartedPlayback
        isPausedByUser = nextSession.isPausedByUser
    }

    val resumePromptGuardInput = ResumePromptGuardInput(
        hasResumeSeekTriggered = resumedFromHistoryVideoId == uiState.currentVideoId && resumePromptLastPositionMs > 0L,
        promptPermanentlyDismissed = resumePromptDismissed,
        isPlayerError = playerErrorMessage != null,
        isBackConfirmVisible = showBackConfirmPrompt,
        isEpisodeSelectorVisible = uiState.selectorVisible,
        isTrackSheetVisible = isTrackSheetVisible,
        isEndOverlayVisible = uiState.pendingEndOverlayKind != null,
        isAutoplayPromptVisible = shouldShowAutoplayPromptCard,
        isPausedByUser = playbackSession.isPausedByUser,
        remainingMs = resumePromptRemainingMs,
    )
    val shouldTickResumePromptCountdown = shouldTickResumePromptCountdown(resumePromptGuardInput)
    val shouldShowResumePromptCard = shouldShowResumePromptCard(resumePromptGuardInput)
    var media3SeekPositionMs by remember(uiState.currentVideoId) { mutableStateOf<Long?>(null) }
    var media3SeekRequestKey by remember(uiState.currentVideoId) { mutableStateOf(0) }
    val media3SubtitleConfigurations = remember(currentEpisode?.subtitleTracks, uiState.baseUrl, accessToken) {
        buildTvMedia3SubtitleConfigurations(
            tracks = currentEpisode?.subtitleTracks.orEmpty(),
            baseUrl = uiState.baseUrl,
            accessToken = accessToken,
        )
    }

    fun reportCurrentEpisodeHistory(completedOverride: Boolean? = null) {
        val videoId = latestUiState.currentVideoId
        if (videoId.isBlank()) {
            return
        }
        reportTvSeriesHistory(
            viewModel = viewModel,
            videoId = videoId,
            mediaPlayer = mediaPlayer,
            media3Snapshot = latestMedia3Snapshot,
            useMedia3Snapshot = shouldUseMedia3SnapshotForTvSeriesHistory(
                currentVideoId = uiState.currentVideoId,
                targetVideoId = videoId,
                currentRouteIsMedia3 = isMedia3Route,
                lastHistoryVideoId = lastHistoryVideoId,
                lastHistoryUsesMedia3Snapshot = lastHistoryUsesMedia3Snapshot,
            ),
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
            reportTvSeriesHistory(
                viewModel = viewModel,
                videoId = lastHistoryVideoId,
                mediaPlayer = mediaPlayer,
                media3Snapshot = latestMedia3Snapshot,
                useMedia3Snapshot = lastHistoryUsesMedia3Snapshot,
            )
        }
        lastHistoryVideoId = uiState.currentVideoId
        lastHistoryUsesMedia3Snapshot = isMedia3Route
        media3Snapshot = TvMedia3PlaybackSnapshot(positionMs = 0L, durationMs = 0L)
        screenPositionMs = 0L
        screenDurationMs = 0L
        resumePromptLastPositionMs = 0L
        resumePromptRemainingMs = 0L
        resumePromptDismissed = false
    }

    LaunchedEffect(uiState.currentSourceUrl, isVlcRoute) {
        val updateDecision = resolveLongFormPlayerUpdate(
            preparedUrl = preparedUrl,
            nextUrl = uiState.currentSourceUrl.takeIf { isVlcRoute },
        )
        if (!isVlcRoute) {
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
            val initialResumePositionMs = if (
                !uiState.startCurrentEpisodeFromBeginning &&
                resumedFromHistoryVideoId != uiState.currentVideoId
            ) {
                currentEpisode?.watchSeconds?.coerceAtLeast(0)?.times(1000L) ?: 0L
            } else {
                0L
            }
            appliedSubtitleSlaveUrl = null
            isVlcPlaying = false
            applyLongFormMediaSource(
                libVLC = libVLC,
                mediaPlayer = mediaPlayer,
                sourceUrl = uiState.currentSourceUrl,
                accessToken = accessToken,
            )
            mediaPlayer.play()
            if (initialResumePositionMs > 0L) {
                delay(250L)
                mediaPlayer.time = initialResumePositionMs
            }
            if (uiState.currentVideoId.isNotBlank()) {
                resumedFromHistoryVideoId = uiState.currentVideoId
                if (shouldTriggerResumePrompt(initialResumePositionMs)) {
                    resumePromptLastPositionMs = initialResumePositionMs
                    resumePromptRemainingMs = TvResumePromptTokens.CountdownDurationMs
                    resumePromptDismissed = false
                }
            }
            preparedUrl = uiState.currentSourceUrl
            hasStartedPlayback = true
            isPausedByUser = false
        }
    }

    var media3AutoStartedSourceUrl by remember(uiState.currentVideoId) { mutableStateOf("") }
    LaunchedEffect(uiState.currentSourceUrl, isMedia3Route) {
        if (
            shouldAutoStartTvDolbyVisionMedia3Playback(
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

    LaunchedEffect(isVlcPlaying, selectedSubtitleTrackId, currentEpisode?.subtitleTracks, uiState.baseUrl, accessToken) {
        if (!isVlcPlaying) return@LaunchedEffect
        val effectiveSubtitleTrackId = normalizeTvSubtitleSelection(selectedSubtitleTrackId)
        val track = resolveSelectedSubtitleTrack(currentEpisode?.subtitleTracks.orEmpty(), effectiveSubtitleTrackId)
        val subtitleUrl = track
            ?.takeIf { it.available && it.url.isNotBlank() }
            ?.let { resolvePlaybackAssetUrl(uiState.baseUrl, it.url) }
            ?.let { appendAccessTokenQuery(it, accessToken) }
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

    LaunchedEffect(playbackSession.hasStartedPlayback, playbackSession.isPausedByUser, isVlcRoute) {
        if (!isVlcRoute || !playbackSession.hasStartedPlayback) {
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
        uiState.currentVideoId,
        isVlcRoute,
        playbackSession.hasStartedPlayback,
        playbackSession.isPausedByUser,
    ) {
        if (!shouldStartPeriodicHistoryReport(
                videoId = uiState.currentVideoId,
                canPlay = isVlcRoute,
                hasStartedPlayback = playbackSession.hasStartedPlayback,
                isPausedByUser = playbackSession.isPausedByUser,
            )
        ) {
            return@LaunchedEffect
        }
        while (true) {
            delay(TvPeriodicHistoryReportIntervalMillis)
            reportTvSeriesHistory(viewModel, uiState.currentVideoId, mediaPlayer)
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

    LaunchedEffect(uiState.currentVideoId, isVlcRoute) {
        while (isVlcRoute) {
            screenPositionMs = mediaPlayer.time.coerceAtLeast(0L)
            screenDurationMs = mediaPlayer.length.coerceAtLeast(0L)
            delay(250L)
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
        showBackConfirmPrompt,
        uiState.selectorVisible,
        isTrackSheetVisible,
        uiState.pendingEndOverlayKind,
        shouldShowAutoplayPromptCard,
    ) {
        if (
            playerErrorMessage != null ||
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

    DisposableEffect(mediaPlayer) {
        onDispose {
            reportTvSeriesHistory(
                viewModel = viewModel,
                videoId = latestCurrentVideoId,
                mediaPlayer = mediaPlayer,
                media3Snapshot = latestMedia3Snapshot,
                useMedia3Snapshot = isMedia3Route,
            )
            mediaPlayer.release()
        }
    }

    DisposableEffect(lifecycleOwner, mediaPlayer, playbackSession.hasStartedPlayback, playbackSession.isPausedByUser, isVlcRoute) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_PAUSE -> {
                    reportTvSeriesHistory(
                        viewModel = viewModel,
                        videoId = latestCurrentVideoId,
                        mediaPlayer = mediaPlayer,
                        media3Snapshot = latestMedia3Snapshot,
                        useMedia3Snapshot = latestIsMedia3Route,
                    )
                    mediaPlayer.pause()
                }

                Lifecycle.Event.ON_RESUME -> {
                    if (playbackSession.shouldResumeOnLifecycle() && isVlcRoute) {
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

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black),
    ) {
        if (isVlcRoute) {
            Box(modifier = Modifier.fillMaxSize()) {
                LongFormVideoPlayer(
                    title = series.title.ifBlank { currentEpisode?.title.orEmpty() },
                    player = mediaPlayer,
                    isFullscreen = true,
                    onBack = onBack,
                    onTogglePlayPause = {
                        updatePlaybackSession(playbackSession.togglePlayPause(canPlay = canPlay))
                    },
                    onToggleFullscreen = {},
                    modifier = Modifier.fillMaxSize(),
                    subtitleTracks = currentEpisode?.subtitleTracks.orEmpty(),
                    selectedSubtitleTrackId = normalizeTvSubtitleSelection(selectedSubtitleTrackId),
                    onSelectSubtitleTrack = {
                        selectedSubtitleTrackId = it ?: ""
                        viewModel.selectSubtitleTrack(it)
                    },
                    selectedAudioTrackId = selectedAudioTrackId,
                    selectedAudioPreference = uiState.selectedAudioPreference,
                    onSelectAudioTrack = { trackId, preference, isUserAction ->
                        selectedAudioTrackId = trackId ?: ""
                        if (isUserAction) {
                            viewModel.selectAudioTrack(trackId, preference)
                        }
                    },
                    showStatusBarPadding = false,
                    tvMode = true,
                    tvSeekStepSeconds = uiState.tvSeekStepSeconds,
                    tvControlsVariant = TvLongFormControlsVariant.SeriesEpisodeRail,
                    seriesTitleForOverlay = series.title,
                    seasonNumber = uiState.selectedSeasonNumber,
                    episodeNumber = uiState.selectedEpisodeNumber,
                    episodeTitle = currentEpisode?.title,
                    episodeRailItems = episodeRailItems,
                    currentEpisodeRailItemId = currentEpisode?.id,
                    onSelectEpisodeRailItem = { selectedItem ->
                        val episodeNumber = selectedSeason(uiState)?.episodes
                            ?.firstOrNull { it.id == selectedItem.id }
                            ?.number
                        if (episodeNumber != null) {
                            selectEpisodeFromPlayer(episodeNumber)
                        }
                    },
                    onEpisodeRailVisibilityChanged = viewModel::setSelectorVisible,
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
                            MediaPlayer.Event.Stopped -> {
                                isPlayerActuallyPlaying = false
                                isVlcPlaying = false
                            }
                            MediaPlayer.Event.EndReached -> {
                                isPlayerActuallyPlaying = false
                                isVlcPlaying = false
                                handlePlaybackEnded()
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
                    playerErrorVisible = playerErrorMessage != null,
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
                    onBackToDetail = { exitToDetailWithDvCover(isMedia3DolbyVisionRoute = false) },
                    modifier = Modifier.fillMaxSize(),
                )
                TvPlayerErrorBanner(
                    message = playerErrorMessage,
                    modifier = Modifier
                        .align(Alignment.BottomCenter)
                        .padding(12.dp),
                )
                if (showBackConfirmPrompt) {
                    TvPlayerBackConfirmPrompt(
                        modifier = Modifier
                            .align(Alignment.BottomCenter)
                            .padding(bottom = 72.dp),
                    )
                }
            }
        } else if (isMedia3Route) {
            Box(modifier = Modifier.fillMaxSize()) {
                TvDolbyVisionMedia3Player(
                    sourceUrl = uiState.currentSourceUrl,
                    mediaId = uiState.currentVideoId,
                    title = series.title.ifBlank { currentEpisode?.title.orEmpty() },
                    accessToken = accessToken,
                    retryKey = routeRetryNonce,
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
                    subtitleConfigurations = media3SubtitleConfigurations,
                    selectedSubtitleTrackId = normalizeTvSubtitleSelection(selectedSubtitleTrackId),
                    selectedAudioTrackId = selectedAudioTrackId,
                    modifier = Modifier.fillMaxSize(),
                    onPlayingChanged = { playing ->
                        isPlayerActuallyPlaying = playing
                        if (playing) {
                            playerErrorMessage = null
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
                        playerErrorMessage = message
                        isPlayerActuallyPlaying = false
                        updatePlaybackSession(playbackSession.copy(hasStartedPlayback = false))
                    },
                    onEnded = ::handlePlaybackEnded,
                    onSnapshotChanged = { snapshot ->
                        media3Snapshot = snapshot
                        screenPositionMs = snapshot.positionMs
                        screenDurationMs = snapshot.durationMs
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
                    onSelectedAudioTrackChanged = { runtimeTrackId ->
                        if (runtimeTrackId != null && runtimeTrackId != selectedAudioTrackId) {
                            selectedAudioTrackId = runtimeTrackId
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
                    seasonNumber = uiState.selectedSeasonNumber,
                    episodeNumber = uiState.selectedEpisodeNumber,
                    episodeTitle = currentEpisode?.title,
                    episodeRailItems = episodeRailItems,
                    currentEpisodeRailItemId = currentEpisode?.id,
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
                        val episodeNumber = selectedSeason(uiState)?.episodes
                            ?.firstOrNull { it.id == selectedItem.id }
                            ?.number
                        if (episodeNumber != null) {
                            selectEpisodeFromPlayer(episodeNumber)
                        }
                    },
                    onEpisodeRailVisibilityChanged = viewModel::setSelectorVisible,
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
                    playerErrorVisible = playerErrorMessage != null,
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
                    onBackToDetail = { exitToDetailWithDvCover(isMedia3DolbyVisionRoute = true) },
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
        } else if (uiState.playbackPreparing) {
            Box(modifier = Modifier.fillMaxSize()) {
                TvPageLoadingState(message = "正在准备当前分集")
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
        if (pendingDvExitToDetail) {
            TvDolbyVisionExitToDetailCover()
        }
    }

}

@Composable
private fun TvPlayerErrorBanner(
    message: String?,
    modifier: Modifier = Modifier,
) {
    if (message.isNullOrBlank()) {
        return
    }
    Surface(
        modifier = modifier,
        color = Color(0xCC2B0F12),
        shape = AppChrome.SurfaceShape,
    ) {
        Text(
            text = message,
            modifier = Modifier.padding(horizontal = 12.dp, vertical = 10.dp),
            color = Color.White,
            style = MaterialTheme.typography.bodySmall,
        )
    }
}

private fun reportTvSeriesHistory(
    viewModel: TvSeriesPlayerViewModel,
    videoId: String,
    mediaPlayer: MediaPlayer,
    media3Snapshot: TvMedia3PlaybackSnapshot = TvMedia3PlaybackSnapshot(positionMs = 0L, durationMs = 0L),
    useMedia3Snapshot: Boolean = false,
    completedOverride: Boolean? = null,
) {
    val snapshot = tvPlaybackHistorySnapshot(
        positionMs = if (useMedia3Snapshot) media3Snapshot.positionMs else mediaPlayer.time,
        durationMs = if (useMedia3Snapshot) media3Snapshot.durationMs else mediaPlayer.length,
    )
    viewModel.reportHistory(videoId, snapshot.watchSeconds, completedOverride ?: snapshot.completed)
}

private fun reportTvSeriesMedia3History(
    viewModel: TvSeriesPlayerViewModel,
    videoId: String,
    playerSnapshot: TvMedia3PlaybackSnapshot,
) {
    val snapshot = tvPlaybackHistorySnapshot(
        positionMs = playerSnapshot.positionMs,
        durationMs = playerSnapshot.durationMs,
    )
    viewModel.reportHistory(videoId, snapshot.watchSeconds, snapshot.completed)
}

internal fun shouldUseMedia3SnapshotForTvSeriesHistory(
    currentVideoId: String,
    targetVideoId: String,
    currentRouteIsMedia3: Boolean,
    lastHistoryVideoId: String,
    lastHistoryUsesMedia3Snapshot: Boolean,
): Boolean =
    when (targetVideoId) {
        currentVideoId -> currentRouteIsMedia3
        lastHistoryVideoId -> lastHistoryUsesMedia3Snapshot
        else -> currentRouteIsMedia3
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

internal fun shouldAutoStartTvDolbyVisionMedia3Playback(
    currentSourceUrl: String,
    isMedia3Route: Boolean,
    autoStartedSourceUrl: String,
): Boolean =
    isMedia3Route &&
        currentSourceUrl.isNotBlank() &&
        autoStartedSourceUrl != currentSourceUrl
