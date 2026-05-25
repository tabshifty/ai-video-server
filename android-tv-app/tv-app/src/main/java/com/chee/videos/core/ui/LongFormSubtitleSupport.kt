package com.chee.videos.core.ui

import com.chee.videos.core.model.SubtitleTrackDto
import com.chee.videos.core.model.TvTrackPreference
import com.chee.videos.core.player.TvLongFormVlcMediaSpec
import com.chee.videos.core.player.buildLongFormMedia
import com.chee.videos.core.player.normalizeLongFormLanguageCode
import com.chee.videos.core.util.UrlBuilder
import java.net.URI
import org.videolan.libvlc.LibVLC
import org.videolan.libvlc.MediaPlayer

internal data class LongFormPlayerUpdateDecision(
    val shouldClear: Boolean,
    val shouldReplaceSource: Boolean,
    val preservePosition: Boolean,
)

fun resolveInitialSubtitleTrackId(tracks: List<SubtitleTrackDto>): String? {
    return tracks.firstOrNull { it.sourceType == "uploaded" && it.isDefault }?.id
        ?: tracks.firstOrNull { it.isDefault }?.id
}

fun resolveSubtitleSelectionOnTrackLoad(
    currentSelection: String?,
    tracks: List<SubtitleTrackDto>,
    hasStartedPlayback: Boolean,
): String? {
    val normalizedSelection = currentSelection?.trim().orEmpty()
    val selectedStillValid = normalizedSelection.isNotBlank() && tracks.any { it.id == normalizedSelection && it.available }
    if (selectedStillValid) {
        return normalizedSelection
    }
    if (hasStartedPlayback) {
        return null
    }
    return resolveInitialSubtitleTrackId(tracks)
}

/**
 * 字幕切换不再触发 setMedia（之前会带 preservePosition reload）。subtitle 改走
 * `mediaPlayer.addSlave(IMedia.Slave.Type.Subtitle, ...)` 路径，由 [[VLC Playing gate]] 控制时机。
 * 本函数仅负责 source URL 的 setMedia 决策。
 */
internal fun resolveLongFormPlayerUpdate(
    preparedUrl: String?,
    nextUrl: String?,
): LongFormPlayerUpdateDecision {
    val normalizedPreparedUrl = preparedUrl?.trim().orEmpty()
    val normalizedNextUrl = nextUrl?.trim().orEmpty()
    if (normalizedNextUrl.isBlank()) {
        return LongFormPlayerUpdateDecision(
            shouldClear = normalizedPreparedUrl.isNotBlank(),
            shouldReplaceSource = false,
            preservePosition = false,
        )
    }
    if (normalizedPreparedUrl.isBlank() || normalizedPreparedUrl != normalizedNextUrl) {
        return LongFormPlayerUpdateDecision(
            shouldClear = false,
            shouldReplaceSource = true,
            preservePosition = false,
        )
    }
    return LongFormPlayerUpdateDecision(
        shouldClear = false,
        shouldReplaceSource = false,
        preservePosition = false,
    )
}

fun resolveSelectedSubtitleTrack(tracks: List<SubtitleTrackDto>, selectedTrackId: String?): SubtitleTrackDto? {
    val normalizedId = selectedTrackId?.trim().orEmpty()
    if (normalizedId.isBlank()) {
        return null
    }
    return tracks.firstOrNull { it.id == normalizedId && it.available && it.url.isNotBlank() }
}

fun subtitleTrackDisplayLabel(track: SubtitleTrackDto): String {
    val label = track.label.trim().ifBlank { track.languageLabel.trim() }.ifBlank { track.languageCode.trim() }
    if (label.isNotBlank()) {
        return label
    }
    return if (track.sourceType == "embedded") "内嵌字幕" else "外挂字幕"
}

/**
 * 仅 setMedia，不在 setMedia 时绑定字幕。字幕通过 [[VLC Playing gate]] 之后由
 * `mediaPlayer.addSlave(IMedia.Slave.Type.Subtitle, url, true)` 注入，避免 LibVLC 在播放器
 * 未真正进入 Playing 状态时丢弃 setAudioTrack / addSlave 调用。
 *
 * LibVLC 的 HTTP 模块不经过 OkHttp interceptor，因此 Authorization Bearer header 不会被自动注入。
 * 当 [accessToken] 非空时，把它以 `?access_token=` query 形式附加到 [sourceUrl] 上，服务端
 * AuthMiddleware 在 Authorization header 缺失时会从 query 读取 token。
 */
