package com.chee.videos.feature.tv

import android.view.KeyEvent as AndroidKeyEvent
import androidx.activity.compose.BackHandler
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInHorizontally
import androidx.compose.animation.slideOutHorizontally
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.focusable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.List
import androidx.compose.material.icons.filled.Audiotrack
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.FastForward
import androidx.compose.material.icons.filled.FastRewind
import androidx.compose.material.icons.filled.Pause
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.SkipNext
import androidx.compose.material.icons.filled.Subtitles
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableLongStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.input.key.onPreviewKeyEvent
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.chee.videos.core.model.SubtitleTrackDto
import com.chee.videos.core.model.TvSubtitlePreferenceMode
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.LongFormAudioTrack
import com.chee.videos.core.ui.subtitleTrackDisplayLabel
import com.chee.videos.core.ui.tvFocusableScaleOnly
import com.chee.videos.core.ui.tryRequestFocus
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlin.math.abs

private val TvLongFormSafeHorizontal = 64.dp
private val TvLongFormSafeVertical = 40.dp
private val TvLongFormPanelWidth = 440.dp
private const val TvLongFormSeekCommitDelayMillis: Long = 300L
private const val TvLongFormSubtitleAutoOptionId = "__tv_subtitle_auto__"
private const val TvLongFormSubtitleOffOptionId = "__tv_subtitle_off__"

internal data class TvLongFormEpisodeOption(
    val id: String,
    val seasonNumber: Int,
    val episodeNumber: Int,
    val title: String,
    val progressPercent: Int,
    val playable: Boolean,
    val current: Boolean,
)

internal data class TvLongFormSeasonOption(
    val number: Int,
    val title: String,
    val episodes: List<TvLongFormEpisodeOption>,
)

@Composable
internal fun TvLongFormStartupLoadingCanvas(
    modifier: Modifier = Modifier,
) {
    var feedbackVisible by remember { mutableStateOf(false) }
    LaunchedEffect(Unit) {
        delay(TvLongFormStartupFeedbackDelayMillis)
        feedbackVisible = true
    }
    Box(
        modifier = modifier
            .fillMaxSize()
            .background(Color.Black),
        contentAlignment = Alignment.Center,
    ) {
        if (feedbackVisible) {
            CircularProgressIndicator(
                color = Color.White,
                strokeWidth = 3.dp,
                modifier = Modifier.size(42.dp),
            )
        }
    }
}

