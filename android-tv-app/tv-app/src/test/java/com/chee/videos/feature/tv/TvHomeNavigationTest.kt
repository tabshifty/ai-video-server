package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvContinueWatchingDto
import com.chee.videos.core.model.TvHomePayload
import com.chee.videos.core.model.TvHomeVideoDto
import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
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
    fun playbackSettingsExposeSeekStepPresets() {
        assertEquals(listOf(5, 10, 15, 20, 30), TvPlaybackSeekStepSetting.allowedSeconds)
        assertEquals(10, TvPlaybackSeekStepSetting.defaultSeconds)
        assertEquals("10秒", TvPlaybackSeekStepSetting.labelFor(10))
        assertEquals(10, TvPlaybackSeekStepSetting.normalize(7))
        assertEquals(20, TvPlaybackSeekStepSetting.normalize(20))
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

    @Test
    fun sideMenuButtonsUseSingleFocusableTargetSoConfirmWorksOnce() {
        val screenPath = Path.of("src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt")

        assertTrue("TV 首页必须存在", screenPath.exists())

        val source = screenPath.readText()
        val sideMenuSource = source.substringAfter("private fun TvHomeSideMenu(")
            .substringBefore("@Composable\nprivate fun TvHomeSideMenuButton(")
        val buttonSource = source.substringAfter("private fun TvHomeSideMenuButton(")
            .substringBefore("@Composable\nprivate fun TvHomeSettingsPanel(")

        assertTrue(
            "IPTV 菜单点击前必须通知 ViewModel 离开目录状态域，避免旧首页/搜索请求在 IPTV 页期间回写首页状态",
            sideMenuSource.contains("if (item == TvHomeMenuItem.Iptv)") &&
                sideMenuSource.contains("onSelect(item)") &&
                sideMenuSource.contains("onOpenIptv()") &&
                sideMenuSource.indexOf("onSelect(item)") < sideMenuSource.indexOf("onOpenIptv()"),
        )
        assertTrue(
            "TV 首页侧边菜单按钮必须使用 tvFocusableGlow 提供唯一焦点目标",
            buttonSource.contains(".tvFocusableGlow("),
        )
        assertFalse(
            "TV 首页侧边菜单按钮不能在 tvFocusableGlow 之后再叠加 .focusable()，否则遥控确认键会先落到重复焦点层，表现为必须按两次才触发",
            buttonSource.contains(".focusable()"),
        )

        val settingsSource = source.substringAfter("@Composable\nprivate fun TvHomeSettingsPanel(")
        assertTrue(settingsSource.contains("播放设置"))
        assertTrue(settingsSource.contains("快进/快退步长"))
    }

    @Test
    fun seriesAutoplaySettingRowProtectsSwitchFromTextCompression() {
        val screenPath = Path.of("src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt")

        assertTrue("TV 首页必须存在", screenPath.exists())

        val source = screenPath.readText()
        val rowSource = source.substringAfter("private fun TvSeriesAutoplaySettingRow(")
            .substringBefore("@Composable\nprivate fun TvSeekStepSettingRow(")

        assertTrue(
            "电视剧自动连播设置行的文案区必须使用 weight(1f) 弹性收缩，避免挤压右侧 Switch",
            rowSource.contains(".weight(1f)"),
        )
        assertTrue(
            "电视剧自动连播设置行标题必须限制单行并省略，避免长文本挤压 Switch",
            rowSource.contains("maxLines = 1") && rowSource.contains("overflow = TextOverflow.Ellipsis"),
        )
        assertTrue(
            "电视剧自动连播设置行右侧 Switch 必须固定占位，避免被左侧文案压缩",
            rowSource.contains("Modifier.width(64.dp)"),
        )
        assertFalse(
            "电视剧自动连播设置行不应固定 68dp 高度，系统字号放大时应允许最小高度向上生长",
            rowSource.contains(".height(68.dp)"),
        )
    }

    @Test
    fun homeShelvesDoNotShowSubtitleCopyUnderTheTitle() {
        val screenPath = Path.of("src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt")

        assertTrue("TV 首页必须存在", screenPath.exists())

        val source = screenPath.readText()
        val shelfSource = source.substringAfter("private fun TvHomeShelf(")
            .substringBefore("@Composable\nprivate fun TvPosterMoreCard(")
        val moreCardSource = source.substringAfter("private fun TvPosterMoreCard(")
            .substringBefore("@Composable\nprivate fun TvHomeShelfCard(")

        assertFalse(
            "货架标题下不应再显示副文案，避免把最近更新/最近播放的语义再解释一遍",
            shelfSource.contains("text = subtitle") || shelfSource.contains("subtitle = \"最近更新\""),
        )
        assertFalse(
            "查看更多卡片不应再显示数量副文案，避免标题下出现多余说明",
            moreCardSource.contains("text = subtitle") || moreCardSource.contains("subtitle = \"共 "),
        )
    }
}
