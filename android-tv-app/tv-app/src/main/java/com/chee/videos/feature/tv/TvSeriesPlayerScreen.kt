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
import com.chee.videos.core.ui.LongFormVideoPlayer
import com.chee.videos.core.ui.TvEpisodeRailItem
import com.chee.videos.core.ui.TvErrorState
import com.chee.videos.core.ui.TvLongFormControlsVariant
import com.chee.videos.core.ui.TvPageLoadingState
import com.chee.videos.core.ui.appendAccessTokenQuery
import com.chee.videos.core.ui.applyLongFormMediaSource
import com.chee.videos.core.ui.resolveLongFormPlayerUpdate
import com.chee.videos.core.ui.resolvePlaybackAssetUrl
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

    val currentEpisode = selectedEpisode(uiState)
    val canPlay = uiState.canPlayCurrentEpisode && uiState.currentSourceUrl.isNotBlank()
    val libVLC = remember { TvVlcLibrary.shared(context) }
    val mediaPlayer = remember(accessToken) { newLongFormMediaPlayer(libVLC) }
    val latestUiState by rememberUpdatedState(uiState)
    val latestCurrentVideoId by rememberUpdatedState(uiState.currentVideoId)
    var hasStartedPlayback by rememberSaveable(uiState.currentVideoId) { mutableStateOf(false) }
    var isPausedByUser by rememberSaveable(uiState.currentVideoId) { mutableStateOf(false) }
    var preparedUrl by remember(uiState.currentVideoId) { mutableStateOf<String?>(null) }
    var appliedSubtitleSlaveUrl by remember(uiState.currentVideoId) { mutableStateOf<String?>(null) }
    var isVlcPlaying by remember(uiState.currentVideoId) { mutableStateOf(false) }
    var selectedSubtitleTrackId by rememberSaveable(uiState.currentVideoId) { mutableStateOf<String?>(null) }
    var selectedAudioTrackId by rememberSaveable(uiState.currentVideoId) { mutableStateOf<String?>(null) }
    var isPlayerActuallyPlaying by remember(uiState.currentVideoId) { mutableStateOf(false) }
    var playerErrorMessage by remember(uiState.currentVideoId) { mutableStateOf<String?>(null) }
    var screenPositionMs by remember(uiState.currentVideoId) { mutableStateOf(0L) }
    var screenDurationMs by remember(uiState.currentVideoId) { mutableStateOf(0L) }
    var lastHistoryVideoId by remember { mutableStateOf("") }
    var resumedFromHistoryVideoId by remember { mutableStateOf("") }
    var resumePromptLastPositionMs by remember(uiState.currentVideoId) { mutableStateOf(0L) }
    var resumePromptRemainingMs by remember(uiState.currentVideoId) { mutableStateOf(0L) }
    var resumePromptDismissed by remember(uiState.currentVideoId) { mutableStateOf(false) }
    var isTrackSheetVisible by remember(uiState.currentVideoId) { mutableStateOf(false) }
    var lastAutoplaySwitchedVideoId by remember { mutableStateOf("") }

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

    fun advanceFromAutoplay() {
        val state = latestUiState
        val videoId = state.currentVideoId
        if (videoId.isBlank() || lastAutoplaySwitchedVideoId == videoId || !state.hasNextPlayableEpisode()) {
            return
        }
        lastAutoplaySwitchedVideoId = videoId
        reportTvSeriesHistory(viewModel, videoId, mediaPlayer, completedOverride = true)
        lastHistoryVideoId = videoId
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
                reportTvSeriesHistory(viewModel, state.currentVideoId, mediaPlayer, completedOverride = true)
                viewModel.showEndOverlay(TvEndOverlayKind.CURRENT_FINISHED)
            }
            else -> {
                reportTvSeriesHistory(viewModel, state.currentVideoId, mediaPlayer, completedOverride = true)
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
            reportTvSeriesHistory(viewModel, lastHistoryVideoId, mediaPlayer)
        }
        lastHistoryVideoId = uiState.currentVideoId
        resumePromptLastPositionMs = 0L
        resumePromptRemainingMs = 0L
        resumePromptDismissed = false
    }

    LaunchedEffect(uiState.currentSourceUrl, canPlay) {
        val updateDecision = resolveLongFormPlayerUpdate(
            preparedUrl = preparedUrl,
            nextUrl = uiState.currentSourceUrl,
        )
        if (!canPlay) {
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

    LaunchedEffect(playbackSession.hasStartedPlayback, playbackSession.isPausedByUser, canPlay) {
        if (!playbackSession.hasStartedPlayback || !canPlay) {
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
        canPlay,
        playbackSession.hasStartedPlayback,
        playbackSession.isPausedByUser,
    ) {
        if (!shouldStartPeriodicHistoryReport(
                videoId = uiState.currentVideoId,
                canPlay = canPlay,
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

    LaunchedEffect(uiState.currentVideoId, canPlay) {
        while (canPlay) {
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
            reportTvSeriesHistory(viewModel, latestCurrentVideoId, mediaPlayer)
            mediaPlayer.release()
        }
    }

    DisposableEffect(lifecycleOwner, mediaPlayer, playbackSession.hasStartedPlayback, playbackSession.isPausedByUser, canPlay) {
        val observer = LifecycleEventObserver { _, event ->
            when (event) {
                Lifecycle.Event.ON_PAUSE -> {
                    reportTvSeriesHistory(viewModel, latestCurrentVideoId, mediaPlayer)
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

    KeepScreenOnEffect(enabled = isPlayerActuallyPlaying)

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color.Black),
    ) {
        if (canPlay) {
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
                            viewModel.selectEpisode(episodeNumber)
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
                    onBackToDetail = onBack,
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
        } else {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                EmptyTvPlayerState()
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

@Composable
private fun EmptyTvPlayerState() {
    Text(
        text = "当前分集暂无可播放视频",
        color = AppChrome.TextSecondary,
        style = MaterialTheme.typography.bodyLarge,
    )
}

private fun reportTvSeriesHistory(
    viewModel: TvSeriesPlayerViewModel,
    videoId: String,
    player: MediaPlayer,
    completedOverride: Boolean? = null,
) {
    val snapshot = tvPlaybackHistorySnapshot(
        positionMs = player.time,
        durationMs = player.length,
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
