package com.chee.videos.core.ui

import androidx.compose.material3.Typography
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.sp

object TvTypographyTokens {
    const val HeroTitleSp: Int = 30
    const val LargeTitleSp: Int = 26
    const val SectionTitleSp: Int = 20
    const val BodySp: Int = 14
    const val HelperSp: Int = 12
    const val CaptionSp: Int = 11
}

val TvTypography: Typography = Typography(
    headlineLarge = TextStyle(
        fontSize = TvTypographyTokens.HeroTitleSp.sp,
        lineHeight = 36.sp,
        letterSpacing = 0.sp,
        fontWeight = FontWeight.SemiBold,
    ),
    headlineMedium = TextStyle(
        fontSize = TvTypographyTokens.LargeTitleSp.sp,
        lineHeight = 32.sp,
        letterSpacing = 0.sp,
        fontWeight = FontWeight.SemiBold,
    ),
    headlineSmall = TextStyle(
        fontSize = 23.sp,
        lineHeight = 29.sp,
        letterSpacing = 0.sp,
        fontWeight = FontWeight.SemiBold,
    ),
    titleLarge = TextStyle(
        fontSize = TvTypographyTokens.SectionTitleSp.sp,
        lineHeight = 26.sp,
        letterSpacing = 0.sp,
        fontWeight = FontWeight.SemiBold,
    ),
    titleMedium = TextStyle(
        fontSize = 17.sp,
        lineHeight = 22.sp,
        letterSpacing = 0.sp,
        fontWeight = FontWeight.Medium,
    ),
    titleSmall = TextStyle(
        fontSize = 15.sp,
        lineHeight = 20.sp,
        letterSpacing = 0.sp,
        fontWeight = FontWeight.Medium,
    ),
    bodyLarge = TextStyle(
        fontSize = TvTypographyTokens.BodySp.sp,
        lineHeight = 20.sp,
        letterSpacing = 0.sp,
        fontWeight = FontWeight.Normal,
    ),
    bodyMedium = TextStyle(
        fontSize = 13.sp,
        lineHeight = 18.sp,
        letterSpacing = 0.sp,
        fontWeight = FontWeight.Normal,
    ),
    bodySmall = TextStyle(
        fontSize = TvTypographyTokens.HelperSp.sp,
        lineHeight = 16.sp,
        letterSpacing = 0.sp,
        fontWeight = FontWeight.Normal,
    ),
    labelLarge = TextStyle(
        fontSize = 13.sp,
        lineHeight = 18.sp,
        letterSpacing = 0.sp,
        fontWeight = FontWeight.Medium,
    ),
    labelMedium = TextStyle(
        fontSize = TvTypographyTokens.HelperSp.sp,
        lineHeight = 16.sp,
        letterSpacing = 0.sp,
        fontWeight = FontWeight.Medium,
    ),
    labelSmall = TextStyle(
        fontSize = TvTypographyTokens.CaptionSp.sp,
        lineHeight = 15.sp,
        letterSpacing = 0.sp,
        fontWeight = FontWeight.Medium,
    ),
)
