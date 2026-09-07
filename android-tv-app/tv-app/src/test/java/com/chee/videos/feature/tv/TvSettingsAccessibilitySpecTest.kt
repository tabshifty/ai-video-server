package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSettingsAccessibilitySpecTest {
    private val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt").readText()

    @Test
    fun `settings can scroll to controls below the viewport`() {
        val panel = source.substringAfter("private fun TvHomeSettingsPanel(")
            .substringBefore("private fun TvSeriesAutoplaySettingRow(")
        assertTrue(panel.contains(".verticalScroll(rememberScrollState())"))
        assertTrue(panel.contains("TvLayoutSpec.scrollBottomSafePaddingDp.dp"))
    }

    @Test
    fun `autoplay has one interactive switch target`() {
        val row = source.substringAfter("private fun TvSeriesAutoplaySettingRow(")
            .substringBefore("private fun TvSeekStepSettingRow(")
        assertTrue(row.contains(".toggleable(value = enabled, role = Role.Switch"))
        assertTrue(row.contains("onCheckedChange = null"))
        assertFalse(row.contains(".clickable"))
    }

    @Test
    fun `seek step exposes mutually exclusive selection`() {
        val row = source.substringAfter("private fun TvSeekStepSettingRow(")
            .substringBefore("private fun TvSettingsActionRow(")
        assertTrue(row.contains(".selectableGroup()"))
        assertTrue(row.contains(".selectable(selected = selected, role = Role.RadioButton)"))
    }
}
