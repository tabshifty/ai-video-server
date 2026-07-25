package com.chee.videos.core.ui.cast

import com.chee.videos.core.model.TvDeviceDto
import com.chee.videos.core.model.TvRemoteSearchContextRequest
import com.chee.videos.core.model.TvRemoteSessionItemDto
import com.chee.videos.core.model.VideoListItemDto

/**
 * 投屏入口与设备选择共享逻辑。搜索结果页与短视频合集页各自提供 items 与 searchContext，
 * 共用同一套设备排序、选择回退、偏好持久化与起始位置解析。见 ADR-0016 / ADR-0017。
 */

private const val ShortRemoteDefaultPageSize = 24

/** 把短视频列表转成投屏会话条目，过滤空 id 并规整标题/缩略图/类型。 */
fun buildShortTvRemoteItems(items: List<VideoListItemDto>): List<TvRemoteSessionItemDto> {
    return items
        .filter { it.id.isNotBlank() }
        .map { item ->
            TvRemoteSessionItemDto(
                videoId = item.id,
                title = item.title.trim(),
                thumbnailPath = item.thumbnailPath?.trim(),
                duration = item.duration,
                type = item.type.trim().ifBlank { "short" },
            )
        }
}

/** 解析投屏起始位置：以当前播放视频 id 在列表中的下标为准，找不到返回 null。 */
fun resolveShortTvRemoteStartIndex(items: List<VideoListItemDto>, playingVideoId: String?): Int? {
    val target = playingVideoId?.trim().orEmpty()
    if (target.isBlank()) {
        return null
    }
    val index = items.indexOfFirst { it.id == target }
    return index.takeIf { it >= 0 }
}

/** 构造关键词搜索上下文；query 为空时返回 null（合集上下文改用 [buildCollectionShortTvRemoteSearchContext]）。 */
fun buildKeywordShortTvRemoteSearchContext(
    activeQuery: String,
    page: Int,
    totalCount: Int,
    pageSize: Int = ShortRemoteDefaultPageSize,
): TvRemoteSearchContextRequest? {
    val query = activeQuery.trim()
    if (query.isBlank()) {
        return null
    }
    return TvRemoteSearchContextRequest(
        query = query,
        type = "short",
        page = page.coerceAtLeast(1),
        pageSize = pageSize.coerceAtLeast(1),
        totalCount = totalCount.coerceAtLeast(0),
    )
}

/**
 * 构造短视频合集投屏上下文：query 为空、type=short、带上 collection_id，让 TV 端按合集补页。
 * 见 ADR-0016（合集补页走 SearchVideos 的 collection 维度）。
 */
fun buildCollectionShortTvRemoteSearchContext(
    collectionId: String,
    page: Int,
    totalCount: Int,
    pageSize: Int = ShortRemoteDefaultPageSize,
): TvRemoteSearchContextRequest? {
    val id = collectionId.trim()
    if (id.isBlank()) {
        return null
    }
    return TvRemoteSearchContextRequest(
        query = "",
        type = "short",
        page = page.coerceAtLeast(1),
        pageSize = pageSize.coerceAtLeast(1),
        totalCount = totalCount.coerceAtLeast(0),
        collectionId = id,
    )
}

fun sortTvDevicesForSelection(items: List<TvDeviceDto>, preferredDeviceId: String?): List<TvDeviceDto> {
    val preferred = preferredDeviceId?.trim().orEmpty()
    return items.sortedWith(
        compareByDescending<TvDeviceDto> { it.isOnline }
            .thenByDescending { it.deviceId == preferred }
            .thenByDescending { it.lastSeenAt.orEmpty() }
            .thenByDescending { it.lastAuthorizedAt.orEmpty() }
            .thenBy { it.deviceName },
    )
}

fun resolveTvDeviceSelection(
    items: List<TvDeviceDto>,
    currentSelectedDeviceId: String?,
    preferredDeviceId: String?,
): String? {
    val currentSelected = currentSelectedDeviceId?.trim().orEmpty()
    if (currentSelected.isNotBlank() && items.any { it.deviceId == currentSelected }) {
        return currentSelected
    }
    val preferred = preferredDeviceId?.trim().orEmpty()
    val onlinePreferred = items.firstOrNull { it.deviceId == preferred && it.isOnline }?.deviceId
    if (onlinePreferred != null) {
        return onlinePreferred
    }
    val firstOnline = items.firstOrNull { it.isOnline }?.deviceId
    if (firstOnline != null) {
        return firstOnline
    }
    return items.firstOrNull { it.deviceId == preferred }?.deviceId
        ?: items.firstOrNull()?.deviceId
}

fun pickTvDeviceForLaunch(items: List<TvDeviceDto>, selectedDeviceId: String): TvDeviceDto? {
    val selected = selectedDeviceId.trim()
    return items.firstOrNull { it.deviceId == selected }
        ?: items.firstOrNull()
}

fun shouldPersistTvDevicePreference(device: TvDeviceDto?): Boolean {
    return device?.isOnline == true
}

fun resolveTvDeviceRequestTarget(
    items: List<TvDeviceDto>,
    selectedDeviceId: String,
): String? {
    return pickTvDeviceForLaunch(items, selectedDeviceId)
        ?.takeIf(::shouldPersistTvDevicePreference)
        ?.deviceId
        ?: pickTvDeviceForLaunch(items, selectedDeviceId)?.deviceId
}
