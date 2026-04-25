package com.chee.videos.core.player

import android.media.MediaCodecList
import android.os.Build
import javax.inject.Inject
import javax.inject.Singleton

enum class PlaybackProfile(val wireValue: String) {
    PRIMARY("primary"),
    COMPAT("compat"),
}

internal fun resolvePreferredLongFormPlaybackProfile(
    deviceSupportsHevc: Boolean,
    isProbablyEmulator: Boolean,
): PlaybackProfile {
    if (!deviceSupportsHevc || isProbablyEmulator) {
        return PlaybackProfile.COMPAT
    }
    return PlaybackProfile.PRIMARY
}

internal fun isProbablyEmulatorDevice(): Boolean {
    val fingerprint = Build.FINGERPRINT.orEmpty()
    val model = Build.MODEL.orEmpty()
    val manufacturer = Build.MANUFACTURER.orEmpty()
    val brand = Build.BRAND.orEmpty()
    val device = Build.DEVICE.orEmpty()
    val product = Build.PRODUCT.orEmpty()
    return listOf(fingerprint, model, manufacturer, brand, device, product)
        .joinToString("|")
        .lowercase()
        .let { normalized ->
            normalized.contains("generic") ||
                normalized.contains("emulator") ||
                normalized.contains("sdk_gphone") ||
                normalized.contains("goldfish") ||
                normalized.contains("ranchu")
        }
}

internal fun deviceSupportsHevcPlayback(): Boolean {
    return runCatching {
        MediaCodecList(MediaCodecList.ALL_CODECS).codecInfos.any { info ->
            !info.isEncoder && info.supportedTypes.any { type ->
                type.equals("video/hevc", ignoreCase = true)
            }
        }
    }.getOrDefault(false)
}

@Singleton
class PlaybackProfileResolver @Inject constructor() {
    fun preferredLongFormProfile(): PlaybackProfile {
        return resolvePreferredLongFormPlaybackProfile(
            deviceSupportsHevc = deviceSupportsHevcPlayback(),
            isProbablyEmulator = isProbablyEmulatorDevice(),
        )
    }
}
