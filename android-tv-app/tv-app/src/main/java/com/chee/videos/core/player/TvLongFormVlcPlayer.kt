package com.chee.videos.core.player

import android.net.Uri
import org.videolan.libvlc.LibVLC
import org.videolan.libvlc.Media
import org.videolan.libvlc.MediaPlayer
import org.videolan.libvlc.interfaces.IMedia

internal data class TvLongFormHwDecoderConfig(
    val enabled: Boolean,
    val force: Boolean,
)

internal data class TvLongFormVlcMediaSpec(
    val sourceUrl: String,
    val subtitleUrl: String? = null,
)

internal fun buildLongFormVlcMediaOptions(): List<String> = listOf(
    ":file-caching=1500",
    ":network-caching=2000",
)

internal fun longFormHwDecoderConfig(): TvLongFormHwDecoderConfig =
    TvLongFormHwDecoderConfig(enabled = true, force = true)

internal fun buildLongFormMedia(
    libVLC: LibVLC,
    spec: TvLongFormVlcMediaSpec,
): Media {
    val media = Media(libVLC, Uri.parse(spec.sourceUrl))
    buildLongFormVlcMediaOptions().forEach(media::addOption)
    val decoder = longFormHwDecoderConfig()
    media.setHWDecoderEnabled(decoder.enabled, decoder.force)
    spec.subtitleUrl?.takeIf { it.isNotBlank() }?.let { subtitleUrl ->
        media.addSlave(IMedia.Slave(IMedia.Slave.Type.Subtitle, 0, subtitleUrl))
    }
    return media
}

fun newLongFormMediaPlayer(libVLC: LibVLC): MediaPlayer = MediaPlayer(libVLC)
