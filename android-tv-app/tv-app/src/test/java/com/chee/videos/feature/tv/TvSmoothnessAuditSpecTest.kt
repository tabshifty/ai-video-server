package com.chee.videos.feature.tv

import java.io.File
import java.nio.file.Path
import kotlin.io.path.readText
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * TV 端 Lazy 列表卫生审计：key 必须稳定且唯一（服务端可返回重名 section/演员，
 * 重复 key 会让 Compose 直接崩溃），混排容器必须声明 contentType，
 * 播放器快照轮询必须做发射端去重。
 */
class TvSmoothnessAuditSpecTest {

    @Test
    fun lazyListKeysAreCollisionHardened() {
        val catalog = Path.of("src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt").readText()
        assertTrue(
            "目录 section 键禁止只用可重名的 section.title，必须带 index 兜底",
            !catalog.contains("key = { _, section -> section.title }"),
        )
        assertTrue(
            catalog.contains("key = { index, section -> \"section-\$index-\${section.title}\" }"),
        )

        val seriesDetail = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt").readText()
        assertTrue(
            "剧集详情演员键禁止 id.ifBlank { name }（空 id 同名演员会重复 key 崩溃）",
            !seriesDetail.contains("key = { it.id.ifBlank { it.name } }"),
        )
        assertTrue(seriesDetail.contains("cast-\$index-"))

        val longFormDetail = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormDetailScreen.kt").readText()
        assertTrue(
            "长视频详情演员行必须带 key",
            !longFormDetail.contains("items(actors.size)") && longFormDetail.contains("actor-\$index-"),
        )

        val iptv = Path.of("src/main/java/com/chee/videos/feature/tv/TvIptvScreen.kt").readText()
        assertTrue(
            "IPTV 频道键必须带分组前缀，防止跨分组频道 id 重复",
            iptv.contains("key = { \"\${group.group}-\${it.id}\" }"),
        )
    }

    @Test
    fun mixedContentContainersDeclareContentType() {
        val catalog = Path.of("src/main/java/com/chee/videos/feature/tv/TvCatalogScreen.kt").readText()
        assertTrue(
            "目录外层 LazyColumn 混排 hero/error/续播/分区/货架/入口多种形态，必须声明 contentType 让同类项复用组合槽",
            Regex("contentType").findAll(catalog).count() >= 6,
        )
        val posterWall = Path.of("src/main/java/com/chee/videos/feature/tv/TvPosterWallScreen.kt").readText()
        assertTrue(
            "海报墙网格的海报卡必须显式声明 contentType，与刷新/错误/加载更多头尾项区隔",
            posterWall.contains("contentType = { _, _ -> \"poster\" }"),
        )
    }

    @Test
    fun featureTvLazyItemsAlwaysCarryKeys() {
        val offenders = File("src/main/java/com/chee/videos/feature/tv")
            .walkTopDown()
            .filter { it.isFile && it.extension == "kt" }
            .flatMap { file ->
                Regex("""items\([a-zA-Z0-9_.]+\.size\)""")
                    .findAll(file.readText())
                    .map { "${file.name}: ${it.value}" }
            }
            .toList()
        assertTrue(
            "feature/tv 下禁止无 key 的 items(list.size) 形态，命中：$offenders",
            offenders.isEmpty(),
        )
    }

    @Test
    fun media3SnapshotPollingDeduplicatesEmissions() {
        val source = Path.of("src/main/java/com/chee/videos/feature/tv/TvLongFormMedia3Player.kt").readText()
        assertTrue(
            "4Hz 快照轮询必须在发射端做结构相等去重（暂停/缓冲态不重复下发相同快照）",
            source.contains("var lastEmitted: TvMedia3PlaybackSnapshot? = null") &&
                source.contains("if (snapshot != lastEmitted)"),
        )
    }
}
