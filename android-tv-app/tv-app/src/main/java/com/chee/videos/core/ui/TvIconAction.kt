package com.chee.videos.core.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.Icon
import androidx.compose.material3.Surface
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Shape
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp

@Composable
fun TvIconActionButton(
    icon: ImageVector,
    contentDescription: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
    iconModifier: Modifier = Modifier,
    size: Dp = 42.dp,
    iconSize: Dp = 24.dp,
    shape: Shape = CircleShape,
    containerColor: Color = AppChrome.Surface.copy(alpha = 0.62f),
    contentColor: Color = AppChrome.TextPrimary,
    focusedScale: Float = 1.1f,
) {
    Surface(
        color = containerColor,
        shape = shape,
        modifier = modifier
            .size(size)
            .tvFocusableGlow(shape = shape, focusedScale = focusedScale)
            .clickable(onClick = onClick),
    ) {
        Box(contentAlignment = Alignment.Center) {
            Icon(
                imageVector = icon,
                contentDescription = contentDescription,
                tint = contentColor,
                modifier = iconModifier.size(iconSize),
            )
        }
    }
}
