@file:OptIn(ExperimentalFoundationApi::class)

package com.chee.videos.core.ui

import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.gestures.BringIntoViewSpec
import kotlin.math.abs

/**
 * TV 端 BringIntoView 行为覆盖。
 *
 * Compose Foundation 1.7 在 leanback 设备上把 LocalBringIntoViewSpec 默认设为
 * androidx.compose.foundation.gestures.PivotBringIntoViewSpec，
 * 该 spec 会把焦点目标硬拉到容器 30% 处——即使焦点目标已经完整可见。这会让
 * TV 首页 LazyColumn 在初始焦点落在 TvFeaturedHero 底部播放按钮时整体上滑约 110dp，
 * 表现为「横幅大海报头部被裁切」。详情页里 LazyRow（演员行）等子滚动同样会触发 pivot。
 *
 * 此 spec 复刻 BringIntoViewSpec.defaultCalculateScrollDistance 的「最少滚动」语义：
 * - 焦点目标完全在可视区内：返回 0，不滚动；
 * - 焦点目标比容器还大且当前已横跨容器：返回 0，不滚动；
 * - 其它情况：选择前缘或后缘中绝对距离更小的那一边，把焦点目标贴到容器边缘所需的最小位移。
 *
 * 通过在 TvShellApp 入口 CompositionLocalProvider(LocalBringIntoViewSpec provides ...)
 * 注入，作用域覆盖所有 TV 屏幕的 LazyColumn / LazyRow / verticalScroll 焦点滚动行为。
 */
val TvMinimalBringIntoViewSpec: BringIntoViewSpec = object : BringIntoViewSpec {
    override fun calculateScrollDistance(
        offset: Float,
        size: Float,
        containerSize: Float,
    ): Float = calculateTvMinimalBringIntoViewScrollDistance(offset, size, containerSize)
}

internal fun calculateTvMinimalBringIntoViewScrollDistance(
    offset: Float,
    size: Float,
    containerSize: Float,
): Float {
    val leadingEdge = offset
    val trailingEdge = offset + size
    if (leadingEdge >= 0f && trailingEdge <= containerSize) return 0f
    if (leadingEdge < 0f && trailingEdge > containerSize) return 0f
    val distanceToLeading = abs(leadingEdge)
    val distanceToTrailing = abs(trailingEdge - containerSize)
    return if (distanceToLeading < distanceToTrailing) {
        leadingEdge
    } else {
        trailingEdge - containerSize
    }
}
