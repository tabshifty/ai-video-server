package com.chee.videos.core.ui

import java.nio.file.Files
import java.nio.file.Path
import kotlin.io.path.extension
import kotlin.io.path.isRegularFile
import kotlin.io.path.readText
import kotlin.io.path.relativeTo
import org.junit.Assert.assertTrue
import org.junit.Assert.fail
import org.junit.Test

class TvNoBareLaunchedEffectFocusRequestAuditTest {

    private val mainSourceRoot: Path = Path.of("src/main/java")

    @Test
    fun `tv main source must not contain LaunchedEffect blocks calling requestFocus directly`() {
        assertTrue("TV main source 目录必须存在", Files.isDirectory(mainSourceRoot))

        val offenders = mutableListOf<String>()
        Files.walk(mainSourceRoot).use { stream ->
            stream
                .filter { it.isRegularFile() && it.extension == "kt" }
                .forEach { file ->
                    val text = file.readText()
                    if (!text.contains("LaunchedEffect")) return@forEach
                    val matches = bareLaunchedEffectFocusRequestMatches(text)
                    matches.forEach { lineNumber ->
                        offenders += "${file.relativeTo(mainSourceRoot)}:$lineNumber"
                    }
                }
        }

        if (offenders.isNotEmpty()) {
            fail(
                "以下位置仍在 LaunchedEffect 内裸调 requestFocus()，必须改用 LaunchedTvInitialFocus：\n" +
                    offenders.joinToString("\n"),
            )
        }
    }
}

internal fun bareLaunchedEffectFocusRequestMatches(source: String): List<Int> {
    val lines = source.lines()
    val matches = mutableListOf<Int>()
    var depth = 0
    var insideLaunchedEffect = false
    var lineNumber = 0
    val launchedEffectRegex = Regex("""\bLaunchedEffect\s*\(""")
    val requestFocusRegex = Regex("""\.requestFocus\s*\(""")
    for (line in lines) {
        lineNumber += 1
        val openCount = line.count { it == '{' }
        val closeCount = line.count { it == '}' }
        if (!insideLaunchedEffect) {
            if (launchedEffectRegex.containsMatchIn(line) && line.contains("{")) {
                insideLaunchedEffect = true
                depth = openCount - closeCount
                if (requestFocusRegex.containsMatchIn(line)) {
                    matches += lineNumber
                }
                if (depth <= 0) {
                    insideLaunchedEffect = false
                    depth = 0
                }
                continue
            }
        } else {
            if (requestFocusRegex.containsMatchIn(line)) {
                matches += lineNumber
            }
            depth += openCount - closeCount
            if (depth <= 0) {
                insideLaunchedEffect = false
                depth = 0
            }
        }
    }
    return matches
}

class BareLaunchedEffectFocusRequestMatcherTest {

    @Test
    fun `flags requestFocus inside LaunchedEffect block`() {
        val source = """
            fun screen() {
                LaunchedEffect(Unit) {
                    focusRequester.requestFocus()
                }
            }
        """.trimIndent()
        assertTrue(bareLaunchedEffectFocusRequestMatches(source).isNotEmpty())
    }

    @Test
    fun `does not flag requestFocus outside LaunchedEffect`() {
        val source = """
            fun screen() {
                onEvent {
                    focusRequester.requestFocus()
                }
            }
        """.trimIndent()
        assertTrue(bareLaunchedEffectFocusRequestMatches(source).isEmpty())
    }

    @Test
    fun `does not flag LaunchedTvInitialFocus blocks`() {
        val source = """
            fun screen() {
                LaunchedTvInitialFocus(Unit) {
                    focusRequester.requestFocus()
                }
            }
        """.trimIndent()
        assertTrue(bareLaunchedEffectFocusRequestMatches(source).isEmpty())
    }

    @Test
    fun `flags requestFocus on the same line as LaunchedEffect`() {
        val source = "LaunchedEffect(Unit) { focusRequester.requestFocus() }"
        assertTrue(bareLaunchedEffectFocusRequestMatches(source).isNotEmpty())
    }
}
