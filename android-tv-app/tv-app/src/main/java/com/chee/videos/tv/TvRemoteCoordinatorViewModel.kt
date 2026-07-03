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

internal fun clearTvRemoteCoordinatorState(
    state: TvRemoteCoordinatorUiState,
    sessionId: String?,
): TvRemoteCoordinatorUiState {
    val target = sessionId?.trim().orEmpty()
    val active = state.activeSessionId?.trim().orEmpty()
    if (active.isBlank()) {
        return state.copy(activeSessionId = null)
    }
    return if (target.isBlank() || active == target) {
        state.copy(activeSessionId = null)
    } else {
        state
    }
}

@HiltViewModel
class TvRemoteCoordinatorViewModel @Inject constructor(
    private val repository: TvRepository,
    private val deviceIdentityProvider: TvDeviceIdentityProvider,
) : ViewModel() {
    private val _uiState = MutableStateFlow(TvRemoteCoordinatorUiState())
    val uiState: StateFlow<TvRemoteCoordinatorUiState> = _uiState.asStateFlow()
    private var pollingJob: Job? = null
    private val dismissedSessionIds = mutableSetOf<String>()
    private val dismissJobs = mutableMapOf<String, Job>()

    fun setPollingEnabled(enabled: Boolean) {
        if (enabled) {
            startPolling()
        } else {
            stopPolling()
        }
    }

    fun refreshNow() {
        viewModelScope.launch {
            refreshSession()
        }
    }

    fun dismissSession(sessionId: String?) {
        val normalizedSessionId = sessionId?.trim().orEmpty()
        if (normalizedSessionId.isNotBlank()) {
            dismissedSessionIds += normalizedSessionId
            ensureDismissJob(normalizedSessionId)
        }
        _uiState.update { state -> clearTvRemoteCoordinatorState(state, sessionId) }
    }

    private fun startPolling() {
        if (pollingJob?.isActive == true) {
            return
        }
        pollingJob?.cancel()
        pollingJob = viewModelScope.launch {
            while (true) {
                refreshSession()
                delay(5_000L)
            }
        }
    }

    private fun stopPolling() {
        pollingJob?.cancel()
        pollingJob = null
    }

    private suspend fun refreshSession() {
        repository.fetchCurrentTvRemoteSession(
            deviceId = deviceIdentityProvider.currentDeviceId(),
            legacyDeviceId = deviceIdentityProvider.legacyDeviceId(),
        )
            .onSuccess { session ->
                val activeSession = session
                    ?.takeIf { current -> current.status == "active" && current.sessionId.isNotBlank() }
                _uiState.update {
                    it.copy(
                        activeSessionId = activeSession
                            ?.sessionId
                            ?.takeUnless { sessionId -> dismissedSessionIds.contains(sessionId) },
                    )
                }
                activeSession?.sessionId?.let { sessionId ->
                    if (dismissedSessionIds.contains(sessionId)) {
                        ensureDismissJob(sessionId)
                    }
                }
                if (session == null || session.status != "active") {
                    dismissedSessionIds.clear()
                    dismissJobs.values.forEach { it.cancel() }
                    dismissJobs.clear()
                }
            }
    }

    private fun ensureDismissJob(sessionId: String) {
        if (dismissJobs[sessionId]?.isActive == true) {
            return
        }
        dismissJobs[sessionId] = viewModelScope.launch {
            while (true) {
                val ended = repository.endTvRemoteSession(sessionId).isSuccess
                if (ended) {
                    dismissedSessionIds.remove(sessionId)
                    dismissJobs.remove(sessionId)
                    return@launch
                }
                delay(1_000L)
            }
        }
    }
}
