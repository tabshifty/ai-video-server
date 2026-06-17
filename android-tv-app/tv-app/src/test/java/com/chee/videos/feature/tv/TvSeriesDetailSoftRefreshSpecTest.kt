package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSeriesDetailSoftRefreshSpecTest {
    @Test
    fun `series detail keeps inline refresh state inside episode pane`() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt").readText()
        val paneSource = source.substringAfter("private fun TvSeriesEpisodePane(")
        val paneBody = paneSource.substringBefore("@Composable\nprivate fun TvSeriesSeasonSelector(")

        assertTrue("电视剧详情已有内容刷新时必须在右侧剧集面板内显示轻量 loading", paneBody.contains("if (refreshing)") && paneBody.contains("TvInlineLoadingState("))
        assertTrue("电视剧详情已有内容失败时必须在右侧剧集面板内显示轻量错误条", paneBody.contains("if (!errorMessage.isNullOrBlank())") && paneBody.contains("TvSeriesInlineError("))
        assertTrue("电视剧详情右侧剧集面板内状态必须位于列表之前，保证旧剧集仍然可见", paneBody.indexOf("TvSeriesInlineError(") < paneBody.indexOf("LazyColumn("))
        assertTrue("电视剧详情右侧剧集面板内 loading 必须位于列表之前，保证旧剧集仍然可见", paneBody.indexOf("TvInlineLoadingState(") < paneBody.indexOf("LazyColumn("))
    }

    @Test
    fun `series detail soft refresh does not steal focus back to play button`() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt").readText()
        val focusSource = source.substringAfter("LaunchedTvInitialFocus(")
            .substringBefore("\n\n    Box(")
        val focusEffectKeys = source.substringAfter("LaunchedTvInitialFocus(")
            .substringBefore(") {")

        assertTrue("电视剧详情首焦点必须使用一次性标记，软刷新完成后不得重新抢回播放按钮", source.contains("var initialFocusRequested"))
        assertTrue("电视剧详情首焦点 effect 不得把 refreshing 作为 key", !focusEffectKeys.contains("uiState.refreshing"))
        assertTrue("电视剧详情首焦点请求成功后必须置位", focusSource.contains("initialFocusRequested = true"))
    }
}
