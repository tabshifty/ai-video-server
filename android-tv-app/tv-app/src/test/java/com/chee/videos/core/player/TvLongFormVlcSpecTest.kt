package com.chee.videos.core.player

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvLongFormVlcSpecTest {
    @Test
    fun longFormVideoPlayer_usesLibVlcInsteadOfMedia3() {
        val source = Path.of("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt").readText()

        assertFalse(source.contains("import androidx.media3."))
        assertFalse(source.contains("ExoPlayer"))
        assertFalse(source.contains("PlayerView"))
        assertFalse(source.contains("MediaItem"))
        assertFalse(source.contains("CaptionStyleCompat"))
        assertFalse(source.contains("PlaybackParameters"))
        assertTrue(source.contains("import org.videolan.libvlc.MediaPlayer"))
        assertTrue(source.contains("VlcLongFormSurface("))
        assertTrue(source.contains("MediaPlayer.EventListener"))
        assertTrue(source.contains("setRate(2f)"))
    }

    @Test
    fun longFormSubtitleSupport_injectsSubtitleWithLibVlcSlave() {
        val source = Path.of("src/main/java/com/chee/videos/core/ui/LongFormSubtitleSupport.kt").readText()
        val playerSource = Path.of("src/main/java/com/chee/videos/core/player/TvLongFormVlcPlayer.kt").readText()

        assertFalse(source.contains("MediaItem.SubtitleConfiguration"))
        assertTrue(source.contains("applyLongFormMediaSource("))
        assertTrue(playerSource.contains("addSlave(IMedia.Slave(IMedia.Slave.Type.Subtitle"))
    }

    @Test
    fun tvBuildFile_keepsLibVlcAndAllowsDvDedicatedMedia3Dependencies() {
        val source = Path.of("build.gradle.kts").readText()

        assertTrue(source.contains("org.videolan.android:libvlc-all"))
        assertTrue(source.contains("androidx.media3:media3-exoplayer"))
        assertTrue(source.contains("androidx.media3:media3-ui"))
        // versionCode/versionName 在 LibVLC 迁移之后会继续随后续任务向上累加，这里只断言不回退到迁移前
        val versionCode = Regex("""versionCode\s*=\s*(\d+)""").find(source)?.groupValues?.get(1)?.toInt()
        assertTrue("versionCode should be ≥ 68 since LibVLC migration baseline", (versionCode ?: 0) >= 68)
        assertTrue(
            "versionName should follow 0.1.x semver since LibVLC migration baseline",
            Regex("""versionName\s*=\s*"0\.1\.(\d+)"""").find(source)?.groupValues?.get(1)?.toIntOrNull()
                ?.let { it >= 68 } == true,
        )
    }

    @Test
    fun tvLongFormScreens_keepMedia3BehindDolbyVisionDedicatedComponent() {
        val longForm = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt").readText()
        val series = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").readText()
        val media3 = Path.of("src/main/java/com/chee/videos/feature/tv/TvDolbyVisionMedia3Player.kt").readText()

        listOf(longForm, series).forEach { source ->
            assertFalse(source.contains("import androidx.media3."))
            assertTrue(source.contains("TvDolbyVisionMedia3Player("))
            assertTrue(source.contains("MediaPlayer"))
        }
        assertTrue(media3.contains("import androidx.media3.exoplayer.ExoPlayer"))
        assertTrue(media3.contains("PlayerView"))
    }
}
