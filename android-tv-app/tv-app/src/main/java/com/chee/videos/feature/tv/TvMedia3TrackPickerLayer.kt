package com.chee.videos.feature.tv

import androidx.compose.runtime.Composable
import com.chee.videos.core.model.SubtitleTrackDto
import com.chee.videos.core.ui.LongFormAudioTrack
import com.chee.videos.core.ui.TvAudioTrackPickerDialog
import com.chee.videos.core.ui.TvSubtitlePickerDialog

internal enum class TvMedia3TrackPickerKind {
    Subtitle,
    Audio,
}

@Composable
internal fun TvMedia3TrackPickerLayer(
    kind: TvMedia3TrackPickerKind?,
    subtitleTracks: List<SubtitleTrackDto>,
    selectedSubtitleTrackId: String?,
    onSelectSubtitleTrack: (String?) -> Unit,
    audioTracks: List<LongFormAudioTrack>,
    selectedAudioTrackId: String?,
    onSelectAudioTrack: (String?) -> Unit,
    onDismissRequest: () -> Unit,
) {
    when (kind) {
        TvMedia3TrackPickerKind.Subtitle -> TvSubtitlePickerDialog(
            subtitleTracks = subtitleTracks,
            selectedSubtitleTrackId = selectedSubtitleTrackId,
            onSelectSubtitleTrack = {
                onSelectSubtitleTrack(it)
                onDismissRequest()
            },
            onDismissRequest = onDismissRequest,
        )

        TvMedia3TrackPickerKind.Audio -> TvAudioTrackPickerDialog(
            audioTracks = audioTracks,
            selectedAudioTrackId = selectedAudioTrackId,
            onSelectAudioTrack = {
                onSelectAudioTrack(it)
                onDismissRequest()
            },
            onDismissRequest = onDismissRequest,
        )

        null -> Unit
    }
}
