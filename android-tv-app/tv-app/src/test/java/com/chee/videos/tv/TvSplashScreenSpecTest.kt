package com.chee.videos.tv

import java.nio.file.Files
import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.fileSize
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSplashScreenSpecTest {
    private val manifestPath = Path.of("src/main/AndroidManifest.xml")
    private val themesPath = Path.of("src/main/res/values/themes.xml")
    private val backgroundPath = Path.of("src/main/res/drawable/tv_splash_background.xml")
    private val splashImagePath = Path.of("src/main/res/drawable-nodpi/tv_splash_image.png")
    private val activityPath = Path.of("src/main/java/com/chee/videos/tv/TvMainActivity.kt")
    private val buildGradlePath = Path.of("build.gradle.kts")

    @Test
    fun tvLauncherUsesSplashStartingThemeThenReturnsToAppTheme() {
        assertTrue("TV Manifest 必须存在", manifestPath.exists())
        assertTrue("TV 主题文件必须存在", themesPath.exists())
        assertTrue("TV MainActivity 必须存在", activityPath.exists())

        val manifest = manifestPath.readText()
        val themes = themesPath.readText()
        val activity = activityPath.readText()

        assertTrue(
            "TV 启动 Activity 必须使用专用 starting theme 承载开屏页图片",
            manifest.contains("android:theme=\"@style/Theme.VideoHome.Starting\""),
        )
        assertTrue(
            "TV starting theme 必须配置窗口背景，而不是在 Compose 内固定等待",
            themes.contains("<style name=\"Theme.VideoHome.Starting\"") &&
                themes.contains("<item name=\"android:windowBackground\">@drawable/tv_splash_background</item>"),
        )
        assertTrue(
            "TV MainActivity onCreate 必须先切回正常 App 主题，避免开屏窗口背景污染后续页面",
            activity.contains("setTheme(R.style.Theme_VideoHome)") &&
                activity.indexOf("setTheme(R.style.Theme_VideoHome)") < activity.indexOf("super.onCreate(savedInstanceState)"),
        )
    }

    @Test
    fun tvSplashImageUsesNodpiWindowBackgroundResource() {
        assertTrue("TV splash 背景 XML 必须存在", backgroundPath.exists())
        assertTrue("TV splash 图片必须放在 drawable-nodpi，避免密度缩放", splashImagePath.exists())
        assertTrue("TV splash 图片不应为空", Files.isRegularFile(splashImagePath) && splashImagePath.fileSize() > 128_000L)

        val background = backgroundPath.readText()
        assertTrue(
            "TV splash 背景必须先铺黑底，再显示用户提供的开屏图片",
            background.contains("<item android:drawable=\"@android:color/black\" />") &&
                background.contains("android:src=\"@drawable/tv_splash_image\""),
        )
    }

    @Test
    fun tvSplashChangeBumpsTvAppVersion() {
        assertTrue("TV App Gradle 配置必须存在", buildGradlePath.exists())
        val gradle = buildGradlePath.readText()

        val versionCode = Regex("""versionCode\s*=\s*(\d+)""").find(gradle)?.groupValues?.get(1)?.toIntOrNull()
        val versionPatch = Regex("""versionName\s*=\s*"0\.1\.(\d+)"""").find(gradle)?.groupValues?.get(1)?.toIntOrNull()

        assertTrue("TV splash 功能更新后 versionCode 不应低于 98", (versionCode ?: 0) >= 98)
        assertTrue("TV splash 功能更新后 versionName patch 不应低于 98", (versionPatch ?: 0) >= 98)
    }
}
