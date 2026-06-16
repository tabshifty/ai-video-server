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
    fun `long form player screen embeds resume prompt via slot only`() {
        // TvResumePromptCard 只能在 resumePromptSlot = { ... } 块里出现，不允许作为 LongFormVideoPlayer 兄弟节点平铺
        assertTrue(
            "TvLongFormPlayerScreen 必须通过 resumePromptSlot 把续播卡内嵌到 LongFormVideoPlayer",
            longFormPlayerScreenSource.contains("resumePromptSlot ="),
        )
        // 简化版的 sibling 检测：TvResumePromptCard 第一处出现之前必须有 "resumePromptSlot ="
        val cardIndex = longFormPlayerScreenSource.indexOf("TvResumePromptCard(")
        val slotIndex = longFormPlayerScreenSource.indexOf("resumePromptSlot =")
        assertTrue(
            "TvLongFormPlayerScreen 的 TvResumePromptCard 必须位于 resumePromptSlot 块之后（即作为 slot 内容）",
            cardIndex == -1 || (slotIndex in 0 until cardIndex),
        )
    }

    @Test
    fun `series player screen embeds resume prompt via slot only`() {
        assertTrue(
            "TvSeriesPlayerScreen 必须通过 resumePromptSlot 把续播卡内嵌到 LongFormVideoPlayer",
            seriesPlayerScreenSource.contains("resumePromptSlot ="),
        )
        val cardIndex = seriesPlayerScreenSource.indexOf("TvResumePromptCard(")
        val slotIndex = seriesPlayerScreenSource.indexOf("resumePromptSlot =")
        assertTrue(
            "TvSeriesPlayerScreen 的 TvResumePromptCard 必须位于 resumePromptSlot 块之后（即作为 slot 内容）",
            cardIndex == -1 || (slotIndex in 0 until cardIndex),
        )
    }

    @Test
    fun `screens propagate overlay visibility into player`() {
        assertTrue(
            "TvLongFormPlayerScreen 必须把 backConfirmPromptVisible 透传给 LongFormVideoPlayer",
            longFormPlayerScreenSource.contains("backConfirmPromptVisible = showBackConfirmPrompt"),
        )
        assertTrue(
            "TvLongFormPlayerScreen 必须把 playerErrorVisible 透传给 LongFormVideoPlayer",
            longFormPlayerScreenSource.contains("playerErrorVisible ="),
        )
        assertTrue(
            "TvSeriesPlayerScreen 必须把 backConfirmPromptVisible 透传给 LongFormVideoPlayer",
            seriesPlayerScreenSource.contains("backConfirmPromptVisible = showBackConfirmPrompt"),
        )
        assertTrue(
            "TvSeriesPlayerScreen 必须把 playerErrorVisible 透传给 LongFormVideoPlayer",
            seriesPlayerScreenSource.contains("playerErrorVisible ="),
        )
    }

    @Test
    fun `screens no longer place resume prompt as sibling fallback path`() {
        // 兜底校验：单片和电视剧 screen 都有 LibVLC / Media3 两个分支，但都必须在 slot 中。
        val longFormCount = Regex("TvResumePromptCard\\(").findAll(longFormPlayerScreenSource).count()
        val seriesCount = Regex("TvResumePromptCard\\(").findAll(seriesPlayerScreenSource).count()
        assertTrue(
            "TvLongFormPlayerScreen 中 TvResumePromptCard 调用应仅出现在 LibVLC / Media3 两个 resumePromptSlot 中，当前 $longFormCount",
            longFormCount <= 2,
        )
        assertTrue(
            "TvSeriesPlayerScreen 中 TvResumePromptCard 调用应仅出现在 LibVLC / Media3 两个 resumePromptSlot 中，当前 $seriesCount",
            seriesCount <= 2,
        )
        assertTrue(
            "TvLongFormPlayerScreen 的 Media3 分支也必须通过 resumePromptSlot 内嵌续播卡",
            longFormPlayerScreenSource.substringAfter("TvSeriesCorePlaybackOverlay(").contains("resumePromptSlot ="),
        )
        assertTrue(
            "TvSeriesPlayerScreen 的 Media3 分支也必须通过 resumePromptSlot 内嵌续播卡",
            seriesPlayerScreenSource.substringAfter("TvSeriesCorePlaybackOverlay(").contains("resumePromptSlot ="),
        )
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
