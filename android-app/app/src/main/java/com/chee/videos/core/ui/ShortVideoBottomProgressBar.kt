package com.chee.videos.core.ui

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.animateDpAsState
import androidx.compose.animation.core.spring
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectHorizontalDragGestures
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.unit.dp

@Composable
fun ShortVideoBottomProgressBar(
    modifier: Modifier = Modifier,
    positionMs: Long,
    durationMs: Long,
    isScrubbing: Boolean,
    scrubTargetMs: Long,
    onScrubStart: () -> Unit,
    onScrubToFraction: (Float) -> Unit,
    onScrubEnd: () -> Unit,
) {
    val barHeight = animateDpAsState(
        targetValue = if (isScrubbing) 6.dp else 2.dp,
        animationSpec = spring(stiffness = 520f),
        label = "shared_short_progress_height",
    )
    val displayPositionMs = if (isScrubbing) scrubTargetMs else positionMs
    val progress = shortProgressFraction(displayPositionMs, durationMs)
    val canSeek = durationMs > 0L

    Column(
        modifier = modifier.fillMaxWidth(),
        horizontalAlignment = Alignment.CenterHorizontally,
    ) {
        AnimatedVisibility(visible = isScrubbing) {
            Surface(
                color = AppChrome.Surface.copy(alpha = 0.96f),
                shape = RoundedCornerShape(12.dp),
                modifier = Modifier.padding(bottom = 8.dp),
                contentColor = AppChrome.TextPrimary,
            ) {
                Text(
                    text = "${formatPlaybackTimeHms(displayPositionMs)} / ${formatPlaybackTimeHms(durationMs)}",
                    color = AppChrome.TextPrimary,
                    style = MaterialTheme.typography.labelMedium,
                    modifier = Modifier.padding(horizontal = 12.dp, vertical = 7.dp),
                )
            }
        }

        Box(
            modifier = Modifier
                .fillMaxWidth()
                .height(if (isScrubbing) 26.dp else 18.dp)
                .pointerInput(canSeek) {
                    if (!canSeek) {
                        return@pointerInput
                    }
                    detectHorizontalDragGestures(
                        onDragStart = { offset ->
                            onScrubStart()
                            val width = size.width.toFloat().coerceAtLeast(1f)
                            onScrubToFraction((offset.x / width).coerceIn(0f, 1f))
                        },
                        onHorizontalDrag = { change, _ ->
                            change.consume()
                            val width = size.width.toFloat().coerceAtLeast(1f)
                            onScrubToFraction((change.position.x / width).coerceIn(0f, 1f))
                        },
                        onDragEnd = onScrubEnd,
                        onDragCancel = onScrubEnd,
                    )
                },
            contentAlignment = Alignment.BottomStart,
        ) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(barHeight.value)
                    .clip(CircleShape)
                    .background(AppChrome.TextPrimary.copy(alpha = 0.18f)),
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxWidth(progress)
                        .fillMaxHeight()
                        .background(AppChrome.AccentStrong),
                )
            }
        }
    }
}

private fun formatPlaybackTimeHms(ms: Long): String {
    val totalSeconds = ms.coerceAtLeast(0L) / 1000L
    val hours = totalSeconds / 3600L
    val minutes = (totalSeconds % 3600L) / 60L
    val seconds = totalSeconds % 60L
    return buildString {
        append(hours.toString().padStart(2, '0'))
        append(':')
        append(minutes.toString().padStart(2, '0'))
        append(':')
        append(seconds.toString().padStart(2, '0'))
    }
}

private fun shortProgressFraction(positionMs: Long, durationMs: Long): Float {
    if (durationMs <= 0L) {
        return 0f
    }
    return (positionMs.toFloat() / durationMs.toFloat()).coerceIn(0f, 1f)
}
