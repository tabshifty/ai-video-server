package com.chee.videos.feature.tv

import com.chee.videos.core.model.UserStateDto
import com.chee.videos.core.model.VideoActorDto
import com.chee.videos.core.model.VideoDetailDto
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertNull
import org.junit.Assert.assertTrue
import org.junit.Test
import java.nio.file.Paths
import kotlin.io.path.exists
import kotlin.io.path.readText

class TvLongFormDetailPresentationTest {

    @Test
    fun `builds immersive hero with year meta backdrop and favorite labels`() {
        val hero = buildTvLongFormDetailHero(
            baseUrl = "https://media.example.com",
            detail = VideoDetailDto(
                id = "movie-1",
                title = "午夜列车",
                description = "一部悬疑长片",
                duration = 7260,
                playUrl = "/api/v1/videos/movie-1/source",
                thumbnailPath = "/poster/movie-1.jpg",
                tags = listOf("悬疑", "惊悚"),
                metadata = mapOf(
                    "release_date" to "2026-02-14",
                    "backdrop_path" to "/backdrop/movie-1.jpg",
                ),
            ),
            videoType = "movie",
        )

        assertEquals("电影", hero.eyebrow)
        assertEquals("午夜列车", hero.title)
        assertEquals("2026 · 2小时1分钟 · 悬疑 · 惊悚", hero.metaLine)
        assertEquals("播放", hero.primaryActionLabel)
        assertEquals("收藏", hero.secondaryActionLabel)
        assertEquals("https://media.example.com/poster/movie-1.jpg", hero.posterUrl)
        assertEquals("https://media.example.com/backdrop/movie-1.jpg", hero.backdropUrl)
        assertFalse(hero.usesPosterAsBackdropFallback)
    }

    @Test
    fun `builds av hero with adult eyebrow`() {
        val hero = buildTvLongFormDetailHero(
            baseUrl = "https://media.example.com",
            detail = VideoDetailDto(
                id = "av-1",
                title = "SNIS-001",
                description = "18+作品",
                duration = 5400,
                thumbnailPath = "/poster/av-1.jpg",
            ),
            videoType = "av",
        )

        assertEquals("18+", hero.eyebrow)
        assertEquals("SNIS-001", hero.title)
    }

    @Test
    fun `builds av hero backdrop from original poster while keeping thumbnail poster`() {
        val hero = buildTvLongFormDetailHero(
            baseUrl = "https://media.example.com",
            detail = VideoDetailDto(
                id = "av-2",
                title = "SNIS-002",
                description = "18+作品",
                duration = 5400,
                thumbnailPath = "/poster/av-2-thumb.jpg",
                metadata = mapOf(
                    "poster_original_path" to "/poster/av-2-original.jpg",
                    "poster_cropped_path" to "/poster/av-2-cropped.jpg",
                ),
            ),
            videoType = "av",
        )

        assertEquals("https://media.example.com/poster/av-2-thumb.jpg", hero.posterUrl)
        assertEquals("https://media.example.com/poster/av-2-original.jpg", hero.backdropUrl)
        assertFalse(hero.usesPosterAsBackdropFallback)
    }

    @Test
    fun `builds poster fallback background when no wide backdrop exists`() {
        val hero = buildTvLongFormDetailHero(
            baseUrl = "https://media.example.com",
            detail = VideoDetailDto(
                id = "movie-2",
                title = "无横幅电影",
                thumbnailPath = "/poster/movie-2.jpg",
            ),
            videoType = "movie",
        )

        assertEquals("https://media.example.com/poster/movie-2.jpg", hero.backdropUrl)
        assertTrue(hero.usesPosterAsBackdropFallback)
    }

    @Test
    fun `builds actor row with avatar and placeholder capped at five`() {
        val hero = buildTvLongFormDetailHero(
            baseUrl = "https://media.example.com",
            detail = VideoDetailDto(
                id = "movie-3",
                title = "群像电影",
                actors = listOf(
                    VideoActorDto(id = "a1", name = "演员一", avatarUrl = "/actors/a1.jpg"),
                    VideoActorDto(id = "a2", name = "演员二"),
                    VideoActorDto(id = "a3", name = "演员三"),
                    VideoActorDto(id = "a4", name = "演员四"),
                    VideoActorDto(id = "a5", name = "演员五"),
                    VideoActorDto(id = "a6", name = "演员六"),
                ),
            ),
            videoType = "movie",
        )

        assertEquals(5, hero.actors.size)
        assertEquals("演员一", hero.actors[0].name)
        assertEquals("https://media.example.com/actors/a1.jpg", hero.actors[0].avatarUrl)
        assertTrue(hero.actors[0].hasAvatar)
        assertEquals("演员二", hero.actors[1].name)
        assertNull(hero.actors[1].avatarUrl)
        assertFalse(hero.actors[1].hasAvatar)
    }

    @Test
    fun `builds cancel favorite label from user state`() {
        val hero = buildTvLongFormDetailHero(
            baseUrl = "https://media.example.com",
            detail = VideoDetailDto(
                id = "movie-4",
                title = "已收藏电影",
                userState = UserStateDto(isFavorited = true),
            ),
            videoType = "movie",
        )

        assertEquals("取消收藏", hero.secondaryActionLabel)
    }

    @Test
    fun `detail screen keeps immersive first screen without share or lower cards`() {
        val screenPath = Paths.get("src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt")
        assertTrue("TV 长视频详情页必须存在", screenPath.exists())

        val source = screenPath.readText()
        assertFalse("电影/18+ 详情页不提供分享入口", source.contains("分享"))
        assertFalse("沉浸式首屏不保留更多信息按钮", source.contains("更多信息"))
        assertFalse("沉浸式首屏不保留下方剧情简介卡片", source.contains("key = \"summary\""))
        assertFalse("沉浸式首屏不保留下方标签卡片", source.contains("key = \"tags\""))
        assertFalse("沉浸式首屏不保留下方播放数据卡片", source.contains("key = \"stats\""))
        assertTrue("详情页必须复用收藏切换逻辑", source.contains("viewModel.toggleFavorite()"))
        assertTrue("详情页演员区必须支持头像图片", source.contains("model = actor.avatarUrl"))
        assertTrue("详情页演员区必须支持无头像占位圆", source.contains("!actor.hasAvatar"))
    }
}
