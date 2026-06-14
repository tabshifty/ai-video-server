package com.chee.videos.core.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.focusable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.focus.onFocusChanged
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.key.Key
import androidx.compose.ui.input.key.KeyEvent
import androidx.compose.ui.input.key.KeyEventType
import androidx.compose.ui.input.key.key
import androidx.compose.ui.input.key.onPreviewKeyEvent
import androidx.compose.ui.input.key.type
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties
import com.chee.videos.core.model.SubtitleTrackDto

internal enum class SubtitlePickerSurface {
    BottomSheet,
    CenterDialog,
}

internal data class SubtitlePickerItem(
    val trackId: String?,
    val label: String,
    val selected: Boolean,
)

internal fun resolveSubtitlePickerSurface(tvMode: Boolean): SubtitlePickerSurface {
    return if (tvMode) SubtitlePickerSurface.CenterDialog else SubtitlePickerSurface.BottomSheet
}

internal fun buildSubtitlePickerItems(
    tracks: List<SubtitleTrackDto>,
    selectedSubtitleTrackId: String?,
): List<SubtitlePickerItem> {
    return buildList {
        add(
            SubtitlePickerItem(
                trackId = null,
                label = "关闭字幕",
                selected = selectedSubtitleTrackId.isNullOrBlank(),
            ),
        )
        tracks.forEach { track ->
            add(
                SubtitlePickerItem(
                    trackId = track.id,
                    label = subtitleTrackDisplayLabel(track),
                    selected = selectedSubtitleTrackId == track.id,
                ),
            )
        }
    }
}

private data class TrackPickerDialogItem(
    val trackId: String?,
    val label: String,
    val supportingText: String = "",
    val selected: Boolean,
)

private val TrackPickerScrim = AppChrome.Canvas.copy(alpha = 0.72f)
private val TrackPickerPanelBorderBrush = Brush.linearGradient(
    listOf(
        AppChrome.TextPrimary.copy(alpha = 0.40f),
        AppChrome.Accent.copy(alpha = 0.40f),
        AppChrome.TextPrimary.copy(alpha = 0.10f),
    ),
)
private val TrackPickerPanelBrush = Brush.linearGradient(
    listOf(
        AppChrome.Surface.copy(alpha = 0.90f),
        AppChrome.SurfaceElevated.copy(alpha = 0.82f),
        AppChrome.Canvas.copy(alpha = 0.90f),
    ),
)

@OptIn(androidx.compose.material3.ExperimentalMaterial3Api::class)
@Composable
internal fun LongFormSubtitleBottomSheet(
    subtitleTracks: List<SubtitleTrackDto>,
    selectedSubtitleTrackId: String?,
    onSelectSubtitleTrack: (String?) -> Unit,
    onDismissRequest: () -> Unit,
) {
    val items = remember(subtitleTracks, selectedSubtitleTrackId) {
        buildSubtitlePickerItems(
            tracks = subtitleTracks,
            selectedSubtitleTrackId = selectedSubtitleTrackId,
        )
    }

    ModalBottomSheet(
        onDismissRequest = onDismissRequest,
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(horizontal = 20.dp, vertical = 8.dp),
            verticalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Text(
                text = "字幕",
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold,
            )
            LazyColumn(
                modifier = Modifier.fillMaxWidth(),
                verticalArrangement = Arrangement.spacedBy(8.dp),
            ) {
                items(items, key = { it.trackId ?: "off" }) { item ->
                    SubtitleOptionRow(
                        label = item.label,
                        selected = item.selected,
                        onClick = {
                            onSelectSubtitleTrack(item.trackId)
                            onDismissRequest()
                        },
                    )
                }
            }
        }
    }
}

@Composable
internal fun TvSubtitlePickerDialog(
    subtitleTracks: List<SubtitleTrackDto>,
    selectedSubtitleTrackId: String?,
    onSelectSubtitleTrack: (String?) -> Unit,
    onDismissRequest: () -> Unit,
) {
    val items = remember(subtitleTracks, selectedSubtitleTrackId) {
        buildSubtitlePickerItems(
            tracks = subtitleTracks,
            selectedSubtitleTrackId = selectedSubtitleTrackId,
        )
    }
    LongFormTrackPickerDialog(
        title = "字幕",
        items = items.map { item ->
            TrackPickerDialogItem(
                trackId = item.trackId,
                label = item.label,
                supportingText = "",
                selected = item.selected,
            )
        },
        fallbackKey = "off",
        onSelectTrack = onSelectSubtitleTrack,
        onDismissRequest = onDismissRequest,
    )
}

