package com.chee.videos.core.player

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvLongFormExoPlayerSpecTest {
    @Test
    fun tvLongFormScreensUseUnifiedMedia3PlayerInsteadOfLibVlc() {
        val longForm = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt").readText()
        val series = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").readText()

        listOf(longForm, series).forEach { source ->
            assertTrue(source.contains("TvLongFormMedia3Player("))
            assertTrue(source.contains("TvSeriesCorePlaybackOverlay("))
            assertFalse(source.contains("org.videolan.libvlc"))
            assertFalse(source.contains("TvVlcLibrary"))
            assertFalse(source.contains("newLongFormMediaPlayer"))
            assertFalse(source.contains("applyLongFormMediaSource("))
            assertFalse(source.contains("appendAccessTokenQuery("))
            assertFalse(source.contains("addSlave("))
            assertFalse(source.contains("LongFormVideoPlayer("))
        }
    }

    @Test
    fun tvLongFormMedia3SubtitleSupportUsesMediaItemAndBearerHeaderPath() {
        val trackSupport = Path.of("src/main/java/com/chee/videos/feature/tv/TvMedia3TrackSupport.kt").readText()
        val player = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt").readText()

        assertTrue(trackSupport.contains("MediaItem.SubtitleConfiguration.Builder"))
        assertTrue(player.contains("DefaultHttpDataSource.Factory()"))
        assertTrue(player.contains("\"Authorization\" to \"Bearer \$accessToken\""))
        assertTrue(player.contains(".setSubtitleConfigurations(subtitleConfigurations)"))
        assertFalse(trackSupport.contains("appendAccessTokenQuery"))
        assertFalse(trackSupport.contains("access_token"))
    }

    @Test
    fun tvBuildFileKeepsIptvLibVlcAndMedia3Dependencies() {
        val source = Path.of("build.gradle.kts").readText()

        assertTrue(source.contains("org.videolan.android:libvlc-all"))
        assertTrue(source.contains("androidx.media3:media3-exoplayer"))
        assertTrue(source.contains("androidx.media3:media3-ui"))
        // versionCode/versionName 会继续随后续任务向上累加，这里只断言不回退到本次迁移前
        val versionCode = Regex("""versionCode\s*=\s*(\d+)""").find(source)?.groupValues?.get(1)?.toInt()
        assertTrue("versionCode should be ≥ 107 since ExoPlayer long-form migration", (versionCode ?: 0) >= 107)
        assertTrue(
            "versionName should follow 0.1.x semver since ExoPlayer long-form migration",
            Regex("""versionName\s*=\s*"0\.1\.(\d+)"""").find(source)?.groupValues?.get(1)?.toIntOrNull()
                ?.let { it >= 107 } == true,
        )
    }

    @Test
    fun tvLongFormMedia3PlayerUsesTextureViewWithoutLegacyPlayerChoice() {
        val player = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt").readText()
        val layout = Path.of("src/main/res/layout/tv_long_form_media3_player_view.xml").readText()

        assertTrue(player.contains("R.layout.tv_long_form_media3_player_view"))
        assertTrue(layout.contains("androidx.media3.ui.PlayerView"))
        assertTrue(layout.contains("app:surface_type=\"texture_view\""))
        assertTrue(layout.contains("app:keep_content_on_player_reset=\"true\""))
        assertTrue(layout.contains("app:shutter_background_color=\"@android:color/transparent\""))
        assertFalse(player.contains("PlayerView(it).apply"))
        assertFalse(player.contains("setShutterBackgroundColor(android.graphics.Color.BLACK)"))
        assertFalse(player.contains("isVlcRoute"))
        assertFalse(player.contains("TvVlcLibrary"))
    }

    @Test
    fun tvLongFormScreensNoLongerUseDvDedicatedMedia3Naming() {
        val longForm = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt").readText()
        val series = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").readText()
        val media3 = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt").readText()

        listOf(longForm, series).forEach { source ->
            assertTrue(source.contains("TvLongFormMedia3Player("))
            assertFalse(source.contains("TvDolbyVisionMedia3Player("))
            assertFalse(source.contains("MEDIA3_DOLBY_VISION"))
            assertFalse(source.contains("isVlcRoute"))
        }
        assertTrue(media3.contains("import androidx.media3.exoplayer.ExoPlayer"))
        assertTrue(media3.contains("PlayerView"))
        assertFalse(media3.contains("杜比视界专用播放链路启动超时"))
    }
}
