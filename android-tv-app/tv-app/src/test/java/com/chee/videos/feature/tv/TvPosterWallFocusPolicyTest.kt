package com.chee.videos.feature.tv

import org.junit.Assert.assertEquals
import org.junit.Test

class TvPosterWallFocusPolicyTest {
    @Test
    fun `return restores an item on later pages`() {
        val ids = (0..71).map { "video-$it" }
        assertEquals(49, resolveTvPosterWallRestoreIndex(ids, "video-49"))
    }

    @Test
    fun `reordered content restores by identity`() {
        assertEquals(0, resolveTvPosterWallRestoreIndex(listOf("c", "b", "a"), "c"))
        assertEquals(2, resolveTvPosterWallRestoreIndex(listOf("a", "b", "c"), "c"))
    }

    @Test
    fun `missing or unsaved target falls back to first item`() {
        assertEquals(0, resolveTvPosterWallRestoreIndex(listOf("a", "b"), "removed"))
        assertEquals(0, resolveTvPosterWallRestoreIndex(listOf("a", "b"), null))
        assertEquals(0, resolveTvPosterWallRestoreIndex(emptyList(), "removed"))
    }
}
