package com.chee.videos.feature.connection

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chee.videos.core.model.ServerEndpoint
import com.chee.videos.core.repository.ServerRepository
import com.chee.videos.core.util.UrlBuilder
import dagger.hilt.android.lifecycle.HiltViewModel
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch

data class ConnectionUiState(
    val scanning: Boolean = false,
    val connecting: Boolean = false,
    val discoveredEndpoints: List<ServerEndpoint> = emptyList(),
    val savedEndpoints: List<ServerEndpoint> = emptyList(),
    val hostInput: String = "",
    val portInput: String = "8080",
    val message: String? = null,
    val messageIsError: Boolean = false,
)

@HiltViewModel
class ConnectionViewModel @Inject constructor(
    private val serverRepository: ServerRepository,
) : ViewModel() {

    private val _uiState = MutableStateFlow(ConnectionUiState())
    val uiState: StateFlow<ConnectionUiState> = _uiState.asStateFlow()

    init {
        viewModelScope.launch {
            serverRepository.endpointsFlow.collect { endpoints ->
                _uiState.update { it.copy(savedEndpoints = endpoints) }
            }
        }
    }

    fun updateHostInput(value: String) {
        _uiState.update { it.copy(hostInput = value) }
    }

    fun updatePortInput(value: String) {
        _uiState.update { it.copy(portInput = value.filter { ch -> ch.isDigit() }.take(5)) }
    }

    fun scanLan() {
        if (_uiState.value.scanning) {
            return
        }
        viewModelScope.launch {
            _uiState.update {
                it.copy(
                    scanning = true,
                    message = "正在扫描局域网服务...",
                    messageIsError = false,
                )
            }
            runCatching {
                serverRepository.scanLocalNetwork()
            }.onSuccess { discovered ->
                _uiState.update {
                    it.copy(
                        scanning = false,
                        discoveredEndpoints = discovered,
                        message = if (discovered.isEmpty()) "未发现服务，请手动输入 IP。" else "已发现 ${discovered.size} 个可用服务。",
                        messageIsError = false,
                    )
                }
            }.onFailure { err ->
                _uiState.update {
                    it.copy(
                        scanning = false,
                        message = "扫描失败：${err.message}",
                        messageIsError = true,
                    )
                }
            }
        }
    }

    fun useEndpoint(baseUrl: String) {
        viewModelScope.launch {
            _uiState.update { it.copy(connecting = true, message = "正在应用地址...", messageIsError = false) }
            runCatching {
                serverRepository.activateEndpoint(baseUrl)
            }.onSuccess {
                _uiState.update { it.copy(connecting = false, message = "已连接：${UrlBuilder.normalizeBaseUrl(baseUrl)}", messageIsError = false) }
            }.onFailure { err ->
                _uiState.update { it.copy(connecting = false, message = "应用地址失败：${err.message}", messageIsError = true) }
            }
        }
    }

    fun manualConnect() {
        if (_uiState.value.connecting) {
            return
        }
        val host = _uiState.value.hostInput.trim()
        val port = _uiState.value.portInput.trim()
        if (host.isBlank()) {
            _uiState.update { it.copy(message = "请输入服务器 IP 或域名", messageIsError = true) }
            return
        }
        val raw = if (port.isBlank()) host else "$host:$port"
        viewModelScope.launch {
            _uiState.update { it.copy(connecting = true, message = "正在测试连接...", messageIsError = false) }
            val ok = serverRepository.testEndpoint(raw)
            if (!ok) {
                _uiState.update { it.copy(connecting = false, message = "连接失败，请确认 IP/端口和服务器状态", messageIsError = true) }
                return@launch
            }
            serverRepository.activateEndpoint(raw)
            _uiState.update { it.copy(connecting = false, message = "连接成功：${UrlBuilder.normalizeBaseUrl(raw)}", messageIsError = false) }
        }
    }

    fun removeSavedEndpoint(baseUrl: String) {
        viewModelScope.launch {
            serverRepository.removeEndpoint(baseUrl)
        }
    }
}
