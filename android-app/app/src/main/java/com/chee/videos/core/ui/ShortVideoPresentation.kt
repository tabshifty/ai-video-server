package com.chee.videos.core.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
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
import com.chee.videos.core.model.VideoFitMode

internal val ShortVideoOverlayButtonBg = AppChrome.SurfaceStrong.copy(alpha = 0.42f)

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
