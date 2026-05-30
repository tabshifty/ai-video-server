package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSeriesEpisodeRailSpecTest {
    @Test
    fun seriesPlayerUsesInlineEpisodeRailInsteadOfBottomSheet() {
        val source = readSource("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt")

        assertFalse(source.contains("ModalBottomSheet("))
        assertFalse(source.contains("LazyColumn("))
        assertFalse(source.contains("text = \"选集播放\""))
        assertTrue(source.contains("tvControlsVariant = TvLongFormControlsVariant.SeriesEpisodeRail"))
        assertTrue(source.contains("episodeRailItems = episodeRailItems"))
        assertTrue(source.contains("currentEpisodeRailItemId = currentEpisode?.id"))
        assertTrue(source.contains("onEpisodeRailVisibilityChanged = viewModel::setSelectorVisible"))
    }

    @Test
    fun longFormPlayerDefinesSeriesEpisodeRailVariantAndDisplayOnlyProgressBar() {
        val source = readSource("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt")

        assertTrue(source.contains("enum class TvLongFormControlsVariant"))
        assertTrue(source.contains("SeriesEpisodeRail"))
        assertTrue(source.contains("TvEpisodeRail("))
        assertTrue(source.contains("TvPlaybackProgressBar("))
        assertTrue(source.contains("episodeRailVisible"))
        assertTrue(source.contains("EnterEpisodeRail"))
        assertTrue(source.contains("ExitEpisodeRail"))
        assertTrue(source.contains("tvControlsVariant == TvLongFormControlsVariant.Default"))
    }

    @Test
    fun episodeRailTooltipTracksFocusAndCurrentEpisodeSelectionHasNoSideEffect() {
        val source = readSource("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt")

        assertTrue(source.contains("visible = focused"))
        assertTrue(source.contains("text = item.title"))
        assertTrue(source.contains("hideAllTvUi()"))
        assertTrue(source.contains("if (item.id != currentEpisodeRailItemId)"))
    }

    private fun readSource(relative: String): String {
        val path = Path.of(relative)
        assertTrue("$relative 必须存在", path.exists())
        return path.readText()
    }
}
