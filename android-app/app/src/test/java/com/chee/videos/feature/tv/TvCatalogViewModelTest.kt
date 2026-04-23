package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class TvCatalogViewModelTest {

    @Test
    fun updateQuery_filtersSeriesByTitle() {
        val viewModel = TvCatalogViewModel()

        viewModel.updateQuery("静默")

        val state = viewModel.uiState.value
        assertEquals("静默", state.query)
        assertEquals(1, state.searchResults.size)
        assertEquals("静默轨道", state.searchResults.first().title)
    }

    @Test
    fun updateQuery_withBlankTextRestoresBrowseMode() {
        val viewModel = TvCatalogViewModel()

        viewModel.updateQuery("夜网")
        viewModel.updateQuery("")

        val state = viewModel.uiState.value
        assertEquals("", state.query)
        assertTrue(state.searchResults.isEmpty())
        assertTrue(state.sections.isNotEmpty())
    }
}
