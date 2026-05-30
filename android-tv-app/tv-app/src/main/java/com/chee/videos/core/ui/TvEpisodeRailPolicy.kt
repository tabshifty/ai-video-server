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

internal fun resolveEpisodeRailFollowScrollFirstVisibleItemIndex(
    items: List<TvEpisodeRailItem>,
    focusedEpisodeId: String?,
    firstVisibleItemIndex: Int,
    lastVisibleItemIndex: Int,
    edgePaddingItemCount: Int = 1,
): Int? {
    if (items.isEmpty() || firstVisibleItemIndex < 0 || lastVisibleItemIndex < firstVisibleItemIndex) {
        return null
    }
    val focusedIndex = resolveEpisodeRailCurrentIndex(items, focusedEpisodeId)
    val safeFirstIndex = firstVisibleItemIndex + edgePaddingItemCount
    val safeLastIndex = lastVisibleItemIndex - edgePaddingItemCount
    val visibleItemCount = lastVisibleItemIndex - firstVisibleItemIndex + 1
    return when {
        focusedIndex < safeFirstIndex -> (focusedIndex - edgePaddingItemCount).coerceAtLeast(0)
        focusedIndex > safeLastIndex -> {
            val trailingVisibleItemCount = (visibleItemCount - edgePaddingItemCount - 1).coerceAtLeast(0)
            (focusedIndex - trailingVisibleItemCount).coerceAtLeast(0)
        }
        else -> null
    }
}

internal fun formatTvEpisodeRailLabel(number: Int): String {
    val normalizedNumber = number.coerceAtLeast(1)
    return "第${formatChinesePositiveInt(normalizedNumber)}集"
}

private fun formatChinesePositiveInt(number: Int): String {
    val digits = arrayOf("零", "一", "二", "三", "四", "五", "六", "七", "八", "九")
    val units = arrayOf("", "十", "百", "千")
    if (number <= 0) {
        return digits[0]
    }
    var remaining = number
    var unitIndex = 0
    var needsZero = false
    val builder = StringBuilder()
    while (remaining > 0) {
        val digit = remaining % 10
        if (digit == 0) {
            if (builder.isNotEmpty()) {
                needsZero = true
            }
        } else {
            if (needsZero) {
                builder.insert(0, digits[0])
                needsZero = false
            }
            builder.insert(0, units[unitIndex])
            builder.insert(0, digits[digit])
        }
        remaining /= 10
        unitIndex += 1
    }
    return builder.toString()
        .let { value ->
            if (value.startsWith("一十")) {
                "十${value.removePrefix("一十")}"
            } else {
                value
            }
        }
        .ifBlank { digits[0] }
}