@Composable
internal fun TvLongFormPlaybackChrome(
    title: String,
    secondaryTitle: String?,
    isPlaying: Boolean,
    positionMs: Long,
    durationMs: Long,
    tvSeekStepSeconds: Int,
    subtitleTracks: List<SubtitleTrackDto>,
    selectedSubtitleTrackId: String?,
    subtitlePreferenceMode: TvSubtitlePreferenceMode,
    audioTracks: List<LongFormAudioTrack>,
    selectedAudioTrackId: String?,
    seasons: List<TvLongFormSeasonOption>,
    currentEpisodeId: String?,
    onTogglePlayPause: () -> Unit,
    onSetPlaybackIntent: (Boolean) -> Unit,
    onSeekTo: (Long) -> Unit,
    onSelectSubtitleTrack: (TvSubtitlePreferenceMode, String?) -> Unit,
    onSelectAudioTrack: (String?) -> Unit,
    onSelectEpisode: (TvLongFormEpisodeOption) -> Unit,
    onPlayNextEpisode: (() -> Unit)?,
    onExitPlayback: () -> Unit,
    modifier: Modifier = Modifier,
    resumeNoticeText: String? = null,
    loadingFeedback: TvLongFormLoadingFeedback = TvLongFormLoadingFeedback.Hidden,
    blockingUiVisible: Boolean = false,
    onInteractionModeChanged: (TvLongFormInteractionMode) -> Unit = {},
) {
    val scope = rememberCoroutineScope()
    val rootFocusRequester = remember { FocusRequester() }
    val timelineFocusRequester = remember { FocusRequester() }
    var mode by remember { mutableStateOf(TvLongFormInteractionMode.Hidden) }
    var lastInteractionAt by remember { mutableLongStateOf(0L) }
    var pendingStepSeek by remember { mutableStateOf<TvChromePendingSeek?>(null) }
    var precisionSeekStartMs by remember { mutableLongStateOf(0L) }
    var precisionSeekTargetMs by remember { mutableLongStateOf(0L) }
    var centerFeedback by remember { mutableStateOf<TvChromeCenterFeedback?>(null) }
    var focusedAction by remember { mutableStateOf(TvLongFormControlAction.PlayPause) }
    var panelSeasonNumber by remember(currentEpisodeId, seasons) {
        mutableStateOf(
            seasons.firstOrNull { season -> season.episodes.any { it.current } }?.number
                ?: seasons.firstOrNull()?.number,
        )
    }
    var seekCommitJob by remember { mutableStateOf<Job?>(null) }
    var centerFeedbackJob by remember { mutableStateOf<Job?>(null) }

    val capabilities = remember(subtitleTracks.size, audioTracks.size, seasons, onPlayNextEpisode) {
        TvLongFormPlaybackCapabilities(
            subtitleTrackCount = subtitleTracks.size,
            audioTrackCount = audioTracks.size,
            hasEpisodePicker = seasons.any { it.episodes.isNotEmpty() },
            hasNextEpisode = onPlayNextEpisode != null,
        )
    }
    val visibleActions = remember(capabilities) { buildTvLongFormVisibleActions(capabilities) }
    val actionFocusRequesters = remember(visibleActions) {
        visibleActions.associateWith { FocusRequester() }
    }
    val safeDurationMs = durationMs.coerceAtLeast(0L)
    val effectivePositionMs = when (mode) {
        TvLongFormInteractionMode.PrecisionSeek -> precisionSeekTargetMs
        else -> pendingStepSeek?.targetMs ?: positionMs.coerceAtLeast(0L)
    }

    fun updateMode(nextMode: TvLongFormInteractionMode) {
        mode = nextMode
        lastInteractionAt = android.os.SystemClock.uptimeMillis()
    }

    fun clampPosition(targetMs: Long): Long = if (safeDurationMs > 0L) {
        targetMs.coerceIn(0L, safeDurationMs)
    } else {
        targetMs.coerceAtLeast(0L)
    }

    fun showCenterFeedback(icon: ImageVector, text: String) {
        centerFeedback = TvChromeCenterFeedback(icon, text)
        centerFeedbackJob?.cancel()
        centerFeedbackJob = scope.launch {
            delay(900L)
            centerFeedback = null
        }
    }

    fun togglePlayback() {
        onTogglePlayPause()
        showCenterFeedback(
            icon = if (isPlaying) Icons.Filled.Pause else Icons.Filled.PlayArrow,
            text = if (isPlaying) "已暂停" else "继续播放",
        )
    }

    fun setPlaybackIntent(shouldPlay: Boolean) {
        onSetPlaybackIntent(shouldPlay)
        showCenterFeedback(
            icon = if (shouldPlay) Icons.Filled.PlayArrow else Icons.Filled.Pause,
            text = if (shouldPlay) "继续播放" else "已暂停",
        )
    }

    fun performStepSeek(deltaMs: Long) {
        val previous = pendingStepSeek
        val anchor = previous?.anchorMs ?: positionMs.coerceAtLeast(0L)
        val currentTarget = previous?.targetMs ?: anchor
        val target = clampPosition(currentTarget + deltaMs)
        pendingStepSeek = TvChromePendingSeek(anchorMs = anchor, targetMs = target)
        showCenterFeedback(
            icon = if (deltaMs >= 0L) Icons.Filled.FastForward else Icons.Filled.FastRewind,
            text = buildString {
                append(if (deltaMs >= 0L) "+" else "-")
                append(formatTvLongFormTime(abs(target - anchor)))
                append("  ")
                append(formatTvLongFormTime(target))
            },
        )
        seekCommitJob?.cancel()
        seekCommitJob = scope.launch {
            delay(TvLongFormSeekCommitDelayMillis)
            val pending = pendingStepSeek ?: return@launch
            onSeekTo(pending.targetMs)
            pendingStepSeek = null
        }
    }

    fun revealControls() {
        updateMode(TvLongFormInteractionMode.Controls)
    }

    fun hideControls() {
        updateMode(TvLongFormInteractionMode.Hidden)
        rootFocusRequester.tryRequestFocus()
    }

    fun beginPrecisionSeek() {
        precisionSeekStartMs = positionMs.coerceAtLeast(0L)
        precisionSeekTargetMs = positionMs.coerceAtLeast(0L)
        updateMode(TvLongFormInteractionMode.PrecisionSeek)
    }

    fun closePanel() {
        updateMode(TvLongFormInteractionMode.Controls)
        actionFocusRequesters[focusedAction]?.tryRequestFocus()
    }

    BackHandler(enabled = !blockingUiVisible) {
        when (resolveTvLongFormBackAction(mode)) {
            TvLongFormBackAction.ClosePanel -> closePanel()
            TvLongFormBackAction.CancelPrecisionSeek -> {
                precisionSeekTargetMs = precisionSeekStartMs
                updateMode(TvLongFormInteractionMode.Controls)
                actionFocusRequesters[TvLongFormControlAction.PlayPause]?.tryRequestFocus()
            }
            TvLongFormBackAction.HideControls -> hideControls()
            TvLongFormBackAction.ExitPlayback -> onExitPlayback()
        }
    }

    LaunchedEffect(mode) {
        onInteractionModeChanged(mode)
    }

    LaunchedEffect(blockingUiVisible) {
        if (blockingUiVisible && mode != TvLongFormInteractionMode.Hidden) {
            updateMode(TvLongFormInteractionMode.Hidden)
        }
    }

    LaunchedEffect(mode, blockingUiVisible) {
        if (blockingUiVisible) return@LaunchedEffect
        when (mode) {
            TvLongFormInteractionMode.Hidden -> rootFocusRequester.tryRequestFocus()
            TvLongFormInteractionMode.Controls -> {
                delay(16L)
                actionFocusRequesters[TvLongFormControlAction.PlayPause]?.tryRequestFocus()
            }
            TvLongFormInteractionMode.PrecisionSeek -> {
                delay(16L)
                timelineFocusRequester.tryRequestFocus()
            }
            else -> Unit
        }
    }

    LaunchedEffect(isPlaying, mode, lastInteractionAt, blockingUiVisible) {
        if (!isPlaying || mode != TvLongFormInteractionMode.Controls || blockingUiVisible) {
            return@LaunchedEffect
        }
        val startedAt = lastInteractionAt
        delay(TvLongFormControlsAutoHideMillis)
        if (
            startedAt == lastInteractionAt &&
            shouldAutoHideTvLongFormControls(
                isPlaying = isPlaying,
                mode = mode,
                idleMs = TvLongFormControlsAutoHideMillis,
            )
        ) {
            hideControls()
        }
    }

    LaunchedEffect(Unit) {
        rootFocusRequester.tryRequestFocus()
    }

    Box(
        modifier = modifier
            .fillMaxSize()
            .focusRequester(rootFocusRequester)
            .focusable()
            .onPreviewKeyEvent { event ->
                val nativeEvent = event.nativeKeyEvent
                if (nativeEvent.action != AndroidKeyEvent.ACTION_DOWN) {
                    return@onPreviewKeyEvent false
                }
                val keyCode = nativeEvent.keyCode
                when (resolveTvLongFormMediaPlaybackCommand(keyCode)) {
                    TvLongFormMediaPlaybackCommand.Play -> {
                        setPlaybackIntent(true)
                        return@onPreviewKeyEvent true
                    }
                    TvLongFormMediaPlaybackCommand.Pause -> {
                        setPlaybackIntent(false)
                        return@onPreviewKeyEvent true
                    }
                    TvLongFormMediaPlaybackCommand.Toggle -> {
                        togglePlayback()
                        return@onPreviewKeyEvent true
                    }
                    null -> Unit
                }
                if (keyCode == AndroidKeyEvent.KEYCODE_MEDIA_REWIND) {
                    performStepSeek(-tvSeekStepSeconds.coerceAtLeast(1) * 1_000L)
                    return@onPreviewKeyEvent true
                }
                if (keyCode == AndroidKeyEvent.KEYCODE_MEDIA_FAST_FORWARD) {
                    performStepSeek(tvSeekStepSeconds.coerceAtLeast(1) * 1_000L)
                    return@onPreviewKeyEvent true
                }
                if (blockingUiVisible || mode != TvLongFormInteractionMode.Hidden) {
                    return@onPreviewKeyEvent false
                }
                when (keyCode) {
                    AndroidKeyEvent.KEYCODE_DPAD_CENTER,
                    AndroidKeyEvent.KEYCODE_ENTER,
                    AndroidKeyEvent.KEYCODE_NUMPAD_ENTER,
                    -> {
                        togglePlayback()
                        revealControls()
                        true
                    }
                    AndroidKeyEvent.KEYCODE_DPAD_DOWN -> {
                        revealControls()
                        true
                    }
                    AndroidKeyEvent.KEYCODE_DPAD_LEFT -> {
                        performStepSeek(-tvSeekStepSeconds.coerceAtLeast(1) * 1_000L)
                        true
                    }
                    AndroidKeyEvent.KEYCODE_DPAD_RIGHT -> {
                        performStepSeek(tvSeekStepSeconds.coerceAtLeast(1) * 1_000L)
                        true
                    }
                    else -> false
                }
            },
    ) {
        AnimatedVisibility(
            visible = mode != TvLongFormInteractionMode.Hidden && !blockingUiVisible,
            enter = fadeIn(tween(180)),
            exit = fadeOut(tween(160)),
            modifier = Modifier.fillMaxSize(),
        ) {
            TvLongFormControlsLayer(
                title = title,
                secondaryTitle = secondaryTitle,
                isPlaying = isPlaying,
                positionMs = effectivePositionMs,
                durationMs = safeDurationMs,
                mode = mode,
                visibleActions = visibleActions,
                focusedAction = focusedAction,
                actionFocusRequesters = actionFocusRequesters,
                timelineFocusRequester = timelineFocusRequester,
                onTimelineFocused = ::beginPrecisionSeek,
                onTimelineKey = { keyCode ->
                    when (keyCode) {
                        AndroidKeyEvent.KEYCODE_DPAD_LEFT -> {
                            precisionSeekTargetMs = clampPosition(
                                precisionSeekTargetMs - tvSeekStepSeconds.coerceAtLeast(1) * 1_000L,
                            )
                            true
                        }
                        AndroidKeyEvent.KEYCODE_DPAD_RIGHT -> {
                            precisionSeekTargetMs = clampPosition(
                                precisionSeekTargetMs + tvSeekStepSeconds.coerceAtLeast(1) * 1_000L,
                            )
                            true
                        }
                        AndroidKeyEvent.KEYCODE_DPAD_CENTER,
                        AndroidKeyEvent.KEYCODE_ENTER,
                        AndroidKeyEvent.KEYCODE_NUMPAD_ENTER,
                        -> {
                            onSeekTo(precisionSeekTargetMs)
                            updateMode(TvLongFormInteractionMode.Controls)
                            actionFocusRequesters[TvLongFormControlAction.PlayPause]?.tryRequestFocus()
                            true
                        }
                        AndroidKeyEvent.KEYCODE_DPAD_DOWN -> {
                            precisionSeekTargetMs = precisionSeekStartMs
                            updateMode(TvLongFormInteractionMode.Controls)
                            actionFocusRequesters[TvLongFormControlAction.PlayPause]?.tryRequestFocus()
                            true
                        }
                        else -> false
                    }
                },
                onActionFocused = { action ->
                    focusedAction = action
                    lastInteractionAt = android.os.SystemClock.uptimeMillis()
                },
                onActionClick = { action ->
                    lastInteractionAt = android.os.SystemClock.uptimeMillis()
                    when (action) {
                        TvLongFormControlAction.Rewind -> performStepSeek(-tvSeekStepSeconds.coerceAtLeast(1) * 1_000L)
                        TvLongFormControlAction.PlayPause -> togglePlayback()
                        TvLongFormControlAction.Forward -> performStepSeek(tvSeekStepSeconds.coerceAtLeast(1) * 1_000L)
                        TvLongFormControlAction.NextEpisode -> onPlayNextEpisode?.invoke()
                        TvLongFormControlAction.Subtitles -> updateMode(TvLongFormInteractionMode.SubtitlePanel)
                        TvLongFormControlAction.AudioTracks -> updateMode(TvLongFormInteractionMode.AudioPanel)
                        TvLongFormControlAction.Episodes -> updateMode(TvLongFormInteractionMode.EpisodePanel)
                    }
                },
            )
        }

        centerFeedback?.let { feedback ->
            TvLongFormCenterFeedback(
                feedback = feedback,
                modifier = Modifier.align(Alignment.Center),
            )
        }

        if (!resumeNoticeText.isNullOrBlank() && mode == TvLongFormInteractionMode.Hidden && !blockingUiVisible) {
            Text(
                text = resumeNoticeText,
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.bodyMedium,
                modifier = Modifier
                    .align(Alignment.BottomStart)
                    .padding(start = TvLongFormSafeHorizontal, bottom = 72.dp)
                    .background(AppChrome.Canvas.copy(alpha = 0.72f), AppChrome.ChipShape)
                    .padding(horizontal = 14.dp, vertical = 9.dp),
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
        }

        when (loadingFeedback) {
            TvLongFormLoadingFeedback.Startup,
            TvLongFormLoadingFeedback.Buffering,
            -> CircularProgressIndicator(
                color = AppChrome.TextPrimary,
                strokeWidth = 3.dp,
                modifier = Modifier
                    .align(Alignment.Center)
                    .size(if (loadingFeedback == TvLongFormLoadingFeedback.Startup) 42.dp else 30.dp),
            )
            else -> Unit
        }

        TvLongFormRightPanel(
            mode = mode,
            subtitleTracks = subtitleTracks,
            selectedSubtitleTrackId = selectedSubtitleTrackId,
            subtitlePreferenceMode = subtitlePreferenceMode,
            audioTracks = audioTracks,
            selectedAudioTrackId = selectedAudioTrackId,
            seasons = seasons,
            selectedSeasonNumber = panelSeasonNumber,
            currentEpisodeId = currentEpisodeId,
            onSelectSubtitleTrack = onSelectSubtitleTrack,
            onSelectAudioTrack = onSelectAudioTrack,
            onSelectSeason = { panelSeasonNumber = it },
            onSelectEpisode = { option ->
                if (option.current) {
                    closePanel()
                } else {
                    onSelectEpisode(option)
                    updateMode(TvLongFormInteractionMode.Hidden)
                }
            },
        )
    }
}

@Composable
private fun TvLongFormControlsLayer(
    title: String,
    secondaryTitle: String?,
    isPlaying: Boolean,
    positionMs: Long,
    durationMs: Long,
    mode: TvLongFormInteractionMode,
    visibleActions: List<TvLongFormControlAction>,
    focusedAction: TvLongFormControlAction,
    actionFocusRequesters: Map<TvLongFormControlAction, FocusRequester>,
    timelineFocusRequester: FocusRequester,
    onTimelineFocused: () -> Unit,
    onTimelineKey: (Int) -> Boolean,
    onActionFocused: (TvLongFormControlAction) -> Unit,
    onActionClick: (TvLongFormControlAction) -> Unit,
) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.verticalGradient(
                    0f to Color.Black.copy(alpha = 0.72f),
                    0.24f to Color.Transparent,
                    0.62f to Color.Transparent,
                    1f to Color.Black.copy(alpha = 0.88f),
                ),
            ),
    ) {
        Column(
            modifier = Modifier
                .align(Alignment.TopStart)
                .padding(horizontal = TvLongFormSafeHorizontal, vertical = TvLongFormSafeVertical)
                .fillMaxWidth(0.72f),
            verticalArrangement = Arrangement.spacedBy(4.dp),
        ) {
            Text(
                text = title,
                color = Color.White,
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.SemiBold,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            if (!secondaryTitle.isNullOrBlank()) {
                Text(
                    text = secondaryTitle,
                    color = Color.White.copy(alpha = 0.74f),
                    style = MaterialTheme.typography.bodyLarge,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }

        Column(
            modifier = Modifier
                .align(Alignment.BottomCenter)
                .fillMaxWidth()
                .padding(horizontal = TvLongFormSafeHorizontal, vertical = TvLongFormSafeVertical),
            verticalArrangement = Arrangement.spacedBy(14.dp),
        ) {
            TvLongFormTimeline(
                positionMs = positionMs,
                durationMs = durationMs,
                seeking = mode == TvLongFormInteractionMode.PrecisionSeek,
                focusRequester = timelineFocusRequester,
                onFocused = onTimelineFocused,
                onKey = onTimelineKey,
            )
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(12.dp),
                verticalAlignment = Alignment.CenterVertically,
            ) {
                visibleActions.forEach { action ->
                    TvLongFormIconButton(
                        action = action,
                        isPlaying = isPlaying,
                        focused = focusedAction == action,
                        focusRequester = actionFocusRequesters.getValue(action),
                        onFocused = { onActionFocused(action) },
                        onMoveFocus = { move ->
                            val next = resolveTvLongFormHorizontalFocus(action, move, visibleActions)
                            actionFocusRequesters[next]?.tryRequestFocus()
                        },
                        onMoveUp = { timelineFocusRequester.tryRequestFocus() },
                        onClick = { onActionClick(action) },
                    )
                }
                Spacer(modifier = Modifier.weight(1f))
                Text(
                    text = focusedAction.displayLabel(isPlaying),
                    color = Color.White.copy(alpha = 0.76f),
                    style = MaterialTheme.typography.labelLarge,
                    maxLines = 1,
                )
            }
        }
    }
}

@Composable
private fun TvLongFormTimeline(
    positionMs: Long,
    durationMs: Long,
    seeking: Boolean,
    focusRequester: FocusRequester,
    onFocused: () -> Unit,
    onKey: (Int) -> Boolean,
) {
    var focused by remember { mutableStateOf(false) }
    val progress = if (durationMs > 0L) {
        (positionMs.toFloat() / durationMs.toFloat()).coerceIn(0f, 1f)
    } else {
        0f
    }
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .focusRequester(focusRequester)
            .onFocusChanged { state ->
                val nowFocused = state.isFocused || state.hasFocus
                if (nowFocused && !focused) onFocused()
                focused = nowFocused
            }
            .onPreviewKeyEvent { event ->
                if (event.nativeKeyEvent.action != AndroidKeyEvent.ACTION_DOWN) false
                else onKey(event.nativeKeyEvent.keyCode)
            }
            .focusable(),
        verticalArrangement = Arrangement.spacedBy(8.dp),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(if (focused || seeking) 8.dp else 5.dp)
                .background(Color.White.copy(alpha = 0.28f), AppChrome.PillShape),
        ) {
            Box(
                modifier = Modifier
                    .fillMaxWidth(progress)
                    .fillMaxHeight()
                    .background(AppChrome.Accent, AppChrome.PillShape),
            )
            if (focused || seeking) {
                Box(
                    modifier = Modifier
                        .align(Alignment.CenterStart)
                        .fillMaxWidth(progress.coerceAtLeast(0.001f))
                        .height(20.dp),
                    contentAlignment = Alignment.CenterEnd,
                ) {
                    Box(
                        modifier = Modifier
                            .size(18.dp)
                            .background(AppChrome.Accent, CircleShape),
                    )
                }
            }
        }
        Row(modifier = Modifier.fillMaxWidth()) {
            Text(
                text = formatTvLongFormTime(positionMs),
                color = Color.White,
                style = MaterialTheme.typography.labelLarge,
            )
            Spacer(modifier = Modifier.weight(1f))
            Text(
                text = formatTvLongFormTime(durationMs),
                color = Color.White.copy(alpha = 0.72f),
                style = MaterialTheme.typography.labelLarge,
            )
        }
    }
}

