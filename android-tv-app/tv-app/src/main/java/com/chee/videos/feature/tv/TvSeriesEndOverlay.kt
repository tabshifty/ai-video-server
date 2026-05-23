package com.chee.videos.feature.tv

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.TweenSpec
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.LaunchedTvInitialFocus
import com.chee.videos.core.ui.TvMotionTokens
import com.chee.videos.core.ui.tryRequestFocus
import com.chee.videos.core.ui.tvFocusableGlow

@Composable
fun TvSeriesEndOverlay(
    kind: TvEndOverlayKind?,
    onPlayNext: () -> Unit,
    onBackToDetail: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val primaryFocusRequester = remember { FocusRequester() }

    LaunchedTvInitialFocus(kind) {
        if (kind != null) {
            primaryFocusRequester.tryRequestFocus()
        }
    }

    AnimatedVisibility(
        visible = kind != null,
        enter = fadeIn(animationSpec = autoplayEndOverlayFadeSpec()),
        exit = fadeOut(animationSpec = autoplayEndOverlayFadeSpec()),
        modifier = modifier,
    ) {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(Color.Black.copy(alpha = 0.62f)),
            contentAlignment = Alignment.Center,
        ) {
            Surface(
                color = AppChrome.Surface.copy(alpha = 0.96f),
                contentColor = AppChrome.TextPrimary,
                shape = AppChrome.SurfaceShape,
                modifier = Modifier.widthIn(min = 360.dp, max = 520.dp),
            ) {
                Column(
                    modifier = Modifier.padding(24.dp),
                    verticalArrangement = Arrangement.spacedBy(18.dp),
                    horizontalAlignment = Alignment.CenterHorizontally,
                ) {
                    Text(
                        text = if (kind == TvEndOverlayKind.CURRENT_FINISHED) "本集已播完" else "全剧已播完",
                        color = AppChrome.TextPrimary,
                        style = MaterialTheme.typography.headlineSmall,
                        fontWeight = FontWeight.Bold,
                    )
                    Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                        if (kind == TvEndOverlayKind.CURRENT_FINISHED) {
                            TvSeriesEndOverlayButton(
                                text = "播放下一集",
                                onClick = onPlayNext,
                                modifier = Modifier.focusRequester(primaryFocusRequester),
                            )
                            TvSeriesEndOverlayButton(
                                text = "返回详情",
                                onClick = onBackToDetail,
                            )
                        } else {
                            TvSeriesEndOverlayButton(
                                text = "返回详情",
                                onClick = onBackToDetail,
                                modifier = Modifier.focusRequester(primaryFocusRequester),
                            )
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun TvSeriesEndOverlayButton(
    text: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    Surface(
        color = AppChrome.SurfaceStrong,
        contentColor = AppChrome.TextPrimary,
        shape = AppChrome.ChipShape,
        modifier = modifier
            .tvFocusableGlow(shape = AppChrome.ChipShape, focusedScale = 1.04f)
            .clickable(onClick = onClick),
    ) {
        Text(
            text = text,
            color = AppChrome.TextPrimary,
            style = MaterialTheme.typography.labelLarge,
            fontWeight = FontWeight.Bold,
            modifier = Modifier.padding(horizontal = 18.dp, vertical = 12.dp),
        )
    }
}

private fun autoplayEndOverlayFadeSpec(): TweenSpec<Float> =
    TweenSpec(
        durationMillis = TvMotionTokens.DurationStandardMs,
        easing = TvMotionTokens.EasingStandard,
    )
