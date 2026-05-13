package com.chee.videos.core.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxWithConstraints
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.heightIn
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.widthIn
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material3.Icon
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.ModalBottomSheet
import androidx.compose.material3.Surface
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.remember
import androidx.compose.runtime.withFrameNanos
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextOverflow
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
                .background(Color(0x99000000))
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
                val panelInteractionSource = remember { MutableInteractionSource() }
                Surface(
                    color = Color(0xF20D1016),
                    shape = RoundedCornerShape(18.dp),
                    modifier = Modifier
                        .fillMaxWidth(0.68f)
                        .widthIn(min = 520.dp, max = 640.dp)
                        .heightIn(max = maxHeight * 0.7f)
                        .clickable(
                            indication = null,
                            interactionSource = panelInteractionSource,
                            onClick = {},
                        ),
                ) {
                    Column(
                        modifier = Modifier.padding(horizontal = 22.dp, vertical = 20.dp),
                        verticalArrangement = Arrangement.spacedBy(14.dp),
                    ) {
                        Text(
                            text = "字幕",
                            color = Color.White,
                            style = MaterialTheme.typography.titleLarge,
                            fontWeight = FontWeight.SemiBold,
                        )
                        LazyColumn(
                            state = listState,
                            modifier = Modifier.fillMaxWidth(),
                            verticalArrangement = Arrangement.spacedBy(10.dp),
                        ) {
                            itemsIndexed(items, key = { _, item -> item.trackId ?: "off" }) { index, item ->
                                val focusRequester = remember { FocusRequester() }
                                SubtitleOptionRow(
                                    label = item.label,
                                    selected = item.selected,
                                    onClick = {
                                        onSelectSubtitleTrack(item.trackId)
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
private fun SubtitleOptionRow(
    label: String,
    selected: Boolean,
    onClick: () -> Unit,
    tvMode: Boolean = false,
    focusRequester: FocusRequester? = null,
) {
    val rowShape = RoundedCornerShape(if (tvMode) 12.dp else 14.dp)
    val rowColor = when {
        selected && tvMode -> Color(0xFF1D67D2)
        selected -> Color(0x26FFFFFF)
        tvMode -> Color(0xFF171C24)
        else -> Color(0x12000000)
    }
    val rowModifier = if (focusRequester != null) {
        Modifier.focusRequester(focusRequester)
    } else {
        Modifier
    }

    if (focusRequester != null) {
        LaunchedEffect(focusRequester, label) {
            withFrameNanos { }
            focusRequester.requestFocus()
        }
    }

    Surface(
        color = rowColor,
        shape = rowShape,
        modifier = Modifier.fillMaxWidth(),
    ) {
        Row(
            modifier = rowModifier
                .fillMaxWidth()
                .padding(horizontal = if (tvMode) 16.dp else 14.dp, vertical = if (tvMode) 14.dp else 12.dp)
                .tvFocusableGlow(shape = rowShape, focusedScale = if (tvMode) 1.04f else 1.02f)
                .clickable(onClick = onClick),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(12.dp),
        ) {
            Text(
                text = label,
                color = Color.White,
                style = if (tvMode) MaterialTheme.typography.bodyLarge else MaterialTheme.typography.bodyMedium,
                maxLines = 1,
                overflow = TextOverflow.Ellipsis,
                modifier = Modifier.weight(1f),
            )
            if (selected) {
                Row(
                    verticalAlignment = Alignment.CenterVertically,
                    horizontalArrangement = Arrangement.spacedBy(6.dp),
                ) {
                    if (tvMode) {
                        Icon(
                            imageVector = Icons.Filled.Check,
                            contentDescription = null,
                            tint = Color.White,
                        )
                    }
                    Text(
                        text = "已选中",
                        color = Color.White.copy(alpha = if (tvMode) 0.92f else 0.72f),
                        style = MaterialTheme.typography.labelSmall,
                    )
                }
            }
        }
    }
}
