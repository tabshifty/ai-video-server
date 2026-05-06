package com.chee.videos.core.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.PlayArrow
import androidx.compose.material3.Icon
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.unit.dp
import androidx.media3.ui.AspectRatioFrameLayout
import com.chee.videos.core.model.ShortPlaybackMode
import com.chee.videos.core.model.VideoFitMode

internal val ShortVideoOverlayButtonBg = AppChrome.SurfaceStrong.copy(alpha = 0.42f)
internal val ShortNonHomeProgressBarBottomSpacing = 12.dp

internal fun shortPosterContentScale(fitMode: VideoFitMode): ContentScale {
    return when (fitMode) {
        VideoFitMode.FILL -> ContentScale.Crop
        VideoFitMode.FIT -> ContentScale.Fit
    }
}

internal fun shortVideoResizeMode(fitMode: VideoFitMode): Int {
    return when (fitMode) {
        VideoFitMode.FILL -> AspectRatioFrameLayout.RESIZE_MODE_ZOOM
        VideoFitMode.FIT -> AspectRatioFrameLayout.RESIZE_MODE_FIT
    }
}

@Composable
internal fun ShortVideoOverlayActionButton(
    icon: ImageVector,
    active: Boolean,
    enabled: Boolean,
    onClick: () -> Unit,
    contentDescription: String,
    modifier: Modifier = Modifier,
) {
    Box(
        modifier = modifier
            .size(44.dp)
            .clip(CircleShape)
            .background(ShortVideoOverlayButtonBg)
            .clickable(
                enabled = enabled,
                interactionSource = remember { MutableInteractionSource() },
                indication = null,
                onClick = onClick,
            ),
        contentAlignment = Alignment.Center,
    ) {
        Icon(
            imageVector = icon,
            contentDescription = contentDescription,
            tint = when {
                !enabled -> Color.White.copy(alpha = 0.45f)
                active -> AppChrome.AccentStrong
                else -> Color.White
            },
        )
    }
}

@Composable
internal fun ShortPlaybackModeToggleButton(
    playbackMode: ShortPlaybackMode,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    enabled: Boolean = true,
) {
    ShortVideoOverlayActionButton(
        icon = Icons.Filled.PlayArrow,
        active = playbackMode == ShortPlaybackMode.AUTO_NEXT,
        enabled = enabled,
        onClick = onClick,
        contentDescription = shortPlaybackModeContentDescription(playbackMode),
        modifier = modifier,
    )
}

internal fun shortPlaybackModeContentDescription(playbackMode: ShortPlaybackMode): String {
    return if (playbackMode == ShortPlaybackMode.LOOP_ONE) {
        "播放模式：循环单视频"
    } else {
        "播放模式：自动播放下一个"
    }
}

internal fun shortPlaybackModeLabel(playbackMode: ShortPlaybackMode): String {
    return if (playbackMode == ShortPlaybackMode.LOOP_ONE) {
        "循环单视频"
    } else {
        "自动播放下一个"
    }
}

internal fun shouldShowShortOverlayProgressBar(currentVideoId: String?): Boolean {
    return !currentVideoId.isNullOrBlank()
}

internal fun Modifier.shortNonHomeProgressBarPadding(): Modifier {
    return this
        .navigationBarsPadding()
        .padding(bottom = ShortNonHomeProgressBarBottomSpacing)
}