@Composable
private fun TvLongFormIconButton(
    action: TvLongFormControlAction,
    isPlaying: Boolean,
    focused: Boolean,
    focusRequester: FocusRequester,
    onFocused: () -> Unit,
    onMoveFocus: (TvLongFormFocusMove) -> Unit,
    onMoveUp: () -> Unit,
    onClick: () -> Unit,
) {
    val icon = action.icon(isPlaying)
    Surface(
        color = if (focused) AppChrome.Accent else Color.Transparent,
        contentColor = if (focused) AppChrome.Canvas else Color.White,
        shape = CircleShape,
        modifier = Modifier
            .size(56.dp)
            .focusRequester(focusRequester)
            .onFocusChanged { state -> if (state.isFocused || state.hasFocus) onFocused() }
            .onPreviewKeyEvent { event ->
                if (event.nativeKeyEvent.action != AndroidKeyEvent.ACTION_DOWN) {
                    return@onPreviewKeyEvent false
                }
                when (event.nativeKeyEvent.keyCode) {
                    AndroidKeyEvent.KEYCODE_DPAD_LEFT -> {
                        onMoveFocus(TvLongFormFocusMove.Left)
                        true
                    }
                    AndroidKeyEvent.KEYCODE_DPAD_RIGHT -> {
                        onMoveFocus(TvLongFormFocusMove.Right)
                        true
                    }
                    AndroidKeyEvent.KEYCODE_DPAD_UP -> {
                        onMoveUp()
                        true
                    }
                    AndroidKeyEvent.KEYCODE_DPAD_DOWN -> true
                    else -> false
                }
            }
            .tvFocusableScaleOnly(focusedScale = 1.08f)
            .clickable(onClick = onClick),
    ) {
        Box(contentAlignment = Alignment.Center) {
            Icon(
                imageVector = icon,
                contentDescription = action.displayLabel(isPlaying),
                tint = if (focused) AppChrome.Canvas else Color.White,
                modifier = Modifier.size(28.dp),
            )
        }
    }
}

