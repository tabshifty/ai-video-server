package com.chee.videos.feature.tv

object TvMockData {
    private val series: List<TvSeriesUiModel> = listOf(
        tvSeries(
            id = "chronicle-city",
            title = "迷雾城档案",
            subtitle = "悬疑 · 犯罪",
            ratingText = "9.1",
            updateText = "更新至 S2·E6",
            description = "一座港口城市接连发生离奇案件，法医、刑警与记者三线交织，逐步揭开二十年前失踪案的真相。",
            tags = listOf("高能反转", "多线叙事", "口碑榜 Top3"),
            cast = listOf("林舟", "周岚", "谢言"),
            seasonCount = 2,
            episodesPerSeason = 8,
            posterSeed = 1,
        ),
        tvSeries(
            id = "winter-atelier",
            title = "冬日制片厂",
            subtitle = "都市 · 职场",
            ratingText = "8.8",
            updateText = "全 24 集",
            description = "年轻导演与纪录片团队在冬季北境追拍真实故事，在行业压力和理想之间寻找平衡。",
            tags = listOf("群像成长", "摄影质感", "年度热议"),
            cast = listOf("宋青", "陆临", "沈雾"),
            seasonCount = 1,
            episodesPerSeason = 24,
            posterSeed = 2,
        ),
        tvSeries(
            id = "silent-orbit",
            title = "静默轨道",
            subtitle = "科幻 · 冒险",
            ratingText = "9.0",
            updateText = "更新至 S1·E10",
            description = "人类首个深空殖民站失联，救援队在残骸与日志中拼凑真相，逐步发现更大的系统性阴谋。",
            tags = listOf("宇宙探索", "硬核设定", "视觉大片"),
            cast = listOf("陈牧", "伊娜", "塔伦"),
            seasonCount = 1,
            episodesPerSeason = 12,
            posterSeed = 3,
        ),
        tvSeries(
            id = "echo-valley",
            title = "回声山谷",
            subtitle = "剧情 · 家庭",
            ratingText = "8.6",
            updateText = "全 16 集",
            description = "三代人回到故乡重建旧宅，隐藏多年的家族秘密在一次山洪后被逐层揭开。",
            tags = listOf("情感细腻", "治愈向", "高分剧情"),
            cast = listOf("夏清", "徐南", "贺川"),
            seasonCount = 1,
            episodesPerSeason = 16,
            posterSeed = 4,
        ),
        tvSeries(
            id = "night-grid",
            title = "夜网协议",
            subtitle = "动作 · 谍战",
            ratingText = "8.9",
            updateText = "更新至 S3·E4",
            description = "情报组织“夜网”在全球城市间展开隐秘交锋，一名双面特工被迫在忠诚与生存间做选择。",
            tags = listOf("高密度动作", "快节奏", "追剧首选"),
            cast = listOf("顾骁", "安洁", "井上遥"),
            seasonCount = 3,
            episodesPerSeason = 10,
            posterSeed = 5,
        ),
    )

    fun allSeries(): List<TvSeriesUiModel> = series

    fun findSeries(seriesId: String): TvSeriesUiModel? = series.firstOrNull { it.id == seriesId }

    fun continueWatching(): TvContinueWatchingUiModel {
        val focus = series.first()
        return TvContinueWatchingUiModel(
            seriesId = focus.id,
            seriesTitle = focus.title,
            seasonNumber = 2,
            episodeNumber = 4,
            episodeTitle = "第4集 暗线浮现",
            progressPercent = 64,
        )
    }

    fun catalogSections(): List<TvCatalogSectionUiModel> {
        val hot = series.take(4)
        val updates = listOf(series[0], series[2], series[4], series[1])
        val classics = listOf(series[3], series[1], series[0], series[2])
        return listOf(
            TvCatalogSectionUiModel(
                title = "热播精选",
                subtitle = "今晚值得连看",
                items = hot,
            ),
            TvCatalogSectionUiModel(
                title = "今日更新",
                subtitle = "新剧情已上线",
                items = updates,
            ),
            TvCatalogSectionUiModel(
                title = "口碑高分",
                subtitle = "评分 8.6+",
                items = classics,
            ),
        )
    }
}

private fun tvSeries(
    id: String,
    title: String,
    subtitle: String,
    ratingText: String,
    updateText: String,
    description: String,
    tags: List<String>,
    cast: List<String>,
    seasonCount: Int,
    episodesPerSeason: Int,
    posterSeed: Int,
): TvSeriesUiModel {
    val seasons = (1..seasonCount).map { season ->
        TvSeasonUiModel(
            id = "$id-s$season",
            number = season,
            title = "第 $season 季",
            overview = "第 $season 季剧情概览",
            episodes = (1..episodesPerSeason).map { episode ->
                TvEpisodeUiModel(
                    id = "$id-s$season-e$episode",
                    number = episode,
                    title = "第${episode}集 ${episodeTitle(episode)}",
                    durationLabel = "${42 + (episode % 6)} 分钟",
                    summary = "围绕关键角色展开的主线推进与情绪爆发，保留多处伏笔。",
                    progressPercent = if (season == 1 && episode == 1) 100 else if (season == 1 && episode <= 3) 48 else 0,
                    videoId = "mock-$id-s$season-e$episode",
                    videoStatus = "ready",
                    playable = true,
                )
            },
        )
    }
    return TvSeriesUiModel(
        id = id,
        title = title,
        subtitle = subtitle,
        ratingText = ratingText,
        updateText = updateText,
        description = description,
        tags = tags,
        cast = cast,
        seasons = seasons,
        posterSeed = posterSeed,
    )
}

private fun episodeTitle(index: Int): String {
    return when (index % 6) {
        0 -> "风暴前夜"
        1 -> "旧案回响"
        2 -> "暗线浮现"
        3 -> "临界时刻"
        4 -> "误导证词"
        else -> "终局回溯"
    }
}
