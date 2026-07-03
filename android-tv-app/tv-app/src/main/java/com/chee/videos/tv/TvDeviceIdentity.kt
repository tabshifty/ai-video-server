package com.chee.videos.tv

import android.content.Context
import android.os.Build
import android.provider.Settings
import com.chee.videos.core.data.AppPreferencesStore
import dagger.hilt.android.qualifiers.ApplicationContext
import java.util.UUID
import javax.inject.Inject
import javax.inject.Singleton

internal fun normalizeStableTvDeviceId(rawId: String): String =
    "android-tv-${rawId.trim()}".lowercase()

internal fun buildLegacyTvDeviceId(device: String, model: String): String =
    "android-tv-${device.trim()}-${model.trim()}".lowercase()

internal fun buildTvDeviceIdFromParts(
    androidId: String?,
    fallbackId: String?,
): String {
    val normalizedAndroidId = androidId?.trim().orEmpty()
    if (normalizedAndroidId.isNotBlank()) {
        return normalizeStableTvDeviceId(normalizedAndroidId)
    }
    val normalizedFallbackId = fallbackId?.trim().orEmpty()
    if (normalizedFallbackId.isNotBlank()) {
        return normalizeStableTvDeviceId(normalizedFallbackId)
    }
    return normalizeStableTvDeviceId(UUID.randomUUID().toString())
}

@Singleton
class TvDeviceIdentityProvider @Inject constructor(
    @ApplicationContext private val appContext: Context,
    private val store: AppPreferencesStore,
) {
    suspend fun currentDeviceId(): String {
        val androidId = Settings.Secure.getString(appContext.contentResolver, Settings.Secure.ANDROID_ID)
        val storedFallbackId = store.readTvDeviceFallbackId()
        val resolved = buildTvDeviceIdFromParts(
            androidId = androidId,
            fallbackId = storedFallbackId,
        )
        if (androidId.isNullOrBlank()) {
            store.saveTvDeviceFallbackId(resolved.removePrefix("android-tv-"))
        }
        return resolved
    }

    fun legacyDeviceId(): String =
        buildLegacyTvDeviceId(Build.DEVICE, Build.MODEL)
}

fun buildTvDeviceName(): String =
    listOf(Build.MANUFACTURER, Build.MODEL).joinToString(" ").trim()
