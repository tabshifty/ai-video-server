@file:androidx.annotation.OptIn(markerClass = [androidx.media3.common.util.UnstableApi::class])

package com.chee.videos.feature.tv

import androidx.compose.runtime.Composable
import androidx.compose.runtime.SideEffect
import androidx.compose.runtime.remember
import androidx.media3.datasource.DefaultHttpDataSource

/** 更新鉴权头而不重建播放源，屏保取图引发正常令牌刷新时仍保留暂停状态和进度。 */
@Composable
internal fun rememberTvPlaybackDataSource(accessToken: String): DefaultHttpDataSource.Factory {
    val factory = remember { DefaultHttpDataSource.Factory().setAllowCrossProtocolRedirects(true) }
    SideEffect {
        factory.setDefaultRequestProperties(
            if (accessToken.isBlank()) emptyMap() else mapOf("Authorization" to "Bearer $accessToken"),
        )
    }
    return factory
}
