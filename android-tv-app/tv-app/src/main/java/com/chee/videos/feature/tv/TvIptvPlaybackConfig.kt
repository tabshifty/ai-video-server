package com.chee.videos.feature.tv

import android.content.Context
import org.videolan.libvlc.LibVLC

internal const val TV_IPTV_NETWORK_CACHING_MS = 4_000

internal fun buildIptvLibVlcArgs(): List<String> = listOf(
    "--network-caching=$TV_IPTV_NETWORK_CACHING_MS",
)

internal fun buildIptvMediaOptions(): List<String> = listOf(
    ":network-caching=$TV_IPTV_NETWORK_CACHING_MS",
)

internal object TvIptvVlcLibrary {
    @Volatile
    private var instance: LibVLC? = null

    fun shared(context: Context): LibVLC {
        return instance ?: synchronized(this) {
            instance ?: LibVLC(context.applicationContext, ArrayList(buildIptvLibVlcArgs())).also {
                instance = it
            }
        }
    }
}
