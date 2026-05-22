package com.chee.videos.core.ui

import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.darkColorScheme
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp

object AppChrome {
    val Canvas = Color(0xFF040508)
    val CanvasRaised = Color(0xFF090B11)
    val Surface = Color(0xFF10141B)
    val SurfaceElevated = Color(0xFF161B24)
    val SurfaceMuted = Color(0xFF1C2330)
    val SurfaceStrong = Color(0xFF20293A)
    val Divider = Color(0x1FFFFFFF)
    val Accent = Color(0xFFE11D48)
    val AccentSoft = Color(0x40E11D48)
    val AccentStrong = Color(0xFFFF5A7A)
    val AccentWarm = Color(0xFFD4A647)
    val TextPrimary = Color(0xFFF8FAFC)
    val TextSecondary = Color(0xFFCCD3DF)
    val TextMuted = Color(0xFFB0BAC8)
    val TextSubtle = Color(0xFF6F7B90)
    val Error = Color(0xFFFF7B8D)

    val PageGradient: Brush
        get() = Brush.verticalGradient(
            colors = listOf(
                Color(0xFF111522),
                CanvasRaised,
                Canvas,
            ),
        )

    val HeroGradient: Brush
        get() = Brush.verticalGradient(
            colors = listOf(
                Color(0xFF181D2A),
                Color(0xFF0B0E15),
                Canvas,
            ),
        )

    val CardShape = RoundedCornerShape(22.dp)
    val SectionShape = RoundedCornerShape(18.dp)
    val PillShape = RoundedCornerShape(999.dp)
}

val AppDarkColors = darkColorScheme(
    primary = AppChrome.Accent,
    secondary = AppChrome.AccentStrong,
    tertiary = AppChrome.AccentWarm,
    background = AppChrome.Canvas,
    surface = AppChrome.Surface,
    surfaceVariant = AppChrome.SurfaceElevated,
    onPrimary = Color.White,
    onSecondary = Color.White,
    onTertiary = AppChrome.Canvas,
    onBackground = AppChrome.TextPrimary,
    onSurface = AppChrome.TextPrimary,
    onSurfaceVariant = AppChrome.TextMuted,
    error = AppChrome.Error,
    onError = AppChrome.Canvas,
)
