package com.chee.videos.core.ui

internal data class HomeContentTabSpec(
    val title: String,
    val type: String,
)

internal data class RootNavigationTabSpec(
    val route: String,
    val label: String,
)

internal val homeContentTabs = listOf(
    HomeContentTabSpec(title = "短视频", type = "short"),
    HomeContentTabSpec(title = "电影", type = "movie"),
    HomeContentTabSpec(title = "电视剧", type = "episode"),
    HomeContentTabSpec(title = "AV", type = "av"),
)

internal val rootNavigationTabs = listOf(
    RootNavigationTabSpec(route = "home", label = "首页"),
    RootNavigationTabSpec(route = "mine", label = "我的"),
)
