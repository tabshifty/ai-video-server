package com.chee.videos.core.ui

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNull
import org.junit.Test

class TvLongFormTitleOverlayDataTest {
    @Test
    fun movieUsesSinglePrimaryLine() {
        val data = buildTvLongFormTitleOverlayData(
            primaryFallback = "电影标题",
            seriesTitle = null,
            seasonNumber = null,
            episodeNumber = null,
            episodeTitle = null,
        )

        assertEquals("电影标题", data.primary)
        assertNull(data.secondary)
    }

    @Test
    fun seriesUsesSeriesTitleAndEpisodeSubtitle() {
        val data = buildTvLongFormTitleOverlayData(
            primaryFallback = "第七集标题",
            seriesTitle = "明朝那些事",
            seasonNumber = 1,
            episodeNumber = 7,
            episodeTitle = "朱元璋登基",
        )

        assertEquals("明朝那些事", data.primary)
        assertEquals("第 1 季 · 第 7 集 朱元璋登基", data.secondary)
    }

    @Test
    fun blankEpisodeTitleDoesNotLeaveTrailingSpace() {
        val data = buildTvLongFormTitleOverlayData(
            primaryFallback = "明朝那些事",
            seriesTitle = "明朝那些事",
            seasonNumber = 1,
            episodeNumber = 7,
            episodeTitle = " ",
        )

        assertEquals("第 1 季 · 第 7 集", data.secondary)
    }

    @Test
    fun nullEpisodeTitleDoesNotLeaveTrailingSpace() {
        val data = buildTvLongFormTitleOverlayData(
            primaryFallback = "明朝那些事",
            seriesTitle = "明朝那些事",
            seasonNumber = 1,
            episodeNumber = 7,
            episodeTitle = null,
        )

        assertEquals("第 1 季 · 第 7 集", data.secondary)
    }

    @Test
    fun blankSeriesTitleFallsBackToPrimary() {
        val data = buildTvLongFormTitleOverlayData(
            primaryFallback = "第七集标题",
            seriesTitle = " ",
            seasonNumber = 1,
            episodeNumber = 7,
            episodeTitle = null,
        )

        assertEquals("第七集标题", data.primary)
        assertEquals("第 1 季 · 第 7 集", data.secondary)
    }
}
