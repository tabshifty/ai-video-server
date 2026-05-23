package com.chee.videos.feature.tv

data class TvNextEpisodeRef(
    val seasonNumber: Int,
    val episodeNumber: Int,
    val title: String,
)

enum class TvEndOverlayKind {
    CURRENT_FINISHED,
    SERIES_FINISHED,
}

object TvSeriesAutoplaySetting {
    const val DEFAULT_ENABLED: Boolean = true

    fun parse(raw: Boolean?): Boolean = raw ?: DEFAULT_ENABLED
}

const val TvAutoplayCountdownSeconds: Int = 10

data class AutoplayPromptGuardInput(
    val isPlaying: Boolean,
    val autoplayEnabled: Boolean,
    val hasNextEpisode: Boolean,
    val isPlayerError: Boolean,
    val isSelectorVisible: Boolean,
    val isBackConfirmVisible: Boolean,
    val isEndOverlayVisible: Boolean,
    val isLoading: Boolean,
    val isCanceledForCurrentEpisode: Boolean,
    val remainingMs: Long,
    val durationMs: Long,
)

fun resolveNextPlayableEpisode(
    series: TvSeriesUiModel,
    currentSeasonNumber: Int,
    currentEpisodeNumber: Int,
): TvNextEpisodeRef? {
    val orderedSeasons = series.seasons
        .mapIndexed { index, season -> season to index }
        .sortedWith(compareBy<Pair<TvSeasonUiModel, Int>> { it.first.number }.thenBy { it.second })

    for ((season, _) in orderedSeasons) {
        if (season.number < currentSeasonNumber) {
            continue
        }

        val orderedEpisodes = season.episodes
            .mapIndexed { index, episode -> episode to index }
            .sortedWith(compareBy<Pair<TvEpisodeUiModel, Int>> { it.first.number }.thenBy { it.second })

        if (season.number == currentSeasonNumber) {
            for ((episode, _) in orderedEpisodes) {
                if (episode.number <= currentEpisodeNumber) {
                    continue
                }
                if (isPlayableEpisode(episode)) {
                    return episode.toNextEpisodeRef(season.number)
                }
            }
            continue
        }

        for ((episode, _) in orderedEpisodes) {
            if (isPlayableEpisode(episode)) {
                return episode.toNextEpisodeRef(season.number)
            }
        }
    }

    return null
}

fun shouldShowAutoplayPromptCard(input: AutoplayPromptGuardInput): Boolean {
    if (!input.isPlaying) return false
    if (!input.autoplayEnabled) return false
    if (!input.hasNextEpisode) return false
    if (input.isPlayerError) return false
    if (input.isSelectorVisible) return false
    if (input.isBackConfirmVisible) return false
    if (input.isEndOverlayVisible) return false
    if (input.isLoading) return false
    if (input.isCanceledForCurrentEpisode) return false
    if (input.durationMs <= 0L) return false
    return input.remainingMs in 1L..10_000L
}

fun shouldHandlePlaybackEnded(
    currentVideoId: String,
    lastAutoplaySwitchedVideoId: String,
): Boolean {
    val current = currentVideoId.trim()
    val lastAutoplay = lastAutoplaySwitchedVideoId.trim()
    if (current.isBlank() || lastAutoplay.isBlank()) return true
    return current == lastAutoplay
}

fun autoplayCountdownTickRemaining(
    remainingMs: Long,
    initialSeconds: Int = TvAutoplayCountdownSeconds,
): Int {
    if (remainingMs <= 0L) return 0
    val roundedSeconds = ((remainingMs + 999L) / 1000L).toInt()
    return roundedSeconds.coerceIn(0, initialSeconds)
}

private fun isPlayableEpisode(episode: TvEpisodeUiModel): Boolean =
    episode.playable && episode.videoId.isNotBlank()

private fun TvEpisodeUiModel.toNextEpisodeRef(seasonNumber: Int): TvNextEpisodeRef =
    TvNextEpisodeRef(
        seasonNumber = seasonNumber,
        episodeNumber = number,
        title = title,
    )
