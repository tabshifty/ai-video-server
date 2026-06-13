package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

class TvStateFeedbackUsageTest {
    @Test
    fun `key tv pages use shared state feedback components`() {
        val catalog = Path.of("src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt").readText()
        val posterWall = Path.of("src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt").readText()
        val longFormDetail = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt").readText()
        val longFormPlayer = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormPlayerScreen.kt").readText()
        val seriesDetail = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt").readText()
        val seriesPlayer = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesPlayerScreen.kt").readText()
        val iptv = Path.of("src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt").readText()

        assertTrue(catalog.contains("TvPageLoadingState("))
        assertTrue(catalog.contains("TvErrorState("))
        assertTrue(catalog.contains("TvEmptyState("))
        assertTrue(posterWall.contains("TvPageLoadingState("))
        assertTrue(posterWall.contains("TvInlineLoadingState("))
        assertTrue(posterWall.contains("TvErrorState("))
        assertTrue(posterWall.contains("TvEmptyState("))
        assertTrue(longFormDetail.contains("TvPageLoadingState("))
        assertTrue(longFormDetail.contains("TvErrorState("))
        assertTrue(longFormPlayer.contains("TvErrorState("))
        assertTrue(longFormPlayer.contains("onAction = viewModel::load"))
        assertTrue(seriesDetail.contains("TvPageLoadingState("))
        assertTrue(seriesDetail.contains("TvErrorState("))
        assertTrue(seriesPlayer.contains("TvErrorState("))
        assertTrue(seriesPlayer.contains("onAction = viewModel::retry"))
        assertTrue(iptv.contains("TvPageLoadingState("))
        assertTrue(iptv.contains("TvErrorState("))
    }
}
