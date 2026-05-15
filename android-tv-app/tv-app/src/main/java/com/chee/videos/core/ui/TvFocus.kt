package com.chee.videos.core.ui

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.border
import androidx.compose.foundation.focusable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.composed
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.foundation.shape.RoundedCornerShape

object TvFocusSafeSpec {
    const val posterCardWidthDp: Float = 146f
    const val posterFocusedScale: Float = 1.04f
    const val focusedBorderWidthDp: Float = 2f
    const val posterFocusSafeSpaceDp: Float = 8f

    fun requiredSafeSpaceDp(
        baseSizeDp: Float,
        focusedScale: Float,
        focusedBorderWidthDp: Float,
    ): Float {
        val scaleOverflow = baseSizeDp * (focusedScale - 1f) / 2f
        return scaleOverflow + focusedBorderWidthDp
    }
}

fun Modifier.tvFocusableGlow(
    enabled: Boolean = true,
    shape: Shape = RoundedCornerShape(20.dp),
    focusedScale: Float = 1.04f,
    focusedBorderWidth: Dp = 2.dp,
): Modifier = composed {
    var isFocused by remember { mutableStateOf(false) }
    val scale by animateFloatAsState(
        targetValue = if (isFocused && enabled) focusedScale else 1f,
        animationSpec = tween(durationMillis = 140),
        label = "tvFocusScale",
    )
    val borderAlpha by animateFloatAsState(
        targetValue = if (isFocused && enabled) 1f else 0f,
        animationSpec = tween(durationMillis = 140),
        label = "tvFocusBorderAlpha",
    )
    this
        .onFocusChanged { state -> isFocused = state.isFocused || state.hasFocus }
        .graphicsLayer {
            scaleX = scale
            scaleY = scale
            shadowElevation = if (isFocused && enabled) 32f else 0f
            this.shape = shape
            clip = false
        }
        .border(
            width = focusedBorderWidth,
            color = androidx.compose.ui.graphics.Color(0xFFFF5A7A).copy(alpha = borderAlpha),
            shape = shape,
        )
        .focusable(enabled = enabled)
}
