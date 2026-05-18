package com.chee.videos.tv

import com.chee.videos.feature.tv.TvLongFormDetailRoutePattern
import com.chee.videos.feature.tv.TvLongFormPlayerRoutePattern
import com.chee.videos.feature.tv.TvCatalogWallRoutePattern
import org.junit.Assert.assertFalse
import org.junit.Test

class TvShellAppRouteVisibilityTest {
    @Test
    fun hidesLegacyGlobalSettingsBecauseHomeOwnsSettingsPanel() {
        assertFalse(shouldShowTvGlobalSettings("tv-home"))
        assertFalse(shouldShowTvGlobalSettings(TvCatalogWallRoutePattern))
        assertFalse(shouldShowTvGlobalSettings(TvLongFormDetailRoutePattern))
        assertFalse(shouldShowTvGlobalSettings(TvLongFormPlayerRoutePattern))
        assertFalse(shouldShowTvGlobalSettings(null))
    }
}
