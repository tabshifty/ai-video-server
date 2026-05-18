package com.chee.videos.feature.tv

import com.chee.videos.tv.TvAccountMenuAction

enum class TvHomeMenuItem(
    val label: String,
    val homeKind: String,
) {
    Series("电视剧", "tv"),
    Movie("电影", "movie"),
    Adult("18+", "av"),
    Iptv("IPTV", ""),
    Search("搜索", ""),
    Settings("设置", ""),
    ;

    val isContentKind: Boolean
        get() = this == Series || this == Movie || this == Adult

    companion object {
        fun defaults(): List<TvHomeMenuItem> = listOf(Series, Movie, Adult, Iptv, Search, Settings)
        fun defaultSelected(): TvHomeMenuItem = Series
    }
}

data class TvTypedHomeSectionUiModel(
    val role: String,
    val title: String,
    val items: List<TvHomeShelfItemUiModel>,
)

enum class TvHomeFocusArea {
    Menu,
    Content,
}

enum class TvHomeFocusDirection {
    Left,
    Right,
    Up,
    Down,
}

data class TvHomeFocusMove(
    val from: TvHomeFocusArea,
    val direction: TvHomeFocusDirection,
)

fun resolveTvHomeFocusMove(area: TvHomeFocusArea, direction: TvHomeFocusDirection): TvHomeFocusArea {
    return when {
        area == TvHomeFocusArea.Menu && direction == TvHomeFocusDirection.Right -> TvHomeFocusArea.Content
        area == TvHomeFocusArea.Content && direction == TvHomeFocusDirection.Left -> TvHomeFocusArea.Menu
        else -> area
    }
}

object TvHomeSettingsPanelAction {
    val menuItem: TvHomeMenuItem = TvHomeMenuItem.Settings

    fun defaults(): List<TvAccountMenuAction> = TvAccountMenuAction.defaults()
}

fun shouldShowTvHomeSideMenu(route: String?): Boolean = route == "tv-home"

fun buildTvTypedHomeSections(
    kind: String,
    featured: TvHomeShelfItemUiModel?,
    recentWatching: List<TvHomeShelfItemUiModel>,
    recentUpdates: List<TvHomeShelfItemUiModel>,
): List<TvTypedHomeSectionUiModel> {
    val normalizedKind = normalizeTvHomeKind(kind)
    return buildList {
        if (featured != null) {
            add(TvTypedHomeSectionUiModel(role = "featured", title = "巨幅推荐", items = listOf(featured)))
        }
        if (recentWatching.isNotEmpty()) {
            add(TvTypedHomeSectionUiModel(role = "recent_watching", title = "最近播放", items = recentWatching))
        }
        if (recentUpdates.isNotEmpty()) {
            add(TvTypedHomeSectionUiModel(role = "recent_updates", title = "最近更新", items = recentUpdates))
        }
        add(
            TvTypedHomeSectionUiModel(
                role = "all",
                title = when (normalizedKind) {
                    "movie" -> "全部电影"
                    "av" -> "全部18+"
                    else -> "全部电视剧"
                },
                items = emptyList(),
            ),
        )
    }
}

fun normalizeTvHomeKind(kind: String): String {
    return when (kind.trim().lowercase()) {
        "movie" -> "movie"
        "av" -> "av"
        else -> "tv"
    }
}
