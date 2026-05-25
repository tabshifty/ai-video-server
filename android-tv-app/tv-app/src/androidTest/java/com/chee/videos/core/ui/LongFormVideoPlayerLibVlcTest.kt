package com.chee.videos.core.ui

import androidx.activity.ComponentActivity
import androidx.compose.material3.MaterialTheme
import androidx.compose.ui.test.junit4.createAndroidComposeRule
import androidx.compose.ui.test.onNodeWithContentDescription
import com.chee.videos.core.player.TvVlcLibrary
import com.chee.videos.core.player.newLongFormMediaPlayer
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.videolan.libvlc.MediaPlayer
import androidx.test.ext.junit.runners.AndroidJUnit4

@RunWith(AndroidJUnit4::class)
class LongFormVideoPlayerLibVlcTest {
    @get:Rule
    val composeRule = createAndroidComposeRule<ComponentActivity>()

    private lateinit var player: MediaPlayer

    @Before
    fun setUp() {
        player = newLongFormMediaPlayer(TvVlcLibrary.shared(composeRule.activity))
    }

    @After
    fun tearDown() {
        player.release()
    }

    @Test
    fun coldStartRendersLibVlcBackedControls() {
        composeRule.setContent {
            MaterialTheme {
                LongFormVideoPlayer(
                    title = "LibVLC 测试视频",
                    player = player,
                    isFullscreen = true,
                    onBack = {},
                    onTogglePlayPause = {},
                    onToggleFullscreen = {},
                    tvMode = true,
                )
            }
        }

        composeRule.waitForIdle()
        composeRule.onNodeWithContentDescription("播放").assertExists()
        composeRule.onNodeWithContentDescription("字幕").assertExists()
    }
}
