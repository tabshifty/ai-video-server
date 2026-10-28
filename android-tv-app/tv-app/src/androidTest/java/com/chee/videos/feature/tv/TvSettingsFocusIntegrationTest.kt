package com.chee.videos.feature.tv

import androidx.activity.compose.setContent
import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.LocalBringIntoViewSpec
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.Modifier
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.graphics.toPixelMap
import androidx.compose.ui.graphics.asAndroidBitmap
import androidx.compose.ui.input.InputMode
import androidx.compose.ui.input.key.Key
import androidx.compose.ui.platform.LocalInputModeManager
import androidx.compose.ui.semantics.SemanticsActions
import androidx.compose.ui.test.ExperimentalTestApi
import androidx.compose.ui.test.assertIsFocused
import androidx.compose.ui.test.assertIsNotSelected
import androidx.compose.ui.test.assertIsSelected
import androidx.compose.ui.test.captureToImage
import androidx.compose.ui.test.isFocused
import androidx.compose.ui.test.junit4.createAndroidComposeRule
import androidx.compose.ui.test.onNodeWithText
import androidx.compose.ui.test.onRoot
import androidx.compose.ui.test.performKeyInput
import androidx.compose.ui.test.performSemanticsAction
import androidx.compose.ui.test.pressKey
import androidx.compose.ui.unit.dp
import androidx.test.ext.junit.runners.AndroidJUnit4
import com.chee.videos.core.ui.AppChrome
import com.chee.videos.core.ui.TvMinimalBringIntoViewSpec
import com.chee.videos.tv.TvMainActivity
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import org.junit.runner.RunWith
import java.io.File

@RunWith(AndroidJUnit4::class)
@OptIn(ExperimentalFoundationApi::class, ExperimentalTestApi::class, ExperimentalComposeUiApi::class)
class TvSettingsFocusIntegrationTest {
    @get:Rule val composeRule = createAndroidComposeRule<TvMainActivity>()

    @Test
    fun remoteFocusHasVisibleBorderWithoutChangingSelectionAndKeepsBottomSpace() {
        val firstAction = FocusRequester()
        composeRule.runOnUiThread {
            composeRule.activity.setContent {
                MaterialTheme {
                    CompositionLocalProvider(LocalBringIntoViewSpec provides TvMinimalBringIntoViewSpec) {
                        Box(Modifier.fillMaxSize().background(AppChrome.Canvas)) {
                            TvHomeSettingsPanel(
                                tvSeekStepSeconds = 10,
                                onSelectTvSeekStepSeconds = {},
                                seriesAutoplayEnabled = true,
                                onSetSeriesAutoplayEnabled = {},
                                onRepair = {}, onLogout = {}, onSwitchServer = {},
                                firstActionFocusRequester = firstAction,
                                modifier = Modifier.fillMaxSize(),
                            )
                        }
                        val input = LocalInputModeManager.current
                        LaunchedEffect(Unit) {
                            input.requestInputMode(InputMode.Keyboard)
                            firstAction.requestFocus()
                        }
                    }
                }
            }
        }
        composeRule.waitUntil(10_000) {
            composeRule.onAllNodes(androidx.compose.ui.test.hasText("5 秒")).fetchSemanticsNodes().isNotEmpty()
        }
        composeRule.onNodeWithText("重新配对").assertIsFocused()
        repeat(7) {
            composeRule.onNode(isFocused()).performKeyInput { pressKey(Key.DirectionDown) }
            composeRule.waitForIdle()
            val focused = composeRule.onNode(isFocused()).fetchSemanticsNode().boundsInRoot
            val root = composeRule.onRoot().fetchSemanticsNode().boundsInRoot
            val safe = with(composeRule.density) { 32.dp.toPx() }
            assertTrue("每次遥控下移后，聚焦项都应完整可见且保留底部安全区", focused.bottom <= root.bottom - safe + 1f)
            assertTrue("聚焦项顶部不可被滚动容器裁切", focused.top >= root.top)
        }
        // 检查生产页面的最后一组控件，而非只检查 padding 字面量。
        val last = composeRule.onNodeWithText("5 秒")
        last.performSemanticsAction(SemanticsActions.RequestFocus)
        composeRule.waitForIdle()
        last.assertIsFocused()
        val viewport = composeRule.onRoot().fetchSemanticsNode().boundsInRoot
        val bounds = last.fetchSemanticsNode().boundsInRoot
        val safePixels = with(composeRule.density) { 32.dp.toPx() }
        assertTrue("末行聚焦后仍应保留底部安全区", viewport.bottom - bounds.bottom >= safePixels - 1f)

        val selected = composeRule.onNodeWithText("10 分钟")
        val target = composeRule.onNodeWithText("15 分钟")
        selected.performSemanticsAction(SemanticsActions.RequestFocus)
        composeRule.waitForIdle()
        selected.assertIsSelected()
        target.assertIsNotSelected()
        val before = target.captureToImage().toPixelMap()
        selected.performKeyInput { pressKey(Key.DirectionRight) }
        composeRule.waitForIdle()
        target.assertIsFocused().assertIsNotSelected()
        selected.assertIsSelected()
        val after = target.captureToImage().toPixelMap()
        fun whitePixels(pixels: androidx.compose.ui.graphics.PixelMap): Int {
            var count = 0
            for (y in 0 until pixels.height / 5) {
                for (x in pixels.width / 4 until pixels.width * 3 / 4) {
                    val color = pixels[x, y]
                    if (color.red > 0.9f && color.green > 0.9f && color.blue > 0.9f) count++
                }
            }
            return count
        }
        assertTrue("遥控移动后应实际绘制浅色焦点边框", whitePixels(after) > whitePixels(before) + 10)
        File(composeRule.activity.cacheDir, "tv-settings-focus.png").outputStream().use {
            composeRule.onRoot().captureToImage().asAndroidBitmap().compress(android.graphics.Bitmap.CompressFormat.PNG, 100, it)
        }
    }
}
