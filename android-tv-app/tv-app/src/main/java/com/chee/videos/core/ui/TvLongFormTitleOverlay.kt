package com.chee.videos.core.ui

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Shadow
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

internal object TvLongFormTitleOverlayTokens {
    const val LeftPaddingDp = 24
    const val TopPaddingDp = 24
    const val PrimaryFontSizeSp = 22
    const val SecondaryFontSizeSp = 16
    const val ShadowOffsetX = 0
    const val ShadowOffsetY = 2
    const val ShadowBlurRadius = 4
    const val SecondaryAlpha = 0.72f

    val PrimaryColor = Color.White
    val SecondaryColor = Color.White.copy(alpha = SecondaryAlpha)
    val ShadowColor = Color.Black.copy(alpha = 0.8f)
}

@Composable
internal fun TvLongFormTitleOverlay(
    data: TvLongFormTitleOverlayData,
    visible: Boolean,
    showStatusBarPadding: Boolean,
    modifier: Modifier = Modifier,
) {
    val topPaddingModifier = if (showStatusBarPadding) {
        Modifier.statusBarsPadding()
    } else {
        Modifier
    }
    AnimatedVisibility(
        visible = visible,
        enter = fadeIn(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
        exit = fadeOut(tween(TvMotionTokens.DurationStandardMs, easing = TvMotionTokens.EasingStandard)),
        modifier = modifier.then(topPaddingModifier),
    ) {
        Column(
            modifier = Modifier.padding(
                start = TvLongFormTitleOverlayTokens.LeftPaddingDp.dp,
                top = TvLongFormTitleOverlayTokens.TopPaddingDp.dp,
            ),
        ) {
            Text(
                text = data.primary,
                color = TvLongFormTitleOverlayTokens.PrimaryColor,
                style = MaterialTheme.typography.titleLarge.copy(
                    fontSize = TvLongFormTitleOverlayTokens.PrimaryFontSizeSp.sp,
                    fontWeight = FontWeight.SemiBold,
                    shadow = titleOverlayShadow(),
                ),
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            data.secondary?.let { secondary ->
                Text(
                    text = secondary,
                    color = TvLongFormTitleOverlayTokens.SecondaryColor,
                    style = MaterialTheme.typography.bodyMedium.copy(
                        fontSize = TvLongFormTitleOverlayTokens.SecondaryFontSizeSp.sp,
                        shadow = titleOverlayShadow(),
                    ),
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
            }
        }
    }
}

private fun titleOverlayShadow(): Shadow {
    return Shadow(
        color = TvLongFormTitleOverlayTokens.ShadowColor,
        offset = Offset(
            TvLongFormTitleOverlayTokens.ShadowOffsetX.toFloat(),
            TvLongFormTitleOverlayTokens.ShadowOffsetY.toFloat(),
        ),
        blurRadius = TvLongFormTitleOverlayTokens.ShadowBlurRadius.toFloat(),
    )
}
