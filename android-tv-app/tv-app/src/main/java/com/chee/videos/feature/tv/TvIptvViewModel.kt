package com.chee.videos.feature.tv

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.model.TvIptvChannelDto
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class TvIptvUiState(
    val loading: Boolean = true,
    val channels: List<TvIptvChannelUiModel> = emptyList(),
    val groups: List<TvIptvChannelGroupUiModel> = emptyList(),
    val currentChannel: TvIptvChannelUiModel? = null,
    val statusMessage: String? = null,
)

@HiltViewModel
class TvIptvViewModel @Inject constructor(
    private val repository: TvRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(TvIptvUiState())
    val uiState: StateFlow<TvIptvUiState> = _uiState.asStateFlow()

    init {
        reload()
    }

    fun reload() {
        viewModelScope.launch {
            _uiState.update { it.copy(loading = true, statusMessage = null) }
            repository.fetchIptvChannels()
                .onSuccess { payload ->
                    val channels = coerceListOrEmpty<TvIptvChannelDto>(payload.channels)
                        .map(::tvIptvChannelToUiModel)
                        .filter { it.id.isNotBlank() && it.name.isNotBlank() && it.url.isNotBlank() }
                    val current = resolveDefaultIptvChannel(channels)
                    _uiState.update {
                        it.copy(
                            loading = false,
                            channels = channels,
                            groups = groupIptvChannels(channels),
                            currentChannel = current,
                            statusMessage = if (current == null) "暂无可播放的 IPTV 频道" else null,
                        )
                    }
                }
                .onFailure { error ->
                    _uiState.update {
                        it.copy(
                            loading = false,
                            channels = emptyList(),
                            groups = emptyList(),
                            currentChannel = null,
                            statusMessage = error.message ?: "IPTV 频道加载失败，请重试",
                        )
                    }
                }
        }
    }

    fun selectChannel(channelId: String) {
        _uiState.update { state ->
            val channel = state.channels.firstOrNull { it.id == channelId } ?: return@update state
            state.copy(currentChannel = channel, statusMessage = null)
        }
    }

    fun stepChannel(step: Int) {
        _uiState.update { state ->
            val channel = resolveIptvChannelAfterStep(
                channels = state.channels,
                currentChannelId = state.currentChannel?.id,
                step = step,
            ) ?: return@update state
            state.copy(currentChannel = channel, statusMessage = null)
        }
    }
}
