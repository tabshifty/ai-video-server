package com.chee.videos.core.ui

internal data class HomeContentTabSpec(
    val title: String,
    val type: String,
)

internal data class RootNavigationTabSpec(
    val route: String,
    val label: String,
)

internal enum class AppNavigationTransitionDirection {
    Forward,
    Backward,
}

internal data class AppNavigationTransitionSpec(
    val durationMillis: Int,
    val fadeStartAlpha: Float,
)

internal val homeContentTabs = listOf(
    HomeContentTabSpec(title = "短视频", type = "short"),
    HomeContentTabSpec(title = "电影", type = "movie"),
    HomeContentTabSpec(title = "电视剧", type = "episode"),
    HomeContentTabSpec(title = "AV", type = "av"),
)

internal val rootNavigationTabs = listOf(
    RootNavigationTabSpec(route = "home", label = "首页"),
    RootNavigationTabSpec(route = "search", label = "搜索"),
    RootNavigationTabSpec(route = "image-collections", label = "图集"),
    RootNavigationTabSpec(route = "mine", label = "我的"),
)

internal fun appNavigationTransitionSpec(): AppNavigationTransitionSpec {
    return AppNavigationTransitionSpec(
        durationMillis = 260,
        fadeStartAlpha = 0.18f,
    )
}

internal fun appNavigationTransitionDirection(
    fromRoute: String?,
    toRoute: String?,
    isPop: Boolean,
): AppNavigationTransitionDirection {
    if (isPop) {
        return AppNavigationTransitionDirection.Backward
    }

    val fromRootIndex = rootNavigationRouteIndex(fromRoute)
    val toRootIndex = rootNavigationRouteIndex(toRoute)
    if (fromRootIndex != null && toRootIndex != null && fromRootIndex != toRootIndex) {
        return if (toRootIndex > fromRootIndex) {
            AppNavigationTransitionDirection.Forward
        } else {
            AppNavigationTransitionDirection.Backward
        }
    }

    return AppNavigationTransitionDirection.Forward
}

private fun rootNavigationRouteIndex(route: String?): Int? {
    return rootNavigationTabs.indexOfFirst { it.route == route }.takeIf { it >= 0 }
}
