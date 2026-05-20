package com.chee.videos.tv

import android.graphics.Color as AndroidColor
import android.os.Bundle
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
}

internal fun shouldSwallowTvComposeHoverExitCrash(err: IllegalStateException): Boolean {
    return err.message == "The ACTION_HOVER_EXIT event was not cleared." &&
        err.stackTrace.any { frame ->
            frame.className == "androidx.compose.ui.platform.AndroidComposeView" &&
                (frame.methodName == "sendHoverExitEvent" || frame.methodName == "dispatchHoverEvent")
        }
}
