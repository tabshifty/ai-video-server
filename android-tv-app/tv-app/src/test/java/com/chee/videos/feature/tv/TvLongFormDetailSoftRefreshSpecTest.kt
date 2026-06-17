package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvLongFormDetailSoftRefreshSpecTest {
    @Test
    fun `long form detail keeps inline refresh state inside glass panel`() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt").readText()
        val panelBody = source.substringAfter("TvDetailGlassPanel(")
            .substringBefore("}\n            }\n        }")

        assertTrue("长视频详情已有内容刷新时必须在底部玻璃面板内显示轻量 loading", panelBody.contains("if (uiState.refreshing)") && panelBody.contains("TvInlineLoadingState("))
        assertTrue("长视频详情已有内容失败时必须在底部玻璃面板内显示轻量错误条", panelBody.contains("if (!uiState.errorMessage.isNullOrBlank())") && panelBody.contains("TvLongFormInlineError("))
        assertTrue("长视频详情轻量状态必须位于操作按钮之前，保证内容区域保持完整", panelBody.indexOf("TvLongFormInlineError(") < panelBody.indexOf("Row(horizontalArrangement = Arrangement.spacedBy(12.dp))"))
    }

    @Test
    fun `long form detail soft refresh does not steal focus back to play button`() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt").readText()
        val focusSource = source.substringAfter("LaunchedTvInitialFocus(")
            .substringBefore("\n\n            Box(")
        val focusEffectKeys = source.substringAfter("LaunchedTvInitialFocus(")
            .substringBefore(") {")

        assertTrue("长视频详情首焦点必须使用一次性标记，软刷新完成后不得重新抢回播放按钮", source.contains("var initialFocusRequested"))
        assertTrue("长视频详情首焦点 effect 不得把 refreshing 作为 key", !focusEffectKeys.contains("uiState.refreshing"))
        assertTrue("长视频详情首焦点请求成功后必须置位", focusSource.contains("initialFocusRequested = true"))
    }
}
