package com.chee.videos.core.ui

import android.os.Build
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.blur
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color

@Composable
fun TvDetailGlassPanel(
    modifier: Modifier = Modifier,
    content: @Composable () -> Unit,
) {
    val supportsBlur = Build.VERSION.SDK_INT >= Build.VERSION_CODES.S
    val scrimColor = if (supportsBlur) {
        TvDetailPanelTokens.ScrimColorBlurred
    } else {
        TvDetailPanelTokens.ScrimColorFallback
    }
    Box(modifier = modifier.fillMaxWidth()) {
        Box(
            modifier = Modifier
                .align(Alignment.TopCenter)
                .fillMaxWidth()
                .height(TvDetailPanelTokens.UpperGradientHeightDp)
                .background(
                    Brush.verticalGradient(
                        colors = listOf(Color.Transparent, scrimColor),
                    ),
                ),
        )
        val panelBaseModifier = if (supportsBlur) {
            Modifier.blur(TvDetailPanelTokens.BlurRadiusDp)
        } else {
            Modifier
        }
        Surface(
            color = scrimColor,
            shape = RoundedCornerShape(
                topStart = TvDetailPanelTokens.TopCornerRadiusDp,
                topEnd = TvDetailPanelTokens.TopCornerRadiusDp,
            ),
            modifier = Modifier
                .fillMaxWidth()
                .then(panelBaseModifier),
        ) {
            content()
        }
    }
}
