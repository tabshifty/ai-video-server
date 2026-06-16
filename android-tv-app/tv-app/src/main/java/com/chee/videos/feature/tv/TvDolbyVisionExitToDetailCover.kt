package com.chee.videos.feature.tv

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color

internal const val TvDolbyVisionExitToDetailCoverDelayMillis = 80L

@Composable
internal fun TvDolbyVisionExitToDetailCover(
    modifier: Modifier = Modifier,
) {
    Box(
        modifier = modifier
            .fillMaxSize()
            .background(Color.Black),
    )
}
