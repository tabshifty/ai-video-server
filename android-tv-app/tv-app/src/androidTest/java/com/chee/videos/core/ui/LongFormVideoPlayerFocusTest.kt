package com.chee.videos.core.ui

import androidx.activity.ComponentActivity
import androidx.compose.material3.MaterialTheme
import androidx.compose.ui.input.key.Key
import androidx.compose.ui.test.ExperimentalTestApi
import androidx.compose.ui.test.assertIsFocused
import androidx.compose.ui.test.assertIsNotFocused
import androidx.compose.ui.test.junit4.createAndroidComposeRule
import androidx.compose.ui.test.onNodeWithContentDescription
import androidx.compose.ui.test.performKeyInput
import androidx.compose.ui.test.pressKey
import androidx.test.ext.junit.runners.AndroidJUnit4
import com.chee.videos.core.player.TvVlcLibrary
import com.chee.videos.core.player.newLongFormMediaPlayer
import org.junit.After
import org.junit.Before
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import org.videolan.libvlc.MediaPlayer

@RunWith(AndroidJUnit4::class)
@OptIn(ExperimentalTestApi::class)
class LongFormVideoPlayerFocusTest {
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
    fun tvMode_initialRenderDoesNotFocusControlButtonUntilDownKey() {
        composeRule.setContent {
            MaterialTheme {
                LongFormVideoPlayer(
                    title = "测试视频",
                    player = player,
                    isFullscreen = false,
                    onBack = {},
                    onTogglePlayPause = {},
                    onToggleFullscreen = {},
                    tvMode = true,
                )
            }
        }

        composeRule.waitForIdle()

        composeRule
            .onNodeWithContentDescription("播放")
            .assertIsNotFocused()

        composeRule.onNodeWithContentDescription("播放").performKeyInput {
            pressKey(Key.DirectionDown)
        }

        composeRule
            .onNodeWithContentDescription("播放")
            .assertIsFocused()
    }

    @Test
    fun tvMode_controlButtonsWrapLeftAndRightWithoutFocusingSlider() {
        composeRule.setContent {
            MaterialTheme {
                LongFormVideoPlayer(
                    title = "测试视频",
                    player = player,
                    isFullscreen = true,
                    onBack = {},
                    onTogglePlayPause = {},
                    onToggleFullscreen = {},
                    tvMode = true,
                    onRequestExitPlayback = {},
                    onExitPlayback = {},
                )
            }
        }

        composeRule.waitForIdle()
        composeRule.onNodeWithContentDescription("播放").performKeyInput {
            pressKey(Key.DirectionDown)
        }

        composeRule.onNodeWithContentDescription("播放").assertIsFocused()
        composeRule.onNodeWithContentDescription("播放").performKeyInput {
            pressKey(Key.DirectionLeft)
        }
        composeRule.onNodeWithContentDescription("退出播放").assertIsFocused()

        composeRule.onNodeWithContentDescription("退出播放").performKeyInput {
            pressKey(Key.DirectionRight)
        }
        composeRule.onNodeWithContentDescription("播放").assertIsFocused()
    }
}
