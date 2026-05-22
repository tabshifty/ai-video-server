package com.chee.videos.core.ui

import android.os.Build
import android.view.HapticFeedbackConstants
import android.view.View
import androidx.compose.animation.core.SpringSpec
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
import androidx.compose.ui.input.key.Key
import androidx.compose.ui.input.key.KeyEventType
import androidx.compose.ui.input.key.key
import androidx.compose.ui.input.key.onPreviewKeyEvent
import androidx.compose.ui.input.key.type
import androidx.compose.ui.platform.LocalView
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
    const val PressedScale: Float = 0.97f
    const val PressDampingRatio: Float = 0.7f
    const val PressStiffness: Float = 720f
}

private val TvPressKeys: Set<Key> = setOf(Key.DirectionCenter, Key.Enter, Key.NumPadEnter)

internal fun isTvPressKey(key: Key): Boolean = key in TvPressKeys

internal fun resolveTvFocusableScaleTarget(
    focused: Boolean,
    pressed: Boolean,
    enabled: Boolean,
    focusedScale: Float,
): Float = when {
    !enabled -> 1f
    pressed -> TvFocusMotionTokens.PressedScale
    focused -> focusedScale
    else -> 1f
}

internal fun tvFocusableScaleSpring(pressed: Boolean): SpringSpec<Float> = if (pressed) {
    spring(
        dampingRatio = TvFocusMotionTokens.PressDampingRatio,
        stiffness = TvFocusMotionTokens.PressStiffness,
    )
} else {
    spring(
        dampingRatio = TvFocusMotionTokens.ScaleDampingRatio,
        stiffness = TvFocusMotionTokens.ScaleStiffness,
    )
}

internal fun performTvPressHapticFeedback(view: View) {
    val constant = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.R) {
        HapticFeedbackConstants.CONFIRM
    } else {
        HapticFeedbackConstants.VIRTUAL_KEY
    }
    view.performHapticFeedback(constant)
}

fun Modifier.tvFocusableGlow(
    enabled: Boolean = true,
    shape: Shape = RoundedCornerShape(20.dp),
    focusedScale: Float = 1.04f,
): Modifier = composed {
    var isFocused by remember { mutableStateOf(false) }
    var isPressed by remember { mutableStateOf(false) }
    val view = LocalView.current
    val scale by animateFloatAsState(
        targetValue = resolveTvFocusableScaleTarget(
            focused = isFocused,
            pressed = isPressed,
            enabled = enabled,
            focusedScale = focusedScale,
        ),
        animationSpec = tvFocusableScaleSpring(isPressed),
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
        .onFocusChanged { state ->
            isFocused = state.isFocused || state.hasFocus
            if (!isFocused) isPressed = false
        }
        .onPreviewKeyEvent { event ->
            if (!enabled || !isFocused) return@onPreviewKeyEvent false
            if (!isTvPressKey(event.key)) return@onPreviewKeyEvent false
            when (event.type) {
                KeyEventType.KeyDown -> isPressed = true
                KeyEventType.KeyUp -> {
                    isPressed = false
                    performTvPressHapticFeedback(view)
                }
                else -> {}
            }
            false
        }
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
    var isPressed by remember { mutableStateOf(false) }
    val view = LocalView.current
    val scale by animateFloatAsState(
        targetValue = resolveTvFocusableScaleTarget(
            focused = isFocused,
            pressed = isPressed,
            enabled = enabled,
            focusedScale = focusedScale,
        ),
        animationSpec = tvFocusableScaleSpring(isPressed),
        label = "tvFocusScaleOnly",
    )
    this
        .onFocusChanged { state ->
            isFocused = state.isFocused || state.hasFocus
            if (!isFocused) isPressed = false
        }
        .onPreviewKeyEvent { event ->
            if (!enabled || !isFocused) return@onPreviewKeyEvent false
            if (!isTvPressKey(event.key)) return@onPreviewKeyEvent false
            when (event.type) {
                KeyEventType.KeyDown -> isPressed = true
                KeyEventType.KeyUp -> {
                    isPressed = false
                    performTvPressHapticFeedback(view)
                }
                else -> {}
            }
            false
        }
        .graphicsLayer {
            scaleX = scale
            scaleY = scale
            shadowElevation = if (isFocused && enabled) 28f else 0f
            this.shape = shape
            clip = false
        }
        .focusable(enabled = enabled)
}
