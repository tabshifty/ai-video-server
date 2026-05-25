package com.chee.videos.core.ui

import android.util.Log
import android.view.KeyEvent as AndroidKeyEvent
import android.graphics.Color as AndroidColor
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.background
import androidx.compose.foundation.focusable
import androidx.compose.foundation.gestures.detectHorizontalDragGestures
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.widthIn
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
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusProperties
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.input.key.onPreviewKeyEvent
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.media3.common.PlaybackParameters
import androidx.media3.common.Player
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.ui.AspectRatioFrameLayout
import androidx.media3.ui.CaptionStyleCompat
import androidx.media3.ui.PlayerView
import coil.compose.AsyncImage
import com.chee.videos.core.model.SubtitleTrackDto
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlin.math.abs

private const val LongFormVideoPlayerLogTag = "LongFormVideoPlayer"

@OptIn(androidx.compose.material3.ExperimentalMaterial3Api::class)
@Composable
fun LongFormVideoPlayer(
    title: String,
    player: ExoPlayer,
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
    onSelectAudioTrack: (String?) -> Unit = {},
    tvMode: Boolean = false,
    tvSeekStepSeconds: Int = 10,
    seriesTitleForOverlay: String? = null,
    seasonNumber: Int? = null,
    episodeNumber: Int? = null,
    episodeTitle: String? = null,
    onOpenEpisodeSelector: (() -> Unit)? = null,
    onNextEpisode: (() -> Unit)? = null,
    onRequestExitPlayback: (() -> Unit)? = null,
    onExitPlayback: (() -> Unit)? = null,
    onTrackSheetVisibilityChanged: (Boolean) -> Unit = {},
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

    var controlsVisible by remember { mutableStateOf(false) }
    var isPlaying by remember { mutableStateOf(player.isPlaying) }
    var durationMs by remember { mutableStateOf(player.duration.coerceAtLeast(0L)) }
    var positionMs by remember { mutableStateOf(player.currentPosition.coerceAtLeast(0L)) }
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
    var audioTracks by remember { mutableStateOf(emptyList<LongFormAudioTrack>()) }
    var hasShownControlsOnce by remember { mutableStateOf(false) }
    var pendingRootFocusRequest by remember { mutableStateOf(false) }
    var pendingPlayPauseFocusRequest by remember { mutableStateOf(false) }
    var focusInControls by remember { mutableStateOf(false) }

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
        return player.duration.coerceAtLeast(0L)
    }

    fun requestRootFocusWhenReady() {
        pendingRootFocusRequest = true
    }

    fun scheduleAutoHideControls() {
        hideControlsJob?.cancel()
        hideControlsJob = scope.launch {
            delay(5000)
            controlsVisible = false
            focusInControls = false
            requestRootFocusWhenReady()
        }
    }

    fun showControlsTemporarily(requestPlayPauseFocus: Boolean = false) {
        controlsVisible = true
        if (requestPlayPauseFocus) {
            pendingPlayPauseFocusRequest = true
        }
        scheduleAutoHideControls()
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
        val current = player.currentPosition.coerceAtLeast(0L)
        val target = (current + deltaMs).coerceAtLeast(0L).let { next ->
            if (duration > 0L) next.coerceAtMost(duration) else next
        }
        player.seekTo(target)
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
            currentPositionMs = player.currentPosition.coerceAtLeast(0L),
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
            player.seekTo(pending.targetPositionMs)
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
                focusInControls = false
                requestRootFocusWhenReady()
                scheduleAutoHideControls()
                true
            }

            TvRemoteKeyAction.EnterFocus -> {
                showControlsTemporarily(requestPlayPauseFocus = true)
                true
            }

            TvRemoteKeyAction.ExitFocus -> {
                requestRootFocusWhenReady()
                focusInControls = false
                true
            }

            TvRemoteKeyAction.TogglePlayPause -> {
                togglePlaybackWithFeedback(showControls = false)
                true
            }

            TvRemoteKeyAction.DismissUi -> {
                controlsVisible = false
                focusInControls = false
                hideControlsJob?.cancel()
                requestRootFocusWhenReady()
                true
            }

            TvRemoteKeyAction.PassThrough -> {
                if (focusInControls) {
                    scheduleAutoHideControls()
                }
                false
            }
        }
    }

    DisposableEffect(player) {
        val listener = object : Player.Listener {
            override fun onIsPlayingChanged(isPlayingValue: Boolean) {
                isPlaying = isPlayingValue
            }

            override fun onEvents(playerObj: Player, events: Player.Events) {
                val duration = playerObj.duration.coerceAtLeast(0L)
                if (duration > 0) {
                    durationMs = duration
                }
                if (!isScrubbing && !draggingSeek) {
                    positionMs = playerObj.currentPosition.coerceAtLeast(0L)
                }
                audioTracks = buildLongFormAudioTracks(playerObj.currentTracks)
            }
        }
        player.addListener(listener)
        isPlaying = player.isPlaying
        durationMs = player.duration.coerceAtLeast(0L)
        positionMs = player.currentPosition.coerceAtLeast(0L)
        audioTracks = buildLongFormAudioTracks(player.currentTracks)
        onDispose {
            player.removeListener(listener)
            if (longPressBoosting) {
                player.playbackParameters = PlaybackParameters(1f)
            }
            hideControlsJob?.cancel()
            hideSeekPreviewJob?.cancel()
            hideCenterFeedbackJob?.cancel()
        }
    }

    LaunchedEffect(player, isScrubbing, draggingSeek) {
        while (true) {
            if (!isScrubbing && !draggingSeek) {
                positionMs = player.currentPosition.coerceAtLeast(0L)
            }
            val duration = player.duration.coerceAtLeast(0L)
            if (duration > 0) {
                durationMs = duration
            }
            delay(250)
        }
    }

    LaunchedEffect(tvMode) {
        if (tvMode) {
            controlsVisible = true
            requestRootFocusWhenReady()
            scheduleAutoHideControls()
        }
    }

    LaunchedEffect(tvMode, controlsVisible) {
        if (!tvMode) {
            return@LaunchedEffect
        }
        if (controlsVisible) {
            hasShownControlsOnce = true
        } else if (hasShownControlsOnce) {
            pendingPlayPauseFocusRequest = false
            focusInControls = false
            requestRootFocusWhenReady()
        }
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

    LaunchedTvInitialFocus(tvMode, controlsVisible, pendingPlayPauseFocusRequest) {
        if (!tvMode || !controlsVisible || !pendingPlayPauseFocusRequest) {
            return@LaunchedTvInitialFocus
        }
        try {
            playPauseFocusRequester.tryRequestFocus()
        } finally {
            pendingPlayPauseFocusRequest = false
        }
    }

    LaunchedEffect(player, audioTracks, selectedAudioTrackId) {
        if (audioTracks.isEmpty()) {
            return@LaunchedEffect
        }
        val resolvedSelection = resolveAudioSelectionOnTrackLoad(
            storedSelection = selectedAudioTrackId,
            tracks = audioTracks,
        )
        if (selectedAudioTrackId?.isNotBlank() == true && resolvedSelection == null) {
            onSelectAudioTrack(null)
        }
        Log.d(
            LongFormVideoPlayerLogTag,
            "applyAudioSelection selectedAudioTrackId=$selectedAudioTrackId resolved=$resolvedSelection " +
                "tracks=${audioTracks.joinToString { "${it.id}:${it.label}/${it.detail}:selected=${it.selected}" }}",
        )
        player.trackSelectionParameters = buildAudioTrackSelectionParameters(
            currentParameters = player.trackSelectionParameters,
            currentTracks = player.currentTracks,
            audioTracks = audioTracks,
            selectedAudioTrackId = resolvedSelection,
        )
        val selectedAfterApply = buildLongFormAudioTracks(player.currentTracks)
            .filter { it.selected }
            .joinToString { it.id }
        Log.d(
            LongFormVideoPlayerLogTag,
            "audioSelectionApplied override=$resolvedSelection selectedAfterApply=$selectedAfterApply",
        )
    }

    val actualDurationMs = effectiveDurationMs()
    val displayPositionMs = when {
        draggingSeek -> dragTargetPositionMs
        isScrubbing -> scrubPositionMs
        else -> positionMs
    }
    val progressValue = if (actualDurationMs > 0) {
        (displayPositionMs.toFloat() / actualDurationMs.toFloat()).coerceIn(0f, 1f)
    } else {
        0f
    }
    val tvControlFocusRequesters = buildList {
        add(playPauseFocusRequester)
        if (tvMode) {
            add(rewindFocusRequester)
        }
        if (tvMode) {
            add(forwardFocusRequester)
        }
        if (tvMode && onOpenEpisodeSelector != null) {
            add(episodeSelectorFocusRequester)
        }
        if (tvMode && onNextEpisode != null) {
            add(nextEpisodeFocusRequester)
        }
        add(subtitleFocusRequester)
        if (tvMode) {
            add(audioTrackFocusRequester)
            add(backDetailFocusRequester)
        }
        if (tvMode && onExitPlayback != null) {
            add(exitPlaybackFocusRequester)
        }
        if (!tvMode) {
            add(fullscreenFocusRequester)
        }
    }
    fun controlFocusModifier(focusRequester: FocusRequester): Modifier {
        val base = Modifier.onFocusChanged { focusState ->
            if (focusState.isFocused) {
                focusInControls = true
            }
        }
        if (!tvMode) {
            return base
        }
        val index = tvControlFocusRequesters.indexOf(focusRequester)
        if (index < 0 || tvControlFocusRequesters.isEmpty()) {
            return base
        }
        val leftRequester = tvControlFocusRequesters[(index - 1 + tvControlFocusRequesters.size) % tvControlFocusRequesters.size]
        val rightRequester = tvControlFocusRequesters[(index + 1) % tvControlFocusRequesters.size]
        return base.focusProperties {
            left = leftRequester
            right = rightRequester
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
            .onPreviewKeyEvent { event ->
                if (!tvMode || event.nativeKeyEvent.action != AndroidKeyEvent.ACTION_DOWN) {
                    return@onPreviewKeyEvent false
                }
                val action = resolveTvRemoteKeyAction(
                    visible = controlsVisible,
                    focusInControls = focusInControls,
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
                            player.playbackParameters = PlaybackParameters(2f)
                            showTransientFeedback(
                                icon = Icons.Filled.FastForward,
                                text = "2.0x 倍速",
                                durationMs = 1_200L,
                            )
                        }
                        tryAwaitRelease()
                        boostJob.cancel()
                        if (longPressBoosting) {
                            player.playbackParameters = PlaybackParameters(1f)
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
                        dragStartPositionMs = player.currentPosition.coerceAtLeast(0L)
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
                        player.seekTo(dragTargetPositionMs)
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
        AndroidView(
            factory = { context ->
                PlayerView(context).apply {
                    useController = false
                    setShutterBackgroundColor(AndroidColor.BLACK)
                    setKeepContentOnPlayerReset(true)
                    resizeMode = AspectRatioFrameLayout.RESIZE_MODE_FIT
                    this.player = player
                    applyLongFormSubtitleStyle()
                }
            },
            update = { view ->
                view.player = player
                view.applyLongFormSubtitleStyle()
            },
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
                    Row(
                        modifier = Modifier.padding(horizontal = 8.dp, vertical = 6.dp),
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(6.dp),
                    ) {
                        CompactPlayerControlButton(
                            icon = if (isPlaying) Icons.Filled.Pause else Icons.Filled.PlayArrow,
                            contentDescription = if (isPlaying) "暂停" else "播放",
                            tvMode = tvMode,
                            focusRequester = playPauseFocusRequester,
                            modifier = controlFocusModifier(playPauseFocusRequester),
                            onClick = { togglePlaybackWithFeedback() },
                        )
                        if (tvMode) {
                            CompactPlayerControlButton(
                                icon = Icons.Filled.FastForward,
                                contentDescription = "快退 ${normalizeTvSeekStepSeconds(tvSeekStepSeconds)} 秒",
                                tvMode = true,
                                focusRequester = rewindFocusRequester,
                                modifier = controlFocusModifier(rewindFocusRequester),
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
                                    player.seekTo(scrubPositionMs)
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
                        if (tvMode) {
                            CompactPlayerControlButton(
                                icon = Icons.Filled.FastForward,
                                contentDescription = "快进 ${normalizeTvSeekStepSeconds(tvSeekStepSeconds)} 秒",
                                tvMode = true,
                                focusRequester = forwardFocusRequester,
                                modifier = controlFocusModifier(forwardFocusRequester),
                                onClick = { performDebouncedStepSeek(tvSeekStepMs) },
                            )
                            onOpenEpisodeSelector?.let { openSelector ->
                                CompactPlayerControlButton(
                                    icon = Icons.Filled.Tv,
                                    contentDescription = "选集",
                                    tvMode = true,
                                    focusRequester = episodeSelectorFocusRequester,
                                    modifier = controlFocusModifier(episodeSelectorFocusRequester),
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
                                    modifier = controlFocusModifier(nextEpisodeFocusRequester),
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
                            modifier = controlFocusModifier(subtitleFocusRequester),
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
                                modifier = controlFocusModifier(audioTrackFocusRequester),
                                onClick = {
                                    audioTrackSheetVisible = true
                                    showControlsTemporarily()
                                },
                            )
                            CompactPlayerControlButton(
                                icon = Icons.AutoMirrored.Filled.ArrowBack,
                                contentDescription = "返回详情",
                                tvMode = true,
                                focusRequester = backDetailFocusRequester,
                                modifier = controlFocusModifier(backDetailFocusRequester),
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
                                    modifier = controlFocusModifier(exitPlaybackFocusRequester),
                                    onClick = {
                                        onRequestExitPlayback?.invoke() ?: exitPlayback()
                                        showControlsTemporarily()
                                    },
                                )
                            }
                        } else {
                            CompactPlayerControlButton(
                                icon = if (isFullscreen) Icons.Filled.FullscreenExit else Icons.Filled.Fullscreen,
                                contentDescription = if (isFullscreen) "退出全屏" else "全屏",
                                focusRequester = fullscreenFocusRequester,
                                modifier = controlFocusModifier(fullscreenFocusRequester),
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
                onSelectAudioTrack = onSelectAudioTrack,
                onDismissRequest = { audioTrackSheetVisible = false },
            )
        }
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

private fun PlayerView.applyLongFormSubtitleStyle() {
    val subtitleView = getSubtitleView() ?: return
    subtitleView.setStyle(buildLongFormSubtitleStyle())
    subtitleView.setApplyEmbeddedStyles(true)
    subtitleView.setApplyEmbeddedFontSizes(true)
}

private fun buildLongFormSubtitleStyle(): CaptionStyleCompat {
    return CaptionStyleCompat(
        0xFFFFFFFF.toInt(),
        0x00000000,
        0x00000000,
        CaptionStyleCompat.EDGE_TYPE_OUTLINE,
        0xB3000000.toInt(),
        null,
    )
}
