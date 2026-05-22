package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class TvCatalogFocusPolicyTest {

    @Test
    fun `prefers continue watching when present`() {
        assertEquals(
            TvCatalogInitialFocusTarget.FEATURED,
            resolveTvCatalogInitialFocusTarget(
                hasFeaturedContent = true,
                hasContinueWatching = true,
                sectionItemCounts = listOf(3, 2),
                tvSeriesCount = 4,
                movieCount = 2,
                avCount = 1,
            ),
        )
    }

    @Test
    fun `falls back to first section item before shelves`() {
        assertEquals(
            TvCatalogInitialFocusTarget.FEATURED,
            resolveTvCatalogInitialFocusTarget(
                hasFeaturedContent = true,
                hasContinueWatching = false,
                sectionItemCounts = listOf(0, 2),
                tvSeriesCount = 4,
                movieCount = 2,
                avCount = 1,
            ),
        )
        assertEquals(
            TvCatalogInitialFocusTarget.FIRST_SECTION_ITEM,
            resolveTvCatalogInitialFocusTarget(
                hasFeaturedContent = false,
                hasContinueWatching = false,
                sectionItemCounts = listOf(0, 2),
                tvSeriesCount = 4,
                movieCount = 2,
                avCount = 1,
            ),
        )
    }

    @Test
    fun `falls back through shelves then menu when content is empty`() {
        assertEquals(
            TvCatalogInitialFocusTarget.TV_SERIES_ITEM,
            resolveTvCatalogInitialFocusTarget(
                hasFeaturedContent = false,
                hasContinueWatching = false,
                sectionItemCounts = emptyList(),
                tvSeriesCount = 1,
                movieCount = 2,
                avCount = 3,
            ),
        )
        assertEquals(
            TvCatalogInitialFocusTarget.MOVIE_ITEM,
            resolveTvCatalogInitialFocusTarget(
                hasFeaturedContent = false,
                hasContinueWatching = false,
                sectionItemCounts = emptyList(),
                tvSeriesCount = 0,
                movieCount = 2,
                avCount = 3,
            ),
        )
        assertEquals(
            TvCatalogInitialFocusTarget.AV_ITEM,
            resolveTvCatalogInitialFocusTarget(
                hasFeaturedContent = false,
                hasContinueWatching = false,
                sectionItemCounts = emptyList(),
                tvSeriesCount = 0,
                movieCount = 0,
                avCount = 3,
            ),
        )
        assertEquals(
            TvCatalogInitialFocusTarget.MENU,
            resolveTvCatalogInitialFocusTarget(
                hasFeaturedContent = false,
                hasContinueWatching = false,
                sectionItemCounts = emptyList(),
                tvSeriesCount = 0,
                movieCount = 0,
                avCount = 0,
            ),
        )
        assertEquals(
            "sections 非空但全为 0 时仍应回退到 MENU",
            TvCatalogInitialFocusTarget.MENU,
            resolveTvCatalogInitialFocusTarget(
                hasFeaturedContent = false,
                hasContinueWatching = false,
                sectionItemCounts = listOf(0, 0),
                tvSeriesCount = 0,
                movieCount = 0,
                avCount = 0,
            ),
        )
    }

    @Test
    fun `empty home focus fallback requests composed menu instead of hidden search`() {
        val screenPath = Path.of("src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt")
        assertTrue("TV 首页必须存在", screenPath.exists())

        val source = screenPath.readText()
        assertTrue(
            "TV 首页空内容启动时必须回退到已组合的左侧菜单焦点，避免请求未绑定搜索框导致启动崩溃；" +
                "并且必须走 tryRequestFocus 这一兜底入口，防止 Compose 1.7 FocusRequester 未挂载时 ISE 透出主 Looper",
            source.contains("TvCatalogInitialFocusTarget.MENU -> menuFocusRequester.tryRequestFocus()"),
        )
    }
}
