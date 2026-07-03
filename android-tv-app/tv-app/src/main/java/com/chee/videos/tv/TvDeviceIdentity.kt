package com.chee.videos.tv

import android.os.Build

fun buildTvDeviceId(): String =
    "android-tv-${Build.DEVICE}-${Build.MODEL}".lowercase()

fun buildTvDeviceName(): String =
    listOf(Build.MANUFACTURER, Build.MODEL).joinToString(" ").trim()