@Composable
private fun TvLongFormCenterFeedback(
    feedback: TvChromeCenterFeedback,
    modifier: Modifier = Modifier,
) {
    Surface(
        color = Color.Black.copy(alpha = 0.70f),
        contentColor = Color.White,
        shape = CircleShape,
        modifier = modifier.size(116.dp),
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center,
        ) {
            Icon(
                imageVector = feedback.icon,
                contentDescription = null,
                tint = Color.White,
                modifier = Modifier.size(36.dp),
            )
            Text(
                text = feedback.text,
                color = Color.White,
                style = MaterialTheme.typography.labelMedium,
                maxLines = 2,
                overflow = TextOverflow.Ellipsis,
            )
        }
    }
}

@Composable
private fun TvLongFormRightPanel(
    mode: TvLongFormInteractionMode,
    subtitleTracks: List<SubtitleTrackDto>,
    selectedSubtitleTrackId: String?,
    subtitlePreferenceMode: TvSubtitlePreferenceMode,
    audioTracks: List<LongFormAudioTrack>,
    selectedAudioTrackId: String?,
    seasons: List<TvLongFormSeasonOption>,
    selectedSeasonNumber: Int?,
    currentEpisodeId: String?,
    onSelectSubtitleTrack: (TvSubtitlePreferenceMode, String?) -> Unit,
    onSelectAudioTrack: (String?) -> Unit,
    onSelectSeason: (Int) -> Unit,
    onSelectEpisode: (TvLongFormEpisodeOption) -> Unit,
) {
    val visible = mode in setOf(
        TvLongFormInteractionMode.SubtitlePanel,
        TvLongFormInteractionMode.AudioPanel,
        TvLongFormInteractionMode.EpisodePanel,
    )
    AnimatedVisibility(
        visible = visible,
        enter = slideInHorizontally(tween(220)) { it } + fadeIn(tween(180)),
        exit = slideOutHorizontally(tween(180)) { it } + fadeOut(tween(140)),
        modifier = Modifier.fillMaxSize(),
    ) {
        Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.CenterEnd) {
            Column(
                modifier = Modifier
                    .fillMaxHeight()
                    .width(TvLongFormPanelWidth)
                    .background(AppChrome.Canvas.copy(alpha = 0.98f))
                    .padding(
                        start = 28.dp,
                        end = TvLongFormSafeHorizontal,
                        top = TvLongFormSafeVertical,
                        bottom = TvLongFormSafeVertical,
                    ),
                verticalArrangement = Arrangement.spacedBy(18.dp),
            ) {
                when (mode) {
                    TvLongFormInteractionMode.SubtitlePanel -> TvLongFormSubtitlePanel(
                        tracks = subtitleTracks,
                        selectedTrackId = selectedSubtitleTrackId,
                        preferenceMode = subtitlePreferenceMode,
                        onSelect = onSelectSubtitleTrack,
                    )
                    TvLongFormInteractionMode.AudioPanel -> TvLongFormAudioPanel(
                        tracks = audioTracks,
                        selectedTrackId = selectedAudioTrackId,
                        onSelect = onSelectAudioTrack,
                    )
                    TvLongFormInteractionMode.EpisodePanel -> TvLongFormEpisodePanel(
                        seasons = seasons,
                        selectedSeasonNumber = selectedSeasonNumber,
                        currentEpisodeId = currentEpisodeId,
                        onSelectSeason = onSelectSeason,
                        onSelectEpisode = onSelectEpisode,
                    )
                    else -> Unit
                }
            }
        }
    }
}

