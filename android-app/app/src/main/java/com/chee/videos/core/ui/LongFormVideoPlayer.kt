package com.chee.videos.core.ui

import android.graphics.Color as AndroidColor
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectHorizontalDragGestures
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material.icons.filled.FastForward
import androidx.compose.material.icons.filled.Fullscreen
import androidx.compose.material.icons.filled.FullscreenExit
import androidx.compose.material.icons.filled.Pause
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material.icons.filled.Subtitles
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
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
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.layout.onSizeChanged
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.media3.common.PlaybackParameters
import androidx.media3.common.Player
import androidx.media3.exoplayer.ExoPlayer
import androidx.media3.ui.AspectRatioFrameLayout
import androidx.media3.ui.PlayerView
import coil.compose.AsyncImage
import com.chee.videos.core.model.SubtitleTrackDto
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlin.math.abs

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
) {
    val scope = rememberCoroutineScope()

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

    var longPressBoosting by remember { mutableStateOf(false) }
    var ignoreTapAfterLongPress by remember { mutableStateOf(false) }

    var showCenterFeedback by remember { mutableStateOf(false) }
    var centerFeedbackText by remember { mutableStateOf("") }
    var centerFeedbackIcon by remember { mutableStateOf(Icons.Filled.PlayArrow) }
    var subtitleSheetVisible by remember { mutableStateOf(false) }

    var hideControlsJob by remember { mutableStateOf<Job?>(null) }
    var hideSeekPreviewJob by remember { mutableStateOf<Job?>(null) }
    var hideCenterFeedbackJob by remember { mutableStateOf<Job?>(null) }

    fun effectiveDurationMs(): Long {
        val fromState = durationMs
        if (fromState > 0) {
            return fromState
        }
        return player.duration.coerceAtLeast(0L)
    }

    fun scheduleAutoHideControls() {
        hideControlsJob?.cancel()
        hideControlsJob = scope.launch {
            delay(5000)
            controlsVisible = false
        }
    }

    fun showControlsTemporarily() {
        controlsVisible = true
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

    fun togglePlaybackWithFeedback() {
        val shouldPause = isPlaying
        onTogglePlayPause()
        showTransientFeedback(
            icon = if (shouldPause) Icons.Filled.Pause else Icons.Filled.PlayArrow,
            text = if (shouldPause) "已暂停" else "继续播放",
        )
        showControlsTemporarily()
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
            }
        }
        player.addListener(listener)
        isPlaying = player.isPlaying
        durationMs = player.duration.coerceAtLeast(0L)
        positionMs = player.currentPosition.coerceAtLeast(0L)
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

    Box(
        modifier = modifier
            .fillMaxSize()
            .background(Color.Black)
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
                }
            },
            update = { view ->
                view.player = player
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
            enter = fadeIn(),
            exit = fadeOut(),
            modifier = Modifier.align(Alignment.Center),
        ) {
            Surface(
                color = Color(0xD610131A),
                shape = RoundedCornerShape(18.dp),
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
            enter = fadeIn(),
            exit = fadeOut(),
            modifier = Modifier.align(Alignment.Center),
        ) {
            Surface(
                color = Color(0xCC0D1016),
                shape = RoundedCornerShape(22.dp),
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
                    )
                    Text(
                        text = centerFeedbackText,
                        color = Color.White,
                        style = MaterialTheme.typography.bodyMedium,
                    )
                }
            }
        }

        AnimatedVisibility(
            visible = controlsVisible,
            enter = fadeIn(),
            exit = fadeOut(),
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
                    shape = RoundedCornerShape(20.dp),
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

        AnimatedVisibility(
            visible = controlsVisible,
            enter = fadeIn(),
            exit = fadeOut(),
            modifier = Modifier.align(Alignment.BottomCenter),
        ) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(horizontal = 12.dp, vertical = 12.dp),
            ) {
                Surface(
                    color = Color(0x910D1016),
                    shape = RoundedCornerShape(20.dp),
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
                            onClick = { togglePlaybackWithFeedback() },
                        )
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
                            modifier = Modifier.weight(1f),
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
                            onClick = {
                                subtitleSheetVisible = true
                                showControlsTemporarily()
                            },
                        )
                        CompactPlayerControlButton(
                            icon = if (isFullscreen) Icons.Filled.FullscreenExit else Icons.Filled.Fullscreen,
                            contentDescription = if (isFullscreen) "退出全屏" else "全屏",
                            onClick = {
                                onToggleFullscreen()
                                showControlsTemporarily()
                            },
                        )
                    }
                }
            }
        }

        if (subtitleSheetVisible) {
            ModalBottomSheet(
                onDismissRequest = { subtitleSheetVisible = false },
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(horizontal = 20.dp, vertical = 8.dp),
                    verticalArrangement = Arrangement.spacedBy(12.dp),
                ) {
                    Text(
                        text = "字幕",
                        style = MaterialTheme.typography.titleMedium,
                        fontWeight = FontWeight.SemiBold,
                    )
                    LazyColumn(
                        modifier = Modifier.fillMaxWidth(),
                        verticalArrangement = Arrangement.spacedBy(8.dp),
                    ) {
                        item {
                            SubtitleOptionRow(
                                label = "关闭字幕",
                                selected = selectedSubtitleTrackId.isNullOrBlank(),
                                onClick = {
                                    onSelectSubtitleTrack(null)
                                    subtitleSheetVisible = false
                                },
                            )
                        }
                        items(subtitleTracks, key = { it.id }) { track ->
                            SubtitleOptionRow(
                                label = subtitleTrackDisplayLabel(track),
                                selected = selectedSubtitleTrackId == track.id,
                                onClick = {
                                    onSelectSubtitleTrack(track.id)
                                    subtitleSheetVisible = false
                                },
                            )
                        }
                    }
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
) {
    Surface(
        color = Color(0x24000000),
        shape = CircleShape,
        modifier = modifier,
    ) {
        IconButton(
            onClick = onClick,
            modifier = Modifier.size(34.dp),
        ) {
            Icon(
                imageVector = icon,
                contentDescription = contentDescription,
                tint = Color.White,
            )
        }
    }
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

@Composable
private fun SubtitleOptionRow(
    label: String,
    selected: Boolean,
    onClick: () -> Unit,
) {
    Surface(
        color = if (selected) Color(0x26FFFFFF) else Color(0x12000000),
        shape = RoundedCornerShape(14.dp),
        modifier = Modifier.fillMaxWidth(),
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 14.dp, vertical = 12.dp)
                .pointerInput(label, selected) {
                    detectTapGestures(onTap = { onClick() })
                },
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Text(
                text = label,
                color = Color.White,
                style = MaterialTheme.typography.bodyMedium,
                modifier = Modifier.weight(1f),
            )
            if (selected) {
                Text(
                    text = "已选中",
                    color = Color.White.copy(alpha = 0.72f),
                    style = MaterialTheme.typography.labelSmall,
                )
            }
        }
    }
}
