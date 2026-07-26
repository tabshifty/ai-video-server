package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvShortFeedScreenSpecTest {
    @Test
    fun shortFeedScreenUsesRootKeyHandlingAndFixedFitPlayback() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/tv/TvShortFeedScreen.kt")
        assertTrue("TV 短视频页必须存在", sourcePath.exists())

        val source = sourcePath.readText()
        assertTrue("短视频页根节点必须接收遥控器焦点", source.contains(".focusable()"))
        assertTrue("短视频页根节点必须处理遥控器按键", source.contains(".onPreviewKeyEvent"))
        assertTrue("短视频页必须处理上键切上一条", source.contains("KEYCODE_DPAD_UP"))
        assertTrue("短视频页必须处理下键切下一条", source.contains("KEYCODE_DPAD_DOWN"))
        assertTrue("短视频页必须处理左键快退", source.contains("KEYCODE_DPAD_LEFT"))
        assertTrue("短视频页必须处理右键快进", source.contains("KEYCODE_DPAD_RIGHT"))
        assertTrue("短视频页必须处理确认键播放/暂停", source.contains("KEYCODE_DPAD_CENTER"))
        assertTrue("短视频页必须处理返回键退出", source.contains("KEYCODE_BACK"))
        assertTrue("短视频页首版必须固定 FIT，不提供铺满切换", source.contains("AspectRatioFrameLayout.RESIZE_MODE_FIT"))
        assertTrue("短视频页必须保留中央暂停/继续提示", source.contains("已暂停") && source.contains("继续播放"))
        assertFalse("短视频页不应再引入旧短视频右侧动作 rail", source.contains("ShortPlaybackModeToggleButton"))
        assertFalse("短视频页不应再复用旧短视频动作按钮组件", source.contains("ShortVideoOverlayActionButton"))
    }

    @Test
    fun shortFeedScreenBackUsesSharedDoublePressConfirm() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/tv/TvShortFeedScreen.kt")
        assertTrue("TV 短视频页必须存在", sourcePath.exists())

        val source = sourcePath.readText()
        assertTrue(
            "短视频页 BACK 必须复用长视频/电视剧双按状态机 resolveTvPlayerBackAction",
            source.contains("resolveTvPlayerBackAction"),
        )
        assertTrue(
            "短视频页必须有统一的双按返回处理函数 handlePlaybackBack",
            source.contains("fun handlePlaybackBack()"),
        )
        assertTrue(
            "短视频页顶层 BackHandler 必须接入双按状态机而非直接 onBack",
            source.contains("BackHandler(onBack = ::handlePlaybackBack)"),
        )
        assertTrue(
            "短视频页 preview 的 BACK 分支必须接入双按状态机而非直接 onBack",
            source.contains("AndroidKeyEvent.KEYCODE_BACK") && source.contains("handlePlaybackBack()"),
        )
        assertTrue(
            "短视频页必须复用既有返回确认提示组件 TvPlayerBackConfirmPrompt",
            source.contains("TvPlayerBackConfirmPrompt"),
        )
        assertFalse(
            "短视频页不应新造并行双按实现（禁止出现独立 backPressTime/lastBackPress 旧式时间戳）",
            source.contains("backPressTime") || source.contains("lastBackPress"),
        )
    }

    @Test
    fun shortFeedUsesSharedPrimarySourceAndDiagnostics() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvShortFeedScreen.kt").readText()

        listOf(
            "buildTvShortPlaybackSourceUrl(normalizedBase, currentVideoId)",
            "TvShortPlaybackDiagnostics()",
            "diagnostics.attach(sharedPlayer)",
            "diagnostics.sourcePrepared(attempt)",
            "diagnostics.detach(sharedPlayer)",
            "createTvShortPlaybackAttempt(",
            "entry = TvShortPlaybackEntry.Local",
        ).forEach { line ->
            assertTrue("TV 本地短视频页必须包含 $line", source.contains(line))
        }
    }
}
