package com.chee.videos.core.ui

import android.app.Activity
import android.content.Context
import android.content.ContextWrapper
import android.view.WindowManager
import android.view.Window
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.remember
import androidx.compose.ui.platform.LocalContext
import java.util.WeakHashMap

private val screenOnOwners = WeakHashMap<Window, MutableSet<Any>>()

@Composable
fun KeepScreenOnEffect(enabled: Boolean) {
    val activity = LocalContext.current.findActivity()
    val owner = remember { Any() }
    DisposableEffect(activity, enabled) {
        val window = activity?.window
        if (window != null && enabled) {
            screenOnOwners.getOrPut(window) { mutableSetOf() }.add(owner)
            window.addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
        }
        onDispose {
            if (window != null && enabled) {
                val owners = screenOnOwners[window]
                owners?.remove(owner)
                if (owners.isNullOrEmpty()) {
                    screenOnOwners.remove(window)
                    window.clearFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
                }
            }
        }
    }
}

private tailrec fun Context.findActivity(): Activity? {
    return when (this) {
        is Activity -> this
        is ContextWrapper -> baseContext.findActivity()
        else -> null
    }
}
