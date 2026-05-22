package com.chee.videos.core.ui

import androidx.compose.animation.core.CubicBezierEasing
import androidx.compose.animation.core.Easing
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import kotlinx.coroutines.delay

object TvListMotionTokens {
    const val StaggerPerItemMs: Long = 35L
    const val StaggerEntryDurationMs: Int = 260
    val StaggerEntryDistanceDp: Dp = 12.dp
    const val StaggerMaxSteps: Int = 12
    val StaggerEntryEasing: Easing = CubicBezierEasing(0.2f, 0f, 0f, 1f)
}

fun tvStaggerEntryDelayMs(
    index: Int,
    perItemDelayMs: Long = TvListMotionTokens.StaggerPerItemMs,
    maxSteps: Int = TvListMotionTokens.StaggerMaxSteps,
): Long {
    val clamped = index.coerceIn(0, maxSteps)
    return clamped * perItemDelayMs
}

@Composable
fun Modifier.tvStaggerEntry(index: Int): Modifier {
    var visible by remember { mutableStateOf(false) }
    LaunchedEffect(Unit) {
        val delayMs = tvStaggerEntryDelayMs(index)
        if (delayMs > 0L) {
            delay(delayMs)
        }
        visible = true
    }
    val progress by animateFloatAsState(
        targetValue = if (visible) 1f else 0f,
        animationSpec = tween(
            durationMillis = TvListMotionTokens.StaggerEntryDurationMs,
            easing = TvListMotionTokens.StaggerEntryEasing,
        ),
        label = "tvStaggerEntry",
    )
    val distancePx = with(LocalDensity.current) {
        TvListMotionTokens.StaggerEntryDistanceDp.toPx()
    }
    return this.graphicsLayer {
        alpha = progress
        translationY = (1f - progress) * distancePx
    }
}