fun applyLongFormMediaSource(
    libVLC: LibVLC,
    mediaPlayer: MediaPlayer,
    sourceUrl: String,
    accessToken: String? = null,
) {
    val authenticatedSourceUrl = appendAccessTokenQuery(sourceUrl, accessToken)
    val media = buildLongFormMedia(
        libVLC = libVLC,
        spec = TvLongFormVlcMediaSpec(sourceUrl = authenticatedSourceUrl),
    )
    mediaPlayer.media = media
    media.release()
}

/**
 * 给 LibVLC 的播放 URL 附加 `access_token` query 参数。
 * 已经包含 `access_token=` 时不重复加。token 空白时直接返回原 URL。
 */
internal fun appendAccessTokenQuery(url: String, accessToken: String?): String {
    val token = accessToken?.trim().orEmpty()
    if (token.isEmpty()) return url
    if (url.contains("access_token=")) return url
    val encoded = java.net.URLEncoder.encode(token, "UTF-8")
    val separator = if (url.contains("?")) "&" else "?"
    return "$url$separator" + "access_token=$encoded"
}

fun resolveSelectedSubtitleTrackByLanguage(
    tracks: List<SubtitleTrackDto>,
    preferredLanguage: String?,
): SubtitleTrackDto? {
    return resolveSelectedSubtitleTrackByPreference(
        tracks = tracks,
        preference = TvTrackPreference(language = preferredLanguage.orEmpty()),
    )
}

fun resolveSelectedSubtitleTrackByPreference(
    tracks: List<SubtitleTrackDto>,
    preference: TvTrackPreference?,
): SubtitleTrackDto? {
    val safePreference = preference ?: return null
    val normalizedPreference = normalizeLongFormLanguageCode(safePreference.language)
    val typePreference = safePreference.type.trim().lowercase()

    if (normalizedPreference.isBlank()) {
        // type-only fallback：language 缺失但 type 不空时按 type 直接匹配。
        // 修复 [[Type-only preference fallback]]：之前直接 return null 会让 isDefault 字幕的偏好永远丢。
        if (typePreference.isBlank()) {
            return null
        }
        return tracks.firstOrNull { it.available && subtitlePreferenceType(it) == typePreference }
    }

    val languageMatches = tracks.filter { track ->
        val language = normalizeLongFormLanguageCode(track.languageCode)
        track.available &&
            (
                language == normalizedPreference ||
                    language.startsWith("$normalizedPreference-") ||
                    normalizedPreference.startsWith("$language-")
                )
    }
    if (languageMatches.isEmpty()) {
        return null
    }
    if (typePreference.isNotBlank()) {
        languageMatches.firstOrNull { subtitlePreferenceType(it) == typePreference }?.let { return it }
    }
    return languageMatches.first()
}

fun buildSubtitleTrackPreference(track: SubtitleTrackDto?): TvTrackPreference? {
    val safeTrack = track ?: return null
    val language = normalizeLongFormLanguageCode(safeTrack.languageCode)
    val type = subtitlePreferenceType(safeTrack)
    val preference = TvTrackPreference(language = language, type = type)
    return preference.takeUnless { it.isBlank() }
}

private fun subtitlePreferenceType(track: SubtitleTrackDto): String {
    val label = listOf(track.label, track.languageLabel, track.sourceType)
        .joinToString(" ")
        .lowercase()
    return when {
        "forced" in label || "强制" in label -> "forced"
        "commentary" in label || "comment" in label || "解说" in label || "评论" in label -> "commentary"
        track.isDefault -> "default"
        else -> ""
    }
}

internal fun resolvePlaybackAssetUrl(baseUrl: String, rawUrl: String): String? {
    val path = rawUrl.trim()
    if (path.isBlank()) {
        return null
    }
    if (path.startsWith("http://") || path.startsWith("https://")) {
        return path
    }
    val candidate = baseUrl.trim()
    if (candidate.startsWith("http://") || candidate.startsWith("https://")) {
        return runCatching { URI.create(candidate).resolve(path).toString() }.getOrNull()
    }
    val normalized = UrlBuilder.normalizeBaseUrl(candidate)
    if (normalized.isBlank()) {
        return null
    }
    return if (path.startsWith('/')) "$normalized$path" else "$normalized/$path"
}
