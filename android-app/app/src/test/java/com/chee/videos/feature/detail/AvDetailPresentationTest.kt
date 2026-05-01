package com.chee.videos.feature.detail

import com.chee.videos.core.model.UserStateDto
import com.chee.videos.core.model.VideoDetailDto
import org.junit.Assert.assertEquals
import org.junit.Test

class AvDetailPresentationTest {

    @Test
    fun `buildAvDetailHeroModel prefers av code and falls back actors from metadata`() {
        val model = buildAvDetailHeroModel(
            VideoDetailDto(
                id = "av-1",
                title = "SSIS-321 绯色样片",
                description = "作品简介",
                duration = 4_800,
                metadata = mapOf(
                    "av_code" to "SSIS-321",
                    "release_date" to "2026-04-01",
                    "actors" to listOf("演员甲", "演员乙"),
                ),
                userState = UserStateDto(),
            ),
        )

        assertEquals("SSIS-321", model.primaryText)
        assertEquals("SSIS-321 绯色样片", model.secondaryTitle)
        assertEquals(listOf("演员甲", "演员乙"), model.actorNames)
    }
}
