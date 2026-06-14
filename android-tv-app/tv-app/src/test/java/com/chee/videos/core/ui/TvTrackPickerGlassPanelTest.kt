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
        assertTrue(text.contains("TrackPickerPanelBorderBrush"))
        assertTrue(text.contains("AppChrome.Accent.copy(alpha = 0.40f)"))
        assertTrue(text.contains("supportingText"))
        assertTrue("音轨/字幕行必须显式处理遥控确认键", text.contains("handleTrackPickerConfirmKey"))
        assertTrue("音轨/字幕行不能复用全局焦点整块 glow，避免播放器选择器出现旧式大面积高亮", !text.contains(".tvFocusableGlow("))
        val optionRowSource = text
            .substringAfter("private fun SubtitleOptionRow(")
            .substringBefore("@Composable\nprivate fun TrackPickerSelectionRail(")
        assertFalse("音轨/字幕行 TV 焦点态不应使用整圈硬描边", optionRowSource.contains("Modifier.border("))
        assertTrue("音轨/字幕行焦点态应使用低饱和暖金背景提亮", text.contains("AppChrome.Accent.copy(alpha = 0.24f)"))
        assertFalse("音轨/字幕选择器不应残留旧蓝青高光", text.contains("39D7E8"))
        assertFalse("音轨/字幕选择器不应残留旧蓝青边缘高光", text.contains("25D9F2"))
        assertTrue("已选中态应使用细色条表达", text.contains("TrackPickerSelectionRail"))

        val audioTrackText = audioTrackSource.readText()
        assertTrue(audioTrackText.contains("自动选择"))
        assertTrue(audioTrackText.contains("跟随视频默认音轨"))
    }
}
