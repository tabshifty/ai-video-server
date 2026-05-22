package com.chee.videos.core.ui

import androidx.compose.material3.Typography
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.sp

/**
 * TV 10-foot 排版地板：所有 Material3 type role 调用点静默继承这里的覆写，避免 200+ 处
 * MaterialTheme.typography.<role> 各自写 fontSize 字面量。新增 role 必须扩这里的 token，不允许
 * 在调用点硬编码 sp。
 *
 * 三档地板：
 * - MainTitleSp = 34（主标题 hero）
 * - SubtitleFloorSp = 22（海报卡 / section / titleSmall 副标题）
 * - HelperFloorSp = 18（body / label / 助记文本）
 */
object TvTypographyTokens {
    const val MainTitleSp: Int = 34
    const val SubtitleFloorSp: Int = 22
    const val HelperFloorSp: Int = 18
    const val TightHelperSp: Int = 18
}

val TvTypography: Typography = Typography(
    headlineLarge = TextStyle(
        fontSize = TvTypographyTokens.MainTitleSp.sp,
        lineHeight = 42.sp,
        letterSpacing = 0.sp,
        fontWeight = FontWeight.SemiBold,
    ),
    headlineMedium = TextStyle(
        fontSize = 30.sp,
        lineHeight = 38.sp,
        letterSpacing = 0.sp,
        fontWeight = FontWeight.SemiBold,
    ),
    headlineSmall = TextStyle(
        fontSize = 28.sp,
        lineHeight = 36.sp,
        letterSpacing = 0.sp,
        fontWeight = FontWeight.SemiBold,
    ),
    titleLarge = TextStyle(
        fontSize = 26.sp,
        lineHeight = 32.sp,
        letterSpacing = 0.sp,
        fontWeight = FontWeight.SemiBold,
    ),
    titleMedium = TextStyle(
        fontSize = 24.sp,
        lineHeight = 30.sp,
        letterSpacing = 0.15.sp,
        fontWeight = FontWeight.Medium,
    ),
    titleSmall = TextStyle(
        fontSize = TvTypographyTokens.SubtitleFloorSp.sp,
        lineHeight = 28.sp,
        letterSpacing = 0.1.sp,
        fontWeight = FontWeight.Medium,
    ),
    bodyLarge = TextStyle(
        fontSize = 20.sp,
        lineHeight = 28.sp,
        letterSpacing = 0.5.sp,
        fontWeight = FontWeight.Normal,
    ),
    bodyMedium = TextStyle(
        fontSize = TvTypographyTokens.HelperFloorSp.sp,
        lineHeight = 24.sp,
        letterSpacing = 0.25.sp,
        fontWeight = FontWeight.Normal,
    ),
    bodySmall = TextStyle(
        fontSize = 18.sp,
        lineHeight = 22.sp,
        letterSpacing = 0.4.sp,
        fontWeight = FontWeight.Normal,
    ),
    labelLarge = TextStyle(
        fontSize = 18.sp,
        lineHeight = 22.sp,
        letterSpacing = 0.1.sp,
        fontWeight = FontWeight.Medium,
    ),
    labelMedium = TextStyle(
        fontSize = 18.sp,
        lineHeight = 22.sp,
        letterSpacing = 0.5.sp,
        fontWeight = FontWeight.Medium,
    ),
    labelSmall = TextStyle(
        fontSize = TvTypographyTokens.TightHelperSp.sp,
        lineHeight = 22.sp,
        letterSpacing = 0.5.sp,
        fontWeight = FontWeight.Medium,
    ),
)
