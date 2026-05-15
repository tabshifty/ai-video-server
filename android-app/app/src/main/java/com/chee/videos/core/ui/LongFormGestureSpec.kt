package com.chee.videos.core.ui

import kotlin.math.roundToInt

internal enum class LongFormVerticalAdjustmentTarget {
    Brightness,
    Volume,
}

internal fun longFormVerticalAdjustmentTarget(
    isFullscreen: Boolean,
    startX: Float,
    widthPx: Float,
): LongFormVerticalAdjustmentTarget? {
    if (!isFullscreen || widthPx <= 0f) {
        return null
    }
    return if (startX < widthPx / 2f) {
        LongFormVerticalAdjustmentTarget.Brightness
    } else {
        LongFormVerticalAdjustmentTarget.Volume
    }
}

internal fun longFormAdjustedPercent(
    startPercent: Float,
    dragDistanceY: Float,
    heightPx: Float,
): Float {
    if (heightPx <= 0f) {
        return startPercent.coerceIn(0f, 1f)
    }
    return (startPercent - dragDistanceY / heightPx).coerceIn(0f, 1f)
}

internal fun longFormBrightnessPercent(windowBrightness: Float): Float {
    if (windowBrightness < 0f) {
        return 0.5f
    }
    return windowBrightness.coerceIn(0.01f, 1f)
}

internal fun longFormVolumePercent(currentVolume: Int, maxVolume: Int): Float {
    if (maxVolume <= 0) {
        return 0f
    }
    return (currentVolume.toFloat() / maxVolume.toFloat()).coerceIn(0f, 1f)
}

internal fun longFormAdjustmentFeedbackText(
    target: LongFormVerticalAdjustmentTarget,
    percent: Float,
): String {
    val label = when (target) {
        LongFormVerticalAdjustmentTarget.Brightness -> "亮度"
        LongFormVerticalAdjustmentTarget.Volume -> "音量"
    }
    val value = (percent.coerceIn(0f, 1f) * 100f).roundToInt()
    return "$label $value%"
}
