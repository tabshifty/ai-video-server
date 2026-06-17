package com.chee.videos.core.ui

import android.view.KeyEvent as AndroidKeyEvent
import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.slideInVertically
import androidx.compose.animation.slideOutVertically
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.focusable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxScope
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.layout.Arrangement
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
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusProperties
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.input.key.onPreviewKeyEvent
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlin.math.abs

@Composable
internal fun TvSeriesCorePlaybackOverlay(
    title: String,
    isPlaying: Boolean,
    positionMs: Long,
    durationMs: Long,
    tvSeekStepSeconds: Int,
    seriesTitleForOverlay: String?,
    seasonNumber: Int?,
    episodeNumber: Int?,
    episodeTitle: String?,
    episodeRailItems: List<TvEpisodeRailItem>,
    currentEpisodeRailItemId: String?,
    onTogglePlayPause: () -> Unit,
    onSeekTo: (Long) -> Unit,
    modifier: Modifier = Modifier,
    showStatusBarPadding: Boolean = false,
    showTrackActions: Boolean = true,
    onSelectEpisodeRailItem: ((TvEpisodeRailItem) -> Unit)? = null,
    onEpisodeRailVisibilityChanged: (Boolean) -> Unit = {},
    onOpenSubtitle: () -> Unit = {},
    onOpenAudioTrack: () -> Unit = {},
    resumePromptVisible: Boolean = false,
    resumePromptSlot: (@Composable BoxScope.() -> Unit)? = null,
    backConfirmPromptVisible: Boolean = false,
    playerErrorVisible: Boolean = false,
    openEpisodeRailRequestKey: Int = 0,
    episodeSwitchState: com.chee.videos.feature.tv.TvEpisodeSwitchUiState? = null,
    onDismissEpisodeSwitchFeedback: () -> Unit = {},
) {
    val scope = rememberCoroutineScope()
    val rootFocusRequester = remember { FocusRequester() }
    val playPauseFocusRequester = remember { FocusRequester() }
    val subtitleFocusRequester = remember { FocusRequester() }
    val audioTrackFocusRequester = remember { FocusRequester() }
    val currentEpisodeRailFocusRequester = remember { FocusRequester() }

    var controlsVisible by remember { mutableStateOf(false) }
    var episodeRailVisible by rememberSaveable(currentEpisodeRailItemId) { mutableStateOf(false) }
    var episodeRailOpenNonce by remember(currentEpisodeRailItemId) { mutableStateOf(0) }
    var focusedEpisodeRailItemId by rememberSaveable(currentEpisodeRailItemId) { mutableStateOf(currentEpisodeRailItemId.orEmpty()) }
    var pendingStepSeek by remember { mutableStateOf<TvPendingStepSeekUpdate?>(null) }
    var seekPreviewText by remember { mutableStateOf("") }
    var showSeekPreview by remember { mutableStateOf(false) }
    var showCenterFeedback by remember { mutableStateOf(false) }
    var centerFeedbackText by remember { mutableStateOf("") }
    var centerFeedbackIcon by remember { mutableStateOf(Icons.Filled.PlayArrow) }
    var persistentCenterFailureMessage by remember { mutableStateOf<String?>(null) }
    var persistentCenterFailureTargetEpisodeNumber by remember { mutableStateOf<Int?>(null) }
    var hasShownControlsOnce by remember { mutableStateOf(false) }
    var pendingRootFocusRequest by remember { mutableStateOf(false) }
    var pendingControlFocusTarget by remember { mutableStateOf<TvControlFocusTarget?>(null) }
    var pendingEpisodeRailFocusRequest by remember { mutableStateOf(false) }
    var playerFocusLayer by remember { mutableStateOf(TvPlayerFocusLayer.Root) }
    var lastFocusedControlTarget by remember { mutableStateOf(TvControlFocusTarget.PlayPause) }

    var hideControlsJob by remember { mutableStateOf<Job?>(null) }
    var hideSeekPreviewJob by remember { mutableStateOf<Job?>(null) }
    var hideCenterFeedbackJob by remember { mutableStateOf<Job?>(null) }
    var commitStepSeekJob by remember { mutableStateOf<Job?>(null) }
    fun effectiveDurationMs(): Long = durationMs.coerceAtLeast(0L)

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
        persistentCenterFailureMessage = null
        persistentCenterFailureTargetEpisodeNumber = null
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

    LaunchedEffect(episodeSwitchState) {
        when (episodeSwitchState) {
            is com.chee.videos.feature.tv.TvEpisodeSwitchUiState.Preparing -> {
                persistentCenterFailureMessage = null
                persistentCenterFailureTargetEpisodeNumber = null
                showTransientFeedback(
                    icon = Icons.Filled.PlayArrow,
                    text = "正在切到第 ${episodeSwitchState.targetEpisodeNumber} 集",
                    durationMs = 1_200L,
                )
            }

            is com.chee.videos.feature.tv.TvEpisodeSwitchUiState.Succeeded -> {
                showTransientFeedback(
                    icon = Icons.Filled.PlayArrow,
                    text = episodeSwitchState.message,
                    durationMs = 900L,
                )
            }

            is com.chee.videos.feature.tv.TvEpisodeSwitchUiState.Canceled -> {
                showTransientFeedback(
                    icon = Icons.Filled.Pause,
                    text = episodeSwitchState.message,
                    durationMs = 1_050L,
                )
            }

            is com.chee.videos.feature.tv.TvEpisodeSwitchUiState.Failed -> {
                hideCenterFeedbackJob?.cancel()
                showCenterFeedback = false
                persistentCenterFailureTargetEpisodeNumber = episodeSwitchState.targetEpisodeNumber
                persistentCenterFailureMessage = episodeSwitchState.message
            }

            null -> {
                persistentCenterFailureMessage = null
                persistentCenterFailureTargetEpisodeNumber = null
            }
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
            currentPositionMs = positionMs.coerceAtLeast(0L),
            durationMs = duration,
            deltaMs = deltaMs,
        )
        pendingStepSeek = pending
        updateSeekPreview(
            anchor = pending.anchorPositionMs,
            target = pending.targetPositionMs,
            duration = duration,
        )
        hideSeekPreviewJob?.cancel()
        commitStepSeekJob?.cancel()
        commitStepSeekJob = scope.launch {
            delay(TvStepSeekDebounceMillis)
            onSeekTo(pending.targetPositionMs)
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
                if (episodeRailItems.isEmpty()) {
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

    LaunchedEffect(episodeRailVisible) {
        onEpisodeRailVisibilityChanged(episodeRailVisible)
    }

    LaunchedEffect(openEpisodeRailRequestKey) {
        if (openEpisodeRailRequestKey <= 0 || episodeRailItems.isEmpty()) {
            return@LaunchedEffect
        }
        handleTvRemoteKeyAction(TvRemoteKeyAction.EnterEpisodeRail)
    }

    LaunchedEffect(Unit) {
        requestRootFocusWhenReady()
    }

    LaunchedEffect(controlsVisible) {
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
        subtitleSheetVisible = false,
        audioTrackSheetVisible = false,
        resumePromptVisible = resumePromptVisible,
        backConfirmPromptVisible = backConfirmPromptVisible,
        playerErrorVisible = playerErrorVisible,
    )
    var previousFocusGuardInput by remember { mutableStateOf(currentFocusGuardInput) }
    LaunchedEffect(currentFocusGuardInput) {
        if (shouldReclaimRootFocus(previousFocusGuardInput, currentFocusGuardInput)) {
            if (episodeRailVisible) {
                previousFocusGuardInput = currentFocusGuardInput
                return@LaunchedEffect
            }
            requestRootFocusWhenReady()
        }
        previousFocusGuardInput = currentFocusGuardInput
    }

    LaunchedTvInitialFocus(true, pendingRootFocusRequest) {
        if (!pendingRootFocusRequest) {
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
            TvControlFocusTarget.Subtitle -> subtitleFocusRequester
            TvControlFocusTarget.AudioTrack -> audioTrackFocusRequester
            else -> playPauseFocusRequester
        }
    }

    LaunchedTvInitialFocus(true, controlsVisible, pendingControlFocusTarget) {
        val target = pendingControlFocusTarget
        if (!controlsVisible || target == null) {
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

    LaunchedTvInitialFocus(true, episodeRailVisible, episodeRailOpenNonce, pendingEpisodeRailFocusRequest) {
        if (!episodeRailVisible || !pendingEpisodeRailFocusRequest) {
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

    val actualDurationMs = effectiveDurationMs()
    val displayPositionMs = pendingStepSeek?.targetPositionMs ?: positionMs.coerceAtLeast(0L)
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
        if (showTrackActions) {
            add(TvControlFocusTarget.Subtitle)
            add(TvControlFocusTarget.AudioTrack)
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
                if (event.nativeKeyEvent.action != AndroidKeyEvent.ACTION_DOWN) {
                    return@onPreviewKeyEvent false
                }
                when (event.nativeKeyEvent.keyCode) {
                    AndroidKeyEvent.KEYCODE_DPAD_LEFT -> requestHorizontalControlFocus(TvControlFocusDirection.Left)
                    AndroidKeyEvent.KEYCODE_DPAD_RIGHT -> requestHorizontalControlFocus(TvControlFocusDirection.Right)
                    AndroidKeyEvent.KEYCODE_DPAD_DOWN -> handleTvRemoteKeyAction(TvRemoteKeyAction.EnterEpisodeRail)
                    else -> false
                }
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
            .background(Color.Transparent)
            .focusRequester(rootFocusRequester)
            .focusable()
            .onFocusChanged { focusState ->
                if (focusState.isFocused) {
                    playerFocusLayer = TvPlayerFocusLayer.Root
                }
                if (
                    !focusState.hasFocus &&
                    !currentFocusGuardInput.anyOverlayVisible() &&
                    playerFocusLayer == TvPlayerFocusLayer.Root &&
                    !episodeRailVisible
                ) {
                    requestRootFocusWhenReady()
                }
            }
            .onPreviewKeyEvent { event ->
                if (event.nativeKeyEvent.action != AndroidKeyEvent.ACTION_DOWN) {
                    return@onPreviewKeyEvent false
                }
                if (currentFocusGuardInput.anyOverlayVisible()) {
                    return@onPreviewKeyEvent false
                }
                val action = resolveTvRemoteKeyAction(
                    visible = controlsVisible,
                    focusLayer = playerFocusLayer,
                    episodeRailEnabled = episodeRailItems.isNotEmpty(),
                    keyCode = event.nativeKeyEvent.keyCode,
                    repeatCount = event.nativeKeyEvent.repeatCount,
                    seekStepSec = tvSeekStepSeconds,
                ) ?: return@onPreviewKeyEvent false
                handleTvRemoteKeyAction(action)
            },
    ) {
        AnimatedVisibility(
            visible = showSeekPreview,
            enter = fadeIn(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
            exit = fadeOut(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
            modifier = Modifier.align(Alignment.Center),
        ) {
            Surface(
                color = PlayerGlassSurfaceStrong,
                shape = AppChrome.SurfaceShape,
            ) {
                Text(
                    text = seekPreviewText,
                    color = AppChrome.TextPrimary,
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
                color = PlayerGlassSurface,
                shape = AppChrome.SurfaceShape,
            ) {
                Row(
                    modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
                    verticalAlignment = Alignment.CenterVertically,
                ) {
                    Icon(
                        imageVector = centerFeedbackIcon,
                        contentDescription = null,
                        tint = AppChrome.TextPrimary,
                        modifier = Modifier.padding(end = 8.dp),
                    )
                    Text(
                        text = centerFeedbackText,
                        color = AppChrome.TextPrimary,
                        style = MaterialTheme.typography.bodyMedium,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                )
            }
        }

        if (!persistentCenterFailureMessage.isNullOrBlank()) {
            Surface(
                color = PlayerGlassSurfaceStrong,
                shape = AppChrome.SurfaceShape,
                modifier = Modifier.align(Alignment.Center),
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
                            text = persistentCenterFailureMessage.orEmpty(),
                            color = AppChrome.TextPrimary,
                            style = MaterialTheme.typography.bodyMedium,
                            maxLines = 2,
                            overflow = TextOverflow.Ellipsis,
                        )
                    }
                    Surface(
                        color = AppChrome.AccentSoft,
                        shape = AppChrome.ChipShape,
                        modifier = Modifier
                            .tvFocusableGlow(shape = AppChrome.ChipShape, focusedScale = 1.04f)
                            .clickable(onClick = onDismissEpisodeSwitchFeedback),
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
                                text = "留在当前分集",
                                color = AppChrome.TextPrimary,
                                style = MaterialTheme.typography.labelLarge,
                            )
                        }
                    }
                }
            }
        }
        }

        TvLongFormTitleOverlay(
            data = titleOverlayData,
            visible = controlsVisible,
            showStatusBarPadding = showStatusBarPadding,
            modifier = Modifier.align(Alignment.TopStart),
        )

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
                    color = PlayerGlassSurface,
                    shape = AppChrome.SurfaceShape,
                    modifier = Modifier.fillMaxWidth(),
                ) {
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
                        label = "TvSeriesCoreBottomPanelPage",
                    ) { page ->
                        when (page) {
                            TvSeriesBottomPanelPage.Controls -> TvSeriesControlsPage(
                                isPlaying = isPlaying,
                                displayPositionMs = displayPositionMs,
                                actualDurationMs = actualDurationMs,
                                progressValue = progressValue,
                                tvMode = true,
                                playPauseFocusRequester = playPauseFocusRequester,
                                subtitleFocusRequester = subtitleFocusRequester,
                                audioTrackFocusRequester = audioTrackFocusRequester,
                                controlFocusModifier = ::controlFocusModifier,
                                showTrackActions = showTrackActions,
                                onTogglePlayPause = { togglePlaybackWithFeedback() },
                                onOpenSubtitle = {
                                    onOpenSubtitle()
                                    showControlsTemporarily()
                                },
                                onOpenAudioTrack = {
                                    onOpenAudioTrack()
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
                }
            }
        }

        resumePromptSlot?.invoke(this)
    }
}
