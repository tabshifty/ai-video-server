package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvHomeInteractionSpecTest {
    private val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt").readText()

    @Test
    fun `content back returns to selected navigation before root exit`() {
        assertTrue(source.contains("BackHandler(enabled = !menuHasFocus)"))
        assertTrue(source.contains("menuFocusRequesters.getValue(uiState.selectedMenu)"))
        assertTrue(source.contains("onFocusChanged = { menuHasFocus = it }"))
    }

    @Test
    fun `loading keeps the same side navigation composed`() {
        val screen = source.substringBefore("private fun TvCatalogSearchBar(")
        assertTrue(Regex("TvHomeSideMenu\\(").findAll(screen).count() == 1)
        assertTrue(screen.indexOf("TvHomeSideMenu(") < screen.indexOf("TvPageLoadingState("))
        assertTrue(screen.contains("uiState.loading || isSearching || menuHasFocus"))
    }

    @Test
    fun `navigation names and selected state are exposed`() {
        val button = source.substringAfter("private fun TvHomeSideMenuButton(")
            .substringBefore("private fun TvHomeSettingsPanel(")
        assertTrue(button.contains("text = item.label"))
        assertTrue(button.contains(".semantics { this.selected = selected }"))
    }

    @Test
    fun `search keeps native text focus and supports ime action`() {
        val searchItem = source.substringAfter("item(key = \"search\")")
            .substringBefore("item(key = \"search-header\")")
        assertFalse(searchItem.contains("tvFocusableScaleOnly"))
        val search = source.substringAfter("private fun TvCatalogSearchBar(")
            .substringBefore("private fun TvHomeSideMenu(")
        assertTrue(search.contains("imeAction = ImeAction.Search"))
        assertTrue(search.contains("keyboardController?.hide()"))
    }

    @Test
    fun `right enters content and a section requester has only one owner`() {
        assertTrue(source.contains("event.key == Key.DirectionRight"))
        assertTrue(source.contains("onEnterContent = ::enterContent"))
        assertTrue(source.contains("if (!uiState.loading && !enterContent()) focusManager.moveFocus(FocusDirection.Right)"))
        assertTrue(source.contains("sectionIndex == uiState.sections.indexOfFirst { it.items.isNotEmpty() }"))
    }
}
