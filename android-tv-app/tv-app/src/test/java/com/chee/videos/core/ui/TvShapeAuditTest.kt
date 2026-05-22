package com.chee.videos.core.ui

import java.nio.file.Files
import java.nio.file.Path
import kotlin.io.path.extension
import kotlin.io.path.isRegularFile
import kotlin.io.path.readText
import kotlin.io.path.relativeTo
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Assert.fail
import org.junit.Test

class TvShapeAuditTest {

    private val mainSourceRoot: Path = Path.of("src/main/java")
    private val appChromePath: Path = mainSourceRoot
        .resolve("com/chee/videos/core/ui/AppChrome.kt")

    private val symmetricRoundedRegex = Regex("""RoundedCornerShape\(\s*(\d+)\.dp\s*\)""")
    private val allowedRadii = setOf(8, 16, 999)
    private val excludedPathPrefixes = listOf(
        "com/chee/videos/MainActivity.kt",
        "com/chee/videos/VideoHomeApp.kt",
        "com/chee/videos/feature/auth/",
        "com/chee/videos/feature/home/",
        "com/chee/videos/feature/mine/",
        "com/chee/videos/feature/player/",
        "com/chee/videos/feature/shorts/",
        "com/chee/videos/feature/shortdiscover/",
        "com/chee/videos/feature/shortsearch/",
        "com/chee/videos/feature/imagecollections/",
        "com/chee/videos/core/ui/ShortVideo",
        "com/chee/videos/core/ui/AppChrome.kt",
    )

    @Test
    fun `tv main source uses only whitelisted symmetric RoundedCornerShape radii`() {
        assertTrue("TV main source 目录必须存在", Files.isDirectory(mainSourceRoot))

        val offenders = mutableListOf<String>()
        Files.walk(mainSourceRoot).use { stream ->
            stream
                .filter { it.isRegularFile() && it.extension == "kt" }
                .forEach { file ->
                    val relative = file.relativeTo(mainSourceRoot).toString().replace('\\', '/')
                    if (excludedPathPrefixes.any { relative.startsWith(it) }) return@forEach
                    file.readText().lineSequence().forEachIndexed { idx, line ->
                        symmetricRoundedRegex.findAll(line).forEach { match ->
                            val radius = match.groupValues[1].toInt()
                            if (radius !in allowedRadii) {
                                offenders += "$relative:${idx + 1}  半径=${radius}dp  ${line.trim()}"
                            }
                        }
                    }
                }
        }

        if (offenders.isNotEmpty()) {
            fail(
                "以下位置出现未列入白名单的对称圆角字面量。" +
                    "白名单：8dp(ChipShape) / 16dp(SurfaceShape) / 999dp(PillShape)；" +
                    "其余请统一走 AppChrome.SurfaceShape：\n" +
                    offenders.joinToString("\n"),
            )
        }
    }

    @Test
    fun `AppChrome exposes the unified shape token set`() {
        assertTrue("AppChrome.kt 必须存在", Files.isRegularFile(appChromePath))
        val source = appChromePath.readText()
        assertTrue("AppChrome 必须暴露 RadiusDp 通用圆角 token", source.contains("RadiusDp"))
        assertTrue("AppChrome 必须暴露 SurfaceShape 通用形状 token", source.contains("SurfaceShape"))
        assertTrue("AppChrome 必须暴露 ChipShape 小 chip 形状 token", source.contains("ChipShape"))
        assertTrue("AppChrome 必须保留 PillShape 胶囊形 token", source.contains("PillShape"))
        assertFalse(
            "AppChrome 必须删除旧 CardShape（22dp 已废弃，统一走 SurfaceShape）",
            source.contains("val CardShape"),
        )
        assertFalse(
            "AppChrome 必须删除旧 SectionShape（18dp 已废弃，统一走 SurfaceShape）",
            source.contains("val SectionShape"),
        )
    }
}
