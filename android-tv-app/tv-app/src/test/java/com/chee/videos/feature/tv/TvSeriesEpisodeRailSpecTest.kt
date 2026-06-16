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
        assertTrue(source.contains("TvSeriesCorePlaybackOverlay("))
        assertTrue(source.contains("episodeRailItems = episodeRailItems"))
        assertTrue(source.contains("currentEpisodeRailItemId = currentEpisode?.id"))
        assertTrue(source.contains("onSelectEpisodeRailItem = { selectedItem ->"))
        assertTrue(source.contains("onEpisodeRailVisibilityChanged = viewModel::setSelectorVisible"))
        assertTrue(source.contains("openEpisodeRailRequestKey = openEpisodeRailRequestKey"))
    }

    @Test
    fun longFormPlayerDefinesSeriesEpisodeRailVariantAndDisplayOnlyProgressBar() {
        val source = readSource("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt")

        assertTrue(source.contains("enum class TvLongFormControlsVariant"))
        assertTrue(source.contains("SeriesEpisodeRail"))
        assertTrue(source.contains("TvEpisodeRail("))
        assertTrue(source.contains("TvSeriesControlsPage("))
        assertTrue(source.contains("TvPlaybackProgressBar("))
        assertTrue(source.contains("episodeRailVisible"))
        assertTrue(source.contains("EnterEpisodeRail"))
        assertTrue(source.contains("ExitEpisodeRail"))
        assertTrue(source.contains("tvControlsVariant == TvLongFormControlsVariant.Default"))
        assertTrue(source.contains("AnimatedContent("))
        assertTrue(source.contains("slideInVertically("))
        assertTrue(source.contains("slideOutVertically("))
        assertTrue(source.contains("for (attempt in 0 until 4)"))
        assertTrue(source.contains("currentEpisodeRailFocusRequester.tryRequestFocus()"))
    }

    @Test
    fun episodeRailTooltipTracksFocusAndCurrentEpisodeSelectionHasNoSideEffect() {
        val source = readSource("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt")
        val episodeRailSource = source.substringAfter("fun TvEpisodeRail(").substringBefore("@Composable\ninternal fun CompactPlayerControlButton(")

        assertTrue(episodeRailSource.contains("if (focused)"))
        assertTrue(episodeRailSource.contains("text = item.title"))
        assertTrue(source.contains("hideAllTvUi()"))
        assertTrue(source.contains("if (item.id != currentEpisodeRailItemId)"))
        assertTrue(episodeRailSource.contains("text = formatTvEpisodeRailLabel(item.number)"))
        assertTrue(episodeRailSource.contains("listState.scrollToItem(nextFirstVisibleItemIndex)"))
        assertFalse(episodeRailSource.contains("animateScrollToItem"))
    }

    @Test
    fun episodeRailUsesStaticHighlightWithoutGlowAndControlsDownEntersRailLocally() {
        val source = readSource("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt")
        val controlsModifierSource = source
            .substringAfter("fun controlFocusModifier(target: TvControlFocusTarget): Modifier {")
            .substringBefore("val titleOverlayData =")
        val episodeRailSource = source
            .substringAfter("fun TvEpisodeRail(")
            .substringBefore("@Composable\ninternal fun CompactPlayerControlButton(")

        assertTrue(controlsModifierSource.contains("AndroidKeyEvent.KEYCODE_DPAD_DOWN"))
        assertTrue(controlsModifierSource.contains("handleTvRemoteKeyAction(TvRemoteKeyAction.EnterEpisodeRail)"))
        assertFalse(episodeRailSource.contains("tvFocusableGlow("))
        assertTrue(episodeRailSource.contains(".focusable(enabled = item.playable)"))
        assertTrue(episodeRailSource.contains("BorderStroke("))
        assertTrue(episodeRailSource.contains("TvEpisodeRailLayoutTokens.ItemSlotWidthDp.dp"))
        assertTrue(episodeRailSource.contains("wrapContentWidth(Alignment.CenterHorizontally, unbounded = true)"))
        assertFalse(episodeRailSource.contains("focusedScale"))
        assertFalse(episodeRailSource.contains("graphicsLayer"))
    }

    @Test
    fun controlsHandleHorizontalFocusLocallyWhenFocused() {
        val source = readSource("src/main/java/com/chee/videos/core/ui/LongFormVideoPlayer.kt")
        val controlsModifierSource = source
            .substringAfter("fun controlFocusModifier(target: TvControlFocusTarget): Modifier {")
            .substringBefore("val titleOverlayData =")

        assertTrue(controlsModifierSource.contains("AndroidKeyEvent.KEYCODE_DPAD_LEFT"))
        assertTrue(controlsModifierSource.contains("AndroidKeyEvent.KEYCODE_DPAD_RIGHT"))
        assertTrue(controlsModifierSource.contains("resolveTvControlHorizontalFocusTarget("))
        assertTrue(controlsModifierSource.contains("requestHorizontalControlFocus("))
        assertTrue(controlsModifierSource.contains("scheduleAutoHideControls()"))
        assertTrue(controlsModifierSource.contains("focusState.isFocused || focusState.hasFocus"))
    }

    private fun readSource(relative: String): String {
        val path = Path.of(relative)
        assertTrue("$relative 必须存在", path.exists())
        return path.readText()
    }
}
