package com.chee.videos.tv

import com.chee.videos.feature.tv.TvCatalogWallRoutePattern
import com.chee.videos.feature.tv.TvLongFormDetailRoutePattern
import com.chee.videos.feature.tv.TvLongFormPlayerRoutePattern
import com.chee.videos.feature.tv.TvPlayerRoutePattern
import com.chee.videos.feature.tv.TvSeriesRoutePattern
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvShellAppBackPolicyTest {
    @Test
    fun handlesPosterWallAndDetailRoutesWithShellBack() {
        assertTrue(shouldHandleTvShellBack(TvCatalogWallRoutePattern))
        assertTrue(shouldHandleTvShellBack(TvLongFormDetailRoutePattern))
        assertTrue(shouldHandleTvShellBack(TvSeriesRoutePattern))

        assertTrue(shouldHandleTvShellBack("tv/wall/recent?title=最近更新"))
        assertTrue(shouldHandleTvShellBack("tv/detail/movie-1?videoType=movie"))
        assertTrue(shouldHandleTvShellBack("tv/series/series-1"))
    }

    @Test
    fun doesNotHandleHomeOrPlayerRoutesWithShellBack() {
        assertFalse(shouldHandleTvShellBack("tv-home"))
        assertFalse(shouldHandleTvShellBack(TvPlayerRoutePattern))
        assertFalse(shouldHandleTvShellBack(TvLongFormPlayerRoutePattern))
        assertFalse(shouldHandleTvShellBack("tv/player/series-1?season=1&episode=2"))
        assertFalse(shouldHandleTvShellBack("tv/long-form-player/movie-1?videoType=movie"))
        assertFalse(shouldHandleTvShellBack(null))
    }

    @Test
    fun onlyHomeRouteHandlesRootExitConfirm() {
        assertTrue(shouldHandleTvRootExitConfirm("tv-home"))

        assertFalse(shouldHandleTvRootExitConfirm(TvCatalogWallRoutePattern))
        assertFalse(shouldHandleTvRootExitConfirm(TvLongFormDetailRoutePattern))
        assertFalse(shouldHandleTvRootExitConfirm(TvSeriesRoutePattern))
        assertFalse(shouldHandleTvRootExitConfirm(TvPlayerRoutePattern))
        assertFalse(shouldHandleTvRootExitConfirm(TvLongFormPlayerRoutePattern))
        assertFalse(shouldHandleTvRootExitConfirm(null))
    }

    @Test
    fun firstRootBackPressShowsExitPrompt() {
        assertEquals(
            TvRootBackAction.ShowPrompt,
            resolveTvRootBackAction(previousPromptUptimeMillis = null, nowUptimeMillis = 1_000L),
        )
    }

    @Test
    fun secondRootBackPressWithinConfirmWindowExits() {
        assertEquals(
            TvRootBackAction.Exit,
            resolveTvRootBackAction(previousPromptUptimeMillis = 1_000L, nowUptimeMillis = 2_999L),
        )
    }

    @Test
    fun rootBackPressAfterConfirmWindowShowsPromptAgain() {
        assertEquals(
            TvRootBackAction.ShowPrompt,
            resolveTvRootBackAction(previousPromptUptimeMillis = 1_000L, nowUptimeMillis = 3_001L),
        )
    }
}
