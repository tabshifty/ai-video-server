package com.chee.videos.tv

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.feature.tv.TvRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlin.coroutines.coroutineContext
import kotlinx.coroutines.CoroutineStart
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
    private val dismissedSessionIds = linkedSetOf<String>()
    private val dismissJobs = mutableMapOf<String, Job>()

    private companion object {
        // dismiss 上报失败时的单轮重试上限与指数退避：避免服务端异常时对同一会话无限 1Hz 重试。
        // 单轮达到上限后任务结束并从 dismissJobs 移除；会话仍留在 dismissedSessionIds 内，
        // 下一次 5s 会话轮询发现该会话仍 active 时会重新拉起新一轮有界重试，
        // 宏观上仍保持「持续上报直到服务端成功结束会话」的语义。
        const val MAX_DISMISS_REPORT_ATTEMPTS = 5
        const val DISMISS_RETRY_BASE_DELAY_MS = 1_000L
        const val DISMISS_RETRY_MAX_DELAY_MS = 16_000L

        // 本地已忽略会话集合的容量上界：同时存在的投放会话极少，超出时按插入序淘汰最旧记录，
        // 防止长时间运行下集合无界增长。
        const val MAX_DISMISSED_SESSION_IDS = 16
    }

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
            while (dismissedSessionIds.size > MAX_DISMISSED_SESSION_IDS) {
                val oldest = dismissedSessionIds.firstOrNull() ?: break
                dismissedSessionIds.remove(oldest)
                dismissJobs.remove(oldest)?.cancel()
            }
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
                    // 迭代快照：Main.immediate 下 cancel() 会同步恢复挂起在 delay 的协程并执行其
                    // finally（移除 map 条目），直接遍历 values 会触发 ConcurrentModificationException
                    dismissJobs.values.toList().forEach { it.cancel() }
                    dismissJobs.clear()
                }
            }
    }

    private fun ensureDismissJob(sessionId: String) {
        if (dismissJobs[sessionId]?.isActive == true) {
            return
        }
        // LAZY 启动：先登记再启动，消除 Main.immediate 下「协程体同步跑完先于登记」导致
        // finally 清理落空、已完成 Job 永久残留在 map 里的竞态
        val job = viewModelScope.launch(start = CoroutineStart.LAZY) {
            try {
                var retryDelayMs = DISMISS_RETRY_BASE_DELAY_MS
                repeat(MAX_DISMISS_REPORT_ATTEMPTS) { attempt ->
                    val ended = repository.endTvRemoteSession(sessionId).isSuccess
                    if (ended) {
                        dismissedSessionIds.remove(sessionId)
                        return@launch
                    }
                    if (attempt < MAX_DISMISS_REPORT_ATTEMPTS - 1) {
                        delay(retryDelayMs)
                        retryDelayMs = (retryDelayMs * 2).coerceAtMost(DISMISS_RETRY_MAX_DELAY_MS)
                    }
                }
            } finally {
                // 身份校验删除：本任务被取消后若同键已注册新任务，不得误删新条目
                dismissJobs.remove(sessionId, coroutineContext[Job])
            }
        }
        dismissJobs[sessionId] = job
        job.start()
    }
}
