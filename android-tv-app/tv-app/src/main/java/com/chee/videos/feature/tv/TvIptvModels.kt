package com.chee.videos.feature.tv

import com.chee.videos.core.model.TvIptvChannelDto

data class TvIptvChannelUiModel(
    val id: String,
    val name: String,
    val url: String,
    val group: String,
    val logoUrl: String? = null,
    val tvgId: String? = null,
    val sortOrder: Int = 0,
)

data class TvIptvChannelGroupUiModel(
    val group: String,
    val channels: List<TvIptvChannelUiModel>,
)

enum class TvIptvRemoteKey {
    Up,
    Down,
    Right,
    Back,
    Ok,
}

enum class TvIptvRemoteAction {
    PreviousChannel,
    NextChannel,
    ShowChannelList,
    ExitPage,
    MoveListFocusUp,
    MoveListFocusDown,
    SelectFocusedChannel,
    CloseChannelList,
}

fun tvIptvChannelToUiModel(dto: TvIptvChannelDto): TvIptvChannelUiModel =
    TvIptvChannelUiModel(
        id = dto.id,
        name = dto.name,
        url = dto.url,
        group = dto.group,
        logoUrl = dto.logoUrl,
        tvgId = dto.tvgId,
        sortOrder = dto.sortOrder,
    )

fun resolveDefaultIptvChannel(channels: List<TvIptvChannelUiModel>): TvIptvChannelUiModel? =
    channels.firstOrNull { it.url.isNotBlank() }

fun resolveIptvChannelAfterStep(
    channels: List<TvIptvChannelUiModel>,
    currentChannelId: String?,
    step: Int,
): TvIptvChannelUiModel? {
    if (channels.isEmpty()) {
        return null
    }
    val currentIndex = channels.indexOfFirst { it.id == currentChannelId }.takeIf { it >= 0 } ?: 0
    val nextIndex = Math.floorMod(currentIndex + step, channels.size)
    return channels[nextIndex]
}

fun groupIptvChannels(channels: List<TvIptvChannelUiModel>): List<TvIptvChannelGroupUiModel> {
    val grouped = linkedMapOf<String, MutableList<TvIptvChannelUiModel>>()
    channels.forEach { channel ->
        val group = channel.group.trim().ifBlank { "未分组" }
        grouped.getOrPut(group) { mutableListOf() } += channel
    }
    return grouped.map { (group, items) -> TvIptvChannelGroupUiModel(group = group, channels = items) }
}

fun resolveIptvPlaybackRemoteAction(key: TvIptvRemoteKey): TvIptvRemoteAction? =
    when (key) {
        TvIptvRemoteKey.Up -> TvIptvRemoteAction.PreviousChannel
        TvIptvRemoteKey.Down -> TvIptvRemoteAction.NextChannel
        TvIptvRemoteKey.Right -> TvIptvRemoteAction.ShowChannelList
        TvIptvRemoteKey.Back -> TvIptvRemoteAction.ExitPage
        TvIptvRemoteKey.Ok -> null
    }

fun resolveIptvChannelListRemoteAction(key: TvIptvRemoteKey): TvIptvRemoteAction? =
    when (key) {
        TvIptvRemoteKey.Up -> TvIptvRemoteAction.MoveListFocusUp
        TvIptvRemoteKey.Down -> TvIptvRemoteAction.MoveListFocusDown
        TvIptvRemoteKey.Ok -> TvIptvRemoteAction.SelectFocusedChannel
        TvIptvRemoteKey.Back -> TvIptvRemoteAction.CloseChannelList
        TvIptvRemoteKey.Right -> null
    }
