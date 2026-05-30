package com.chee.videos.core.ui

data class TvEpisodeRailItem(
    val id: String,
    val number: Int,
    val title: String,
    val playable: Boolean,
    val current: Boolean,
)

internal fun resolveEpisodeRailCurrentIndex(
    items: List<TvEpisodeRailItem>,
    currentEpisodeId: String?,
): Int =
    items.indexOfFirst { it.id == currentEpisodeId }.takeIf { it >= 0 } ?: 0

internal fun resolveEpisodeRailInitialFirstVisibleItemIndex(
    items: List<TvEpisodeRailItem>,
    currentEpisodeId: String?,
    leadingItemCount: Int = 3,
): Int {
    if (items.isEmpty()) {
        return 0
    }
    val currentIndex = resolveEpisodeRailCurrentIndex(items, currentEpisodeId)
    return (currentIndex - leadingItemCount).coerceAtLeast(0)
}
