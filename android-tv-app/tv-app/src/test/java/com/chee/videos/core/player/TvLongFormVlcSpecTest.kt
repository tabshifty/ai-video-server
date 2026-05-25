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
    fun tvBuildFile_removesMedia3Dependencies() {
        val source = Path.of("build.gradle.kts").readText()

        assertFalse(source.contains("media3-"))
        assertTrue(source.contains("versionCode = 68"))
        assertTrue(source.contains("versionName = \"0.1.68\""))
    }

    @Test
    fun tvLongFormScreens_useLibVlcMediaPlayer() {
        val longForm = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt").readText()
        val series = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").readText()

        listOf(longForm, series).forEach { source ->
            assertFalse(source.contains("ExoPlayer"))
            assertFalse(source.contains("DefaultHttpDataSource"))
            assertFalse(source.contains("DefaultMediaSourceFactory"))
            assertTrue(source.contains("MediaPlayer"))
        }
    }
}
