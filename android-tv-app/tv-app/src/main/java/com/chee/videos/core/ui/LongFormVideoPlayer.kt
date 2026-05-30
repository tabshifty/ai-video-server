package com.chee.videos.core.ui

import android.util.Log
import android.view.KeyEvent as AndroidKeyEvent
import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.focusable
import androidx.compose.foundation.gestures.detectHorizontalDragGestures
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxScope
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.layout.wrapContentWidth
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.FastForward
import androidx.compose.material.icons.filled.Fullscreen
import androidx.compose.material.icons.filled.FullscreenExit
import androidx.compose.material.icons.filled.GraphicEq
import androidx.compose.material.icons.filled.Pause
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.SkipNext
import androidx.compose.material.icons.filled.Subtitles
import androidx.compose.material.icons.filled.Tv
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.foundation.lazy.LazyRow
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material3.Slider
import androidx.compose.material3.SliderDefaults
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.focus.focusProperties
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.input.key.onPreviewKeyEvent
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import coil.compose.AsyncImage
import com.chee.videos.core.model.SubtitleTrackDto
import com.chee.videos.core.model.TvTrackPreference
import com.chee.videos.core.player.VlcLongFormSurface
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlin.math.abs
import org.videolan.libvlc.MediaPlayer

private const val LongFormVideoPlayerLogTag = "LongFormVideoPlayer"

enum class TvLongFormControlsVariant {
    Default,
    SeriesEpisodeRail,
}

internal enum class TvControlFocusTarget {
    PlayPause,
    Rewind,
    Forward,
    EpisodeSelector,
    NextEpisode,
    Subtitle,
    AudioTrack,
    BackDetail,
    ExitPlayback,
    Fullscreen,
}

internal enum class TvControlFocusDirection {
    Left,
    Right,
}

internal fun resolveTvControlHorizontalFocusTarget(
    current: TvControlFocusTarget,
    direction: TvControlFocusDirection,
    targets: List<TvControlFocusTarget>,
): TvControlFocusTarget? {
    val index = targets.indexOf(current)
    if (index < 0 || targets.isEmpty()) {
        return null
    }
    return when (direction) {
        TvControlFocusDirection.Left -> targets[(index - 1 + targets.size) % targets.size]
        TvControlFocusDirection.Right -> targets[(index + 1) % targets.size]
    }
}

private enum class TvSeriesBottomPanelPage(val depth: Int) {
    Controls(0),
    EpisodeRail(1),
}

private object TvEpisodeRailLayoutTokens {
    const val ItemSlotWidthDp: Int = 120
    const val TooltipHeightDp: Int = 44
    const val TooltipMaxWidthDp: Int = 220
}

