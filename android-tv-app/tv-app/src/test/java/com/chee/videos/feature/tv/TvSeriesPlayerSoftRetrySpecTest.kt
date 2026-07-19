package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSeriesPlayerSoftRetrySpecTest {

    private val source = Path.of(
        "src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt",
    ).readText()

    @Test
    fun `剧集播放器接入完整软重试状态与取消令牌`() {
        assertTrue(source.contains("var activeSoftRetryAttemptKey by remember(uiState.currentVideoId)"))
        assertTrue(source.contains("var ignoredRetryAttemptKey by remember(uiState.currentVideoId)"))
        assertTrue(source.contains("var cancelPrepareRequestKey by remember(uiState.currentVideoId)"))
        assertTrue(source.contains("var softRetryUiState by remember(uiState.currentVideoId)"))
        assertTrue(source.contains("softRetryUiState = TvLongFormSoftRetryUiState.Preparing(nextRetryKey)"))
        assertTrue(source.contains("ignoredRetryAttemptKey = preparingState.retryKey"))
        assertTrue(source.contains("cancelPrepareRequestKey += 1"))
        assertTrue(source.contains("cancelPrepareRequestKey = cancelPrepareRequestKey"))
        assertTrue(source.contains("cancelPrepareRetryKey = ignoredRetryAttemptKey"))
        assertTrue(source.contains("shouldIgnoreTvLongFormRetryError(ignoredRetryAttemptKey, eventRetryKey)"))
        assertTrue(source.contains("onPlayingChanged = { playing, eventRetryKey ->"))
        assertTrue(source.contains("eventRetryKey == activeRetryKey"))
        assertTrue(source.contains("onError = { message, eventRetryKey ->"))
        assertTrue(source.contains("if (eventRetryKey == routeRetryNonce)"))
    }

    @Test
    fun `软重试 BACK 与焦点守卫接入聚合状态`() {
        assertTrue(source.contains("when (resolveSeriesSoftRetryBackAction(softRetryUiState))"))
        assertTrue(source.contains("softRetryUiState is TvLongFormSoftRetryUiState.Failed"))
        assertTrue(source.contains("playerErrorVisible = overlayPlayerErrorVisible"))
    }

    @Test
    fun `软失败主动作 requester 位于 focusable 节点之前`() {
        val buttonSource = source
            .substringAfter("private fun TvSeriesPlayerSoftRetryActionButton(")
        val requesterIndex = buttonSource.indexOf("modifier = modifier")
        val focusableIndex = buttonSource.indexOf(".tvFocusableScaleOnly")

        assertTrue("软重试动作按钮必须存在", buttonSource.isNotBlank())
        assertTrue("动作按钮必须接收外部 modifier", requesterIndex >= 0)
        assertTrue("动作按钮必须保留 TV 焦点节点", focusableIndex >= 0)
        assertTrue(source.contains("modifier = Modifier.focusRequester(retryFocusRequester)"))
        assertTrue(
            "外部 focusRequester 必须排在 focusable 节点之前",
            requesterIndex < focusableIndex,
        )
    }
}
