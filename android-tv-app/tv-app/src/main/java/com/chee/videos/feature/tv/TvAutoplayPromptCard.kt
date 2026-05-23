package com.chee.videos.feature.tv

import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.TweenSpec
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.LaunchedTvInitialFocus
import com.chee.videos.core.ui.TvMotionTokens
import com.chee.videos.core.ui.tryRequestFocus
import com.chee.videos.core.ui.tvFocusableGlow

@Composable
fun TvAutoplayPromptCard(
    nextEpisodeRef: TvNextEpisodeRef,
    visible: Boolean,
    remainingSeconds: Int,
    onPlayNow: () -> Unit,
    onCancel: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val playNowFocusRequester = remember { FocusRequester() }

    LaunchedTvInitialFocus(visible, nextEpisodeRef.seasonNumber, nextEpisodeRef.episodeNumber) {
        if (visible) {
            playNowFocusRequester.tryRequestFocus()
        }
    }

    AnimatedVisibility(
        visible = visible,
        enter = fadeIn(animationSpec = autoplayPromptFadeSpec()),
        exit = fadeOut(animationSpec = autoplayPromptFadeSpec()),
        modifier = modifier,
    ) {
        Surface(
            color = AppChrome.SurfaceMuted.copy(alpha = 0.94f),
            contentColor = AppChrome.TextPrimary,
            shape = AppChrome.SurfaceShape,
            modifier = Modifier.widthIn(min = 320.dp, max = 380.dp),
        ) {
            Column(
                modifier = Modifier.padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(12.dp),
            ) {
                Text(
                    text = "即将播放 · 第 ${nextEpisodeRef.episodeNumber} 集 ${nextEpisodeRef.title}",
                    color = AppChrome.TextPrimary,
                    style = MaterialTheme.typography.bodyLarge,
                    fontWeight = FontWeight.SemiBold,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                    TvAutoplayActionButton(
                        text = "立即播放 (${remainingSeconds.coerceAtLeast(1)})",
                        onClick = onPlayNow,
                        modifier = Modifier.focusRequester(playNowFocusRequester),
                    )
                    TvAutoplayActionButton(
                        text = "取消本次",
                        onClick = onCancel,
                    )
                }
            }
        }
    }
}

@Composable
private fun TvAutoplayActionButton(
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
            modifier = Modifier.padding(horizontal = 14.dp, vertical = 10.dp),
        )
    }
}

private fun autoplayPromptFadeSpec(): TweenSpec<Float> =
    TweenSpec(
        durationMillis = TvMotionTokens.DurationStandardMs,
        easing = TvMotionTokens.EasingStandard,
    )
