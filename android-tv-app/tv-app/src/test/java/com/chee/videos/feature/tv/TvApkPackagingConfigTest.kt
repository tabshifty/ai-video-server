package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvApkPackagingConfigTest {
    private val buildGradlePath = Path.of("build.gradle.kts")

    @Test
    fun tvApkBuildSplitsOnlyArmAbiPackagesWithoutUniversalApk() {
        assertTrue("TV App Gradle 配置必须存在", buildGradlePath.exists())

        val gradle = buildGradlePath.readText()
        assertTrue(
            "TV APK 必须启用 ABI split，避免把四套 VLC native so 打进同一个 APK",
            gradle.contains("splits") && gradle.contains("abi"),
        )
        assertTrue(
            "TV APK ABI split 只应包含 ARM 32/64 位",
            gradle.contains("include(\"armeabi-v7a\", \"arm64-v8a\")"),
        )
        assertTrue(
            "TV APK 不应生成 x86/x86_64 分包",
            gradle.contains("exclude(\"x86\", \"x86_64\")"),
        )
        assertTrue(
            "TV APK 不应生成 universal APK，分发时按 ARM ABI 选择安装包",
            gradle.contains("isUniversalApk = false"),
        )
        assertFalse(
            "TV APK ABI 拆包后不应继续包含通用 APK 开关",
            gradle.contains("isUniversalApk = true"),
        )
    }

    @Test
    fun releaseBuildEnablesCodeAndResourceShrinkWithoutRemovingLibVlc() {
        assertTrue("TV App Gradle 配置必须存在", buildGradlePath.exists())

        val gradle = buildGradlePath.readText()
        assertTrue(
            "Release 必须开启 R8，以降低最终 APK 体积",
            gradle.contains("isMinifyEnabled = true"),
        )
        assertTrue(
            "Release 必须开启资源瘦身，以降低最终 APK 体积",
            gradle.contains("isShrinkResources = true"),
        )
        assertTrue(
            "IPTV 播放兼容性依赖必须保留 LibVLC",
            gradle.contains("implementation(\"org.videolan.android:libvlc-all:3.6.0\")"),
        )
    }
}