@Composable
private fun TvLongFormSubtitlePanel(
    tracks: List<SubtitleTrackDto>,
    selectedTrackId: String?,
    preferenceMode: TvSubtitlePreferenceMode,
    onSelect: (TvSubtitlePreferenceMode, String?) -> Unit,
) {
    Text("字幕", color = Color.White, style = MaterialTheme.typography.headlineSmall, fontWeight = FontWeight.SemiBold)
    val options = remember(tracks, selectedTrackId, preferenceMode) {
        listOf<TvChromePanelOption>(
            TvChromePanelOption(
                TvLongFormSubtitleAutoOptionId,
                "自动选择",
                "跟随视频默认字幕",
                preferenceMode == TvSubtitlePreferenceMode.AUTO,
            ),
            TvChromePanelOption(
                TvLongFormSubtitleOffOptionId,
                "关闭字幕",
                "",
                preferenceMode == TvSubtitlePreferenceMode.OFF,
            ),
        ) + tracks.map {
            TvChromePanelOption(
                it.id,
                subtitleTrackDisplayLabel(it),
                "",
                preferenceMode == TvSubtitlePreferenceMode.SPECIFIC && it.id == selectedTrackId,
            )
        }
    }
    TvLongFormOptionList(
        options = options,
        onSelect = { option ->
            when (option.id) {
                TvLongFormSubtitleAutoOptionId -> onSelect(TvSubtitlePreferenceMode.AUTO, null)
                TvLongFormSubtitleOffOptionId -> onSelect(TvSubtitlePreferenceMode.OFF, null)
                else -> onSelect(TvSubtitlePreferenceMode.SPECIFIC, option.id)
            }
        },
    )
}

