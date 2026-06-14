package com.chee.videos.core.ui

import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.darkColorScheme
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp

object AppChrome {
    val Canvas = Color(0xFF040508)
    val CanvasRaised = Color(0xFF080A0D)
    val Surface = Color(0x7810161F)
    val SurfaceElevated = Color(0x98241C11)
    val SurfaceMuted = Color(0xAA141820)
    val SurfaceStrong = Color(0xFF30291E)
    val Divider = Color(0x33E8B85B)
    val Accent = Color(0xFFE8B85B)
    val AccentSoft = Color(0x40E8B85B)
    val AccentStrong = Color(0xFFEFC463)
    val AccentWarm = Color(0xFFD6A64F)
    val TextPrimary = Color(0xFFF7F1E6)
    val TextSecondary = Color(0xFFD7DCE5)
    val TextMuted = Color(0xFFB0BAC8)
    val TextSubtle = Color(0xFF8D96A4)
    val Error = Color(0xFFFF8A7A)

    val PageGradient: Brush
        get() = Brush.verticalGradient(
            colors = listOf(
                Color(0xFF10161F),
                CanvasRaised,
                Canvas,
            ),
        )

    val HeroGradient: Brush
        get() = Brush.verticalGradient(
            colors = listOf(
                Color(0xFF20180E),
                Color(0xFF0A0D12),
                Canvas,
            ),
        )

    val RadiusDp: Dp = 8.dp
    val SurfaceShape: Shape = RoundedCornerShape(RadiusDp)
    val ChipRadiusDp: Dp = 8.dp
    val ChipShape: Shape = RoundedCornerShape(ChipRadiusDp)
    val PillShape: Shape = RoundedCornerShape(999.dp)
}

val AppDarkColors = darkColorScheme(
    primary = AppChrome.Accent,
    secondary = AppChrome.AccentStrong,
    tertiary = AppChrome.AccentWarm,
    background = AppChrome.Canvas,
    surface = AppChrome.Surface,
    surfaceVariant = AppChrome.SurfaceElevated,
    onPrimary = AppChrome.Canvas,
    onSecondary = AppChrome.Canvas,
    onTertiary = AppChrome.Canvas,
    onBackground = AppChrome.TextPrimary,
    onSurface = AppChrome.TextPrimary,
    onSurfaceVariant = AppChrome.TextMuted,
    error = AppChrome.Error,
    onError = AppChrome.Canvas,
)
