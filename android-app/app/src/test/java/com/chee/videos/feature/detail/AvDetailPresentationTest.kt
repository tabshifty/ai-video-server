package com.chee.videos.feature.detail

import com.chee.videos.core.model.UserStateDto
import com.chee.videos.core.model.VideoActorDto
import com.chee.videos.core.model.VideoDetailDto
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
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

    @Test
    fun `buildAvDetailActorModels prefers api actors and resolves relative avatar url`() {
        val models = buildAvDetailActorModels(
            baseUrl = "http://127.0.0.1:8080",
            detail = VideoDetailDto(
                id = "av-2",
                title = "IPZZ-777 夜色样片",
                actors = listOf(
                    VideoActorDto(
                        id = "actor-1",
                        name = "演员甲",
                        avatarUrl = "/api/v1/actors/actor-1/avatar",
                    ),
                ),
                metadata = mapOf("actors" to listOf("元数据演员")),
                userState = UserStateDto(),
            ),
        )

        assertEquals(1, models.size)
        assertEquals("actor-1", models[0].id)
        assertEquals("演员甲", models[0].name)
        assertEquals("http://127.0.0.1:8080/api/v1/actors/actor-1/avatar", models[0].avatarUrl)
        assertTrue(models[0].hasAvatar)
        assertTrue(models[0].canOpenDetail)
    }

    @Test
    fun `buildAvDetailActorModels falls back to metadata names without avatar`() {
        val models = buildAvDetailActorModels(
            baseUrl = "http://127.0.0.1:8080",
            detail = VideoDetailDto(
                id = "av-3",
                title = "MIDV-123 晨雾样片",
                metadata = mapOf("actors" to listOf("演员乙", "演员丙")),
                userState = UserStateDto(),
            ),
        )

        assertEquals(listOf("演员乙", "演员丙"), models.map { it.name })
        assertNull(models[0].id)
        assertNull(models[0].avatarUrl)
        assertFalse(models[0].hasAvatar)
        assertFalse(models[0].canOpenDetail)
        assertNull(models[1].id)
        assertNull(models[1].avatarUrl)
        assertFalse(models[1].hasAvatar)
        assertFalse(models[1].canOpenDetail)
    }
}