@Composable
private fun TvLongFormAudioPanel(
    tracks: List<LongFormAudioTrack>,
    selectedTrackId: String?,
    onSelect: (String?) -> Unit,
) {
    Text("音轨", color = Color.White, style = MaterialTheme.typography.headlineSmall, fontWeight = FontWeight.SemiBold)
    val options = remember(tracks, selectedTrackId) {
        listOf(
            TvChromePanelOption(null, "自动选择", "跟随视频默认音轨", selectedTrackId.isNullOrBlank()),
        ) + tracks.map {
            TvChromePanelOption(it.id, it.label, it.detail, it.id == selectedTrackId)
        }
    }
    TvLongFormOptionList(options = options, onSelect = { onSelect(it.id) })
}

@Composable
private fun TvLongFormOptionList(
    options: List<TvChromePanelOption>,
    onSelect: (TvChromePanelOption) -> Unit,
) {
    val selectedIndex = options.indexOfFirst { it.selected }.coerceAtLeast(0)
    val initialFocusRequester = remember { FocusRequester() }
    val listState = rememberLazyListState(initialFirstVisibleItemIndex = selectedIndex)
    LaunchedEffect(options) {
        delay(32L)
        initialFocusRequester.tryRequestFocus()
    }
    LazyColumn(
        state = listState,
        verticalArrangement = Arrangement.spacedBy(8.dp),
        modifier = Modifier.fillMaxWidth(),
    ) {
        itemsIndexed(options, key = { _, option -> option.id ?: "auto" }) { index, option ->
            TvLongFormPanelRow(
                title = option.title,
                supporting = option.supporting,
                selected = option.selected,
                enabled = true,
                onClick = { onSelect(option) },
                modifier = if (index == selectedIndex) {
                    Modifier.focusRequester(initialFocusRequester)
                } else {
                    Modifier
                },
            )
        }
    }
}

