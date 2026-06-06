package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
import org.junit.Assert.assertEquals
import org.junit.Assert.assertFalse
import org.junit.Assert.assertTrue
import org.junit.Test

class TvSeriesDetailActionSpecTest {
    @Test
    fun `series detail uses shared tv icon action for back action`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt")
        assertTrue("电视剧详情页必须存在", sourcePath.exists())

        val source = sourcePath.readText()
        assertTrue("电视剧详情页返回操作应复用共享 TV 图标操作", source.contains("TvIconActionButton("))
        assertFalse("电视剧详情页不应导入默认 Material IconButton", source.contains("import androidx.compose.material3.IconButton"))
        assertFalse("电视剧详情页不应导入默认 Material Button", source.contains("import androidx.compose.material3.Button"))
        assertFalse("电视剧详情页不应导入默认 Material TextButton", source.contains("import androidx.compose.material3.TextButton"))
        assertFalse("电视剧详情页不应使用默认 Material IconButton", source.contains("IconButton("))
        assertFalse("电视剧详情页不应使用默认 Material TextButton", source.contains("TextButton("))
        assertTrue("电视剧详情页播放、季选择和集选择必须保留遥控焦点状态", source.contains("onFocusChanged"))
        assertTrue("电视剧详情页参考图焦点应使用暖金色，而不是旧版青蓝 glow", source.contains("ReferenceGold"))
        assertFalse("电视剧详情页参考图还原不应继续套用旧版共享青蓝焦点 glow", source.contains(".tvFocusableGlow("))
    }

    @Test
    fun `series detail matches immersive reference body layout`() {
        val sourcePath = Path.of("src/main/java/com/chee/videos/feature/tv/TvSeriesDetailScreen.kt")
        assertTrue("电视剧详情页必须存在", sourcePath.exists())

        val source = sourcePath.readText()
        assertTrue("电视剧详情页主体应使用全屏背景层", source.contains("TvSeriesDetailBackdrop("))
        assertTrue("电视剧详情页主体应拆成左侧信息面板", source.contains("TvSeriesHeroPane("))
        assertTrue("电视剧详情页标题必须独立成参考图式大标题块", source.contains("TvSeriesTitleBlock("))
        assertTrue("电视剧详情页主体应拆成右侧剧集面板", source.contains("TvSeriesEpisodePane("))
        assertTrue("电视剧详情页应保留演员区", source.contains("TvSeriesCastRow("))
        assertTrue("电视剧详情页右侧剧集应使用横向缩略图卡片", source.contains("TvSeriesEpisodeCard("))
        assertTrue("电视剧详情页操作区应接近参考图的主按钮 + 次按钮组合", source.contains("TvSeriesReferenceActionRow("))
        assertTrue("电视剧详情页应提供参考图中的次操作“我的片单”", source.contains("我的片单"))
        assertTrue("电视剧详情页应有参考图同类的 4K 角标", source.contains("\"4K\""))
        assertTrue("右侧标题文案必须是中文“剧集”", source.contains("\"剧集\""))
        assertTrue("主按钮文案必须使用中文播放集数", source.contains("播放第"))
        assertTrue("电视剧详情页大标题应使用正字距拉开以贴近参考图", source.contains("letterSpacing"))
        assertTrue("左侧信息区必须有独立宽度，避免被右侧剧集面板吞掉", source.contains("HeroInfoWidthDp"))
        assertTrue("右侧剧集面板宽度必须按 TV density 收窄，给左侧信息留出空间", source.contains("const val EpisodePaneWidthDp = 430"))
        assertTrue("左侧信息区仍应复用参考图式标题拆分规则", source.contains("splitTvSeriesReferenceTitle("))
        assertTrue("电视剧详情页标题必须使用左对齐列", source.contains("horizontalAlignment = Alignment.Start"))
        assertTrue("电视剧详情页标题文本必须左对齐", source.contains("textAlign = TextAlign.Start"))
        assertTrue("电视剧详情页标题应合并成单一展示标题，不能再渲染顶部“剧集”顶标", source.contains("displayTitle"))
        assertTrue("剧情摘要至少显示四行", source.contains("minLines = 4"))
        assertTrue("剧情摘要最多固定四行避免挤压操作区", source.contains("maxLines = 4"))
        assertFalse("左侧信息区不能再渲染标题上方的“剧集”顶标", source.contains("text = formatTvSeriesReferenceTitle(titleParts.eyebrow"))
        assertFalse("左侧信息区不能再包含 18+ 年龄角标组件", source.contains("TvSeriesAgeBadge("))
        assertFalse("左侧信息区不能再显示 18+ 年龄文案", source.contains("text = \"18+\""))
        assertTrue("左侧评分区必须包含参考图式 IMDb 角标", source.contains("\"IMDb\""))
        assertTrue("参考图还原必须降低标题字号，不再沿用旧大字号布局", source.contains("titleParts.stretched -> 44.sp"))
        assertTrue("参考图还原必须降低右侧分集卡高度", source.contains("const val EpisodeCardHeightDp = 84"))
        assertFalse("左侧信息区不能继续使用会溢出 TV 逻辑宽度的 560dp 固定宽度", source.contains("width(560.dp)"))
        assertFalse("本次不保留左侧竖向导航栏", source.contains("TvSeriesReferenceSideRail("))
        assertFalse("本次不保留顶部导航栏", source.contains("TvSeriesReferenceTopNav("))
        assertFalse("电视剧详情页不应继续保留旧版剧情简介独立区块", source.contains("key = \"tags\""))
        assertFalse("电视剧详情页不应继续保留旧版季选择独立区块", source.contains("key = \"seasons\""))
        assertFalse("电视剧详情页不应继续保留旧版选集网格独立区块", source.contains("key = \"episodes\""))
    }

    @Test
    fun `series detail title formatting keeps chinese title horizontal`() {
        assertEquals("主角", formatTvSeriesReferenceTitle("主角", stretched = false))
        assertEquals("剧集", formatTvSeriesReferenceTitle("剧集", stretched = false))
        assertEquals("T H E", formatTvSeriesReferenceTitle("THE", stretched = true))
    }
}
