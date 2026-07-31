package com.chee.videos.core.ui

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class LongFormVideoPlayerSpecTest {

    private val playerSource: String by lazy {
        java.nio.file.Path.of(
            "src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt",
        ).toFile().readText()
    }

    private val longFormPlayerScreenSource: String by lazy {
        java.nio.file.Path.of(
            "src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt",
        ).toFile().readText()
    }

    private val seriesPlayerScreenSource: String by lazy {
        java.nio.file.Path.of(
            "src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt",
        ).toFile().readText()
    }

    @Test
    fun `player references PlayerFocusGuardInput aggregate`() {
        assertTrue(
            "LongFormVideoPlayer 必须使用 PlayerFocusGuardInput 聚合 overlay 可见性",
            playerSource.contains("PlayerFocusGuardInput("),
        )
    }

    @Test
    fun `player calls shouldReclaimRootFocus on transitions`() {
        assertTrue(
            "LongFormVideoPlayer 必须在 overlay 跃迁时调 shouldReclaimRootFocus",
            playerSource.contains("shouldReclaimRootFocus("),
        )
    }

    @Test
    fun `player exposes resumePromptSlot parameter`() {
        assertTrue(
            "LongFormVideoPlayer 必须暴露 resumePromptSlot 槽位让父级把续播卡内嵌进来",
            playerSource.contains("resumePromptSlot"),
        )
        assertTrue(
            "LongFormVideoPlayer 必须暴露 resumePromptVisible 让聚合 guard 感知续播卡可见性",
            playerSource.contains("resumePromptVisible"),
        )
        assertTrue(
            "LongFormVideoPlayer 必须暴露 backConfirmPromptVisible 让聚合 guard 感知返回二次确认可见性",
            playerSource.contains("backConfirmPromptVisible"),
        )
        assertTrue(
            "LongFormVideoPlayer 必须暴露 playerErrorVisible 让聚合 guard 感知错误浮层可见性",
            playerSource.contains("playerErrorVisible"),
        )
    }

    @Test
    fun `player root box has onFocusChanged null-focus guard`() {
        assertTrue(
            "LongFormVideoPlayer 根 Box 必须挂 onFocusChanged 做最后一道焦点真空兜底",
            playerSource.contains(".onFocusChanged"),
        )
        assertTrue(
            "onFocusChanged 兜底必须依赖 anyOverlayVisible 判断当前是否还有 overlay",
            playerSource.contains("anyOverlayVisible()"),
        )
    }

    @Test
    fun `tv long form screens use the shared chrome and lightweight resume notice`() {
        listOf(longFormPlayerScreenSource, seriesPlayerScreenSource).forEach { source ->
            assertTrue(source.contains("TvLongFormPlaybackChrome("))
            assertTrue(source.contains("resumeNoticeText ="))
            assertTrue(source.contains("blockingUiVisible ="))
            assertFalse(source.contains("TvResumePromptCard("))
            assertFalse(source.contains("TvSeriesCorePlaybackOverlay("))
            assertFalse(source.contains("showBackConfirmPrompt"))
        }
    }

    @Test
    fun `player guard input includes all six overlay slots`() {
        val guardSource = java.nio.file.Path.of(
            "src/main/java/com/chee/videos/core/ui/LongFormPlayerFocusGuard.kt",
        ).toFile().readText()
        listOf(
            "controlsVisible",
            "subtitleSheetVisible",
            "audioTrackSheetVisible",
            "resumePromptVisible",
            "backConfirmPromptVisible",
            "playerErrorVisible",
        ).forEach { field ->
            assertTrue(
                "PlayerFocusGuardInput 必须包含 $field 字段",
                guardSource.contains(field),
            )
        }
        assertFalse(
            "controlsVisible 不应被算作 overlay（anyOverlayVisible 中不能出现 controlsVisible）",
            guardSource.contains("controlsVisible ||"),
        )
    }
}
