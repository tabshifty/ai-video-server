package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvContinueWatchingDto
import com.chee.videos.core.model.TvHomePayload
import com.chee.videos.core.model.TvHomeVideoDto
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Rule
import org.junit.Test
import kotlinx.coroutines.test.runTest

class TvHomeNavigationTest {
    @get:Rule
    val mainDispatcherRule = MainDispatcherRule()

    @Test
    fun menuDefaultsToSeriesWithFixedOrderAndAvLabel() {
        assertEquals(
            listOf("电视剧", "电影", "18+", "IPTV", "搜索", "设置"),
            TvHomeMenuItem.defaults().map { it.label },
        )
        assertEquals(TvHomeMenuItem.Series, TvHomeMenuItem.defaultSelected())
        assertEquals("tv", TvHomeMenuItem.Series.homeKind)
        assertEquals("movie", TvHomeMenuItem.Movie.homeKind)
        assertEquals("av", TvHomeMenuItem.Adult.homeKind)
        assertFalse(TvHomeMenuItem.Iptv.isContentKind)
    }

    @Test
    fun selectingTypeMenuRequestsMatchingKindAndMapsAdultToAv() = runTest {
        val repository = FakeTvRepository(
            homePayloads = mapOf(
                "tv" to TvHomePayload(kind = "tv"),
                "movie" to TvHomePayload(kind = "movie"),
                "av" to TvHomePayload(kind = "av"),
            ),
        )
        val viewModel = TvCatalogViewModel(repository = repository)

        viewModel.awaitIdle()
        viewModel.selectMenu(TvHomeMenuItem.Movie)
        viewModel.awaitIdle()
        viewModel.selectMenu(TvHomeMenuItem.Adult)
        viewModel.awaitIdle()

        assertEquals(listOf("tv", "movie", "av"), repository.homeRequests.map { it.kind })
        assertEquals(TvHomeMenuItem.Adult, viewModel.uiState.value.selectedMenu)
        assertEquals("18+", viewModel.uiState.value.selectedMenu.label)
    }

    @Test
    fun searchMenuDoesNotRequestTypedHomeAgainAndUsesSearchPayload() = runTest {
        val repository = FakeTvRepository(homePayload = TvHomePayload())
        val viewModel = TvCatalogViewModel(repository = repository)

        viewModel.awaitIdle()
        viewModel.selectMenu(TvHomeMenuItem.Search)
        viewModel.updateQuery("午夜")
        viewModel.awaitIdle()

        assertEquals(listOf("tv"), repository.homeRequests.map { it.kind })
        assertEquals(listOf("午夜"), repository.searchRequests.map { it.query })
        assertEquals(TvHomeMenuItem.Search, viewModel.uiState.value.selectedMenu)
    }

    @Test
    fun typedHomeSectionsFollowFeaturedRecentWatchingRecentUpdatesAllAndSkipEmptySections() {
        val sections = buildTvTypedHomeSections(
            kind = "movie",
            featured = TvHomeShelfItemUiModel(id = "movie-1", type = "movie", title = "午夜列车", description = "悬疑长片"),
            recentWatching = listOf(
                TvHomeShelfItemUiModel(id = "movie-2", type = "movie", title = "旧片续看", description = "继续播放"),
            ),
            recentUpdates = emptyList(),
        )

        assertEquals(listOf("巨幅推荐", "最近播放", "全部电影"), sections.map { it.title })
        assertEquals("featured", sections[0].role)
        assertEquals("recent_watching", sections[1].role)
        assertEquals("all", sections[2].role)
    }

    @Test
    fun settingsPanelUsesExistingAccountActions() {
        assertEquals(
            listOf("重新配对", "退出登录", "切换服务器"),
            TvHomeSettingsPanelAction.defaults().map { it.label },
        )
        assertEquals(TvHomeMenuItem.Settings, TvHomeSettingsPanelAction.menuItem)
    }

    @Test
    fun focusPolicyMovesPredictablyBetweenSideMenuAndContent() {
        assertEquals(
            TvHomeFocusArea.Content,
            resolveTvHomeFocusMove(TvHomeFocusArea.Menu, TvHomeFocusDirection.Right),
        )
        assertEquals(
            TvHomeFocusArea.Menu,
            resolveTvHomeFocusMove(TvHomeFocusArea.Content, TvHomeFocusDirection.Left),
        )
        assertEquals(
            TvHomeFocusArea.Menu,
            resolveTvHomeFocusMove(TvHomeFocusArea.Menu, TvHomeFocusDirection.Left),
        )
        assertFalse(shouldShowTvHomeSideMenu("tv/player/7/1/1"))
        assertFalse(shouldShowTvHomeSideMenu("tv/detail/movie-1?type=movie"))
        assertTrue(shouldShowTvHomeSideMenu("tv-home"))
    }
}
