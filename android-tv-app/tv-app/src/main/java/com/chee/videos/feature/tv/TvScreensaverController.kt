package com.chee.videos.feature.tv

import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.SideEffect
import androidx.compose.runtime.remember
import androidx.compose.runtime.staticCompositionLocalOf
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update

/** 所有写入均来自主线程；显示状态单独存放，不能反过来重启闲置协程。 */
@Singleton
class TvScreensaverController @Inject constructor() {
    private val _runtime = MutableStateFlow(TvScreensaverRuntime())
    val runtime = _runtime.asStateFlow()
    private val _showing = MutableStateFlow(false)
    val showing = _showing.asStateFlow()
    private val playbackOwners = mutableSetOf<Any>()

    fun userActivity() {
        _showing.value = false
        _runtime.update { it.copy(interactionSequence = it.interactionSequence + 1) }
    }

    fun setSession(sessionKey: String?) {
        if (_runtime.value.sessionKey == sessionKey) return
        _showing.value = false
        _runtime.update { it.copy(sessionKey = sessionKey, interactionSequence = it.interactionSequence + 1) }
    }

    fun setForeground(foreground: Boolean) {
        if (_runtime.value.foreground == foreground) return
        _showing.value = false
        _runtime.update { it.copy(foreground = foreground, interactionSequence = it.interactionSequence + 1) }
    }

    internal fun setShowing(showing: Boolean) {
        _showing.value = showing
    }

    internal fun setPlaybackBusy(owner: Any, busy: Boolean) {
        if (busy) playbackOwners.add(owner) else playbackOwners.remove(owner)
        val blocked = playbackOwners.isNotEmpty()
        if (blocked) _showing.value = false
        if (_runtime.value.playbackBusy == blocked) return
        _runtime.update { it.copy(playbackBusy = blocked, interactionSequence = it.interactionSequence + 1) }
    }
}

internal val LocalTvScreensaverController = staticCompositionLocalOf<TvScreensaverController?> { null }

@Composable
internal fun TvScreensaverPlaybackEffect(busy: Boolean) {
    val controller = LocalTvScreensaverController.current ?: return
    val owner = remember { Any() }
    SideEffect { controller.setPlaybackBusy(owner, busy) }
    DisposableEffect(controller, owner) {
        onDispose { controller.setPlaybackBusy(owner, false) }
    }
}

/** 记录被用于唤醒的完整输入序列，抬起及长按重复不能穿透到底层页面。 */
internal class TvScreensaverInputGate {
    private val consumedKeys = mutableSetOf<Int>()
    private var consumingTouch = false

    fun key(keyCode: Int, down: Boolean, showing: Boolean, systemKey: Boolean): Boolean {
        if (systemKey) return false
        if (keyCode in consumedKeys) {
            if (!down) consumedKeys.remove(keyCode)
            return true
        }
        if (!showing) return false
        if (down) consumedKeys.add(keyCode)
        return true
    }

    fun touch(start: Boolean, end: Boolean, showing: Boolean): Boolean {
        if (start) consumingTouch = showing
        if (showing) consumingTouch = true
        val consumed = consumingTouch
        if (end) consumingTouch = false
        return consumed
    }

    fun reset() {
        consumedKeys.clear()
        consumingTouch = false
    }
}
