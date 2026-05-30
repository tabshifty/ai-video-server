package com.chee.videos.feature.tv

internal const val TV_IPTV_NETWORK_CACHING_MS = 4_000

internal fun buildIptvLibVlcArgs(): List<String> = listOf(
    "--network-caching=$TV_IPTV_NETWORK_CACHING_MS",
)

internal fun buildIptvMediaOptions(): List<String> = listOf(
    ":network-caching=$TV_IPTV_NETWORK_CACHING_MS",
)

