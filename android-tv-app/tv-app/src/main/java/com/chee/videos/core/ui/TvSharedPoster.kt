package com.chee.videos.core.ui

import androidx.compose.animation.AnimatedContentScope
import androidx.compose.animation.ExperimentalSharedTransitionApi
import androidx.compose.animation.SharedTransitionScope
import androidx.compose.runtime.Composable
import androidx.compose.runtime.ProvidableCompositionLocal
import androidx.compose.runtime.compositionLocalOf
import androidx.compose.ui.Modifier

@OptIn(ExperimentalSharedTransitionApi::class)
val LocalTvSharedTransitionScope: ProvidableCompositionLocal<SharedTransitionScope?> =
    compositionLocalOf { null }

val LocalTvAnimatedContentScope: ProvidableCompositionLocal<AnimatedContentScope?> =
    compositionLocalOf { null }

private const val TvSeriesPosterSharedKeyPrefix = "tv-series-poster-"

@OptIn(ExperimentalSharedTransitionApi::class)
@Composable
fun Modifier.tvSharedSeriesPoster(seriesId: String): Modifier {
    val sharedScope = LocalTvSharedTransitionScope.current ?: return this
    val animatedScope = LocalTvAnimatedContentScope.current ?: return this
    return with(sharedScope) {
        this@tvSharedSeriesPoster.sharedElement(
            state = rememberSharedContentState(key = "$TvSeriesPosterSharedKeyPrefix$seriesId"),
            animatedVisibilityScope = animatedScope,
        )
    }
}
