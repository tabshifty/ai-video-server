package com.chee.videos.core.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.ColumnScope
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material.icons.filled.Tv
import androidx.compose.material.icons.filled.Warning
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp

@Composable
fun TvPageLoadingState(
    message: String,
    modifier: Modifier = Modifier,
) {
    TvStateContainer(modifier = modifier) {
        TvStatePanel {
            CircularProgressIndicator(
                modifier = Modifier.size(32.dp),
                color = TvFocusGlowColor,
                strokeWidth = 2.5.dp,
            )
            Text(
                text = message,
                color = AppChrome.TextSecondary,
                style = MaterialTheme.typography.bodyMedium,
                textAlign = TextAlign.Center,
            )
        }
    }
}

@Composable
fun TvInlineLoadingState(
    message: String,
    modifier: Modifier = Modifier,
) {
    Row(
        modifier = modifier.padding(vertical = 8.dp),
        horizontalArrangement = Arrangement.Center,
        verticalAlignment = Alignment.CenterVertically,
    ) {
        CircularProgressIndicator(
            modifier = Modifier.size(18.dp),
            color = TvFocusGlowColor,
            strokeWidth = 2.dp,
        )
        Text(
            text = message,
            color = AppChrome.TextMuted,
            style = MaterialTheme.typography.bodySmall,
            modifier = Modifier.padding(start = 8.dp),
        )
    }
}

@Composable
fun TvEmptyState(
    title: String,
    message: String,
    actionLabel: String? = null,
    onAction: (() -> Unit)? = null,
    modifier: Modifier = Modifier,
) {
    TvStateContainer(modifier = modifier) {
        TvStatePanel {
            TvStateIcon(Icons.Filled.Tv)
            TvStateTexts(title = title, message = message)
            TvStateAction(actionLabel = actionLabel, onAction = onAction)
        }
    }
}

@Composable
fun TvErrorState(
    title: String = "加载失败",
    message: String,
    actionLabel: String = "重试",
    onAction: () -> Unit,
    modifier: Modifier = Modifier,
) {
    TvStateContainer(modifier = modifier) {
        TvStatePanel {
            TvStateIcon(Icons.Filled.Warning, tint = AppChrome.Error)
            TvStateTexts(title = title, message = message)
            TvStateAction(actionLabel = actionLabel, onAction = onAction)
        }
    }
}

@Composable
private fun TvStateContainer(
    modifier: Modifier,
    content: @Composable () -> Unit,
) {
    Box(
        modifier = modifier.fillMaxSize(),
        contentAlignment = Alignment.Center,
    ) {
        content()
    }
}

@Composable
private fun TvStatePanel(content: @Composable ColumnScope.() -> Unit) {
    Surface(
        color = Color(0xD610131A),
        shape = AppChrome.SurfaceShape,
    ) {
        Column(
            modifier = Modifier.padding(horizontal = 28.dp, vertical = 24.dp),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(10.dp),
            content = content,
        )
    }
}

@Composable
private fun TvStateIcon(
    imageVector: ImageVector,
    tint: Color = AppChrome.TextMuted,
) {
    Icon(
        imageVector = imageVector,
        contentDescription = null,
        tint = tint,
        modifier = Modifier.size(36.dp),
    )
}

@Composable
private fun TvStateTexts(title: String, message: String) {
    Text(
        text = title,
        color = AppChrome.TextPrimary,
        style = MaterialTheme.typography.titleMedium,
        fontWeight = FontWeight.SemiBold,
        textAlign = TextAlign.Center,
    )
    Text(
        text = message,
        color = AppChrome.TextSecondary,
        style = MaterialTheme.typography.bodyMedium,
        textAlign = TextAlign.Center,
    )
}

@Composable
private fun TvStateAction(actionLabel: String?, onAction: (() -> Unit)?) {
    if (actionLabel == null || onAction == null) {
        return
    }
    Surface(
        color = Color(0x2239D7E8),
        shape = AppChrome.SurfaceShape,
        modifier = Modifier
            .tvFocusableGlow(shape = AppChrome.SurfaceShape, focusedScale = 1.04f)
            .clickable(onClick = onAction),
    ) {
        Row(
            modifier = Modifier.padding(horizontal = 16.dp, vertical = 12.dp),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
            verticalAlignment = Alignment.CenterVertically,
        ) {
            Icon(Icons.Filled.Refresh, contentDescription = null, tint = Color.White, modifier = Modifier.size(20.dp))
            Text(text = actionLabel, color = Color.White, style = MaterialTheme.typography.bodyMedium)
        }
    }
}