@Composable
private fun TvLongFormEpisodePanel(
    seasons: List<TvLongFormSeasonOption>,
    selectedSeasonNumber: Int?,
    currentEpisodeId: String?,
    onSelectSeason: (Int) -> Unit,
    onSelectEpisode: (TvLongFormEpisodeOption) -> Unit,
) {
    Text("选集", color = Color.White, style = MaterialTheme.typography.headlineSmall, fontWeight = FontWeight.SemiBold)
    val seasonFocusRequesters = remember(seasons) {
        seasons.associate { it.number to FocusRequester() }
    }
    LazyRow(horizontalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.fillMaxWidth()) {
        itemsIndexed(seasons, key = { _, season -> season.number }) { index, season ->
            val selected = season.number == selectedSeasonNumber
            Surface(
                color = if (selected) AppChrome.Accent else AppChrome.SurfaceElevated,
                contentColor = if (selected) AppChrome.Canvas else Color.White,
                shape = AppChrome.ChipShape,
                modifier = Modifier
                    .focusRequester(seasonFocusRequesters.getValue(season.number))
                    .onPreviewKeyEvent { event ->
                        if (event.nativeKeyEvent.action != AndroidKeyEvent.ACTION_DOWN) {
                            return@onPreviewKeyEvent false
                        }
                        val targetIndex = when (event.nativeKeyEvent.keyCode) {
                            AndroidKeyEvent.KEYCODE_DPAD_LEFT -> (index - 1).coerceAtLeast(0)
                            AndroidKeyEvent.KEYCODE_DPAD_RIGHT -> (index + 1).coerceAtMost(seasons.lastIndex)
                            else -> return@onPreviewKeyEvent false
                        }
                        seasonFocusRequesters[seasons[targetIndex].number]?.tryRequestFocus()
                        true
                    }
                    .tvFocusableScaleOnly(focusedScale = 1.04f)
                    .clickable { onSelectSeason(season.number) },
            ) {
                Text(
                    text = season.title.ifBlank { "第 ${season.number} 季" },
                    color = if (selected) AppChrome.Canvas else Color.White,
                    style = MaterialTheme.typography.labelLarge,
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 11.dp),
                    maxLines = 1,
                )
            }
        }
    }
    val episodes = seasons.firstOrNull { it.number == selectedSeasonNumber }?.episodes.orEmpty()
    val initialEpisodeIndex = episodes.indexOfFirst { it.id == currentEpisodeId || it.current }
        .takeIf { it >= 0 && episodes[it].playable }
        ?: episodes.indexOfFirst { it.playable }.coerceAtLeast(0)
    val initialFocusRequester = remember(selectedSeasonNumber) { FocusRequester() }
    val episodeListState = rememberLazyListState(initialFirstVisibleItemIndex = initialEpisodeIndex)
    LaunchedEffect(episodes, initialEpisodeIndex) {
        if (episodes.getOrNull(initialEpisodeIndex)?.playable == true) {
            episodeListState.scrollToItem(initialEpisodeIndex)
            delay(32L)
            initialFocusRequester.tryRequestFocus()
        }
    }
    LazyColumn(
        state = episodeListState,
        verticalArrangement = Arrangement.spacedBy(8.dp),
        modifier = Modifier.fillMaxWidth(),
    ) {
        itemsIndexed(episodes, key = { _, episode -> episode.id }) { index, episode ->
            TvLongFormPanelRow(
                title = "第 ${episode.episodeNumber} 集  ${episode.title}".trim(),
                supporting = when {
                    !episode.playable -> "暂不可播放"
                    episode.progressPercent > 0 -> "已观看 ${episode.progressPercent.coerceIn(0, 100)}%"
                    else -> ""
                },
                selected = episode.id == currentEpisodeId || episode.current,
                enabled = episode.playable,
                onClick = { onSelectEpisode(episode) },
                progressPercent = episode.progressPercent,
                modifier = if (index == initialEpisodeIndex && episode.playable) {
                    Modifier.focusRequester(initialFocusRequester)
                } else {
                    Modifier
                },
            )
        }
    }
}

