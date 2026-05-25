package com.chee.videos.core.player

import android.content.Context
import org.videolan.libvlc.LibVLC

object TvVlcLibrary {
    @Volatile
    private var instance: LibVLC? = null

    fun shared(context: Context): LibVLC {
        return instance ?: synchronized(this) {
            instance ?: LibVLC(context.applicationContext, ArrayList(buildTvSharedVlcArgs())).also {
                instance = it
            }
        }
    }
}

internal fun buildTvSharedVlcArgs(): List<String> = listOf(
    "--no-drop-late-frames",
    "--no-skip-frames",
    "--rtsp-tcp",
)