@OptIn(androidx.compose.material3.ExperimentalMaterial3Api::class)
@Composable
fun LongFormVideoPlayer(
    title: String,
    player: MediaPlayer,
    isFullscreen: Boolean,
    onBack: () -> Unit,
    onTogglePlayPause: () -> Unit,
    onToggleFullscreen: () -> Unit,
    modifier: Modifier = Modifier,
    showStatusBarPadding: Boolean = false,
    posterUrl: String? = null,
    showPoster: Boolean = false,
    subtitleTracks: List<SubtitleTrackDto> = emptyList(),
    selectedSubtitleTrackId: String? = null,
    onSelectSubtitleTrack: (String?) -> Unit = {},
    selectedAudioTrackId: String? = null,
    selectedAudioPreference: TvTrackPreference? = null,
    onSelectAudioTrack: (trackId: String?, preference: TvTrackPreference?, isUserAction: Boolean) -> Unit = { _, _, _ -> },
    tvMode: Boolean = false,
    tvSeekStepSeconds: Int = 10,
    tvControlsVariant: TvLongFormControlsVariant = TvLongFormControlsVariant.Default,
    seriesTitleForOverlay: String? = null,
    seasonNumber: Int? = null,
    episodeNumber: Int? = null,
    episodeTitle: String? = null,
    episodeRailItems: List<TvEpisodeRailItem> = emptyList(),
    currentEpisodeRailItemId: String? = null,
    onSelectEpisodeRailItem: ((TvEpisodeRailItem) -> Unit)? = null,
    onEpisodeRailVisibilityChanged: (Boolean) -> Unit = {},
    onOpenEpisodeSelector: (() -> Unit)? = null,
    onNextEpisode: (() -> Unit)? = null,
    onRequestExitPlayback: (() -> Unit)? = null,
    onExitPlayback: (() -> Unit)? = null,
    onTrackSheetVisibilityChanged: (Boolean) -> Unit = {},
    onVlcEvent: (MediaPlayer.Event) -> Unit = {},
    resumePromptVisible: Boolean = false,
    resumePromptSlot: (@Composable BoxScope.() -> Unit)? = null,
    backConfirmPromptVisible: Boolean = false,
    playerErrorVisible: Boolean = false,
) {
    val scope = rememberCoroutineScope()
    val rootFocusRequester = remember { FocusRequester() }
    val playPauseFocusRequester = remember { FocusRequester() }
    val rewindFocusRequester = remember { FocusRequester() }
    val forwardFocusRequester = remember { FocusRequester() }
    val episodeSelectorFocusRequester = remember { FocusRequester() }
    val nextEpisodeFocusRequester = remember { FocusRequester() }
    val subtitleFocusRequester = remember { FocusRequester() }
    val audioTrackFocusRequester = remember { FocusRequester() }
    val backDetailFocusRequester = remember { FocusRequester() }
    val exitPlaybackFocusRequester = remember { FocusRequester() }
    val fullscreenFocusRequester = remember { FocusRequester() }
    val currentEpisodeRailFocusRequester = remember { FocusRequester() }

    var controlsVisible by remember { mutableStateOf(false) }
    var episodeRailVisible by rememberSaveable(currentEpisodeRailItemId, tvControlsVariant) { mutableStateOf(false) }
    var episodeRailOpenNonce by remember(currentEpisodeRailItemId, tvControlsVariant) { mutableStateOf(0) }
    var focusedEpisodeRailItemId by rememberSaveable(currentEpisodeRailItemId) { mutableStateOf(currentEpisodeRailItemId.orEmpty()) }
    var isPlaying by remember { mutableStateOf(player.isPlaying) }
    var isVlcPlaying by remember { mutableStateOf(player.isPlaying) }
    var durationMs by remember { mutableStateOf(player.length.coerceAtLeast(0L)) }
    var positionMs by remember { mutableStateOf(player.time.coerceAtLeast(0L)) }
    var isScrubbing by remember { mutableStateOf(false) }
    var scrubPositionMs by remember { mutableStateOf(0L) }
    var viewWidthPx by remember { mutableStateOf(1f) }

    var draggingSeek by remember { mutableStateOf(false) }
    var dragStartPositionMs by remember { mutableStateOf(0L) }
    var dragTargetPositionMs by remember { mutableStateOf(0L) }
    var dragDistancePx by remember { mutableStateOf(0f) }
    var seekPreviewText by remember { mutableStateOf("") }
    var showSeekPreview by remember { mutableStateOf(false) }
    var pendingStepSeek by remember { mutableStateOf<TvPendingStepSeekUpdate?>(null) }

    var longPressBoosting by remember { mutableStateOf(false) }
    var ignoreTapAfterLongPress by remember { mutableStateOf(false) }

    var showCenterFeedback by remember { mutableStateOf(false) }
    var centerFeedbackText by remember { mutableStateOf("") }
    var centerFeedbackIcon by remember { mutableStateOf(Icons.Filled.PlayArrow) }
    var subtitleSheetVisible by remember { mutableStateOf(false) }
    var audioTrackSheetVisible by remember { mutableStateOf(false) }
    val isTrackSheetVisible = subtitleSheetVisible || audioTrackSheetVisible
    LaunchedEffect(isTrackSheetVisible) {
        onTrackSheetVisibilityChanged(isTrackSheetVisible)
    }
    LaunchedEffect(episodeRailVisible) {
        onEpisodeRailVisibilityChanged(episodeRailVisible)
    }
    var audioTracks by remember { mutableStateOf(emptyList<LongFormAudioTrack>()) }
    var hasShownControlsOnce by remember { mutableStateOf(false) }
    var pendingRootFocusRequest by remember { mutableStateOf(false) }
    var pendingControlFocusTarget by remember { mutableStateOf<TvControlFocusTarget?>(null) }
    var pendingEpisodeRailFocusRequest by remember { mutableStateOf(false) }
    var playerFocusLayer by remember { mutableStateOf(TvPlayerFocusLayer.Root) }
    var lastFocusedControlTarget by remember { mutableStateOf(TvControlFocusTarget.PlayPause) }
    val latestOnVlcEvent by rememberUpdatedState(onVlcEvent)

    var hideControlsJob by remember { mutableStateOf<Job?>(null) }
    var hideSeekPreviewJob by remember { mutableStateOf<Job?>(null) }
    var hideCenterFeedbackJob by remember { mutableStateOf<Job?>(null) }
    var commitStepSeekJob by remember { mutableStateOf<Job?>(null) }
    val tvSeekStepMs = normalizeTvSeekStepSeconds(tvSeekStepSeconds) * 1_000L

    fun effectiveDurationMs(): Long {
        val fromState = durationMs
        if (fromState > 0) {
            return fromState
        }
        return player.length.coerceAtLeast(0L)
    }

    fun requestRootFocusWhenReady() {
        pendingRootFocusRequest = true
    }

    fun hideEpisodeRail() {
        episodeRailVisible = false
        pendingEpisodeRailFocusRequest = false
    }

    fun requestControlFocusWhenReady(target: TvControlFocusTarget) {
        pendingControlFocusTarget = target
    }

    fun scheduleAutoHideControls() {
        hideControlsJob?.cancel()
        if (episodeRailVisible) {
            return
        }
        hideControlsJob = scope.launch {
            delay(5000)
            controlsVisible = false
            hideEpisodeRail()
            playerFocusLayer = TvPlayerFocusLayer.Root
            requestRootFocusWhenReady()
        }
    }

    fun showControlsTemporarily(requestControlFocus: TvControlFocusTarget? = null) {
        controlsVisible = true
        if (requestControlFocus != null) {
            requestControlFocusWhenReady(requestControlFocus)
        }
        scheduleAutoHideControls()
    }

    fun hideAllTvUi() {
        controlsVisible = false
        hideEpisodeRail()
        playerFocusLayer = TvPlayerFocusLayer.Root
        hideControlsJob?.cancel()
        requestRootFocusWhenReady()
    }

    fun showTransientFeedback(
        icon: ImageVector,
        text: String,
        durationMs: Long = 900L,
    ) {
        centerFeedbackIcon = icon
        centerFeedbackText = text
        showCenterFeedback = true
        hideCenterFeedbackJob?.cancel()
        hideCenterFeedbackJob = scope.launch {
            delay(durationMs)
            showCenterFeedback = false
        }
    }

    fun togglePlaybackWithFeedback(showControls: Boolean = true) {
        val shouldPause = isPlaying
        onTogglePlayPause()
        showTransientFeedback(
            icon = if (shouldPause) Icons.Filled.Pause else Icons.Filled.PlayArrow,
            text = if (shouldPause) "已暂停" else "继续播放",
        )
        if (showControls) {
            showControlsTemporarily()
        }
    }

    fun performStepSeek(deltaMs: Long, showControls: Boolean = true) {
        val duration = effectiveDurationMs()
        val current = player.time.coerceAtLeast(0L)
        val target = (current + deltaMs).coerceAtLeast(0L).let { next ->
            if (duration > 0L) next.coerceAtMost(duration) else next
        }
        player.time = target
        positionMs = target
        val delta = target - current
        val direction = if (delta >= 0) "快进" else "快退"
        val sign = if (delta >= 0) "+" else "-"
        seekPreviewText = buildString {
            append(direction)
            append(" ")
            append(sign)
            append(formatPlaybackTime(abs(delta)))
            append("\n")
            append(formatPlaybackTime(target))
            if (duration > 0) {
                append(" / ")
                append(formatPlaybackTime(duration))
            }
        }
        showSeekPreview = true
        hideSeekPreviewJob?.cancel()
        hideSeekPreviewJob = scope.launch {
            delay(900)
            showSeekPreview = false
        }
        if (showControls) {
            showControlsTemporarily()
        }
    }

    fun updateSeekPreview(anchor: Long, target: Long, duration: Long) {
        val delta = target - anchor
        val direction = if (delta >= 0) "快进" else "快退"
        val sign = if (delta >= 0) "+" else "-"
        seekPreviewText = buildString {
            append(direction)
            append(" ")
            append(sign)
            append(formatPlaybackTime(abs(delta)))
            append("\n")
            append(formatPlaybackTime(target))
            if (duration > 0) {
                append(" / ")
                append(formatPlaybackTime(duration))
            }
        }
        showSeekPreview = true
    }

    fun performDebouncedStepSeek(deltaMs: Long, showControls: Boolean = true) {
        val duration = effectiveDurationMs()
        val pending = resolveTvPendingStepSeek(
            previous = pendingStepSeek,
            currentPositionMs = player.time.coerceAtLeast(0L),
            durationMs = duration,
            deltaMs = deltaMs,
        )
        pendingStepSeek = pending
        positionMs = pending.targetPositionMs
        updateSeekPreview(
            anchor = pending.anchorPositionMs,
            target = pending.targetPositionMs,
            duration = duration,
        )
        hideSeekPreviewJob?.cancel()
        commitStepSeekJob?.cancel()
        commitStepSeekJob = scope.launch {
            delay(TvStepSeekDebounceMillis)
            player.time = pending.targetPositionMs
            positionMs = pending.targetPositionMs
            pendingStepSeek = null
            hideSeekPreviewJob = scope.launch {
                delay(900)
                showSeekPreview = false
            }
        }
        if (showControls) {
            showControlsTemporarily()
        }
    }

    fun handleTvRemoteKeyAction(action: TvRemoteKeyAction): Boolean {
        return when (action) {
            is TvRemoteKeyAction.Seek -> {
                performDebouncedStepSeek(
                    deltaMs = action.deltaMs,
                    showControls = false,
                )
                controlsVisible = true
                hideEpisodeRail()
                playerFocusLayer = TvPlayerFocusLayer.Root
                requestRootFocusWhenReady()
                scheduleAutoHideControls()
                true
            }

            TvRemoteKeyAction.EnterFocus -> {
                hideEpisodeRail()
                playerFocusLayer = TvPlayerFocusLayer.Controls
                showControlsTemporarily(requestControlFocus = TvControlFocusTarget.PlayPause)
                true
            }

            TvRemoteKeyAction.ExitFocus -> {
                hideEpisodeRail()
                playerFocusLayer = TvPlayerFocusLayer.Root
                requestRootFocusWhenReady()
                true
            }

            TvRemoteKeyAction.EnterEpisodeRail -> {
                if (tvControlsVariant != TvLongFormControlsVariant.SeriesEpisodeRail || episodeRailItems.isEmpty()) {
                    return false
                }
                controlsVisible = true
                episodeRailVisible = true
                episodeRailOpenNonce += 1
                focusedEpisodeRailItemId = currentEpisodeRailItemId.orEmpty()
                pendingEpisodeRailFocusRequest = true
                playerFocusLayer = TvPlayerFocusLayer.EpisodeRail
                hideControlsJob?.cancel()
                true
            }

            TvRemoteKeyAction.ExitEpisodeRail -> {
                hideEpisodeRail()
                playerFocusLayer = TvPlayerFocusLayer.Controls
                requestControlFocusWhenReady(lastFocusedControlTarget)
                scheduleAutoHideControls()
                true
            }

            TvRemoteKeyAction.TogglePlayPause -> {
                togglePlaybackWithFeedback(showControls = false)
                true
            }

            TvRemoteKeyAction.DismissUi -> {
                hideAllTvUi()
                true
            }

            TvRemoteKeyAction.PassThrough -> {
                if (playerFocusLayer == TvPlayerFocusLayer.Controls && !episodeRailVisible) {
                    scheduleAutoHideControls()
                }
                false
            }
        }
    }

    DisposableEffect(player) {
        val listener = object : MediaPlayer.EventListener {
            override fun onEvent(event: MediaPlayer.Event) {
                latestOnVlcEvent(event)
                when (event.type) {
                    MediaPlayer.Event.Playing -> {
                        isPlaying = true
                        isVlcPlaying = true
                    }
                    MediaPlayer.Event.Paused -> isPlaying = false
                    MediaPlayer.Event.Stopped,
                    MediaPlayer.Event.EndReached,
                    MediaPlayer.Event.EncounteredError -> {
                        isPlaying = false
                        isVlcPlaying = false
                    }
                    MediaPlayer.Event.TimeChanged -> if (!isScrubbing && !draggingSeek && pendingStepSeek == null) {
                        positionMs = player.time.coerceAtLeast(0L)
                    }
                    MediaPlayer.Event.LengthChanged -> {
                        val duration = player.length.coerceAtLeast(0L)
                        if (duration > 0) durationMs = duration
                    }
                }
                audioTracks = buildLongFormAudioTracksFromVlc(player.audioTracks, player.audioTrack)
            }
        }
        player.setEventListener(listener)
        isPlaying = player.isPlaying
        durationMs = player.length.coerceAtLeast(0L)
        positionMs = player.time.coerceAtLeast(0L)
        audioTracks = buildLongFormAudioTracksFromVlc(player.audioTracks, player.audioTrack)
        onDispose {
            player.setEventListener(null)
            if (longPressBoosting) {
                player.setRate(1f)
            }
            hideControlsJob?.cancel()
            hideSeekPreviewJob?.cancel()
            hideCenterFeedbackJob?.cancel()
        }
    }

    LaunchedEffect(player, isScrubbing, draggingSeek, pendingStepSeek) {
        while (true) {
            if (!isScrubbing && !draggingSeek && pendingStepSeek == null) {
                positionMs = player.time.coerceAtLeast(0L)
            }
            val duration = player.length.coerceAtLeast(0L)
            if (duration > 0) {
                durationMs = duration
            }
            delay(250)
        }
    }

    LaunchedEffect(tvMode, tvControlsVariant) {
        if (tvMode) {
            requestRootFocusWhenReady()
            if (tvControlsVariant == TvLongFormControlsVariant.SeriesEpisodeRail) {
                controlsVisible = false
                hideControlsJob?.cancel()
            } else {
                controlsVisible = true
                scheduleAutoHideControls()
            }
        }
    }

    LaunchedEffect(tvMode, controlsVisible) {
        if (!tvMode) {
            return@LaunchedEffect
        }
        if (controlsVisible) {
            hasShownControlsOnce = true
        } else if (hasShownControlsOnce) {
            pendingControlFocusTarget = null
            hideEpisodeRail()
            playerFocusLayer = TvPlayerFocusLayer.Root
            requestRootFocusWhenReady()
        }
    }

    val currentFocusGuardInput = PlayerFocusGuardInput(
        controlsVisible = controlsVisible,
        subtitleSheetVisible = subtitleSheetVisible,
        audioTrackSheetVisible = audioTrackSheetVisible,
        resumePromptVisible = resumePromptVisible,
        backConfirmPromptVisible = backConfirmPromptVisible,
        playerErrorVisible = playerErrorVisible,
    )
    var previousFocusGuardInput by remember { mutableStateOf(currentFocusGuardInput) }
    LaunchedEffect(currentFocusGuardInput) {
        if (tvMode && shouldReclaimRootFocus(previousFocusGuardInput, currentFocusGuardInput)) {
            if (episodeRailVisible) {
                previousFocusGuardInput = currentFocusGuardInput
                return@LaunchedEffect
            }
            requestRootFocusWhenReady()
        }
        previousFocusGuardInput = currentFocusGuardInput
    }

    LaunchedTvInitialFocus(tvMode, pendingRootFocusRequest) {
        if (!tvMode || !pendingRootFocusRequest) {
            return@LaunchedTvInitialFocus
        }
        try {
            rootFocusRequester.tryRequestFocus()
        } finally {
            pendingRootFocusRequest = false
        }
    }

    fun resolveControlFocusRequester(target: TvControlFocusTarget): FocusRequester {
        return when (target) {
            TvControlFocusTarget.PlayPause -> playPauseFocusRequester
            TvControlFocusTarget.Rewind -> rewindFocusRequester
            TvControlFocusTarget.Forward -> forwardFocusRequester
            TvControlFocusTarget.EpisodeSelector -> episodeSelectorFocusRequester
            TvControlFocusTarget.NextEpisode -> nextEpisodeFocusRequester
            TvControlFocusTarget.Subtitle -> subtitleFocusRequester
            TvControlFocusTarget.AudioTrack -> audioTrackFocusRequester
            TvControlFocusTarget.BackDetail -> backDetailFocusRequester
            TvControlFocusTarget.ExitPlayback -> exitPlaybackFocusRequester
            TvControlFocusTarget.Fullscreen -> fullscreenFocusRequester
        }
    }

    LaunchedTvInitialFocus(tvMode, controlsVisible, pendingControlFocusTarget) {
        val target = pendingControlFocusTarget
        if (!tvMode || !controlsVisible || target == null) {
            return@LaunchedTvInitialFocus
        }
        for (attempt in 0 until 4) {
            if (resolveControlFocusRequester(target).tryRequestFocus()) {
                break
            }
            if (attempt < 3) {
                delay(16)
            }
        }
        pendingControlFocusTarget = null
    }

    LaunchedTvInitialFocus(tvMode, episodeRailVisible, episodeRailOpenNonce, pendingEpisodeRailFocusRequest) {
        if (!tvMode || !episodeRailVisible || !pendingEpisodeRailFocusRequest) {
            return@LaunchedTvInitialFocus
        }
        for (attempt in 0 until 4) {
            if (currentEpisodeRailFocusRequester.tryRequestFocus()) {
                break
            }
            if (attempt < 3) {
                delay(16)
            }
        }
        pendingEpisodeRailFocusRequest = false
    }

    LaunchedEffect(player, audioTracks, selectedAudioTrackId, selectedAudioPreference, isVlcPlaying) {
        if (!isVlcPlaying || audioTracks.isEmpty()) {
            return@LaunchedEffect
        }
        val resolvedSelection = resolveAudioSelectionOnTrackLoad(
            currentSelection = selectedAudioTrackId,
            storedPreference = selectedAudioPreference,
            tracks = audioTracks,
        )
        if (selectedAudioTrackId?.isNotBlank() == true && resolvedSelection == null) {
            // 已有选择已经失效（track id 在新 media 不存在且 preference 也匹配不到）→ 清掉本地状态，
            // isUserAction=false 让父级不要写 DataStore（这不是用户的主动操作）。
            onSelectAudioTrack(null, null, false)
        }
        Log.d(
            LongFormVideoPlayerLogTag,
            "applyAudioSelection selectedAudioTrackId=$selectedAudioTrackId selectedAudioPreference=$selectedAudioPreference " +
                "resolved=$resolvedSelection " +
                "tracks=${audioTracks.joinToString { "${it.id}:${it.label}/${it.detail}:selected=${it.selected}" }}",
        )
        val selected = audioTracks.firstOrNull { it.id == resolvedSelection }
        if (selected != null) {
            // 只在我们解析出明确的目标轨时才覆盖 LibVLC 的当前选择。
            // LibVLC 3.x 把 audioTrack = -1 解释为"禁用音频"，不是"自动"，
            // 所以首次进入（无 preference + 无 selection）必须保留 LibVLC 自己挑的默认轨。
            player.audioTrack = selected.vlcTrackId
        }
        val selectedAfterApply = buildLongFormAudioTracksFromVlc(player.audioTracks, player.audioTrack)
            .filter { it.selected }
            .joinToString { it.id }
        Log.d(
            LongFormVideoPlayerLogTag,
            "audioSelectionApplied override=$resolvedSelection selectedAfterApply=$selectedAfterApply",
        )

        // F3：resolvedSelection 是父级当前未感知到的新选择（由 preference 推导而来），
        // 通过 isUserAction=false 回灌给父级，picker 的 selectedAudioTrackId 不再撒谎。
        if (resolvedSelection != null &&
            resolvedSelection != selectedAudioTrackId?.takeIf { it.isNotBlank() }
        ) {
            val preferenceForResolved = selected?.let(::buildAudioTrackPreference)
            onSelectAudioTrack(resolvedSelection, preferenceForResolved, false)
        }
    }

    val actualDurationMs = effectiveDurationMs()
    val displayPositionMs = when {
        draggingSeek -> dragTargetPositionMs
        isScrubbing -> scrubPositionMs
        pendingStepSeek != null -> pendingStepSeek?.targetPositionMs ?: positionMs
        else -> positionMs
    }
    val progressValue = if (actualDurationMs > 0) {
        (displayPositionMs.toFloat() / actualDurationMs.toFloat()).coerceIn(0f, 1f)
    } else {
        0f
    }
    val seriesBottomPanelPage = if (episodeRailVisible) {
        TvSeriesBottomPanelPage.EpisodeRail
    } else {
        TvSeriesBottomPanelPage.Controls
    }
    val tvControlFocusTargets = buildList {
        add(TvControlFocusTarget.PlayPause)
        if (tvMode && tvControlsVariant == TvLongFormControlsVariant.Default) {
            add(TvControlFocusTarget.Rewind)
            add(TvControlFocusTarget.Forward)
        }
        if (tvMode && tvControlsVariant == TvLongFormControlsVariant.Default && onOpenEpisodeSelector != null) {
            add(TvControlFocusTarget.EpisodeSelector)
        }
        if (tvMode && tvControlsVariant == TvLongFormControlsVariant.Default && onNextEpisode != null) {
            add(TvControlFocusTarget.NextEpisode)
        }
        add(TvControlFocusTarget.Subtitle)
        if (tvMode) {
            add(TvControlFocusTarget.AudioTrack)
        }
        if (tvMode && tvControlsVariant == TvLongFormControlsVariant.Default) {
            add(TvControlFocusTarget.BackDetail)
            if (onExitPlayback != null) {
                add(TvControlFocusTarget.ExitPlayback)
            }
        }
        if (!tvMode) {
            add(TvControlFocusTarget.Fullscreen)
        }
    }
    fun controlFocusModifier(target: TvControlFocusTarget): Modifier {
        fun requestHorizontalControlFocus(direction: TvControlFocusDirection): Boolean {
            val nextTarget = resolveTvControlHorizontalFocusTarget(
                current = target,
                direction = direction,
                targets = tvControlFocusTargets,
            ) ?: return false
            val moved = resolveControlFocusRequester(nextTarget).tryRequestFocus()
            if (moved) {
                playerFocusLayer = TvPlayerFocusLayer.Controls
                lastFocusedControlTarget = nextTarget
                scheduleAutoHideControls()
            }
            return moved
        }

        val base = Modifier
            .onFocusChanged { focusState ->
                if (focusState.isFocused || focusState.hasFocus) {
                    playerFocusLayer = TvPlayerFocusLayer.Controls
                    lastFocusedControlTarget = target
                }
            }
            .onPreviewKeyEvent { event ->
                if (!tvMode || event.nativeKeyEvent.action != AndroidKeyEvent.ACTION_DOWN) {
                    return@onPreviewKeyEvent false
                }
                if (tvControlsVariant != TvLongFormControlsVariant.SeriesEpisodeRail || episodeRailItems.isEmpty()) {
                    return@onPreviewKeyEvent when (event.nativeKeyEvent.keyCode) {
                        AndroidKeyEvent.KEYCODE_DPAD_LEFT -> requestHorizontalControlFocus(TvControlFocusDirection.Left)
                        AndroidKeyEvent.KEYCODE_DPAD_RIGHT -> requestHorizontalControlFocus(TvControlFocusDirection.Right)
                        else -> false
                    }
                }
                when (event.nativeKeyEvent.keyCode) {
                    AndroidKeyEvent.KEYCODE_DPAD_LEFT -> requestHorizontalControlFocus(TvControlFocusDirection.Left)
                    AndroidKeyEvent.KEYCODE_DPAD_RIGHT -> requestHorizontalControlFocus(TvControlFocusDirection.Right)
                    AndroidKeyEvent.KEYCODE_DPAD_DOWN -> handleTvRemoteKeyAction(TvRemoteKeyAction.EnterEpisodeRail)
                    else -> false
                }
            }
        if (!tvMode) {
            return base
        }
        if (tvControlFocusTargets.indexOf(target) < 0 || tvControlFocusTargets.isEmpty()) {
            return base
        }
        val leftTarget = resolveTvControlHorizontalFocusTarget(
            current = target,
            direction = TvControlFocusDirection.Left,
            targets = tvControlFocusTargets,
        ) ?: return base
        val rightTarget = resolveTvControlHorizontalFocusTarget(
            current = target,
            direction = TvControlFocusDirection.Right,
            targets = tvControlFocusTargets,
        ) ?: return base
        return base.focusProperties {
            left = resolveControlFocusRequester(leftTarget)
            right = resolveControlFocusRequester(rightTarget)
        }
    }
    val titleOverlayData = buildTvLongFormTitleOverlayData(
        primaryFallback = title,
        seriesTitle = seriesTitleForOverlay,
        seasonNumber = seasonNumber,
        episodeNumber = episodeNumber,
        episodeTitle = episodeTitle,
    )

    Box(
        modifier = modifier
            .fillMaxSize()
            .background(Color.Black)
            .focusRequester(rootFocusRequester)
            .focusable()
            .onFocusChanged { focusState ->
                if (tvMode && focusState.isFocused) {
                    playerFocusLayer = TvPlayerFocusLayer.Root
                }
                if (tvMode &&
                    !focusState.hasFocus &&
                    !currentFocusGuardInput.anyOverlayVisible() &&
                    playerFocusLayer == TvPlayerFocusLayer.Root &&
                    !episodeRailVisible
                ) {
                    requestRootFocusWhenReady()
                }
            }
            .onPreviewKeyEvent { event ->
                if (!tvMode || event.nativeKeyEvent.action != AndroidKeyEvent.ACTION_DOWN) {
                    return@onPreviewKeyEvent false
                }
                // 任何 overlay（续播卡 / 字幕音轨 picker / 返回二次确认 / 错误浮层）可见时，
                // player 路由不消费按键，让 key event 流向 overlay 自身的 clickable / focusable
                // 处理——否则续播卡的"继续观看"按钮按 CENTER 会被 player 拦下来切播放暂停。
                if (currentFocusGuardInput.anyOverlayVisible()) {
                    return@onPreviewKeyEvent false
                }
                val action = resolveTvRemoteKeyAction(
                    visible = controlsVisible,
                    focusLayer = playerFocusLayer,
                    episodeRailEnabled = tvControlsVariant == TvLongFormControlsVariant.SeriesEpisodeRail && episodeRailItems.isNotEmpty(),
                    keyCode = event.nativeKeyEvent.keyCode,
                    repeatCount = event.nativeKeyEvent.repeatCount,
                    seekStepSec = tvSeekStepSeconds,
                ) ?: return@onPreviewKeyEvent false
                handleTvRemoteKeyAction(
                    action = action,
                )
            }
            .onSizeChanged {
                viewWidthPx = it.width.toFloat().coerceAtLeast(1f)
            }
            .pointerInput(player, isPlaying) {
                detectTapGestures(
                    onTap = {
                        if (ignoreTapAfterLongPress) {
                            ignoreTapAfterLongPress = false
                            return@detectTapGestures
                        }
                        controlsVisible = !controlsVisible
                        if (controlsVisible) {
                            scheduleAutoHideControls()
                        } else {
                            hideControlsJob?.cancel()
                        }
                    },
                    onDoubleTap = {
                        if (ignoreTapAfterLongPress) {
                            ignoreTapAfterLongPress = false
                            return@detectTapGestures
                        }
                        togglePlaybackWithFeedback()
                    },
                    onPress = {
                        longPressBoosting = false
                        val boostJob = scope.launch {
                            delay(350)
                            longPressBoosting = true
                            ignoreTapAfterLongPress = true
                            player.setRate(2f)
                            showTransientFeedback(
                                icon = Icons.Filled.FastForward,
                                text = "2.0x 倍速",
                                durationMs = 1_200L,
                            )
                        }
                        tryAwaitRelease()
                        boostJob.cancel()
                        if (longPressBoosting) {
                            player.setRate(1f)
                            longPressBoosting = false
                        }
                    },
                )
            }
            .pointerInput(player, viewWidthPx) {
                detectHorizontalDragGestures(
                    onDragStart = {
                        draggingSeek = true
                        dragDistancePx = 0f
                        dragStartPositionMs = player.time.coerceAtLeast(0L)
                        dragTargetPositionMs = dragStartPositionMs
                        showSeekPreview = true
                        showControlsTemporarily()
                    },
                    onHorizontalDrag = { change, dragAmount ->
                        change.consume()
                        dragDistancePx += dragAmount
                        val deltaMs = (dragDistancePx / viewWidthPx) * 120_000f
                        val duration = effectiveDurationMs()
                        val target = (dragStartPositionMs + deltaMs.toLong()).coerceAtLeast(0L)
                        dragTargetPositionMs = if (duration > 0) {
                            target.coerceAtMost(duration)
                        } else {
                            target
                        }
                        val delta = dragTargetPositionMs - dragStartPositionMs
                        val direction = if (delta >= 0) "快进" else "快退"
                        val sign = if (delta >= 0) "+" else "-"
                        seekPreviewText = buildString {
                            append(direction)
                            append(" ")
                            append(sign)
                            append(formatPlaybackTime(abs(delta)))
                            append("\n")
                            append(formatPlaybackTime(dragTargetPositionMs))
                            if (duration > 0) {
                                append(" / ")
                                append(formatPlaybackTime(duration))
                            }
                        }
                    },
                    onDragEnd = {
                        draggingSeek = false
                        player.time = dragTargetPositionMs
                        positionMs = dragTargetPositionMs
                        hideSeekPreviewJob?.cancel()
                        hideSeekPreviewJob = scope.launch {
                            delay(900)
                            showSeekPreview = false
                        }
                        scheduleAutoHideControls()
                    },
                    onDragCancel = {
                        draggingSeek = false
                        hideSeekPreviewJob?.cancel()
                        hideSeekPreviewJob = scope.launch {
                            delay(500)
                            showSeekPreview = false
                        }
                    },
                )
            },
    ) {
        VlcLongFormSurface(
            mediaPlayer = player,
            modifier = Modifier.fillMaxSize(),
        )

        if (!posterUrl.isNullOrBlank() && showPoster) {
            AsyncImage(
                model = posterUrl,
                contentDescription = "$title 封面",
                contentScale = ContentScale.Crop,
                modifier = Modifier.fillMaxSize(),
            )
        }

        AnimatedVisibility(
            visible = showSeekPreview || draggingSeek,
            enter = fadeIn(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
            exit = fadeOut(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
            modifier = Modifier.align(Alignment.Center),
        ) {
            Surface(
                color = Color(0xD610131A),
                shape = AppChrome.SurfaceShape,
            ) {
                Text(
                    text = seekPreviewText,
                    color = Color.White,
                    style = MaterialTheme.typography.bodySmall,
                    modifier = Modifier
                        .padding(horizontal = 14.dp, vertical = 10.dp)
                        .widthIn(min = 112.dp),
                )
            }
        }

        AnimatedVisibility(
            visible = showCenterFeedback,
            enter = fadeIn(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
            exit = fadeOut(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
            modifier = Modifier.align(Alignment.Center),
        ) {
            Surface(
                color = Color(0xCC0D1016),
                shape = AppChrome.SurfaceShape,
            ) {
                Row(
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
                    horizontalArrangement = Arrangement.spacedBy(8.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Icon(
                        imageVector = centerFeedbackIcon,
                        contentDescription = null,
                        tint = Color.White,
                        modifier = Modifier.size(22.dp),
                    )
                    Text(
                        text = centerFeedbackText,
                        color = Color.White,
                        style = MaterialTheme.typography.bodyMedium,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                }
            }
        }

        if (tvMode) {
            TvLongFormTitleOverlay(
                data = titleOverlayData,
                visible = controlsVisible,
                showStatusBarPadding = showStatusBarPadding,
                modifier = Modifier.align(Alignment.TopStart),
            )
        } else {
            AnimatedVisibility(
                visible = controlsVisible,
                enter = fadeIn(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
                exit = fadeOut(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
                modifier = Modifier.align(Alignment.TopCenter),
            ) {
                val topPaddingModifier = if (showStatusBarPadding) {
                    Modifier.statusBarsPadding()
                } else {
                    Modifier
                }
                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .then(topPaddingModifier)
                        .padding(horizontal = 12.dp, vertical = 12.dp),
                ) {
                    Surface(
                        color = Color(0x8C0D1016),
                        shape = AppChrome.SurfaceShape,
                        modifier = Modifier.fillMaxWidth(),
                    ) {
                        Row(
                            modifier = Modifier.padding(horizontal = 8.dp, vertical = 6.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(8.dp),
                        ) {
                            CompactPlayerControlButton(
                                icon = Icons.AutoMirrored.Filled.ArrowBack,
                                contentDescription = "返回",
                                tvMode = tvMode,
                                onClick = {
                                    showControlsTemporarily()
                                    onBack()
                                },
                            )
                            Text(
                                text = title,
                                color = Color.White,
                                style = MaterialTheme.typography.titleSmall,
                                fontWeight = FontWeight.SemiBold,
                                maxLines = 1,
                                overflow = TextOverflow.Ellipsis,
                                modifier = Modifier.weight(1f),
                            )
                        }
                    }
                }
            }
        }

        AnimatedVisibility(
            visible = controlsVisible,
            enter = fadeIn(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
            exit = fadeOut(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
            modifier = Modifier.align(Alignment.BottomCenter),
        ) {
            val isSeriesEpisodeRailVariant = tvMode && tvControlsVariant == TvLongFormControlsVariant.SeriesEpisodeRail
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 12.dp, vertical = 12.dp),
            ) {
                Surface(
                    color = Color(0x910D1016),
                    shape = AppChrome.SurfaceShape,
                    modifier = Modifier.fillMaxWidth(),
                ) {
                    if (isSeriesEpisodeRailVariant) {
                        AnimatedContent(
                            targetState = seriesBottomPanelPage,
                            transitionSpec = {
                                if (targetState.depth > initialState.depth) {
                                    slideInVertically(
                                        animationSpec = tween(
                                            TvMotionTokens.DurationStandardMs,
                                            easing = TvMotionTokens.EasingStandard,
                                        ),
                                    ) { fullHeight -> fullHeight / 2 } togetherWith
                                        slideOutVertically(
                                            animationSpec = tween(
                                                TvMotionTokens.DurationStandardMs,
                                                easing = TvMotionTokens.EasingStandard,
                                            ),
                                        ) { fullHeight -> -fullHeight / 2 }
                                } else {
                                    slideInVertically(
                                        animationSpec = tween(
                                            TvMotionTokens.DurationStandardMs,
                                            easing = TvMotionTokens.EasingStandard,
                                        ),
                                    ) { fullHeight -> -fullHeight / 2 } togetherWith
                                        slideOutVertically(
                                            animationSpec = tween(
                                                TvMotionTokens.DurationStandardMs,
                                                easing = TvMotionTokens.EasingStandard,
                                            ),
                                        ) { fullHeight -> fullHeight / 2 }
                                }
                            },
                            label = "TvSeriesBottomPanelPage",
                        ) { page ->
                            when (page) {
                                TvSeriesBottomPanelPage.Controls -> TvSeriesControlsPage(
                                    isPlaying = isPlaying,
                                    displayPositionMs = displayPositionMs,
                                    actualDurationMs = actualDurationMs,
                                    progressValue = progressValue,
                                    tvMode = tvMode,
                                    playPauseFocusRequester = playPauseFocusRequester,
                                    subtitleFocusRequester = subtitleFocusRequester,
                                    audioTrackFocusRequester = audioTrackFocusRequester,
                                    controlFocusModifier = ::controlFocusModifier,
                                    onTogglePlayPause = { togglePlaybackWithFeedback() },
                                    onOpenSubtitle = {
                                        subtitleSheetVisible = true
                                        showControlsTemporarily()
                                    },
                                    onOpenAudioTrack = {
                                        audioTrackSheetVisible = true
                                        showControlsTemporarily()
                                    },
                                )

                                TvSeriesBottomPanelPage.EpisodeRail -> TvEpisodeRail(
                                    items = episodeRailItems,
                                    currentEpisodeId = currentEpisodeRailItemId,
                                    focusedEpisodeId = focusedEpisodeRailItemId,
                                    openNonce = episodeRailOpenNonce,
                                    currentEpisodeFocusRequester = currentEpisodeRailFocusRequester,
                                    onFocusedEpisodeChanged = { episodeId ->
                                        focusedEpisodeRailItemId = episodeId
                                        playerFocusLayer = TvPlayerFocusLayer.EpisodeRail
                                    },
                                    onSelectEpisode = { item ->
                                        hideAllTvUi()
                                        if (item.id != currentEpisodeRailItemId) {
                                            onSelectEpisodeRailItem?.invoke(item)
                                        }
                                    },
                                )
                            }
                        }
                    } else {
                        Row(
                            modifier = Modifier.padding(horizontal = 8.dp, vertical = 8.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(6.dp),
                        ) {
                            CompactPlayerControlButton(
                                icon = if (isPlaying) Icons.Filled.Pause else Icons.Filled.PlayArrow,
                                contentDescription = if (isPlaying) "暂停" else "播放",
                                tvMode = tvMode,
                                focusRequester = playPauseFocusRequester,
                                modifier = controlFocusModifier(TvControlFocusTarget.PlayPause),
                                onClick = { togglePlaybackWithFeedback() },
                            )
                            if (tvMode && tvControlsVariant == TvLongFormControlsVariant.Default) {
                                CompactPlayerControlButton(
                                    icon = Icons.Filled.FastForward,
                                    contentDescription = "快退 ${normalizeTvSeekStepSeconds(tvSeekStepSeconds)} 秒",
                                    tvMode = true,
                                    focusRequester = rewindFocusRequester,
                                    modifier = controlFocusModifier(TvControlFocusTarget.Rewind),
                                    onClick = { performDebouncedStepSeek(-tvSeekStepMs) },
                                    reverseMirror = true,
                                )
                            }
                            Text(
                                text = formatPlaybackTime(displayPositionMs),
                                color = Color.White.copy(alpha = 0.86f),
                                style = MaterialTheme.typography.labelSmall,
                            )
                            Slider(
                                value = progressValue,
                                onValueChange = { value ->
                                    val duration = effectiveDurationMs()
                                    if (duration <= 0) {
                                        return@Slider
                                    }
                                    isScrubbing = true
                                    scrubPositionMs = (duration * value).toLong().coerceIn(0L, duration)
                                    showControlsTemporarily()
                                },
                                onValueChangeFinished = {
                                    if (isScrubbing) {
                                        player.time = scrubPositionMs
                                        positionMs = scrubPositionMs
                                    }
                                    isScrubbing = false
                                    scheduleAutoHideControls()
                                },
                                colors = SliderDefaults.colors(
                                    thumbColor = Color.White,
                                    activeTrackColor = Color.White.copy(alpha = 0.92f),
                                    inactiveTrackColor = Color.White.copy(alpha = 0.22f),
                                ),
                                modifier = Modifier
                                    .weight(1f)
                                    .focusProperties { canFocus = false },
                            )
                            Text(
                                text = if (actualDurationMs > 0) {
                                    formatPlaybackTime(actualDurationMs)
                                } else {
                                    "--:--"
                                },
                                color = Color.White.copy(alpha = 0.68f),
                                style = MaterialTheme.typography.labelSmall,
                            )
                            if (tvMode && tvControlsVariant == TvLongFormControlsVariant.Default) {
                                CompactPlayerControlButton(
                                    icon = Icons.Filled.FastForward,
                                    contentDescription = "快进 ${normalizeTvSeekStepSeconds(tvSeekStepSeconds)} 秒",
                                    tvMode = true,
                                    focusRequester = forwardFocusRequester,
                                    modifier = controlFocusModifier(TvControlFocusTarget.Forward),
                                    onClick = { performDebouncedStepSeek(tvSeekStepMs) },
                                )
                                onOpenEpisodeSelector?.let { openSelector ->
                                    CompactPlayerControlButton(
                                        icon = Icons.Filled.Tv,
                                        contentDescription = "选集",
                                        tvMode = true,
                                        focusRequester = episodeSelectorFocusRequester,
                                        modifier = controlFocusModifier(TvControlFocusTarget.EpisodeSelector),
                                        onClick = {
                                            openSelector()
                                            showControlsTemporarily()
                                        },
                                    )
                                }
                                onNextEpisode?.let { nextEpisode ->
                                    CompactPlayerControlButton(
                                        icon = Icons.Filled.SkipNext,
                                        contentDescription = "下一集",
                                        tvMode = true,
                                        focusRequester = nextEpisodeFocusRequester,
                                        modifier = controlFocusModifier(TvControlFocusTarget.NextEpisode),
                                        onClick = {
                                            nextEpisode()
                                            showControlsTemporarily()
                                        },
                                    )
                                }
                            }
                            CompactPlayerControlButton(
                                icon = Icons.Filled.Subtitles,
                                contentDescription = "字幕",
                                tvMode = tvMode,
                                focusRequester = subtitleFocusRequester,
                                modifier = controlFocusModifier(TvControlFocusTarget.Subtitle),
                                onClick = {
                                    subtitleSheetVisible = true
                                    showControlsTemporarily()
                                },
                            )
                            if (tvMode) {
                                CompactPlayerControlButton(
                                    icon = Icons.Filled.GraphicEq,
                                    contentDescription = "音轨",
                                    tvMode = true,
                                    focusRequester = audioTrackFocusRequester,
                                    modifier = controlFocusModifier(TvControlFocusTarget.AudioTrack),
                                    onClick = {
                                        audioTrackSheetVisible = true
                                        showControlsTemporarily()
                                    },
                                )
                                if (tvControlsVariant == TvLongFormControlsVariant.Default) {
                                    CompactPlayerControlButton(
                                        icon = Icons.AutoMirrored.Filled.ArrowBack,
                                        contentDescription = "返回详情",
                                        tvMode = true,
                                        focusRequester = backDetailFocusRequester,
                                        modifier = controlFocusModifier(TvControlFocusTarget.BackDetail),
                                        onClick = {
                                            onRequestExitPlayback?.invoke() ?: onBack()
                                            showControlsTemporarily()
                                        },
                                    )
                                    onExitPlayback?.let { exitPlayback ->
                                        CompactPlayerControlButton(
                                            icon = Icons.Filled.FullscreenExit,
                                            contentDescription = "退出播放",
                                            tvMode = true,
                                            focusRequester = exitPlaybackFocusRequester,
                                            modifier = controlFocusModifier(TvControlFocusTarget.ExitPlayback),
                                            onClick = {
                                                onRequestExitPlayback?.invoke() ?: exitPlayback()
                                                showControlsTemporarily()
                                            },
                                        )
                                    }
                                }
                            } else {
                                CompactPlayerControlButton(
                                    icon = if (isFullscreen) Icons.Filled.FullscreenExit else Icons.Filled.Fullscreen,
                                    contentDescription = if (isFullscreen) "退出全屏" else "全屏",
                                    focusRequester = fullscreenFocusRequester,
                                    modifier = controlFocusModifier(TvControlFocusTarget.Fullscreen),
                                    onClick = {
                                        onToggleFullscreen()
                                        showControlsTemporarily()
                                    },
                                )
                            }
                        }
                    }
                }
            }
        }

        if (subtitleSheetVisible) {
            val dismissSubtitlePicker = { subtitleSheetVisible = false }
            when (resolveSubtitlePickerSurface(tvMode)) {
                SubtitlePickerSurface.CenterDialog -> TvSubtitlePickerDialog(
                    subtitleTracks = subtitleTracks,
                    selectedSubtitleTrackId = selectedSubtitleTrackId,
                    onSelectSubtitleTrack = onSelectSubtitleTrack,
                    onDismissRequest = dismissSubtitlePicker,
                )

                SubtitlePickerSurface.BottomSheet -> LongFormSubtitleBottomSheet(
                    subtitleTracks = subtitleTracks,
                    selectedSubtitleTrackId = selectedSubtitleTrackId,
                    onSelectSubtitleTrack = onSelectSubtitleTrack,
                    onDismissRequest = dismissSubtitlePicker,
                )
            }
        }
        if (audioTrackSheetVisible) {
            TvAudioTrackPickerDialog(
                audioTracks = audioTracks,
                selectedAudioTrackId = selectedAudioTrackId,
                onSelectAudioTrack = { trackId ->
                    val preference = buildAudioTrackPreference(audioTracks.firstOrNull { it.id == trackId })
                    // 用户主动从 picker 选轨：isUserAction=true，父级 save DataStore
                    onSelectAudioTrack(trackId, preference, true)
                },
                onDismissRequest = { audioTrackSheetVisible = false },
            )
        }

        resumePromptSlot?.invoke(this)
    }
}

internal const val TvStepSeekDebounceMillis = 300L

internal data class TvPendingStepSeekUpdate(
    val anchorPositionMs: Long,
    val targetPositionMs: Long,
    val accumulatedDeltaMs: Long,
    val shouldCommitImmediately: Boolean,
)

internal fun resolveTvPendingStepSeek(
    previous: TvPendingStepSeekUpdate?,
    currentPositionMs: Long,
    durationMs: Long,
    deltaMs: Long,
): TvPendingStepSeekUpdate {
    val anchor = previous?.anchorPositionMs ?: currentPositionMs.coerceAtLeast(0L)
    val rawTarget = (previous?.targetPositionMs ?: anchor) + deltaMs
    val target = rawTarget.coerceAtLeast(0L).let { next ->
        if (durationMs > 0L) next.coerceAtMost(durationMs) else next
    }
    return TvPendingStepSeekUpdate(
        anchorPositionMs = anchor,
        targetPositionMs = target,
        accumulatedDeltaMs = target - anchor,
        shouldCommitImmediately = false,
    )
}

internal fun normalizeTvSeekStepSeconds(seconds: Int): Int {
    return when (seconds) {
        5, 10, 15, 20, 30 -> seconds
        else -> 10
    }
}

@Composable
private fun TvPlaybackProgressBar(
    progressValue: Float,
    modifier: Modifier = Modifier,
) {
    Box(
        modifier = modifier
            .height(6.dp)
            .background(Color.White.copy(alpha = 0.18f), shape = AppChrome.PillShape),
    ) {
        Box(
            modifier = Modifier
                .fillMaxWidth(progressValue.coerceIn(0f, 1f))
                .height(6.dp)
                .background(Color.White.copy(alpha = 0.92f), shape = AppChrome.PillShape),
        )
    }
}

@Composable
private fun TvSeriesControlsPage(
    isPlaying: Boolean,
    displayPositionMs: Long,
    actualDurationMs: Long,
    progressValue: Float,
    tvMode: Boolean,
    playPauseFocusRequester: FocusRequester,
    subtitleFocusRequester: FocusRequester,
    audioTrackFocusRequester: FocusRequester,
    controlFocusModifier: (TvControlFocusTarget) -> Modifier,
    onTogglePlayPause: () -> Unit,
    onOpenSubtitle: () -> Unit,
    onOpenAudioTrack: () -> Unit,
) {
    Row(
        modifier = Modifier.padding(horizontal = 8.dp, vertical = 8.dp),
        verticalAlignment = Alignment.CenterVertically,
        horizontalArrangement = Arrangement.spacedBy(6.dp),
    ) {
        CompactPlayerControlButton(
            icon = if (isPlaying) Icons.Filled.Pause else Icons.Filled.PlayArrow,
            contentDescription = if (isPlaying) "暂停" else "播放",
            tvMode = tvMode,
            focusRequester = playPauseFocusRequester,
            modifier = controlFocusModifier(TvControlFocusTarget.PlayPause),
            onClick = onTogglePlayPause,
        )
        Text(
            text = formatPlaybackTime(displayPositionMs),
            color = Color.White.copy(alpha = 0.86f),
            style = MaterialTheme.typography.labelSmall,
        )
        TvPlaybackProgressBar(
            progressValue = progressValue,
            modifier = Modifier
                .weight(1f)
                .focusProperties { canFocus = false },
        )
        Text(
            text = if (actualDurationMs > 0) {
                formatPlaybackTime(actualDurationMs)
            } else {
                "--:--"
            },
            color = Color.White.copy(alpha = 0.68f),
            style = MaterialTheme.typography.labelSmall,
        )
        CompactPlayerControlButton(
            icon = Icons.Filled.Subtitles,
            contentDescription = "字幕",
            tvMode = tvMode,
            focusRequester = subtitleFocusRequester,
            modifier = controlFocusModifier(TvControlFocusTarget.Subtitle),
            onClick = onOpenSubtitle,
        )
        CompactPlayerControlButton(
            icon = Icons.Filled.GraphicEq,
            contentDescription = "音轨",
            tvMode = true,
            focusRequester = audioTrackFocusRequester,
            modifier = controlFocusModifier(TvControlFocusTarget.AudioTrack),
            onClick = onOpenAudioTrack,
        )
    }
}

@Composable
private fun TvEpisodeRail(
    items: List<TvEpisodeRailItem>,
    currentEpisodeId: String?,
    focusedEpisodeId: String?,
    openNonce: Int,
    currentEpisodeFocusRequester: FocusRequester,
    onFocusedEpisodeChanged: (String) -> Unit,
    onSelectEpisode: (TvEpisodeRailItem) -> Unit,
) {
    if (items.isEmpty()) {
        return
    }
    val initialFirstVisibleItemIndex = remember(items, currentEpisodeId, openNonce) {
        resolveEpisodeRailInitialFirstVisibleItemIndex(items, currentEpisodeId)
    }
    val listState = rememberLazyListState(initialFirstVisibleItemIndex = initialFirstVisibleItemIndex)
    var skipFollowScrollOnOpen by remember(openNonce) { mutableStateOf(true) }
    LaunchedEffect(items, focusedEpisodeId, openNonce) {
        if (skipFollowScrollOnOpen) {
            skipFollowScrollOnOpen = false
            return@LaunchedEffect
        }
        val visibleItems = listState.layoutInfo.visibleItemsInfo
        if (visibleItems.isEmpty()) {
            return@LaunchedEffect
        }
        val nextFirstVisibleItemIndex = resolveEpisodeRailFollowScrollFirstVisibleItemIndex(
            items = items,
            focusedEpisodeId = focusedEpisodeId,
            firstVisibleItemIndex = visibleItems.firstOrNull()?.index ?: initialFirstVisibleItemIndex,
            lastVisibleItemIndex = visibleItems.lastOrNull()?.index ?: initialFirstVisibleItemIndex,
        ) ?: return@LaunchedEffect
        listState.scrollToItem(nextFirstVisibleItemIndex)
    }

    LazyRow(
        state = listState,
        contentPadding = PaddingValues(horizontal = 8.dp),
        horizontalArrangement = Arrangement.spacedBy(6.dp),
        modifier = Modifier.fillMaxWidth(),
    ) {
        items(items, key = { item -> item.id }) { item ->
            val focused = item.id == focusedEpisodeId
            val current = item.id == currentEpisodeId
            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(6.dp),
                modifier = Modifier.width(TvEpisodeRailLayoutTokens.ItemSlotWidthDp.dp),
            ) {
                Box(
                    modifier = Modifier
                        .width(TvEpisodeRailLayoutTokens.ItemSlotWidthDp.dp)
                        .height(TvEpisodeRailLayoutTokens.TooltipHeightDp.dp),
                    contentAlignment = Alignment.BottomCenter,
                ) {
                    if (focused) {
                        Box(
                            modifier = Modifier.wrapContentWidth(Alignment.CenterHorizontally, unbounded = true),
                            contentAlignment = Alignment.Center,
                        ) {
                            Surface(
                                color = Color(0xE8141820),
                                shape = AppChrome.SurfaceShape,
                            ) {
                                Text(
                                    text = item.title,
                                    color = Color.White,
                                    style = MaterialTheme.typography.labelMedium,
                                    maxLines = 1,
                                    overflow = TextOverflow.Ellipsis,
                                    modifier = Modifier
                                        .padding(horizontal = 12.dp, vertical = 8.dp)
                                        .widthIn(max = TvEpisodeRailLayoutTokens.TooltipMaxWidthDp.dp),
                                )
                            }
                        }
                    }
                }
                Surface(
                    color = when {
                        focused -> AppChrome.SurfaceMuted.copy(alpha = 0.98f)
                        current -> AppChrome.SurfaceStrong.copy(alpha = 0.96f)
                        item.playable -> AppChrome.SurfaceElevated.copy(alpha = 0.92f)
                        else -> AppChrome.Surface.copy(alpha = 0.66f)
                    },
                    border = when {
                        focused -> BorderStroke(1.dp, AppChrome.TextSecondary.copy(alpha = 0.72f))
                        current -> BorderStroke(1.dp, AppChrome.Divider.copy(alpha = 0.88f))
                        else -> null
                    },
                    shape = AppChrome.ChipShape,
                    modifier = Modifier
                        .width(TvEpisodeRailLayoutTokens.ItemSlotWidthDp.dp)
                        .then(if (current) Modifier.focusRequester(currentEpisodeFocusRequester) else Modifier)
                        .onFocusChanged { focusState ->
                            if (focusState.isFocused) {
                                onFocusedEpisodeChanged(item.id)
                            }
                        }
                        .focusProperties { canFocus = item.playable }
                        .focusable(enabled = item.playable)
                        .clickable(enabled = item.playable) { onSelectEpisode(item) },
                ) {
                    Text(
                        text = formatTvEpisodeRailLabel(item.number),
                        color = when {
                            item.playable -> AppChrome.TextPrimary
                            else -> AppChrome.TextSubtle
                        },
                        style = MaterialTheme.typography.labelLarge,
                        fontWeight = if (current) FontWeight.Bold else FontWeight.SemiBold,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                        modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
                    )
                }
            }
        }
    }
}

@Composable
private fun CompactPlayerControlButton(
    icon: ImageVector,
    contentDescription: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    tvMode: Boolean = false,
    focusRequester: FocusRequester? = null,
    reverseMirror: Boolean = false,
) {
    val buttonModifier = if (focusRequester != null) {
        modifier.focusRequester(focusRequester)
    } else {
        modifier
    }
    TvIconActionButton(
        icon = icon,
        contentDescription = contentDescription,
        onClick = onClick,
        modifier = buttonModifier,
        iconModifier = Modifier.graphicsLayer {
            if (reverseMirror) {
                scaleX = -1f
            }
        },
        size = if (tvMode) 42.dp else 34.dp,
        iconSize = if (tvMode) 24.dp else 20.dp,
        containerColor = Color(0x24000000),
        contentColor = Color.White,
        focusedScale = if (tvMode) 1.12f else 1.04f,
    )
}

private fun formatPlaybackTime(ms: Long): String {
    if (ms <= 0) {
        return "00:00"
    }
    val totalSeconds = (ms / 1000L).toInt()
    val hours = totalSeconds / 3600
    val minutes = (totalSeconds % 3600) / 60
    val seconds = totalSeconds % 60
    return if (hours > 0) {
        String.format("%d:%02d:%02d", hours, minutes, seconds)
    } else {
        String.format("%02d:%02d", minutes, seconds)
    }
}
