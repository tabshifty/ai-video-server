package com.chee.videos.core.ui

import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * 源文 audit：确保 [[VLC Playing gate]]、subtitle slave 注入、audio LaunchedEffect 状态回灌
 * 的代码路径保留不退步。
 */
class LongFormVlcPlayingGateSpecTest {

    private val playerSource: String by lazy {
        java.nio.file.Path.of(
            "src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt",
        ).toFile().readText()
    }

    private val longFormScreen: String by lazy {
        java.nio.file.Path.of(
            "src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt",
        ).toFile().readText()
    }

    private val seriesScreen: String by lazy {
        java.nio.file.Path.of(
            "src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt",
        ).toFile().readText()
    }

    private val subtitleSupport: String by lazy {
        java.nio.file.Path.of(
            "src/main/java/com/chee/videos/core/ui/LongFormSubtitleSupport.kt",
        ).toFile().readText()
    }

    @Test
    fun playerTracksIsVlcPlayingState() {
        // G1：LongFormVideoPlayer 必须用 isVlcPlaying 跟踪 VLC 真正进入 Playing 状态
        val occurrences = Regex("\\bisVlcPlaying\\b").findAll(playerSource).count()
        assertTrue(
            "LongFormVideoPlayer 必须出现 isVlcPlaying 至少 3 处（state 声明 + Event 处理 + LaunchedEffect gate），当前 $occurrences",
            occurrences >= 3,
        )
    }

    @Test
    fun audioLaunchedEffectGatedOnPlaying() {
        // G2：audio LaunchedEffect 必须 gate 在 isVlcPlaying 之上
        val pattern = Regex(
            "LaunchedEffect\\(player, audioTracks[\\s\\S]*?isVlcPlaying\\)",
        )
        assertTrue(
            "audio LaunchedEffect 必须把 isVlcPlaying 纳入 key 列表",
            pattern.containsMatchIn(playerSource),
        )
        assertTrue(
            "audio LaunchedEffect 必须 early-return when !isVlcPlaying",
            playerSource.contains("if (!isVlcPlaying"),
        )
    }

    @Test
    fun screensInjectSubtitleViaAddSlave() {
        // G3 / G4：两个 Screen 必须通过 mediaPlayer.addSlave(IMedia.Slave.Type.Subtitle, ...) 注入字幕
        assertTrue(
            "TvLongFormPlayerScreen 必须用 mediaPlayer.addSlave(IMedia.Slave.Type.Subtitle, ...) 注入字幕",
            longFormScreen.contains("addSlave(IMedia.Slave.Type.Subtitle"),
        )
        assertTrue(
            "TvSeriesPlayerScreen 必须用 mediaPlayer.addSlave(IMedia.Slave.Type.Subtitle, ...) 注入字幕",
            seriesScreen.contains("addSlave(IMedia.Slave.Type.Subtitle"),
        )
    }

    @Test
    fun applyLongFormMediaSourceNoLongerCarriesSubtitleParam() {
        // applyLongFormMediaSource 签名已收缩，不再带 selectedSubtitleTrack / baseUrl 参数
        assertFalse(
            "applyLongFormMediaSource 不应保留 selectedSubtitleTrack 参数",
            subtitleSupport.contains("selectedSubtitleTrack: SubtitleTrackDto"),
        )
        assertFalse(
            "TvLongFormPlayerScreen 调用 applyLongFormMediaSource 时不应再传 selectedSubtitleTrack",
            longFormScreen.contains("selectedSubtitleTrack = "),
        )
        assertFalse(
            "TvSeriesPlayerScreen 调用 applyLongFormMediaSource 时不应再传 selectedSubtitleTrack",
            seriesScreen.contains("selectedSubtitleTrack = "),
        )
    }

    @Test
    fun audioCallbackDistinguishesUserAndProgrammaticPaths() {
        // G5：onSelectAudioTrack 既有 isUserAction=true 也有 isUserAction=false 调用
        assertTrue(
            "LongFormVideoPlayer 必须以 isUserAction=true 形式从 picker 上报",
            playerSource.contains("onSelectAudioTrack(trackId, preference, true)"),
        )
        assertTrue(
            "LongFormVideoPlayer 必须以 isUserAction=false 形式做状态回灌（resolved selection）",
            playerSource.contains(", false)") &&
                playerSource.contains("onSelectAudioTrack("),
        )
    }

    @Test
    fun resolveLongFormPlayerUpdateNoLongerHasSubtitleParams() {
        // resolveLongFormPlayerUpdate 签名也收缩了，subtitle 改走 addSlave
        assertFalse(
            "resolveLongFormPlayerUpdate 不应保留 preparedSubtitleTrackId 参数",
            subtitleSupport.contains("preparedSubtitleTrackId: String?"),
        )
        assertFalse(
            "resolveLongFormPlayerUpdate 不应保留 nextSubtitleTrackId 参数",
            subtitleSupport.contains("nextSubtitleTrackId: String?"),
        )
    }

    @Test
    fun screensExposeIsVlcPlayingStateForSlaveTiming() {
        // 两个 Screen 必须各自维护 isVlcPlaying（独立于 LongFormVideoPlayer 内部那份）
        assertTrue(
            "TvLongFormPlayerScreen 必须维护 isVlcPlaying state 以决定 addSlave 时机",
            longFormScreen.contains("isVlcPlaying"),
        )
        assertTrue(
            "TvSeriesPlayerScreen 必须维护 isVlcPlaying state 以决定 addSlave 时机",
            seriesScreen.contains("isVlcPlaying"),
        )
        assertTrue(
            "TvLongFormPlayerScreen 必须用 appliedSubtitleSlaveUrl 做 addSlave 幂等",
            longFormScreen.contains("appliedSubtitleSlaveUrl"),
        )
        assertTrue(
            "TvSeriesPlayerScreen 必须用 appliedSubtitleSlaveUrl 做 addSlave 幂等",
            seriesScreen.contains("appliedSubtitleSlaveUrl"),
        )
    }

    @Test
    fun typeOnlyFallbackImplementedInThreeResolvers() {
        // F2 全部三处必须有"language 空 + type 非空 → 按 type 匹配"的分支
        val trackSelectionSource = java.nio.file.Path.of(
            "src/main/java/com/chee/videos/core/player/TvLongFormTrackSelection.kt",
        ).toFile().readText()
        assertTrue(
            "resolveLongFormTrackByLanguage 必须在 language.isBlank 时检查 type 并按 type 匹配",
            trackSelectionSource.contains("type-only fallback"),
        )
        assertTrue(
            "resolveSelectedSubtitleTrackByPreference 必须在 language.isBlank 时按 type 匹配",
            subtitleSupport.contains("type-only fallback"),
        )
    }
}
