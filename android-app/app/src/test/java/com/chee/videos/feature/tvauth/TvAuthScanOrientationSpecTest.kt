package com.chee.videos.feature.tvauth

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * 扫码登录 TV 不应被第三方库 zxing-android-embedded 默认 CaptureActivity 的横屏方向锁绑架。
 *
 * 根因：库默认 CaptureActivity 在其 Manifest 里声明了 sensorLandscape，app 若不注册自定义
 * CaptureActivity 覆盖，扫码页就被锁成横屏；setOrientationLocked(false) 只抑制运行时
 * setRequestedOrientation，清不掉 Manifest 级方向锁。本 spec 守住“扫码走自定义竖屏 Activity”
 * 这一行为契约，防回归到库默认横屏。
 */
class TvAuthScanOrientationSpecTest {

    @Test
    fun `scan uses a custom capture activity instead of the library default`() {
        val appSource = Path.of("src/main/java/com/chee/videos/VideoHomeApp.kt").readText()

        assertTrue(
            "扫码必须注册自定义 CaptureActivity，不得用库默认横屏 Activity",
            appSource.contains("setCaptureActivity("),
        )
    }

    @Test
    fun `custom capture activity is declared portrait in the manifest`() {
        val manifest = Path.of("src/main/AndroidManifest.xml").readText()

        assertTrue(
            "Manifest 必须声明扫码用的自定义 CaptureActivity",
            manifest.contains("CaptureActivity"),
        )
        assertTrue(
            "扫码 Activity 必须显式锁定竖屏，不得沿用库默认 sensorLandscape",
            manifest.contains("android:screenOrientation=\"portrait\""),
        )
    }

    @Test
    fun `custom capture activity subclass exists and extends the library base`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/tvauth/PortraitCaptureActivity.kt")
        assertTrue("必须存在自定义扫码 Activity 源文件", sourcePath.exists())

        val source = sourcePath.readText()
        assertTrue(
            "自定义扫码 Activity 必须继承库 CaptureActivity 基类",
            source.contains(": CaptureActivity(") || source.contains(": CaptureActivity {") ||
                source.contains("CaptureActivity()"),
        )
        assertFalse(
            "自定义扫码 Activity 不得调用 setOrientationLocked，方向交由 Manifest 锁定",
            source.contains(".setOrientationLocked("),
        )
    }
}
