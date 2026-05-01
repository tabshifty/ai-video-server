package com.chee.videos.feature.home

import com.chee.videos.core.model.VideoListItemDto
import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class HomeAvPresentationTest {

    @Test
    fun `prefers metadata av code over title for primary label`() {
        val model = buildAvCatalogCardModel(
            VideoListItemDto(
                id = "1",
                title = "SSIS-111 夜色引力",
                type = "av",
                duration = 3_600,
                metadata = mapOf("av_code" to "IPX-777"),
            ),
        )

        assertEquals("IPX-777", model.primaryText)
        assertEquals("SSIS-111 夜色引力", model.secondaryText)
    }

    @Test
    fun `extracts normalized av code from title when metadata is absent`() {
        val model = buildAvCatalogCardModel(
            VideoListItemDto(
                id = "2",
                title = "ssis 123 清晨边界",
                type = "av",
                duration = 4_200,
            ),
        )

        assertEquals("SSIS-123", model.primaryText)
        assertEquals("ssis 123 清晨边界", model.secondaryText)
    }

    @Test
    fun `falls back to full title when no av code can be resolved`() {
        val model = buildAvCatalogCardModel(
            VideoListItemDto(
                id = "3",
                title = "只剩标题没有番号",
                type = "av",
                duration = 1_800,
            ),
        )

        assertEquals("只剩标题没有番号", model.primaryText)
        assertNull(model.secondaryText)
    }
}
