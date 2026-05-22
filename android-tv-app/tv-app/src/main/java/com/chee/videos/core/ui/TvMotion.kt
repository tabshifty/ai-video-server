package com.chee.videos.core.ui

import androidx.compose.animation.core.CubicBezierEasing
import androidx.compose.animation.core.Easing

object TvMotionTokens {
    const val DurationFastMs: Int = 200
    const val DurationStandardMs: Int = 240
    const val DurationEmphasizedMs: Int = 260

    val EasingStandard: Easing = CubicBezierEasing(0.2f, 0f, 0f, 1f)
}
