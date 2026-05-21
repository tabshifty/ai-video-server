package com.chee.videos.core.ui

import java.nio.file.Files
import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvTrackPickerGlassPanelTest {
    private val source = Path.of(
        "..",
        "tv-app",
        "src",
        "main",
        "java",
        "com",
        "chee",
        "videos",
        "core",
        "ui",
        "SubtitlePicker.kt",
    )
    private val audioTrackSource = Path.of(
        "..",
        "tv-app",
        "src",
        "main",
        "java",
        "com",
        "chee",
        "videos",
        "core",
        "ui",
        "LongFormAudioTrackSupport.kt",
    )

    @Test
    fun tvTrackPickerUsesNightGlassPanelInsteadOfPlainDarkDialog() {
        assertTrue("SubtitlePicker.kt 不存在", Files.exists(source))
        val text = source.readText()

        assertTrue(text.contains("NightGlassTrackPickerPanel"))
        assertTrue(text.contains("夜台玻璃"))
        assertTrue(text.contains("Brush.linearGradient"))
        assertTrue(text.contains("Color(0x6625D9F2)"))
        assertTrue(text.contains("Color(0x1AFFFFFF)"))
        assertTrue(text.contains("supportingText"))
        assertTrue("音轨/字幕行必须显式处理遥控确认键", text.contains("handleTrackPickerConfirmKey"))
        assertTrue("音轨/字幕行不能复用全局粉红焦点边框", !text.contains(".tvFocusableGlow("))
        val optionRowSource = text
            .substringAfter("private fun SubtitleOptionRow(")
            .substringBefore("@Composable\nprivate fun TrackPickerSelectionRail(")
        assertFalse("音轨/字幕行 TV 焦点态不应使用整圈硬描边", optionRowSource.contains("Modifier.border("))
        assertTrue("音轨/字幕行焦点态应使用低饱和蓝青背景提亮", text.contains("Color(0x2E39D7E8)"))
        assertTrue("已选中态应使用细色条表达", text.contains("TrackPickerSelectionRail"))

        val audioTrackText = audioTrackSource.readText()
        assertTrue(audioTrackText.contains("自动选择"))
        assertTrue(audioTrackText.contains("跟随视频默认音轨"))
    }
}
