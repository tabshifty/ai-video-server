package com.chee.videos.tv

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * TV 启动器身份审计：自适应图标 + TV banner + 深色窗口底色是应用在 leanback
 * 启动器上的门面，禁止回退到系统占位图标或无 banner 状态。
 */
class TvLauncherIdentitySpecTest {

    private val manifest: String by lazy {
        Path.of("src/main/AndroidManifest.xml").readText()
    }

    @Test
    fun manifestWiresLauncherIdentity() {
        assertTrue("应用图标必须使用自适应图标", manifest.contains("android:icon=\"@mipmap/ic_launcher\""))
        assertTrue("圆形图标必须使用自适应图标", manifest.contains("android:roundIcon=\"@mipmap/ic_launcher_round\""))
        assertTrue("必须声明 TV banner（leanback 启动器 320x180dp 卡片）", manifest.contains("android:banner=\"@drawable/tv_banner\""))
        assertFalse("禁止回退到系统占位图标", manifest.contains("@android:drawable/sym_def_app_icon"))
    }

    @Test
    fun adaptiveIconAndBannerResourcesExist() {
        assertTrue(Path.of("src/main/res/mipmap-anydpi-v26/ic_launcher.xml").exists())
        assertTrue(Path.of("src/main/res/mipmap-anydpi-v26/ic_launcher_round.xml").exists())
        assertTrue(Path.of("src/main/res/drawable/ic_launcher_glyph.xml").exists())
        assertTrue(Path.of("src/main/res/drawable/ic_launcher_background.xml").exists())
        assertTrue(Path.of("src/main/res/drawable/tv_banner.xml").exists())
        val icon = Path.of("src/main/res/mipmap-anydpi-v26/ic_launcher.xml").readText()
        assertTrue("图标必须是 adaptive-icon 结构", icon.contains("<adaptive-icon"))
    }

    @Test
    fun launcherArtUsesOnlyAppChromeFamilyColors() {
        // 图标与 banner 的色值必须来自 AppChrome 家族（暖金 + 暗玻璃），禁止引入新色
        val allowed = listOf("040508", "080A0D", "10161F", "E8B85B", "EFC463", "D6A64F", "00000000")
        listOf(
            "src/main/res/drawable/ic_launcher_glyph.xml",
            "src/main/res/drawable/ic_launcher_background.xml",
            "src/main/res/drawable/tv_banner.xml",
        ).forEach { path ->
            val colors = Regex("#[0-9A-Fa-f]{6,8}").findAll(Path.of(path).readText()).map { it.value }.toList()
            colors.forEach { color ->
                assertTrue(
                    "$path 出现非 AppChrome 家族色 $color",
                    allowed.any { color.uppercase().endsWith(it.uppercase()) },
                )
            }
        }
    }

    @Test
    fun appThemeKeepsDarkWindowBackground() {
        val themes = Path.of("src/main/res/values/themes.xml").readText()
        assertTrue(
            "Theme.VideoHome 必须声明深色 windowBackground（防止切回应用主题时白闪）",
            themes.contains("<item name=\"android:windowBackground\">@color/tv_window_background</item>"),
        )
        val colors = Path.of("src/main/res/values/colors.xml").readText()
        assertTrue(
            "tv_window_background 必须等于 AppChrome.Canvas（#FF040508），两处值必须同步",
            colors.contains("<color name=\"tv_window_background\">#FF040508</color>"),
        )
        val chrome = Path.of("src/main/java/com/chee/videos/core/ui/AppChrome.kt").readText()
        assertTrue(
            "AppChrome.Canvas 与 XML 侧 tv_window_background 必须保持 0xFF040508",
            chrome.contains("val Canvas = Color(0xFF040508)"),
        )
    }
}
