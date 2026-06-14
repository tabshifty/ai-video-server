package com.chee.videos.tv

import android.graphics.Color as AndroidColor
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.util.Log
import android.view.KeyEvent
import android.view.MotionEvent
import androidx.activity.ComponentActivity
import androidx.activity.SystemBarStyle
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import com.chee.videos.tv.R
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class TvMainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        setTheme(R.style.Theme_VideoHome)
        super.onCreate(savedInstanceState)
        installMainLooperHoverExitGuard()
        enableEdgeToEdge(
            statusBarStyle = SystemBarStyle.dark(AndroidColor.TRANSPARENT),
            navigationBarStyle = SystemBarStyle.dark(AndroidColor.TRANSPARENT),
        )
        setContent {
            TvShellApp()
        }
    }

    override fun dispatchKeyEvent(event: KeyEvent): Boolean {
        return try {
            super.dispatchKeyEvent(event)
        } catch (err: IllegalStateException) {
            if (shouldSwallowTvComposeFocusRequesterCrash(err)) {
                Log.w(TAG, "swallowed FocusRequester crash in dispatchKeyEvent — prevents input-dispatch ANR", err)
                false
            } else {
                throw err
            }
        }
    }

    override fun dispatchGenericMotionEvent(event: MotionEvent): Boolean {
        return try {
            super.dispatchGenericMotionEvent(event)
        } catch (err: IllegalStateException) {
            if (shouldSwallowTvComposeHoverExitCrash(err)) {
                true
            } else {
                throw err
            }
        }
    }

    private fun installMainLooperHoverExitGuard() {
        if (mainLooperGuardInstalled) return
        mainLooperGuardInstalled = true
        Handler(Looper.getMainLooper()).post {
            while (true) {
                try {
                    Looper.loop()
                    return@post
                } catch (err: Throwable) {
                    if (err is IllegalStateException) {
                        if (shouldSwallowTvComposeHoverExitCrash(err)) {
                            Log.w(TAG, "swallowed Compose ACTION_HOVER_EXIT crash on main looper", err)
                            continue
                        }
                        if (shouldSwallowTvComposeFocusRequesterCrash(err)) {
                            Log.w(TAG, "swallowed Compose FocusRequester-not-initialized crash on main looper", err)
                            continue
                        }
                    }
                    throw err
                }
            }
        }
    }

    companion object {
        private const val TAG = "TvMainActivity"

        @Volatile
        private var mainLooperGuardInstalled = false
    }
}

internal fun shouldSwallowTvComposeHoverExitCrash(err: IllegalStateException): Boolean {
    if (err.message != "The ACTION_HOVER_EXIT event was not cleared.") return false
    return err.stackTrace.any { frame ->
        frame.className == "androidx.compose.ui.platform.AndroidComposeView" &&
            (
                frame.methodName == "sendHoverExitEvent" ||
                    frame.methodName.startsWith("sendHoverExitEvent\$lambda") ||
                    frame.methodName == "dispatchHoverEvent" ||
                    frame.methodName.startsWith("dispatchHoverEvent\$lambda")
            )
    }
}

internal fun shouldSwallowTvComposeFocusRequesterCrash(err: IllegalStateException): Boolean {
    val message = err.message ?: return false
    if (!message.contains("FocusRequester is not initialized")) return false
    // 消息匹配是主安全网；栈帧只用来确认异常来自 Compose focus 子系统、不是业务代码自己抛的同样消息。
    // 不再绑定到具体 FocusRequester 方法名（findFocusTargetNode$ui_release / focus$ui_release / requestFocus 等），
    // 因为 Compose 1.7+ 在异步路径（FocusOwnerImpl.focusSearch-ULY8qGw、AndroidComposeView$keyInputModifier$1.invoke-ZmokQxo
    // 等）抛出同样异常时不一定保留 FocusRequester 自身栈帧；放宽到 androidx.compose.ui.focus.* 包前缀让兜底更鲁棒。
    return err.stackTrace.any { frame ->
        frame.className.startsWith("androidx.compose.ui.focus.")
    }
}
