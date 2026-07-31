package com.chee.videos.core.data

import androidx.datastore.core.DataStore
import androidx.datastore.preferences.core.Preferences
import androidx.datastore.preferences.core.booleanPreferencesKey
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import com.chee.videos.core.model.ServerEndpoint
import com.chee.videos.core.model.ShortPlaybackMode
import com.chee.videos.core.model.SessionTokens
import com.chee.videos.core.model.TvTrackPreference
import com.chee.videos.core.model.VideoFitMode
import com.chee.videos.core.util.UrlBuilder
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import javax.inject.Inject
import javax.inject.Singleton
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.flow.map

@Singleton
class AppPreferencesStore @Inject constructor(
    private val dataStore: DataStore<Preferences>,
    private val gson: Gson,
) {
    private object Keys {
        val activeBaseUrl = stringPreferencesKey("active_base_url")
        val endpoints = stringPreferencesKey("server_endpoints")
        val accessToken = stringPreferencesKey("access_token")
        val refreshToken = stringPreferencesKey("refresh_token")
        val shortFitMode = stringPreferencesKey("short_fit_mode")
        val shortDiscoverFitMode = stringPreferencesKey("short_discover_fit_mode")
        val unifiedShortFitMode = stringPreferencesKey("unified_short_fit_mode")
        val shortPlaybackMode = stringPreferencesKey("short_playback_mode")
        val legacyTvSubtitleTrackPreferences = stringPreferencesKey("tv_subtitle_preferences")
        val legacyTvAudioTrackPreferences = stringPreferencesKey("tv_audio_preferences")
        val legacyTvSubtitleLanguagePreferences = stringPreferencesKey("tv_subtitle_language_preferences")
        val legacyTvAudioLanguagePreferences = stringPreferencesKey("tv_audio_language_preferences")
        val tvAccountSubtitlePreference = stringPreferencesKey("tv_account_subtitle_preference")
        val tvAccountAudioPreference = stringPreferencesKey("tv_account_audio_preference")
        val tvSeekStepSeconds = stringPreferencesKey("tv_seek_step_seconds")
        val tvSeriesAutoplayEnabled = booleanPreferencesKey("tv_series_autoplay_enabled")
        val tvDeviceFallbackId = stringPreferencesKey("tv_device_fallback_id")
    }

    val activeBaseUrlFlow: Flow<String?> = dataStore.data.map { prefs ->
        prefs[Keys.activeBaseUrl]?.takeIf { it.isNotBlank() }
    }

    val accessTokenFlow: Flow<String?> = dataStore.data.map { prefs ->
        prefs[Keys.accessToken]?.takeIf { it.isNotBlank() }
    }

    val refreshTokenFlow: Flow<String?> = dataStore.data.map { prefs ->
        prefs[Keys.refreshToken]?.takeIf { it.isNotBlank() }
    }

    val shortFitModeFlow: Flow<VideoFitMode> = dataStore.data.map { prefs ->
        VideoFitMode.fromRaw(prefs[Keys.shortFitMode])
    }

    val shortDiscoverFitModeFlow: Flow<VideoFitMode> = dataStore.data.map { prefs ->
        VideoFitMode.fromRaw(prefs[Keys.shortDiscoverFitMode])
    }

    val unifiedShortFitModeFlow: Flow<VideoFitMode> = dataStore.data.map { prefs ->
        VideoFitMode.fromRaw(prefs[Keys.unifiedShortFitMode])
    }

    val shortPlaybackModeFlow: Flow<ShortPlaybackMode> = dataStore.data.map { prefs ->
        ShortPlaybackMode.fromRaw(prefs[Keys.shortPlaybackMode])
    }

    val tvSeekStepSecondsFlow: Flow<Int> = dataStore.data.map { prefs ->
        normalizeTvSeekStepSeconds(prefs[Keys.tvSeekStepSeconds]?.toIntOrNull())
    }

    val tvSeriesAutoplayEnabledFlow: Flow<Boolean> = dataStore.data.map { prefs ->
        prefs[Keys.tvSeriesAutoplayEnabled] ?: true
    }

    val endpointsFlow: Flow<List<ServerEndpoint>> = dataStore.data.map { prefs ->
        decodeEndpoints(prefs[Keys.endpoints].orEmpty())
    }

    suspend fun readActiveBaseUrl(): String? = activeBaseUrlFlow.first()

    suspend fun readAccessToken(): String? = accessTokenFlow.first()

    suspend fun readRefreshToken(): String? = refreshTokenFlow.first()

    suspend fun saveTokens(tokens: SessionTokens) {
        dataStore.edit { prefs ->
            prefs[Keys.accessToken] = tokens.accessToken
            prefs[Keys.refreshToken] = tokens.refreshToken
        }
    }

    suspend fun clearTokens() {
        dataStore.edit { prefs ->
            prefs.remove(Keys.accessToken)
            prefs.remove(Keys.refreshToken)
            clearTvTrackPreferences(prefs)
        }
    }

    suspend fun clearActiveServerAndTokens() {
        dataStore.edit { prefs ->
            prefs.remove(Keys.activeBaseUrl)
            prefs.remove(Keys.accessToken)
            prefs.remove(Keys.refreshToken)
            clearTvTrackPreferences(prefs)
        }
    }

    suspend fun saveShortFitMode(mode: VideoFitMode) {
        dataStore.edit { prefs ->
            prefs[Keys.shortFitMode] = mode.rawValue
        }
    }

    suspend fun saveShortDiscoverFitMode(mode: VideoFitMode) {
        dataStore.edit { prefs ->
            prefs[Keys.shortDiscoverFitMode] = mode.rawValue
        }
    }

    suspend fun saveUnifiedShortFitMode(mode: VideoFitMode) {
        dataStore.edit { prefs ->
            prefs[Keys.unifiedShortFitMode] = mode.rawValue
        }
    }

    suspend fun saveShortPlaybackMode(mode: ShortPlaybackMode) {
        dataStore.edit { prefs ->
            prefs[Keys.shortPlaybackMode] = mode.rawValue
        }
    }

    suspend fun readTvSeekStepSeconds(): Int = tvSeekStepSecondsFlow.first()

    suspend fun saveTvSeekStepSeconds(seconds: Int) {
        dataStore.edit { prefs ->
            prefs[Keys.tvSeekStepSeconds] = normalizeTvSeekStepSeconds(seconds).toString()
        }
    }

    suspend fun readTvSeriesAutoplayEnabled(): Boolean? =
        dataStore.data.first()[Keys.tvSeriesAutoplayEnabled]

    suspend fun saveTvSeriesAutoplayEnabled(enabled: Boolean) {
        dataStore.edit { prefs ->
            prefs[Keys.tvSeriesAutoplayEnabled] = enabled
        }
    }

    suspend fun readTvDeviceFallbackId(): String? =
        dataStore.data.first()[Keys.tvDeviceFallbackId]?.takeIf { it.isNotBlank() }

    suspend fun saveTvDeviceFallbackId(value: String) {
        val normalized = value.trim()
        if (normalized.isBlank()) {
            return
        }
        dataStore.edit { prefs ->
            prefs[Keys.tvDeviceFallbackId] = normalized
        }
    }

    suspend fun readTvSubtitlePreference(): TvTrackPreference? {
        clearLegacyTvTrackPreferences()
        return dataStore.data.first()[Keys.tvAccountSubtitlePreference]
            ?.let(::decodeTrackPreference)
    }

    suspend fun saveTvSubtitlePreference(preference: TvTrackPreference?) {
        clearLegacyTvTrackPreferences()
        saveTrackPreference(Keys.tvAccountSubtitlePreference, preference)
    }

    suspend fun readTvAudioPreference(): TvTrackPreference? {
        clearLegacyTvTrackPreferences()
        return dataStore.data.first()[Keys.tvAccountAudioPreference]
            ?.let(::decodeTrackPreference)
    }

    suspend fun saveTvAudioPreference(preference: TvTrackPreference?) {
        clearLegacyTvTrackPreferences()
        saveTrackPreference(Keys.tvAccountAudioPreference, preference)
    }

    suspend fun setActiveBaseUrl(baseUrl: String) {
        val normalized = UrlBuilder.normalizeBaseUrl(baseUrl)
        dataStore.edit { prefs ->
            if (normalized.isBlank()) {
                prefs.remove(Keys.activeBaseUrl)
            } else {
                prefs[Keys.activeBaseUrl] = normalized
            }
        }
    }

    suspend fun upsertEndpoint(baseUrl: String): ServerEndpoint {
        val normalized = UrlBuilder.normalizeBaseUrl(baseUrl)
        val now = System.currentTimeMillis()
        val endpoint = ServerEndpoint(baseUrl = normalized, lastSuccessAt = now)
        dataStore.edit { prefs ->
            val current = decodeEndpoints(prefs[Keys.endpoints].orEmpty()).toMutableList()
            val next = current
                .filterNot { it.baseUrl == normalized }
                .plus(endpoint)
                .sortedByDescending { it.lastSuccessAt }
                .take(20)
            prefs[Keys.endpoints] = encodeEndpoints(next)
        }
        return endpoint
    }

    suspend fun removeEndpoint(baseUrl: String) {
        val normalized = UrlBuilder.normalizeBaseUrl(baseUrl)
        dataStore.edit { prefs ->
            val next = decodeEndpoints(prefs[Keys.endpoints].orEmpty())
                .filterNot { it.baseUrl == normalized }
            prefs[Keys.endpoints] = encodeEndpoints(next)
            if (prefs[Keys.activeBaseUrl] == normalized) {
                prefs.remove(Keys.activeBaseUrl)
                prefs.remove(Keys.accessToken)
                prefs.remove(Keys.refreshToken)
            }
        }
    }

    private fun encodeEndpoints(endpoints: List<ServerEndpoint>): String {
        return gson.toJson(endpoints)
    }

    private fun decodeEndpoints(raw: String): List<ServerEndpoint> {
        if (raw.isBlank()) {
            return emptyList()
        }
        return try {
            val type = object : TypeToken<List<ServerEndpoint>>() {}.type
            gson.fromJson<List<ServerEndpoint>>(raw, type)
                ?.filter { it.baseUrl.isNotBlank() }
                ?.sortedByDescending { it.lastSuccessAt }
                ?: emptyList()
        } catch (_: Exception) {
            emptyList()
        }
    }

    private fun decodeTrackPreference(raw: String): TvTrackPreference? {
        if (raw.isBlank()) {
            return null
        }
        return try {
            gson.fromJson(raw, TvTrackPreference::class.java)?.let(::normalizeTvTrackPreference)
        } catch (_: Exception) {
            null
        }
    }

    private suspend fun saveTrackPreference(
        prefsKey: Preferences.Key<String>,
        preference: TvTrackPreference?,
    ) {
        val normalizedPreference = preference?.let(::normalizeTvTrackPreference)
        dataStore.edit { prefs ->
            if (normalizedPreference == null) {
                prefs.remove(prefsKey)
            } else {
                prefs[prefsKey] = gson.toJson(normalizedPreference)
            }
        }
    }

    private suspend fun clearLegacyTvTrackPreferences() {
        dataStore.edit { prefs ->
            prefs.remove(Keys.legacyTvSubtitleTrackPreferences)
            prefs.remove(Keys.legacyTvAudioTrackPreferences)
            prefs.remove(Keys.legacyTvSubtitleLanguagePreferences)
            prefs.remove(Keys.legacyTvAudioLanguagePreferences)
        }
    }

    private fun clearTvTrackPreferences(prefs: androidx.datastore.preferences.core.MutablePreferences) {
        prefs.remove(Keys.legacyTvSubtitleTrackPreferences)
        prefs.remove(Keys.legacyTvAudioTrackPreferences)
        prefs.remove(Keys.legacyTvSubtitleLanguagePreferences)
        prefs.remove(Keys.legacyTvAudioLanguagePreferences)
        prefs.remove(Keys.tvAccountSubtitlePreference)
        prefs.remove(Keys.tvAccountAudioPreference)
    }

    private fun normalizeTvTrackPreference(preference: TvTrackPreference): TvTrackPreference? {
        val normalized = TvTrackPreference(
            language = preference.language.trim(),
            type = preference.type.trim().lowercase(),
        )
        return normalized.takeUnless { it.isBlank() }
    }

    private fun normalizeTvSeekStepSeconds(seconds: Int?): Int {
        return when (seconds) {
            5, 10, 15, 20, 30 -> seconds
            else -> 10
        }
    }
}
