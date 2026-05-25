package com.chee.videos.core.ui

/**
 * 聚合 TV 长视频播放器内可能持焦的叠加层可见性。任意 overlay 从 true 跃迁到 false 且当前没有其他
 * overlay 仍在显示时，需要把焦点回收到 player 根 Box，避免 [[TV 长视频焦点真空]]。
 *
 * `controlsVisible` 不算 overlay（它是播放器自身的 controls 区域），但它的 true→false 跃迁同样
 * 需要触发兜底，因为控制条 auto-hide 之后焦点可能停在已经 dispose 的 button 上。
 */
internal data class PlayerFocusGuardInput(
    val controlsVisible: Boolean,
    val subtitleSheetVisible: Boolean,
    val audioTrackSheetVisible: Boolean,
    val resumePromptVisible: Boolean,
    val backConfirmPromptVisible: Boolean,
    val playerErrorVisible: Boolean,
) {
    fun anyOverlayVisible(): Boolean =
        subtitleSheetVisible ||
            audioTrackSheetVisible ||
            resumePromptVisible ||
            backConfirmPromptVisible ||
            playerErrorVisible
}

/**
 * 判断从 [previous] 到 [current] 之间，是否需要把焦点回收到 player 根 Box。
 *
 * 触发条件：任一可见性字段从 true 跃迁到 false，且 [current] 下没有其他 overlay 仍在显示。
 */
internal fun shouldReclaimRootFocus(
    previous: PlayerFocusGuardInput,
    current: PlayerFocusGuardInput,
): Boolean {
    val transitionedFalse =
        (previous.subtitleSheetVisible && !current.subtitleSheetVisible) ||
            (previous.audioTrackSheetVisible && !current.audioTrackSheetVisible) ||
            (previous.resumePromptVisible && !current.resumePromptVisible) ||
            (previous.backConfirmPromptVisible && !current.backConfirmPromptVisible) ||
            (previous.playerErrorVisible && !current.playerErrorVisible) ||
            (previous.controlsVisible && !current.controlsVisible)
    return transitionedFalse && !current.anyOverlayVisible()
}
