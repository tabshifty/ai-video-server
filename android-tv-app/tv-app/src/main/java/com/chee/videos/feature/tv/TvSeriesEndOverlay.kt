package com.chee.videos.feature.tv

import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier

@Composable
fun TvSeriesEndOverlay(
    kind: TvEndOverlayKind?,
    onPlayNext: () -> Unit,
    onReplayCurrent: () -> Unit,
    onBackToDetail: () -> Unit,
    modifier: Modifier = Modifier,
) {
    val currentKind = kind ?: return
    TvLongFormCompletionOverlay(
        headline = if (currentKind == TvEndOverlayKind.CURRENT_FINISHED) "本集已播完" else "全剧已播完",
        title = if (currentKind == TvEndOverlayKind.CURRENT_FINISHED) "可以继续观看下一集" else "感谢观看",
        replayLabel = if (currentKind == TvEndOverlayKind.CURRENT_FINISHED) "播放下一集" else "重播本集",
        onBackToDetail = onBackToDetail,
        onReplay = if (currentKind == TvEndOverlayKind.CURRENT_FINISHED) onPlayNext else onReplayCurrent,
        modifier = modifier,
    )
}
