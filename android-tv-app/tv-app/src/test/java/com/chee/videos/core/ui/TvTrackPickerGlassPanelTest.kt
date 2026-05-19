package com.chee.videos.core.ui

import java.nio.file.Files
import java.nio.file.Path
import kotlin.io.path.readText
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

        val audioTrackText = audioTrackSource.readText()
        assertTrue(audioTrackText.contains("自动选择"))
        assertTrue(audioTrackText.contains("跟随视频默认音轨"))
    }
}
