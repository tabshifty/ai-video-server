package com.chee.videos.tv

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.feature.tv.TvRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class TvRemoteCoordinatorUiState(
    val activeSessionId: String? = null,
)

@HiltViewModel
class TvRemoteCoordinatorViewModel @Inject constructor(
    private val repository: TvRepository,
) : ViewModel() {
    private val _uiState = MutableStateFlow(TvRemoteCoordinatorUiState())
    val uiState: StateFlow<TvRemoteCoordinatorUiState> = _uiState.asStateFlow()
    private var pollingJob: Job? = null

    init {
        startPolling()
    }

    fun refreshNow() {
        viewModelScope.launch {
            refreshSession()
        }
    }

    private fun startPolling() {
        pollingJob?.cancel()
        pollingJob = viewModelScope.launch {
            while (true) {
                refreshSession()
                delay(5_000L)
            }
        }
    }

    private suspend fun refreshSession() {
        repository.fetchCurrentTvRemoteSession(buildTvDeviceId())
            .onSuccess { session ->
                _uiState.update {
                    it.copy(
                        activeSessionId = session
                            ?.takeIf { current -> current.status == "active" && current.sessionId.isNotBlank() }
                            ?.sessionId,
                    )
                }
            }
    }
}
