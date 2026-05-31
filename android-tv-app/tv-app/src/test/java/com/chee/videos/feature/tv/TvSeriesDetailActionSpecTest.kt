package com.chee.videos.feature.tv

import java.nio.file.Path
import kotlin.io.path.exists
import kotlin.io.path.readText
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
        assertTrue("电视剧详情页播放、季选择和集选择仍应使用共享焦点视觉", source.contains(".tvFocusableGlow("))
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
        assertFalse("电视剧详情页不应继续保留旧版剧情简介独立区块", source.contains("key = \"tags\""))
        assertFalse("电视剧详情页不应继续保留旧版季选择独立区块", source.contains("key = \"seasons\""))
        assertFalse("电视剧详情页不应继续保留旧版选集网格独立区块", source.contains("key = \"episodes\""))
    }
}