@Composable
internal fun TvAudioTrackPickerDialog(
    audioTracks: List<LongFormAudioTrack>,
    selectedAudioTrackId: String?,
    onSelectAudioTrack: (String?) -> Unit,
    onDismissRequest: () -> Unit,
) {
    val items = remember(audioTracks, selectedAudioTrackId) {
        buildAudioTrackPickerItems(
            tracks = audioTracks,
            selectedAudioTrackId = selectedAudioTrackId,
        )
    }
    LongFormTrackPickerDialog(
        title = "音轨",
        items = items.map { item ->
            TrackPickerDialogItem(
                trackId = item.trackId,
                label = item.label,
                supportingText = item.detail,
                selected = item.selected,
            )
        },
        fallbackKey = "auto",
        onSelectTrack = onSelectAudioTrack,
        onDismissRequest = onDismissRequest,
    )
}

@Composable
private fun LongFormTrackPickerDialog(
    title: String,
    items: List<TrackPickerDialogItem>,
    fallbackKey: String,
    onSelectTrack: (String?) -> Unit,
    onDismissRequest: () -> Unit,
) {
    val selectedIndex = items.indexOfFirst { it.selected }.let { if (it >= 0) it else 0 }
    val listState = rememberLazyListState(initialFirstVisibleItemIndex = selectedIndex)

    Dialog(
        onDismissRequest = onDismissRequest,
        properties = DialogProperties(usePlatformDefaultWidth = false),
    ) {
        val scrimInteractionSource = remember { MutableInteractionSource() }
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(TrackPickerScrim)
                .clickable(
                    indication = null,
                    interactionSource = scrimInteractionSource,
                    onClick = onDismissRequest,
                ),
            contentAlignment = Alignment.Center,
        ) {
            BoxWithConstraints(
                modifier = Modifier.fillMaxSize(),
                contentAlignment = Alignment.Center,
            ) {
                NightGlassTrackPickerPanel(
                    modifier = Modifier.heightIn(max = maxHeight * 0.7f),
                ) {
                    Column(
                        modifier = Modifier.padding(horizontal = 22.dp, vertical = 20.dp),
                        verticalArrangement = Arrangement.spacedBy(14.dp),
                    ) {
                        Text(
                            text = title,
                            color = AppChrome.TextPrimary,
                            style = MaterialTheme.typography.titleLarge,
                            fontWeight = FontWeight.SemiBold,
                        )
                        LazyColumn(
                            state = listState,
                            modifier = Modifier.fillMaxWidth(),
                            verticalArrangement = Arrangement.spacedBy(10.dp),
                        ) {
                            itemsIndexed(items, key = { _, item -> item.trackId ?: fallbackKey }) { index, item ->
                                val focusRequester = remember { FocusRequester() }
                                SubtitleOptionRow(
                                    label = item.label,
                                    supportingText = item.supportingText,
                                    selected = item.selected,
                                    onClick = {
                                        onSelectTrack(item.trackId)
                                        onDismissRequest()
                                    },
                                    tvMode = true,
                                    focusRequester = if (index == selectedIndex) focusRequester else null,
                                )
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun NightGlassTrackPickerPanel(
    modifier: Modifier = Modifier,
    content: @Composable () -> Unit,
) {
    // 夜台玻璃：深色半透明基底、暖金高光边缘和低饱和焦点光感。
    val panelInteractionSource = remember { MutableInteractionSource() }
    Box(
        modifier = modifier
            .fillMaxWidth(0.68f)
            .widthIn(min = 520.dp, max = 640.dp)
            .border(
                width = 1.dp,
                brush = TrackPickerPanelBorderBrush,
                shape = AppChrome.SurfaceShape,
            )
            .background(
                brush = TrackPickerPanelBrush,
                shape = AppChrome.SurfaceShape,
            )
            .clickable(
                indication = null,
                interactionSource = panelInteractionSource,
                onClick = {},
            ),
    ) {
        content()
    }
}

@Composable
private fun SubtitleOptionRow(
    label: String,
    supportingText: String = "",
    selected: Boolean,
    onClick: () -> Unit,
    tvMode: Boolean = false,
    focusRequester: FocusRequester? = null,
) {
    val rowShape = AppChrome.SurfaceShape
    var focused by remember { androidx.compose.runtime.mutableStateOf(false) }
    val rowColor = when {
        focused && tvMode -> AppChrome.Accent.copy(alpha = 0.24f)
        selected && tvMode -> AppChrome.Accent.copy(alpha = 0.18f)
        selected -> AppChrome.TextPrimary.copy(alpha = 0.15f)
        tvMode -> AppChrome.TextPrimary.copy(alpha = 0.10f)
        else -> AppChrome.Canvas.copy(alpha = 0.07f)
    }
    val interactionSource = remember { MutableInteractionSource() }
    val rowModifier = if (focusRequester != null) {
        Modifier.focusRequester(focusRequester)
    } else {
        Modifier
    }

    if (focusRequester != null) {
        LaunchedTvInitialFocus(focusRequester, label) {
            focusRequester.tryRequestFocus()
        }
    }

    Surface(
        color = rowColor,
        shape = rowShape,
        modifier = Modifier
            .fillMaxWidth()
            .then(if (tvMode) Modifier.clip(rowShape) else Modifier),
    ) {
        Row(
            modifier = rowModifier
                .fillMaxWidth()
                .onFocusChanged { focused = it.isFocused || it.hasFocus }
                .then(if (tvMode) Modifier.onPreviewKeyEvent { handleTrackPickerConfirmKey(it, onClick) } else Modifier)
                .padding(horizontal = if (tvMode) 16.dp else 14.dp, vertical = if (tvMode) 14.dp else 12.dp)
                .focusable(enabled = tvMode)
                .clickable(
                    interactionSource = interactionSource,
                    indication = null,
                    onClick = onClick,
                ),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            if (tvMode) {
                TrackPickerSelectionRail(
                    selected = selected,
                    focused = focused,
                )
            }
            Column(modifier = Modifier.weight(1f)) {
                Text(
                    text = label,
                    color = AppChrome.TextPrimary,
                    style = if (tvMode) MaterialTheme.typography.bodyLarge else MaterialTheme.typography.bodyMedium,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis,
                )
                if (supportingText.isNotBlank()) {
                    Text(
                        text = supportingText,
                        color = AppChrome.TextMuted,
                        style = MaterialTheme.typography.labelMedium,
                        maxLines = 1,
                        overflow = TextOverflow.Ellipsis,
                    )
                }
            }
            if (selected) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    if (tvMode) {
                        Icon(
                            imageVector = Icons.Filled.Check,
                            contentDescription = null,
                            tint = AppChrome.TextPrimary,
                        )
                    }
                    Text(
                        text = "已选中",
                        color = if (tvMode) AppChrome.TextPrimary else AppChrome.TextMuted,
                        style = MaterialTheme.typography.labelSmall,
                    )
                }
            }
        }
    }
}

@Composable
private fun TrackPickerSelectionRail(
    selected: Boolean,
    focused: Boolean,
) {
    val color = when {
        selected -> AppChrome.Accent
        focused -> AppChrome.Accent.copy(alpha = 0.52f)
        else -> Color.Transparent
    }
    Box(
        modifier = Modifier
            .width(3.dp)
            .heightIn(min = 28.dp)
            .background(color = color, shape = AppChrome.PillShape),
    )
}

private fun handleTrackPickerConfirmKey(
    event: KeyEvent,
    onConfirm: () -> Unit,
): Boolean {
    if (event.type != KeyEventType.KeyDown) {
        return false
    }
    return when (event.key) {
        Key.DirectionCenter,
        Key.Enter,
        Key.NumPadEnter,
        -> {
            onConfirm()
            true
        }
        else -> false
    }
}
