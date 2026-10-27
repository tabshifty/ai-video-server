@file:androidx.annotation.OptIn(markerClass = [androidx.media3.common.util.UnstableApi::class])

package com.chee.videos.feature.tv

import android.net.Uri
import android.view.KeyEvent
import android.view.WindowManager
import androidx.activity.compose.setContent
import androidx.compose.material3.Button
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Text
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.mutableStateOf
import androidx.compose.ui.Modifier
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.input.InputMode
import androidx.compose.ui.platform.LocalInputModeManager
import androidx.compose.ui.test.assertIsFocused
import androidx.compose.ui.test.junit4.createAndroidComposeRule
import androidx.compose.ui.test.onNodeWithText
import androidx.media3.datasource.DataSpec
import androidx.media3.datasource.DefaultHttpDataSource
import androidx.test.ext.junit.runners.AndroidJUnit4
import com.chee.videos.core.ui.KeepScreenOnEffect
import com.chee.videos.tv.TvMainActivity
import java.net.InetAddress
import java.net.ServerSocket
import java.util.concurrent.Executors
import java.util.concurrent.TimeUnit
import org.junit.Assert.*
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith

@RunWith(AndroidJUnit4::class)
@OptIn(ExperimentalComposeUiApi::class)
class TvScreensaverIntegrationTest {
    @get:Rule val composeRule = createAndroidComposeRule<TvMainActivity>()

    @Test
    fun wakeRemotePressDoesNotClickOriginalButtonOrLoseFocus() {
        val focus = FocusRequester()
        var clicks = 0
        composeRule.runOnUiThread {
            composeRule.activity.setContent {
                MaterialTheme {
                    val inputMode = LocalInputModeManager.current
                    Button(onClick = { clicks++ }, modifier = Modifier.focusRequester(focus)) {
                        Text("原页面按钮")
                    }
                    LaunchedEffect(Unit) {
                        inputMode.requestInputMode(InputMode.Keyboard)
                        focus.requestFocus()
                    }
                }
            }
        }
        composeRule.onNodeWithText("原页面按钮").assertIsFocused()
        composeRule.runOnIdle {
            val activity = composeRule.activity
            activity.screensaver.setShowing(true)
            assertTrue(activity.dispatchKeyEvent(KeyEvent(KeyEvent.ACTION_DOWN, KeyEvent.KEYCODE_DPAD_CENTER)))
            assertFalse(activity.screensaver.showing.value)
            assertTrue(activity.dispatchKeyEvent(KeyEvent(KeyEvent.ACTION_DOWN, KeyEvent.KEYCODE_DPAD_CENTER)))
            assertTrue(activity.dispatchKeyEvent(KeyEvent(KeyEvent.ACTION_UP, KeyEvent.KEYCODE_DPAD_CENTER)))
            assertEquals(0, clicks)
        }
        composeRule.onNodeWithText("原页面按钮").assertIsFocused()
        composeRule.runOnIdle {
            composeRule.activity.dispatchKeyEvent(KeyEvent(KeyEvent.ACTION_DOWN, KeyEvent.KEYCODE_DPAD_CENTER))
            composeRule.activity.dispatchKeyEvent(KeyEvent(KeyEvent.ACTION_UP, KeyEvent.KEYCODE_DPAD_CENTER))
        }
        composeRule.runOnIdle { assertEquals(1, clicks) }
    }

    @Test
    fun playbackAndScreensaverKeepScreenOnIndependently() {
        val playback = mutableStateOf(true)
        val screen = mutableStateOf(false)
        val playbackPresent = mutableStateOf(true)
        composeRule.runOnUiThread {
            composeRule.activity.setContent {
                if (playbackPresent.value) KeepScreenOnEffect(playback.value)
                KeepScreenOnEffect(screen.value)
            }
        }
        fun assertFlag(expected: Boolean) = composeRule.runOnIdle {
            val flags = composeRule.activity.window.attributes.flags
            assertEquals(expected, flags and WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON != 0)
        }
        assertFlag(true)
        composeRule.runOnIdle { screen.value = true; playback.value = false }
        assertFlag(true)
        composeRule.runOnIdle { playbackPresent.value = false }
        assertFlag(true)
        composeRule.runOnIdle { screen.value = false }
        assertFlag(false)
    }

    @Test
    fun tokenRefreshUpdatesExistingMediaDataSourceWithoutReplacingFactory() {
        val token = mutableStateOf("before-refresh")
        var factory: DefaultHttpDataSource.Factory? = null
        composeRule.runOnUiThread {
            composeRule.activity.setContent { factory = rememberTvPlaybackDataSource(token.value) }
        }
        composeRule.waitForIdle()
        val original = checkNotNull(factory)
        val source = original.createDataSource()
        val executor = Executors.newSingleThreadExecutor()
        try {
            ServerSocket(0, 1, InetAddress.getByName("127.0.0.1")).use { server ->
                fun readAuthorization(): String? {
                    val received = executor.submit<String?> {
                        server.accept().use { socket ->
                            socket.soTimeout = 5_000
                            val reader = socket.getInputStream().bufferedReader()
                            var authorization: String? = null
                            while (true) {
                                val line = reader.readLine() ?: break
                                if (line.isEmpty()) break
                                if (line.startsWith("Authorization:", ignoreCase = true)) {
                                    authorization = line.substringAfter(':').trim()
                                }
                            }
                            socket.getOutputStream().write(
                                "HTTP/1.1 200 OK\r\nContent-Length: 1\r\nConnection: close\r\n\r\nx".toByteArray(),
                            )
                            authorization
                        }
                    }
                    source.open(DataSpec(Uri.parse("http://127.0.0.1:${server.localPort}/video")))
                    source.read(ByteArray(1), 0, 1)
                    source.close()
                    return received.get(5, TimeUnit.SECONDS)
                }
                assertEquals("Bearer before-refresh", readAuthorization())
                composeRule.runOnIdle { token.value = "after-refresh" }
                composeRule.runOnIdle { assertSame(original, factory) }
                assertEquals("Bearer after-refresh", readAuthorization())
            }
        } finally {
            source.close()
            executor.shutdownNow()
        }
    }
}
