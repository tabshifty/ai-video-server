package com.chee.videos.feature.tv

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.LaunchedTvInitialFocus
import com.chee.videos.core.ui.tryRequestFocus

@Composable
internal fun TvLongFormCompletionOverlay(
    headline: String,
    title: String,
    replayLabel: String,
    onBackToDetail: () -> Unit,
    onReplay: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val backFocusRequester = remember { FocusRequester() }
    LaunchedTvInitialFocus(headline, title) {
        backFocusRequester.tryRequestFocus()
    }

    Box(
        modifier = modifier
            .fillMaxSize()
            .background(Color.Black),
        contentAlignment = Alignment.Center,
    ) {
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(16.dp),
            modifier = Modifier.padding(horizontal = 64.dp, vertical = 40.dp),
        ) {
            Text(
                text = headline,
                color = AppChrome.TextPrimary,
                style = MaterialTheme.typography.headlineMedium,
                fontWeight = FontWeight.Bold,
            )
            Text(
                text = title,
                color = AppChrome.TextSecondary,
                style = MaterialTheme.typography.titleMedium,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
            )
            Row(
                horizontalArrangement = Arrangement.spacedBy(12.dp),
                modifier = Modifier.padding(top = 12.dp),
            ) {
                TvLongFormCompletionAction(
                    label = "返回详情",
                    onClick = onBackToDetail,
                    modifier = Modifier.focusRequester(backFocusRequester),
                )
                TvLongFormCompletionAction(
                    label = replayLabel,
                    onClick = onReplay,
                )
            }
        }
    }
}

@Composable
private fun TvLongFormCompletionAction(
    label: String,
    onClick: () -> Unit,
    modifier: Modifier = Modifier,
) {
    var focused by remember { mutableStateOf(false) }
    Surface(
        color = if (focused) AppChrome.Accent else AppChrome.SurfaceStrong,
        contentColor = if (focused) AppChrome.Canvas else AppChrome.TextPrimary,
        shape = AppChrome.ChipShape,
        modifier = modifier
            .onFocusChanged { focused = it.isFocused || it.hasFocus }
            .clickable(onClick = onClick),
    ) {
        Text(
            text = label,
            style = MaterialTheme.typography.labelLarge,
            fontWeight = FontWeight.Bold,
            modifier = Modifier.padding(horizontal = 20.dp, vertical = 12.dp),
        )
    }
}