@Composable
private fun TvLongFormPanelRow(
    title: String,
    supporting: String,
    selected: Boolean,
    enabled: Boolean,
    onClick: () -> Unit,
    progressPercent: Int = 0,
    modifier: Modifier = Modifier,
) {
    var focused by remember { mutableStateOf(false) }
    Surface(
        color = if (focused) AppChrome.SurfaceElevated else Color.Transparent,
        contentColor = if (enabled) Color.White else AppChrome.TextMuted,
        shape = AppChrome.SurfaceShape,
        border = if (focused) BorderStroke(2.dp, AppChrome.Accent) else null,
        modifier = modifier
            .fillMaxWidth()
            .onFocusChanged { focused = it.isFocused || it.hasFocus }
            .onPreviewKeyEvent { event ->
                event.nativeKeyEvent.action == AndroidKeyEvent.ACTION_DOWN &&
                    event.nativeKeyEvent.keyCode in setOf(
                        AndroidKeyEvent.KEYCODE_DPAD_LEFT,
                        AndroidKeyEvent.KEYCODE_DPAD_RIGHT,
                    )
            }
            .focusable(enabled)
            .clickable(enabled = enabled, onClick = onClick),
    ) {
        Column(
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 12.dp),
            verticalArrangement = Arrangement.spacedBy(5.dp),
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(
                    text = title,
                    color = if (enabled) Color.White else AppChrome.TextMuted,
                    style = MaterialTheme.typography.bodyLarge,
                    modifier = Modifier.weight(1f),
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                if (selected) {
                    Icon(
                        imageVector = Icons.Filled.Check,
                        contentDescription = "当前选中",
                        tint = AppChrome.Accent,
                        modifier = Modifier.size(22.dp),
                    )
                }
            }
            if (supporting.isNotBlank()) {
                Text(
                    text = supporting,
                    color = AppChrome.TextMuted,
                    style = MaterialTheme.typography.labelMedium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
            if (progressPercent > 0) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .height(3.dp)
                        .background(Color.White.copy(alpha = 0.18f), AppChrome.PillShape),
                ) {
                    Box(
                        modifier = Modifier
                            .fillMaxWidth(progressPercent.coerceIn(0, 100) / 100f)
                            .fillMaxHeight()
                            .background(AppChrome.Accent, AppChrome.PillShape),
                    )
                }
            }
        }
    }
}

private data class TvChromePendingSeek(val anchorMs: Long, val targetMs: Long)
private data class TvChromeCenterFeedback(val icon: ImageVector, val text: String)
private data class TvChromePanelOption(
    val id: String?,
    val title: String,
    val supporting: String,
    val selected: Boolean,
)

private fun TvLongFormControlAction.icon(isPlaying: Boolean): ImageVector = when (this) {
    TvLongFormControlAction.Rewind -> Icons.Filled.FastRewind
    TvLongFormControlAction.PlayPause -> if (isPlaying) Icons.Filled.Pause else Icons.Filled.PlayArrow
    TvLongFormControlAction.Forward -> Icons.Filled.FastForward
    TvLongFormControlAction.NextEpisode -> Icons.Filled.SkipNext
    TvLongFormControlAction.Subtitles -> Icons.Filled.Subtitles
    TvLongFormControlAction.AudioTracks -> Icons.Filled.Audiotrack
    TvLongFormControlAction.Episodes -> Icons.AutoMirrored.Filled.List
}

private fun TvLongFormControlAction.displayLabel(isPlaying: Boolean): String = when (this) {
    TvLongFormControlAction.Rewind -> "快退"
    TvLongFormControlAction.PlayPause -> if (isPlaying) "暂停" else "播放"
    TvLongFormControlAction.Forward -> "快进"
    TvLongFormControlAction.NextEpisode -> "下一集"
    TvLongFormControlAction.Subtitles -> "字幕"
    TvLongFormControlAction.AudioTracks -> "音轨"
    TvLongFormControlAction.Episodes -> "选集"
}

internal fun formatTvLongFormTime(positionMs: Long): String {
    val totalSeconds = (positionMs.coerceAtLeast(0L) / 1_000L)
    val hours = totalSeconds / 3_600L
    val minutes = (totalSeconds % 3_600L) / 60L
    val seconds = totalSeconds % 60L
    return if (hours > 0L) {
        "%d:%02d:%02d".format(hours, minutes, seconds)
    } else {
        "%02d:%02d".format(minutes, seconds)
    }
}
