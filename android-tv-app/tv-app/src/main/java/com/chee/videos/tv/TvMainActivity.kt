package com.chee.videos.tv

import android.graphics.Color as AndroidColor
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.util.Log
import android.view.MotionEvent
import androidx.activity.ComponentActivity
import androidx.activity.SystemBarStyle
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import dagger.hilt.android.AndroidEntryPoint

@AndroidEntryPoint
class TvMainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
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
                    if (err is IllegalStateException && shouldSwallowTvComposeHoverExitCrash(err)) {
                        Log.w(TAG, "swallowed Compose ACTION_HOVER_EXIT crash on main looper", err)
                        continue
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
