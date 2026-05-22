package com.chee.videos.core.ui

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.spring
import androidx.compose.foundation.background
import androidx.compose.foundation.focusable
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.composed
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.unit.dp

val TvFocusGlowColor = Color(0xFF39D7E8)
private val TvFocusGlowSurface = Color(0x2639D7E8)

object TvFocusSafeSpec {
    const val posterCardWidthDp: Float = 146f
    const val posterFocusedScale: Float = 1.04f
    const val focusedHaloPaddingDp: Float = 2f
    const val posterFocusSafeSpaceDp: Float = 8f

    fun requiredSafeSpaceDp(
        baseSizeDp: Float,
        focusedScale: Float,
        focusedHaloPaddingDp: Float,
    ): Float {
        val scaleOverflow = baseSizeDp * (focusedScale - 1f) / 2f
        return scaleOverflow + focusedHaloPaddingDp
    }
}

object TvFocusMotionTokens {
    const val ScaleDampingRatio: Float = 0.8f
    const val ScaleStiffness: Float = 380f
    const val SurfaceDampingRatio: Float = 1f
    const val SurfaceStiffness: Float = 620f
}

fun Modifier.tvFocusableGlow(
    enabled: Boolean = true,
    shape: Shape = RoundedCornerShape(20.dp),
    focusedScale: Float = 1.04f,
): Modifier = composed {
    var isFocused by remember { mutableStateOf(false) }
    val scale by animateFloatAsState(
        targetValue = if (isFocused && enabled) focusedScale else 1f,
        animationSpec = spring(
            dampingRatio = TvFocusMotionTokens.ScaleDampingRatio,
            stiffness = TvFocusMotionTokens.ScaleStiffness,
        ),
        label = "tvFocusScale",
    )
    val surfaceAlpha by animateFloatAsState(
        targetValue = if (isFocused && enabled) 1f else 0f,
        animationSpec = spring(
            dampingRatio = TvFocusMotionTokens.SurfaceDampingRatio,
            stiffness = TvFocusMotionTokens.SurfaceStiffness,
        ),
        label = "tvFocusSurfaceAlpha",
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
        .background(color = TvFocusGlowSurface.copy(alpha = surfaceAlpha), shape = shape)
        .focusable(enabled = enabled)
}

fun Modifier.tvFocusableScaleOnly(
    enabled: Boolean = true,
    shape: Shape = RoundedCornerShape(20.dp),
    focusedScale: Float = 1.04f,
): Modifier = composed {
    var isFocused by remember { mutableStateOf(false) }
    val scale by animateFloatAsState(
        targetValue = if (isFocused && enabled) focusedScale else 1f,
        animationSpec = spring(
            dampingRatio = TvFocusMotionTokens.ScaleDampingRatio,
            stiffness = TvFocusMotionTokens.ScaleStiffness,
        ),
        label = "tvFocusScaleOnly",
    )
    this
        .onFocusChanged { state -> isFocused = state.isFocused || state.hasFocus }
        .graphicsLayer {
            scaleX = scale
            scaleY = scale
            shadowElevation = if (isFocused && enabled) 28f else 0f
            this.shape = shape
            clip = false
        }
        .focusable(enabled = enabled)
}
